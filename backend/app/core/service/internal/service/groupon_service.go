package service

import (
	"context"
	"fmt"
	"time"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/grouponactivity"
	"go-wind-shop/app/core/service/internal/data/ent/groupongroup"
	"go-wind-shop/app/core/service/internal/data/ent/order"
	"go-wind-shop/app/core/service/internal/data/ent/paymentrefund"
	"go-wind-shop/app/core/service/internal/data/ent/paymenttransaction"

	"google.golang.org/protobuf/types/known/emptypb"

	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"

	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

type GrouponService struct {
	catalogV1.UnimplementedGrouponActivityServiceServer

	log *logx.Helper

	repo          *data.GrouponActivityRepo
	entClient     *entCrud.EntClient[*ent.Client]
	refundService *PaymentRefundService
	taskScheduler TaskScheduler
}

func NewGrouponService(ctx *bootstrap.Context, repo *data.GrouponActivityRepo, entClient *entCrud.EntClient[*ent.Client]) *GrouponService {
	return &GrouponService{
		log:       logx.New(ctx.NewLoggerHelper("groupon/service/core-service")),
		repo:      repo,
		entClient: entClient,
	}
}

// RegisterRefundService 由 wiring 在构造完成后注入（构造期 refundService 尚未
// 创建——refund 依赖 paymentTransaction 依赖 order 依赖 groupon，构造图成环）。
func (s *GrouponService) RegisterRefundService(refundService *PaymentRefundService) {
	s.refundService = refundService
}

// RegisterTaskScheduler 由 server/asynq_server.go 在启动期注入 asynq server。
func (s *GrouponService) RegisterTaskScheduler(ts TaskScheduler) {
	s.taskScheduler = ts
}

func (s *GrouponService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListGrouponActivityResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *GrouponService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountGrouponActivityResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}
	return &catalogV1.CountGrouponActivityResponse{Count: uint64(count)}, nil
}

func (s *GrouponService) Get(ctx context.Context, req *catalogV1.GetGrouponActivityRequest) (*catalogV1.GrouponActivity, error) {
	return s.repo.Get(ctx, req)
}

func (s *GrouponService) Create(ctx context.Context, req *catalogV1.CreateGrouponActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}
	if err := s.validate(req.Data); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GrouponService) Update(ctx context.Context, req *catalogV1.UpdateGrouponActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}
	if err := s.validate(req.Data); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GrouponService) Delete(ctx context.Context, req *catalogV1.DeleteGrouponActivityRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GrouponService) validate(data *catalogV1.GrouponActivity) error {
	if data.GetSkuId() == 0 {
		return catalogV1.ErrorBadRequest("groupon requires sku_id")
	}
	if data.GetGroupPrice() <= 0 {
		return catalogV1.ErrorBadRequest("group_price must be positive")
	}
	if data.GetGroupSize() < 2 {
		data.GroupSize = trans.Ptr(int32(2))
	}
	if data.GetGroupValidHours() <= 0 {
		data.GroupValidHours = trans.Ptr(int32(24))
	}
	if data.GetValidFrom() != nil && data.GetValidUntil() != nil &&
		data.GetValidFrom().AsTime().After(data.GetValidUntil().AsTime()) {
		return catalogV1.ErrorBadRequest("valid_from must not be after valid_until")
	}
	return nil
}

// ListActiveGroupons 当前生效中的拼团活动（C端商品页拼团价展示）。
func (s *GrouponService) ListActiveGroupons(ctx context.Context, req *catalogV1.ListActiveGrouponsRequest) (*catalogV1.ListGrouponActivityResponse, error) {
	now := time.Now()
	list, err := s.entClient.Client().GrouponActivity.Query().
		Where(
			grouponactivity.StatusEQ(grouponactivity.StatusGrouponStatusActive),
			grouponactivity.ValidFromLTE(now),
			grouponactivity.ValidUntilGTE(now),
		).
		All(ctx)
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("query failed")
	}
	items := make([]*catalogV1.GrouponActivity, 0, len(list))
	for _, it := range list {
		items = append(items, s.toActivityDTO(it))
	}
	return &catalogV1.ListGrouponActivityResponse{Items: items, Total: uint64(len(items))}, nil
}

