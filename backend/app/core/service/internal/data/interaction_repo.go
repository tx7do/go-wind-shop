package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	entCrud "github.com/tx7do/go-crud/entgo"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/commentlike"
	"go-wind-shop/app/core/service/internal/data/ent/interactioncounter"

	interactionV1 "go-wind-shop/api/gen/go/interaction/service/v1"
)

// InteractionRepo 是评论点赞 ledger 与 interaction_counter 计数表的唯一写入方。
//
// 设计要点：
//   - viewer 用户身份由 service 层从鉴权上下文提取后传入，repo 不接受客户端
//     传入的 user_id。
//   - Like/Unlike 在单个 ent.Tx 内同时操作 comment_like ledger 表与
//     interaction_counter 计数表（upsert 计数行），保证原子性。
//   - 计数统一存于 interaction_counter 表（target_type/target_id/metric/count），
//     comment 表上的散落计数列已移除。
//   - 重复点赞走幂等：先 Exist 检查，已存在则 no-op，不依赖约束错误判断。
//   - shop 无 post/watch，仅支持 comment like。
//   - 跨租户隔离由 mixin + 查询带 tenant_id 保证。
type InteractionRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper
}

func NewInteractionRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *InteractionRepo {
	return &InteractionRepo{
		entClient: entClient,
		log:       ctx.NewLoggerHelper("interaction/repo/core-service"),
	}
}

// txn 在事务内执行 fn，沿用跨表事务惯例。
func (r *InteractionRepo) txn(ctx context.Context, fn func(tx *ent.Tx) error) (err error) {
	var tx *ent.Tx
	tx, err = r.entClient.Client().Tx(ctx)
	if err != nil {
		r.log.Errorf("start transaction failed: %s", err.Error())
		return interactionV1.ErrorInternalServerError("start transaction failed")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.log.Errorf("transaction rollback failed: %s", rollbackErr.Error())
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			r.log.Errorf("transaction commit failed: %s", commitErr.Error())
			err = interactionV1.ErrorInternalServerError("transaction commit failed")
		}
	}()
	return fn(tx)
}

// Like 点赞 comment。幂等：已点赞则 no-op。
// 返回操作后的点赞状态与最新计数。
func (r *InteractionRepo) Like(ctx context.Context, viewerUserID uint32, targetType interactionV1.TargetType, targetID uint32) (bool, int32, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant || viewerUserID == 0 {
		return false, 0, interactionV1.ErrorUnauthorized("viewer identity required")
	}

	switch targetType {
	case interactionV1.TargetType_TARGET_TYPE_COMMENT:
		likeCount, liked, err := r.likeTarget(ctx, tid, viewerUserID, targetID, targetType)
		if err != nil {
			return false, 0, err
		}
		return liked, likeCount, nil

	default:
		return false, 0, interactionV1.ErrorBadRequest("invalid target type")
	}
}

// Unlike 取消点赞 comment。幂等：未点赞则 no-op。
func (r *InteractionRepo) Unlike(ctx context.Context, viewerUserID uint32, targetType interactionV1.TargetType, targetID uint32) (bool, int32, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant || viewerUserID == 0 {
		return false, 0, interactionV1.ErrorUnauthorized("viewer identity required")
	}

	switch targetType {
	case interactionV1.TargetType_TARGET_TYPE_COMMENT:
		likeCount, liked, err := r.unlikeTarget(ctx, tid, viewerUserID, targetID, targetType)
		if err != nil {
			return false, 0, err
		}
		return liked, likeCount, nil

	default:
		return false, 0, interactionV1.ErrorBadRequest("invalid target type")
	}
}

// readCount 读取 interaction_counter 表中 (tenant, target_type, target_id, metric) 的当前计数。
// 不存在行则返回 0。
func (r *InteractionRepo) readCount(ctx context.Context, tid, targetID uint32, targetType interactionV1.TargetType, metric interactionV1.CounterMetric) (int32, error) {
	entity, err := r.entClient.Client().InteractionCounter.Query().
		Where(
			interactioncounter.TenantIDEQ(tid),
			interactioncounter.TargetTypeEQ(uint8(targetType)),
			interactioncounter.TargetIDEQ(targetID),
			interactioncounter.MetricEQ(uint8(metric)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		r.log.Errorf("query interaction_counter count failed: %s", err.Error())
		return 0, interactionV1.ErrorInternalServerError("query interaction_counter count failed")
	}
	if entity.Count == nil {
		return 0, nil
	}
	return int32(*entity.Count), nil
}

// likeTarget 在单个 tx 内：写 comment_like ledger + 在 interaction_counter
// 表中 upsert 计数（LIKE 指标 +1）。幂等：若 ledger 行已存在则 no-op。
func (r *InteractionRepo) likeTarget(ctx context.Context, tid, viewerUserID, targetID uint32, targetType interactionV1.TargetType) (int32, bool, error) {
	exists, err := r.ledgerExists(ctx, tid, viewerUserID, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
	if err != nil {
		return 0, false, err
	}
	if exists {
		// 幂等：已点赞，no-op，但仍返回当前状态
		likeCount, qerr := r.readCount(ctx, tid, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
		if qerr != nil {
			return 0, false, qerr
		}
		return likeCount, true, nil
	}

	txErr := r.txn(ctx, func(tx *ent.Tx) error {
		if err := r.createLedgerRow(ctx, tx, tid, viewerUserID, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE); err != nil {
			return err
		}
		// 在 interaction_counter 表中 upsert 计数（+1）
		if err := r.adjustCounterRow(ctx, tx, tid, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE, 1); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return 0, false, txErr
	}

	likeCount, err := r.readCount(ctx, tid, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
	if err != nil {
		return 0, false, err
	}
	return likeCount, true, nil
}

// unlikeTarget 在单个 tx 内：删 ledger + 在 interaction_counter 表中递减计数。
// 幂等：若 ledger 行不存在则 no-op。
func (r *InteractionRepo) unlikeTarget(ctx context.Context, tid, viewerUserID, targetID uint32, targetType interactionV1.TargetType) (int32, bool, error) {
	exists, err := r.ledgerExists(ctx, tid, viewerUserID, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
	if err != nil {
		return 0, false, err
	}
	if !exists {
		// 幂等：未点赞，no-op
		likeCount, qerr := r.readCount(ctx, tid, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
		if qerr != nil {
			return 0, false, qerr
		}
		return likeCount, false, nil
	}

	txErr := r.txn(ctx, func(tx *ent.Tx) error {
		affected, err := r.deleteLedgerRow(ctx, tx, tid, viewerUserID, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
		if err != nil {
			return err
		}
		if affected == 0 {
			// 并发删除，幂等处理
			return nil
		}
		// 在 interaction_counter 表中递减计数（-1）
		if err := r.adjustCounterRow(ctx, tx, tid, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE, -1); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return 0, false, txErr
	}

	likeCount, err := r.readCount(ctx, tid, targetID, targetType, interactionV1.CounterMetric_COUNTER_METRIC_LIKE)
	if err != nil {
		return 0, false, err
	}
	return likeCount, false, nil
}

// ledgerExists 查询 (tenant, viewer, comment) 的 comment_like ledger 行是否存在。
func (r *InteractionRepo) ledgerExists(ctx context.Context, tid, viewerUserID, targetID uint32, targetType interactionV1.TargetType, metric interactionV1.CounterMetric) (bool, error) {
	switch metric {
	case interactionV1.CounterMetric_COUNTER_METRIC_LIKE:
		switch targetType {
		case interactionV1.TargetType_TARGET_TYPE_COMMENT:
			exists, err := r.entClient.Client().CommentLike.Query().
				Where(
					commentlike.TenantIDEQ(tid),
					commentlike.UserIDEQ(viewerUserID),
					commentlike.CommentIDEQ(targetID),
				).
				Exist(ctx)
			if err != nil {
				r.log.Errorf("query comment_like exist failed: %s", err.Error())
				return false, interactionV1.ErrorInternalServerError("query comment_like exist failed")
			}
			return exists, nil

		default:
			return false, interactionV1.ErrorBadRequest("invalid target type")
		}

	default:
		return false, interactionV1.ErrorBadRequest("invalid metric")
	}
}

// createLedgerRow 在 tx 内创建 comment_like ledger 行。
func (r *InteractionRepo) createLedgerRow(ctx context.Context, tx *ent.Tx, tid, viewerUserID, targetID uint32, targetType interactionV1.TargetType, metric interactionV1.CounterMetric) error {
	switch metric {
	case interactionV1.CounterMetric_COUNTER_METRIC_LIKE:
		switch targetType {
		case interactionV1.TargetType_TARGET_TYPE_COMMENT:
			if _, err := tx.CommentLike.Create().
				SetTenantID(tid).
				SetUserID(viewerUserID).
				SetCommentID(targetID).
				Save(ctx); err != nil {
				r.log.Errorf("insert comment_like failed: %s", err.Error())
				return interactionV1.ErrorInternalServerError("insert comment_like failed")
			}
			return nil

		default:
			return interactionV1.ErrorBadRequest("invalid target type")
		}

	default:
		return interactionV1.ErrorBadRequest("invalid metric")
	}
}

// deleteLedgerRow 在 tx 内删除 comment_like ledger 行，返回受影响行数。
func (r *InteractionRepo) deleteLedgerRow(ctx context.Context, tx *ent.Tx, tid, viewerUserID, targetID uint32, targetType interactionV1.TargetType, metric interactionV1.CounterMetric) (int, error) {
	switch metric {
	case interactionV1.CounterMetric_COUNTER_METRIC_LIKE:
		switch targetType {
		case interactionV1.TargetType_TARGET_TYPE_COMMENT:
			affected, err := tx.CommentLike.Delete().
				Where(
					commentlike.TenantIDEQ(tid),
					commentlike.UserIDEQ(viewerUserID),
					commentlike.CommentIDEQ(targetID),
				).
				Exec(ctx)
			if err != nil {
				r.log.Errorf("delete comment_like failed: %s", err.Error())
				return 0, interactionV1.ErrorInternalServerError("delete comment_like failed")
			}
			return affected, nil

		default:
			return 0, interactionV1.ErrorBadRequest("invalid target type")
		}

	default:
		return 0, interactionV1.ErrorBadRequest("invalid metric")
	}
}

// adjustCounterRow 在 tx 内 upsert interaction_counter 表的计数行。
//
// 行不存在时按 delta 创建（仅当 delta>0）；存在时按 delta 递增，归 0 则删行。
// 复合 unique 索引 (tenant, target_type, target_id, metric) 保证单目标单 metric 仅一行。
func (r *InteractionRepo) adjustCounterRow(ctx context.Context, tx *ent.Tx, tid, targetID uint32, targetType interactionV1.TargetType, metric interactionV1.CounterMetric, delta int64) error {
	existing, err := tx.InteractionCounter.Query().
		Where(
			interactioncounter.TenantIDEQ(tid),
			interactioncounter.TargetTypeEQ(uint8(targetType)),
			interactioncounter.TargetIDEQ(targetID),
			interactioncounter.MetricEQ(uint8(metric)),
		).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			r.log.Errorf("query interaction_counter row failed: %s", err.Error())
			return interactionV1.ErrorInternalServerError("query interaction_counter row failed")
		}
		// 行不存在
		if delta <= 0 {
			// 无行可递减，幂等 no-op
			return nil
		}
		if _, cerr := tx.InteractionCounter.Create().
			SetTenantID(tid).
			SetTargetType(uint8(targetType)).
			SetTargetID(targetID).
			SetMetric(uint8(metric)).
			SetCount(delta).
			Save(ctx); cerr != nil {
			r.log.Errorf("create interaction_counter row failed: %s", cerr.Error())
			return interactionV1.ErrorInternalServerError("create interaction_counter row failed")
		}
		return nil
	}

	// 行存在，递增
	newVal := *existing.Count + delta
	if newVal <= 0 {
		// 计数归 0，删行
		if derr := tx.InteractionCounter.DeleteOneID(existing.ID).Exec(ctx); derr != nil {
			r.log.Errorf("delete interaction_counter row failed: %s", derr.Error())
			return interactionV1.ErrorInternalServerError("delete interaction_counter row failed")
		}
		return nil
	}
	if _, uerr := tx.InteractionCounter.UpdateOneID(existing.ID).SetCount(newVal).Save(ctx); uerr != nil {
		r.log.Errorf("update interaction_counter row failed: %s", uerr.Error())
		return interactionV1.ErrorInternalServerError("update interaction_counter row failed")
	}
	return nil
}

// GetCounts 批量查询 interaction_counter 表中指定 (target_type, target_ids, metrics) 的计数。
// 返回 target_id → metric → count 的嵌套 map。未记录的 (target, metric) 不出现在响应中。
func (r *InteractionRepo) GetCounts(ctx context.Context, targetType interactionV1.TargetType, targetIDs []uint32, metrics []interactionV1.CounterMetric) (map[uint32]*interactionV1.CountMap, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant {
		// 未登录（SystemViewer，tenant_id==0）或无 viewer context：返回空结果
		// 而非 401。计数查询按 tenant 隔离，无 tenant 时无数据可返回，不抛错
		// 以免打断公开页渲染。登录用户由 hasTenant==true 分支按 tenant 过滤。
		return map[uint32]*interactionV1.CountMap{}, nil
	}
	if len(targetIDs) == 0 || len(metrics) == 0 {
		return map[uint32]*interactionV1.CountMap{}, nil
	}

	// 转换 metrics 为 uint8 谓词列表
	metricPreds := make([]uint8, 0, len(metrics))
	for _, m := range metrics {
		if m == interactionV1.CounterMetric_COUNTER_METRIC_UNSPECIFIED {
			continue
		}
		metricPreds = append(metricPreds, uint8(m))
	}
	if len(metricPreds) == 0 {
		return map[uint32]*interactionV1.CountMap{}, nil
	}

	rows, err := r.entClient.Client().InteractionCounter.Query().
		Where(
			interactioncounter.TenantIDEQ(tid),
			interactioncounter.TargetTypeEQ(uint8(targetType)),
			interactioncounter.TargetIDIn(targetIDs...),
			interactioncounter.MetricIn(metricPreds...),
		).
		All(ctx)
	if err != nil {
		r.log.Errorf("query interaction_counter batch failed: %s", err.Error())
		return nil, interactionV1.ErrorInternalServerError("query interaction_counter batch failed")
	}

	result := make(map[uint32]*interactionV1.CountMap, len(rows))
	for _, row := range rows {
		if row.TargetID == nil || row.Metric == nil || row.Count == nil {
			continue
		}
		tidVal := *row.TargetID
		metricVal := interactionV1.CounterMetric(*row.Metric)
		countVal := *row.Count
		cm, ok := result[tidVal]
		if !ok {
			cm = &interactionV1.CountMap{}
			result[tidVal] = cm
		}
		cm.Counts = append(cm.Counts, &interactionV1.MetricCount{
			Metric: metricVal,
			Count:  countVal,
		})
	}
	return result, nil
}

// GetInteractionStatus 批量查询当前 viewer 对指定评论的点赞状态。
func (r *InteractionRepo) GetInteractionStatus(ctx context.Context, viewerUserID uint32, targetType interactionV1.TargetType, targetIDs []uint32) (map[uint32]*interactionV1.InteractionStatus, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant || viewerUserID == 0 {
		return nil, interactionV1.ErrorUnauthorized("viewer identity required")
	}

	result := make(map[uint32]*interactionV1.InteractionStatus, len(targetIDs))
	for _, id := range targetIDs {
		result[id] = &interactionV1.InteractionStatus{Liked: false}
	}

	switch targetType {
	case interactionV1.TargetType_TARGET_TYPE_COMMENT:
		likedRows, err := r.entClient.Client().CommentLike.Query().
			Where(
				commentlike.TenantIDEQ(tid),
				commentlike.UserIDEQ(viewerUserID),
				commentlike.CommentIDIn(targetIDs...),
			).
			All(ctx)
		if err != nil {
			r.log.Errorf("query comment_like status failed: %s", err.Error())
			return nil, interactionV1.ErrorInternalServerError("query comment_like status failed")
		}
		for _, row := range likedRows {
			if row.CommentID != nil {
				if _, ok := result[*row.CommentID]; ok {
					result[*row.CommentID].Liked = true
				}
			}
		}

	default:
		return nil, interactionV1.ErrorBadRequest("invalid target type")
	}

	return result, nil
}

// PurgeTargetInteractions 清除单条评论上的全部点赞 ledger，
// 并在同一事务内把对应 interaction_counter 行归零删除。
// tenant 从 viewer context 提取，targetType/targetID 来自运营请求。
// 返回被删除的 ledger 行数。
func (r *InteractionRepo) PurgeTargetInteractions(ctx context.Context, targetType interactionV1.TargetType, targetID uint32) (uint32, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant {
		return 0, interactionV1.ErrorUnauthorized("viewer identity required")
	}

	var affected int
	txErr := r.txn(ctx, func(tx *ent.Tx) error {
		// 删该 target 在 comment_like 表的全部行
		switch targetType {
		case interactionV1.TargetType_TARGET_TYPE_COMMENT:
			n, err := tx.CommentLike.Delete().
				Where(
					commentlike.TenantIDEQ(tid),
					commentlike.CommentIDEQ(targetID),
				).
				Exec(ctx)
			if err != nil {
				r.log.Errorf("purge comment_like by target failed: %s", err.Error())
				return interactionV1.ErrorInternalServerError("purge comment_like failed")
			}
			affected += n
		default:
			return interactionV1.ErrorBadRequest("invalid target type")
		}

		// 同步 counter：对该 target 涉及的 metric，把计数归零（adjustCounterRow 的
		// delta = -当前计数 触发“归0删行”分支）。
		metrics := []interactionV1.CounterMetric{
			interactionV1.CounterMetric_COUNTER_METRIC_LIKE,
		}
		for _, m := range metrics {
			existing, qerr := tx.InteractionCounter.Query().
				Where(
					interactioncounter.TenantIDEQ(tid),
					interactioncounter.TargetTypeEQ(uint8(targetType)),
					interactioncounter.TargetIDEQ(targetID),
					interactioncounter.MetricEQ(uint8(m)),
				).
				Only(ctx)
			if qerr != nil {
				if ent.IsNotFound(qerr) {
					continue
				}
				r.log.Errorf("query counter row for purge failed: %s", qerr.Error())
				return interactionV1.ErrorInternalServerError("query counter row failed")
			}
			if existing.Count == nil || *existing.Count <= 0 {
				continue
			}
			if aerr := r.adjustCounterRow(ctx, tx, tid, targetID, targetType, m, -*existing.Count); aerr != nil {
				return aerr
			}
		}
		return nil
	})
	if txErr != nil {
		return 0, txErr
	}
	return uint32(affected), nil
}

// purgeBatchSize 是 PurgeUserInteractions 分批删除的批大小。
// 每批一个独立短事务，避免长事务锁表。
const purgeBatchSize = 200

// PurgeUserInteractions 清除指定用户在全站的全部评论点赞 ledger，分批短事务。
// 每批：取一批该用户的 comment_like 行 → 在独立事务内删该批 + 对每个受影响 target
// 回滚 counter（adjustCounterRow delta=-1）。
// tenant 从 viewer context 提取，userID 来自运营请求。
// 返回累计被删除的 ledger 行数。中途失败则返回已删行数 + 错误。
func (r *InteractionRepo) PurgeUserInteractions(ctx context.Context, userID uint32) (uint32, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant {
		return 0, interactionV1.ErrorUnauthorized("viewer identity required")
	}

	var totalAffected uint32

	// comment_likes（LIKE metric, target=comment）
	nComment, err := r.purgeUserBatchFromLedger(ctx, tid, userID, "comment_like",
		func(tx *ent.Tx, tid uint32, uid, targetID uint32) error {
			return r.adjustCounterRow(ctx, tx, tid, targetID, interactionV1.TargetType_TARGET_TYPE_COMMENT, interactionV1.CounterMetric_COUNTER_METRIC_LIKE, -1)
		},
		func(ctx context.Context, tid, uid uint32, offset, limit int) (ids []uint32, err error) {
			rows, qerr := r.entClient.Client().CommentLike.Query().
				Where(
					commentlike.TenantIDEQ(tid),
					commentlike.UserIDEQ(uid),
				).
				Limit(limit).
				Offset(offset).
				All(ctx)
			if qerr != nil {
				return nil, qerr
			}
			for _, row := range rows {
				if row.CommentID != nil {
					ids = append(ids, *row.CommentID)
				}
			}
			return ids, nil
		},
		func(tx *ent.Tx, ctx context.Context, tid, uid, targetID uint32) (int, error) {
			return tx.CommentLike.Delete().
				Where(
					commentlike.TenantIDEQ(tid),
					commentlike.UserIDEQ(uid),
					commentlike.CommentIDEQ(targetID),
				).
				Exec(ctx)
		},
	)
	if err != nil {
		r.log.Errorf("purge user comment_like batch failed: %s", err.Error())
		return totalAffected, err
	}
	totalAffected += nComment

	return totalAffected, nil
}

// purgeUserBatchFromLedger 分批删除某用户在某 ledger 表的全部行。
// 返回累计被删除的行数。
// queryBatch: 取一批行的 targetID 列表（带 tenant+user 谓词，limit/offset 分页）
// deleteOne: 在 tx 内删单个 (user, target) ledger 行，返回受影响行数
// counterAdjust: 在同一 tx 内对受影响 target 回滚 counter（delta=-1）
// 每批一个独立 txn。命中 0 行时停止该表。
func (r *InteractionRepo) purgeUserBatchFromLedger(
	ctx context.Context,
	tid uint32,
	uid uint32,
	label string,
	counterAdjust func(tx *ent.Tx, tid uint32, uid, targetID uint32) error,
	queryBatch func(ctx context.Context, tid, uid uint32, offset, limit int) ([]uint32, error),
	deleteOne func(tx *ent.Tx, ctx context.Context, tid, uid, targetID uint32) (int, error),
) (uint32, error) {
	var affected uint32
	for {
		// 每批从头查（offset=0）：上一批删了行后，剩余行前移到头部。
		ids, qerr := queryBatch(ctx, tid, uid, 0, purgeBatchSize)
		if qerr != nil {
			r.log.Errorf("query %s batch failed: %s", label, qerr.Error())
			return affected, interactionV1.ErrorInternalServerError("query ledger batch failed")
		}
		if len(ids) == 0 {
			return affected, nil
		}
		var batchDeleted uint32
		txErr := r.txn(ctx, func(tx *ent.Tx) error {
			for _, targetID := range ids {
				n, derr := deleteOne(tx, ctx, tid, uid, targetID)
				if derr != nil {
					return derr
				}
				if n > 0 {
					affected += uint32(n)
					batchDeleted += uint32(n)
					if aerr := counterAdjust(tx, tid, uid, targetID); aerr != nil {
						return aerr
					}
				}
			}
			return nil
		})
		if txErr != nil {
			return affected, txErr
		}
		// 本批未删任何行却仍查到 ids：避免死循环，退出。
		if batchDeleted == 0 {
			return affected, nil
		}
	}
}

// ResetCounter 按 ledger 真实计数重算指定 (target, metric) 的 interaction_counter 行。
// recount==0 → 删行；recount>0 → set 绝对值；行不存在 + recount>0 → 创建。
// 用于修复 counter 与 ledger 不一致的漂移。
func (r *InteractionRepo) ResetCounter(ctx context.Context, targetType interactionV1.TargetType, targetID uint32, metric interactionV1.CounterMetric) (int64, error) {
	tid, hasTenant := maybeTenantFromViewer(ctx)
	if !hasTenant {
		return 0, interactionV1.ErrorUnauthorized("viewer identity required")
	}

	// 按 ledger 真实计数
	var recount int64
	switch metric {
	case interactionV1.CounterMetric_COUNTER_METRIC_LIKE:
		switch targetType {
		case interactionV1.TargetType_TARGET_TYPE_COMMENT:
			c, err := r.entClient.Client().CommentLike.Query().
				Where(
					commentlike.TenantIDEQ(tid),
					commentlike.CommentIDEQ(targetID),
				).
				Count(ctx)
			if err != nil {
				r.log.Errorf("count comment_like for recount failed: %s", err.Error())
				return 0, interactionV1.ErrorInternalServerError("count ledger failed")
			}
			recount = int64(c)
		default:
			return 0, interactionV1.ErrorBadRequest("invalid target type")
		}
	default:
		return 0, interactionV1.ErrorBadRequest("invalid metric")
	}

	// 同步 counter 行到 recount。用 adjustCounterRow 把现有计数推到 recount：
	// delta = recount - 现有计数。现有计数在 tx 内查。
	txErr := r.txn(ctx, func(tx *ent.Tx) error {
		var cur int64
		existing, qerr := tx.InteractionCounter.Query().
			Where(
				interactioncounter.TenantIDEQ(tid),
				interactioncounter.TargetTypeEQ(uint8(targetType)),
				interactioncounter.TargetIDEQ(targetID),
				interactioncounter.MetricEQ(uint8(metric)),
			).
			Only(ctx)
		if qerr != nil {
			if !ent.IsNotFound(qerr) {
				r.log.Errorf("query counter row for reset failed: %s", qerr.Error())
				return interactionV1.ErrorInternalServerError("query counter row failed")
			}
			cur = 0
		} else if existing.Count != nil {
			cur = *existing.Count
		}
		delta := recount - cur
		if delta == 0 {
			return nil
		}
		return r.adjustCounterRow(ctx, tx, tid, targetID, targetType, metric, delta)
	})
	if txErr != nil {
		return 0, txErr
	}
	return recount, nil
}