// ListGroups 活动下的团列表（status 可选过滤）。
func (s *GrouponService) ListGroups(ctx context.Context, req *catalogV1.ListGrouponGroupsRequest) (*catalogV1.ListGrouponGroupResponse, error) {
	if req.GetActivityId() == 0 {
		return nil, catalogV1.ErrorBadRequest("activity_id is required")
	}
	q := s.entClient.Client().GrouponGroup.Query().Where(groupongroup.ActivityIDEQ(req.GetActivityId()))
	switch req.GetStatus() {
	case "OPENING":
		q = q.Where(groupongroup.StatusEQ(groupongroup.StatusGrouponGroupStatusOpening))
	case "SUCCESS":
		q = q.Where(groupongroup.StatusEQ(groupongroup.StatusGrouponGroupStatusSuccess))
	case "FAILED":
		q = q.Where(groupongroup.StatusEQ(groupongroup.StatusGrouponGroupStatusFailed))
	}
	list, err := q.Order(ent.Desc(groupongroup.FieldID)).All(ctx)
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("query failed")
	}
	items := make([]*catalogV1.GrouponGroup, 0, len(list))
	for _, g := range list {
		items = append(items, s.toGroupDTO(g))
	}
	return &catalogV1.ListGrouponGroupResponse{Items: items, Total: uint64(len(items))}, nil
}

// GetGroup 团详情。
func (s *GrouponService) GetGroup(ctx context.Context, req *catalogV1.GetGrouponGroupRequest) (*catalogV1.GrouponGroup, error) {
	g, err := s.entClient.Client().GrouponGroup.Query().
		Where(groupongroup.IDEQ(req.GetId())).
		Only(ctx)
	if err != nil {
		return nil, catalogV1.ErrorNotFound("groupon group not found")
	}
	return s.toGroupDTO(g), nil
}

// ApplyGrouponInTx 下单事务内的拼团处理（订单服务调用）：
//   - 校验活动（ACTIVE + 窗口 + SKU 匹配）；
//   - 参团：行锁锁定团，校验 OPENING/未过期/未满员，member_count+1；
//   - 开团：返回 isNewGroup=true，由调用方在订单落库后调 CreateGroupForLeaderTx。
//
// 返回 (团ID[参团时], 成团价, 是否开新团, 错误)。拼团价为结算最高优先级。
func (s *GrouponService) ApplyGrouponInTx(ctx context.Context, tx *ent.Tx, activityId, joinGroupId, userId, skuId uint32) (groupId uint32, groupPrice int64, isNewGroup bool, err error) {
	act, aErr := tx.GrouponActivity.Query().
		Where(grouponactivity.IDEQ(activityId)).
		ForUpdate().
		Only(ctx)
	if aErr != nil || act == nil {
		return 0, 0, false, orderV1.ErrorBadRequest("groupon activity not found")
	}
	if act.Status == nil || *act.Status != grouponactivity.StatusGrouponStatusActive {
		return 0, 0, false, orderV1.ErrorBadRequest("groupon activity is not active")
	}
	now := time.Now()
	if (act.ValidFrom != nil && now.Before(*act.ValidFrom)) || (act.ValidUntil != nil && now.After(*act.ValidUntil)) {
		return 0, 0, false, orderV1.ErrorBadRequest("groupon activity is not in valid window")
	}
	if act.SkuID == nil || *act.SkuID != skuId {
		return 0, 0, false, orderV1.ErrorBadRequest("groupon activity does not match sku")
	}

	if joinGroupId != 0 {
		g, gErr := tx.GrouponGroup.Query().
			Where(groupongroup.IDEQ(joinGroupId)).
			ForUpdate().
			Only(ctx)
		if gErr != nil || g == nil {
			return 0, 0, false, orderV1.ErrorBadRequest("groupon group not found")
		}
		if g.Status == nil || *g.Status != groupongroup.StatusGrouponGroupStatusOpening {
			return 0, 0, false, orderV1.ErrorBadRequest("groupon group is not opening")
		}
		if g.ExpireAt != nil && now.After(*g.ExpireAt) {
			return 0, 0, false, orderV1.ErrorBadRequest("groupon group expired")
		}
		if g.ActivityID == nil || *g.ActivityID != activityId {
			return 0, 0, false, orderV1.ErrorBadRequest("groupon group does not match activity")
		}
		if g.MemberCount >= g.TargetSize {
			return 0, 0, false, orderV1.ErrorBadRequest("groupon group is full")
		}
		if uErr := tx.GrouponGroup.UpdateOneID(joinGroupId).SetMemberCount(g.MemberCount + 1).Exec(ctx); uErr != nil {
			return 0, 0, false, orderV1.ErrorInternalServerError("join groupon group failed")
		}
		price := int64(0)
		if act.GroupPrice != nil {
			price = *act.GroupPrice
		}
		return joinGroupId, price, false, nil
	}

	price := int64(0)
	if act.GroupPrice != nil {
		price = *act.GroupPrice
	}
	return 0, price, true, nil
}

// CreateGroupForLeaderTx 订单落库后为团长创建团（同事务；订单 ID 此时已知）。
func (s *GrouponService) CreateGroupForLeaderTx(ctx context.Context, tx *ent.Tx, activityId, leaderOrderId uint32) (groupId uint32, err error) {
	act, aErr := tx.GrouponActivity.Query().Where(grouponactivity.IDEQ(activityId)).Only(ctx)
	if aErr != nil || act == nil {
		return 0, orderV1.ErrorBadRequest("groupon activity not found")
	}
	hours := int32(24)
	if act.GroupValidHours > 0 {
		hours = act.GroupValidHours
	}
	target := int32(2)
	if act.GroupSize >= 2 {
		target = act.GroupSize
	}
	expire := time.Now().Add(time.Duration(hours) * time.Hour)
	status := groupongroup.StatusGrouponGroupStatusOpening
	// 先落团再查 ID（避免 OnConflict+ID 兼容性问题）
	cErr := tx.GrouponGroup.Create().
		SetActivityID(activityId).
		SetLeaderOrderID(leaderOrderId).
		SetMemberCount(1).
		SetPaidCount(0).
		SetTargetSize(target).
		SetStatus(status).
		SetExpireAt(expire).
		SetCreatedAt(time.Now()).
		Exec(ctx)
	if cErr != nil {
		s.log.Errorf("create groupon group exec failed: %v", cErr)
		return 0, orderV1.ErrorInternalServerError("create groupon group failed")
	}
	// 回查拿自增 ID（团长订单唯一 → 天然定位）
	g, qErr := tx.GrouponGroup.Query().
		Where(
			groupongroup.LeaderOrderIDEQ(leaderOrderId),
			groupongroup.ActivityIDEQ(activityId),
		).
		First(ctx)
	if qErr != nil || g == nil {
		s.log.Errorf("create groupon group query back failed: %v", qErr)
		return 0, orderV1.ErrorInternalServerError("create groupon group lookup failed")
	}
	gid := g.ID
	if uErr := tx.Order.UpdateOneID(leaderOrderId).
		SetNillableGrouponGroupID(&gid).
		SetUpdatedAt(time.Now()).
		Exec(ctx); uErr != nil {
		return 0, orderV1.ErrorInternalServerError("bind groupon group to leader order failed")
	}
	s.log.Infof("groupon group created: group=%d activity=%d leader_order=%d expire=%s", gid, activityId, leaderOrderId, expire.Format(time.RFC3339))
	return gid, nil
}

// HandleGrouponSweep 拼团清扫周期任务处理器（asynq 注册，SystemViewer 上下文）。
//
// 对 OPENING 团实时统计已支付成员数（PAID/FULFILLED/CLOSED）：
//   - 达标 → 翻 SUCCESS（成员订单照常履约）
//   - 过期未成团 → 翻 FAILED，并对已支付成员订单自动退款（创建+确认走既有
//     退款自动化：渠道退款/流水 REFUNDED/库存回补/站内信/积分回冲）+ 关闭订单
//
// 未支付成员订单不在此处理——order_sweep 的超时取消会兜底。
func (s *GrouponService) HandleGrouponSweep(taskType string, taskData *task.GrouponSweepTaskData) error {
	ctx := appViewer.NewSystemViewerContext(context.Background())
	s.log.Infow("msg", "groupon sweep started", "task_type", taskType)

	opening, err := s.entClient.Client().GrouponGroup.Query().
		Where(groupongroup.StatusEQ(groupongroup.StatusGrouponGroupStatusOpening)).
		All(ctx)
	if err != nil {
		s.log.Errorf("groupon sweep query opening groups failed: %v", err)
		return err
	}

	success, failed := 0, 0
	for _, g := range opening {
		if g.ID == 0 || g.ActivityID == nil {
			continue
		}
		paid, pErr := s.entClient.Client().Order.Query().
			Where(
				order.GrouponGroupIDEQ(g.ID),
				order.StatusIn(
					order.StatusOrderStatusPaid,
					order.StatusOrderStatusFulfilled,
					order.StatusOrderStatusClosed,
				),
			).
			Count(ctx)
		if pErr != nil {
			s.log.Errorf("groupon sweep count paid for group [%d] failed: %v", g.ID, pErr)
			continue
		}

		expired := g.ExpireAt != nil && time.Now().After(*g.ExpireAt)
		target := int(g.TargetSize)
		if target < 2 {
			target = 2
		}

		if paid >= target {
			done := groupongroup.StatusGrouponGroupStatusSuccess
			if uErr := s.entClient.Client().GrouponGroup.UpdateOneID(g.ID).
				SetStatus(done).
				SetPaidCount(int32(paid)).
				SetUpdatedAt(time.Now()).
				Exec(ctx); uErr != nil {
				s.log.Errorf("groupon sweep mark SUCCESS for group [%d] failed: %v", g.ID, uErr)
				continue
			}
			s.log.Infow("msg", "groupon group succeeded", "group_id", g.ID, "paid", paid)
			success++
			continue
		}
		if !expired {
			_ = s.entClient.Client().GrouponGroup.UpdateOneID(g.ID).
				SetPaidCount(int32(paid)).
				Exec(ctx)
			continue
		}

		failedStatus := groupongroup.StatusGrouponGroupStatusFailed
		if uErr := s.entClient.Client().GrouponGroup.UpdateOneID(g.ID).
			SetStatus(failedStatus).
			SetPaidCount(int32(paid)).
			SetUpdatedAt(time.Now()).
			Exec(ctx); uErr != nil {
			s.log.Errorf("groupon sweep mark FAILED for group [%d] failed: %v", g.ID, uErr)
			continue
		}
		s.autoRefundPaidMembers(ctx, g.ID)
		failed++
	}

	s.log.Infow("msg", "groupon sweep done", "opening", len(opening), "success", success, "failed", failed)
	return nil
}

// autoRefundPaidMembers 对拼团失败团的已支付成员订单执行自动退款 + 关闭订单。
func (s *GrouponService) autoRefundPaidMembers(ctx context.Context, groupId uint32) {
	members, err := s.entClient.Client().Order.Query().
		Where(
			order.GrouponGroupIDEQ(groupId),
			order.StatusEQ(order.StatusOrderStatusPaid),
		).
		All(ctx)
	if err != nil {
		s.log.Errorf("groupon refund query members for group [%d] failed: %v", groupId, err)
		return
	}

	for _, o := range members {
		if o.ID == 0 {
			continue
		}
		txs, tErr := s.entClient.Client().PaymentTransaction.Query().
			Where(
				paymenttransaction.OrderIDEQ(o.ID),
				paymenttransaction.StatusEQ(paymenttransaction.StatusPaymentTransactionStatusSucceeded),
			).
			All(ctx)
		if tErr != nil {
			s.log.Errorf("groupon refund query tx for order [%d] failed: %v", o.ID, tErr)
			continue
		}

		for _, txEnt := range txs {
			txId := txEnt.ID
			if txId == 0 {
				continue
			}
			if cErr := s.refundService.CreateAuto(ctx, txId, itDerefInt64(txEnt.Amount), "拼团失败自动退款"); cErr != nil {
				s.log.Errorf("groupon auto refund create for order [%d] failed: %v", o.ID, cErr)
				continue
			}
			refundId, gErr := s.latestRefundId(ctx, txId)
			if gErr != nil || refundId == 0 {
				continue
			}
			if aErr := s.refundService.ApproveAuto(ctx, refundId); aErr != nil {
				s.log.Errorf("groupon auto refund approve [%d] failed: %v", refundId, aErr)
				continue
			}
		}

		// 订单关闭（PAID→CANCELLED，直接 ent 更新带状态条件；库存已由退款
		// 联动回补，不走 CancelOrder 以免二次回补）
		n, uErr := s.entClient.Client().Order.Update().
			SetStatus(order.StatusOrderStatusCancelled).
			SetUpdatedAt(time.Now()).
			Where(
				order.IDEQ(o.ID),
				order.StatusEQ(order.StatusOrderStatusPaid),
			).
			Save(ctx)
		if uErr != nil || n == 0 {
			s.log.Errorf("groupon close order [%d] failed: %v (n=%d)", o.ID, uErr, n)
			continue
		}
		s.log.Infow("msg", "groupon failed member order refunded and closed", "group_id", groupId, "order_id", o.ID)
	}
}

func (s *GrouponService) latestRefundId(ctx context.Context, txId uint32) (uint32, error) {
	rf, err := s.entClient.Client().PaymentRefund.Query().
		Where(paymentrefund.TransactionIDEQ(txId)).
		Order(ent.Desc(paymentrefund.FieldID)).
		First(ctx)
	if err != nil || rf == nil {
		return 0, fmt.Errorf("refund lookup failed: %v", err)
	}
	return rf.ID, nil
}

func (s *GrouponService) toActivityDTO(it *ent.GrouponActivity) *catalogV1.GrouponActivity {
	dto := &catalogV1.GrouponActivity{
		Id:              trans.Ptr(it.ID),
		Name:            trans.Ptr(it.Name),
		SkuId:           it.SkuID,
		GroupPrice:      it.GroupPrice,
		GroupSize:       trans.Ptr(it.GroupSize),
		GroupValidHours: trans.Ptr(it.GroupValidHours),
		PerUserLimit:    trans.Ptr(it.PerUserLimit),
		ValidFrom:       timeToTspb(it.ValidFrom),
		ValidUntil:      timeToTspb(it.ValidUntil),
		Currency:        trans.Ptr(it.Currency),
		TenantId:        it.TenantID,
	}
	if it.Status != nil {
		if v, ok := catalogV1.GrouponActivity_Status_value[string(*it.Status)]; ok {
			dto.Status = catalogV1.GrouponActivity_Status(v).Enum()
		}
	}
	if it.CreatedAt != nil {
		dto.CreatedAt = timeToTspb(it.CreatedAt)
	}
	if it.UpdatedAt != nil {
		dto.UpdatedAt = timeToTspb(it.UpdatedAt)
	}
	return dto
}

func (s *GrouponService) toGroupDTO(g *ent.GrouponGroup) *catalogV1.GrouponGroup {
	dto := &catalogV1.GrouponGroup{
		Id:            trans.Ptr(g.ID),
		ActivityId:    g.ActivityID,
		LeaderOrderId: g.LeaderOrderID,
		MemberCount:   trans.Ptr(g.MemberCount),
		PaidCount:     trans.Ptr(g.PaidCount),
		TargetSize:    trans.Ptr(g.TargetSize),
		ExpireAt:      timeToTspb(g.ExpireAt),
	}
	if g.Status != nil {
		if v, ok := catalogV1.GrouponGroup_Status_value[string(*g.Status)]; ok {
			dto.Status = catalogV1.GrouponGroup_Status(v).Enum()
		}
	}
	if g.CreatedAt != nil {
		dto.CreatedAt = timeToTspb(g.CreatedAt)
	}
	return dto
}

func timeToTspb(t *time.Time) *tspb.Timestamp {
	if t == nil {
		return nil
	}
	return tspb.New(*t)
}

func itDerefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
