package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/cart"
	"go-wind-shop/app/core/service/internal/data/ent/cartitem"
	"go-wind-shop/app/core/service/internal/data/ent/coupontemplate"
	"go-wind-shop/app/core/service/internal/data/ent/skuprice"
	"go-wind-shop/app/core/service/internal/data/ent/usercoupon"

	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/go-crud/viewer"

	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"

	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
)

type UserCouponService struct {
	couponV1.UnimplementedUserCouponServiceServer

	log *log.Helper

	repo *data.UserCouponRepo

	couponTemplateRepo *data.CouponTemplateRepo

	entClient *entCrud.EntClient[*ent.Client]
}

func NewUserCouponService(
	ctx *bootstrap.Context,
	repo *data.UserCouponRepo,
	couponTemplateRepo *data.CouponTemplateRepo,
	entClient *entCrud.EntClient[*ent.Client],
) *UserCouponService {
	return &UserCouponService{
		log:                ctx.NewLoggerHelper("user-coupon/service/core-service"),
		repo:               repo,
		couponTemplateRepo: couponTemplateRepo,
		entClient:          entClient,
	}
}

func (s *UserCouponService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListUserCouponResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *UserCouponService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.CountUserCouponResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &couponV1.CountUserCouponResponse{
		Count: uint64(count),
	}, nil
}

func (s *UserCouponService) Get(ctx context.Context, req *couponV1.GetUserCouponRequest) (*couponV1.UserCoupon, error) {
	return s.repo.Get(ctx, req)
}

func (s *UserCouponService) Create(ctx context.Context, req *couponV1.CreateUserCouponRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	// per-user 限领校验：该用户该模板已领张数（含所有状态，防止领→用→退→再领绕限）
	// >= max_redemptions_per_user 则拒。
	// 注意：此校验无事务包裹，存在 TOCTOU 竞态。但发放是 admin 低频操作，
	// 且核销侧 redeemCouponInTx 的事务内行锁 count 会兜底——竞态最多导致
	// "多领了几张但用不出去"，可接受。
	tmplId := req.Data.GetCouponTemplateId()
	uid := req.Data.GetUserId()
	if tmplId != 0 && uid != 0 {
		tmpl, gErr := s.couponTemplateRepo.Get(ctx, &couponV1.GetCouponTemplateRequest{
			QueryBy: &couponV1.GetCouponTemplateRequest_Id{Id: tmplId},
		})
		if gErr != nil || tmpl == nil {
			return nil, couponV1.ErrorBadRequest("coupon template not found")
		}
		if tmpl.GetMaxRedemptionsPerUser() > 0 {
			issuedCount, cErr := s.entClient.Client().UserCoupon.Query().
				Where(
					usercoupon.And(
						usercoupon.UserIDEQ(uid),
						usercoupon.CouponTemplateIDEQ(tmplId),
					),
				).
				Count(ctx)
			if cErr != nil {
				return nil, couponV1.ErrorInternalServerError("per-user issuance check failed")
			}
			if issuedCount >= int(tmpl.GetMaxRedemptionsPerUser()) {
				return nil, couponV1.ErrorBadRequest("coupon per-user issuance limit reached")
			}
		}
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Claim 买家自助领取公开可领模板。
//
// 原子性：仿 order_service.redeemCouponInTx 的 ForUpdate + tx 范式。
//   - user_id 从 viewer 强制（不从请求取，仿 interaction_admin_service L48-55）。
//   - ForUpdate 锁模板行，锁内校验 claimable / status / 有效窗口 / 全局核销上限。
//   - per-user 限领校验在 tx 内 Count（status 不过滤——含所有状态，防领→用→退→再领绕限）。
//   - tx 内 Create user_coupon（status 强制 UNUSED）。UserPrivacy 的 Create 分支会强制 user_id=viewer，双重保护。
//   - 不增 redeemed_count（那是核销计数，核销时才增）。
//   - 任一校验失败 → return err → defer Rollback（fail-closed）。
func (s *UserCouponService) Claim(ctx context.Context, req *couponV1.ClaimCouponRequest) (*emptypb.Empty, error) {
	// 0. 从 viewer 取 user_id（不从请求取）。
	vc, exist := viewer.FromContext(ctx)
	if !exist {
		return nil, couponV1.ErrorForbidden("missing viewer context")
	}
	uid := vc.UserID()
	if uid == 0 {
		return nil, couponV1.ErrorForbidden("viewer has no user id")
	}
	uidU32 := uint32(uid)
	tmplId := req.GetCouponTemplateId()
	if tmplId == 0 {
		return nil, couponV1.ErrorBadRequest("invalid template id")
	}

	// 1. 开事务。
	tx, err := s.entClient.Client().Tx(ctx)
	if err != nil {
		s.log.Errorf("claim coupon open tx failed: %v", err)
		return nil, couponV1.ErrorInternalServerError("claim failed")
	}
	defer func() { _ = tx.Rollback() }()

	// 2. ForUpdate 锁定模板行。
	tmpl, tErr := tx.CouponTemplate.Query().
		Where(coupontemplate.IDEQ(tmplId)).
		ForUpdate().
		Only(ctx)
	if tErr != nil || tmpl == nil {
		s.log.Warnf("claim: template [%d] not found: %v", tmplId, tErr)
		return nil, couponV1.ErrorBadRequest("coupon template not found")
	}

	// 3. 锁内校验。
	// 3a. claimable == true。
	if tmpl.Claimable == nil || !*tmpl.Claimable {
		return nil, couponV1.ErrorBadRequest("coupon template is not claimable")
	}
	// 3b. status == ACTIVE。
	if tmpl.Status == nil || *tmpl.Status != coupontemplate.StatusCouponTemplateStatusActive {
		return nil, couponV1.ErrorBadRequest("coupon template is inactive")
	}
	// 3c. 有效窗口。
	if !couponApplicableNowEntity(tmpl, time.Now()) {
		return nil, couponV1.ErrorBadRequest("coupon is not within its valid window")
	}
	// 3d. 全局核销上限（防领了用不出去）。
	if tmpl.MaxRedemptions != nil && *tmpl.MaxRedemptions > 0 {
		if tmpl.RedeemedCount == nil || *tmpl.RedeemedCount >= *tmpl.MaxRedemptions {
			return nil, couponV1.ErrorBadRequest("coupon redemption limit reached")
		}
	}

	// 4. per-user 限领校验（tx 内 Count，防 TOCTOU）。
	//    status 不过滤——含所有状态（UNUSED/USED/EXPIRED），与 admin Create 路径一致：
	//    防领→用→退→再领绕 per-user 限领。
	//    UserPrivacy 按 caller user_id 自动注入 WHERE（双重保护）。
	if tmpl.MaxRedemptionsPerUser != nil && *tmpl.MaxRedemptionsPerUser > 0 {
		issuedCount, cErr := tx.UserCoupon.Query().
			Where(
				usercoupon.And(
					usercoupon.UserIDEQ(uidU32),
					usercoupon.CouponTemplateIDEQ(tmplId),
				),
			).
			Count(ctx)
		if cErr != nil {
			s.log.Errorf("claim: per-user issuance count failed for user [%d] template [%d]: %v", uidU32, tmplId, cErr)
			return nil, couponV1.ErrorInternalServerError("per-user issuance check failed")
		}
		if issuedCount >= int(*tmpl.MaxRedemptionsPerUser) {
			s.log.Warnf("claim: per-user issuance limit reached for user [%d] template [%d]: %d >= %d", uidU32, tmplId, issuedCount, *tmpl.MaxRedemptionsPerUser)
			return nil, couponV1.ErrorBadRequest("coupon per-user issuance limit reached")
		}
	}

	// 5. tx 内 Create user_coupon（status 强制 UNUSED，user_id 强制 viewer）。
	//    UserPrivacy 的 Create 分支会强制 user_id=viewer，此处显式设双重保护。
	unusedStatus := usercoupon.StatusUserCouponStatusUnused
	now := time.Now()
	if cErr := tx.UserCoupon.Create().
		SetNillableUserID(&uidU32).
		SetNillableCouponTemplateID(&tmplId).
		SetNillableStatus(&unusedStatus).
		SetNillableTenantID(tmpl.TenantID).
		SetNillableCreatedBy(&uidU32).
		SetCreatedAt(now).
		Exec(ctx); cErr != nil {
		s.log.Errorf("claim: insert user_coupon failed for user [%d] template [%d]: %v", uidU32, tmplId, cErr)
		return nil, couponV1.ErrorInternalServerError("claim failed")
	}

	// 6. 提交事务。
	if cErr := tx.Commit(); cErr != nil {
		s.log.Errorf("claim: commit tx failed: %v", cErr)
		return nil, couponV1.ErrorInternalServerError("claim failed")
	}

	s.log.Infof("claim: user [%d] claimed template [%d]", uidU32, tmplId)
	return &emptypb.Empty{}, nil
}

func (s *UserCouponService) Update(ctx context.Context, req *couponV1.UpdateUserCouponRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *UserCouponService) Delete(ctx context.Context, req *couponV1.DeleteUserCouponRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Quote 试算（报价预览）。
//
// 在当前登录用户的购物车快照上计算折前/折后金额，不持锁、不落库。
// 用户隔离由 UserPrivacy（user_coupon / cart / cartitem）自动按 viewer.user_id 注入
// WHERE 条件，跨用户 user_coupon_id 查不到 → 返回 applicable=false。本层 additionally
// 显式带 user_id/tenant_id 过滤购物车，与 OrderService.Create 的双重保护模式一致。
//
// 最终抵扣以下单时 Create 事务内校验为准；预览非预留，高并发下可能被他人先核销。
func (s *UserCouponService) Quote(ctx context.Context, req *couponV1.QuoteRequest) (*couponV1.QuoteResponse, error) {
	// 默认不适用响应。
	resp := &couponV1.QuoteResponse{
		Applicable:        boolPtr(false),
		PreDiscountTotal:  int64Ptr(0),
		Discount:          int64Ptr(0),
		PostDiscountTotal: int64Ptr(0),
	}

	couponId := req.GetUserCouponId()
	if couponId == 0 {
		return resp, nil
	}

	// 1. 读券实例（只读快照）。UserPrivacy 按 viewer.user_id 自动注入 WHERE，
	//    跨用户 coupon_id 命中不到 → 返回不适用。
	uc, err := s.repo.Get(ctx, &couponV1.GetUserCouponRequest{
		QueryBy: &couponV1.GetUserCouponRequest_Id{Id: couponId},
	})
	if err != nil || uc == nil {
		return resp, nil
	}
	if uc.GetStatus() != couponV1.UserCoupon_UNUSED {
		return resp, nil
	}

	// 2. 读关联模板（只读快照）。TenantPrivacy 按 viewer.tenant_id 自动注入 WHERE。
	tmpl, err := s.couponTemplateRepo.Get(ctx, &couponV1.GetCouponTemplateRequest{
		QueryBy: &couponV1.GetCouponTemplateRequest_Id{Id: uc.GetCouponTemplateId()},
	})
	if err != nil || tmpl == nil {
		return resp, nil
	}
	if tmpl.GetStatus() != couponV1.CouponTemplate_ACTIVE {
		return resp, nil
	}
	if !couponApplicableNow(tmpl, time.Now()) {
		return resp, nil
	}
	if tmpl.GetMaxRedemptions() > 0 && tmpl.GetRedeemedCount() >= tmpl.GetMaxRedemptions() {
		return resp, nil
	}

	// per-user 限用预览校验：与 redeemCouponInTx 对称——该用户该模板已 USED 的券数
	// >= max_redemptions_per_user 则不适用。Quote 是只读快照不持锁，高并发下可能
	// 脱节，但 redeemCouponInTx 的事务内行锁 count 会兜底。
	if tmpl.GetMaxRedemptionsPerUser() > 0 {
		uid, ok := viewerUserIDFromContext(ctx)
		if !ok {
			return resp, nil
		}
		usedCount, cErr := s.entClient.Client().UserCoupon.Query().
			Where(
				usercoupon.And(
					usercoupon.UserIDEQ(uid),
					usercoupon.CouponTemplateIDEQ(uc.GetCouponTemplateId()),
					usercoupon.StatusEQ(usercoupon.StatusUserCouponStatusUsed),
				),
			).
			Count(ctx)
		if cErr != nil || usedCount >= int(tmpl.GetMaxRedemptionsPerUser()) {
			return resp, nil
		}
	}

	// 3. 读当前 viewer 的购物车（只读，无锁无扣减），算折前总额。
	//    与 OrderService.Create 的 cart-iter 同源，仅累加价格行，不改库存。
	preDiscountTotal, cartErr := s.computeCartPreDiscountTotal(ctx)
	if cartErr != nil || preDiscountTotal <= 0 {
		return resp, nil
	}

	// 币种须与模板一致。当前系统仅支持 CNY（Currency mixin 默认值），故结算币固定 CNY。
	if tmpl.GetCurrency() != "" && tmpl.GetCurrency() != "CNY" {
		return resp, nil
	}

	// 4. 计算抵扣（与 Create 共用 computeDiscount，单一真相源）。
	discount, applicable := computeDiscount(preDiscountTotal, discountParams{
		discountType:       tmpl.GetDiscountType(),
		discountValue:      tmpl.GetDiscountValue(),
		discountPercentage: tmpl.GetDiscountPercentage(),
	})
	if !applicable || discount <= 0 {
		return resp, nil
	}

	finalTotal := preDiscountTotal - discount
	if finalTotal < 0 {
		finalTotal = 0
	}

	resp.Applicable = boolPtr(true)
	resp.PreDiscountTotal = int64Ptr(preDiscountTotal)
	resp.Discount = int64Ptr(discount)
	resp.PostDiscountTotal = int64Ptr(finalTotal)
	return resp, nil
}

// computeCartPreDiscountTotal 只读遍历当前 viewer 的购物车，累加各 SKU 在结算币下的
// 小计。与 OrderService.Create 的 cart-iter 同源，但不扣库存、不持锁。
//
// 用户隔离双重保护：ent UserPrivacy 自动注入 user_id=viewer 过滤，本层 additionally
// 显式带 tenant_id/user_id 条件（与 OrderService.Create 一致）。
func (s *UserCouponService) computeCartPreDiscountTotal(ctx context.Context) (int64, error) {
	uid, ok := viewerUserIDFromContext(ctx)
	if !ok {
		return 0, couponV1.ErrorInternalServerError("viewer missing")
	}
	tid := tenantIDFromContext(ctx)
	if tid == 0 {
		return 0, couponV1.ErrorInternalServerError("viewer tenant missing")
	}

	// 查购物车（双重保护：显式 + privacy 自动）。
	cartEnt, cErr := s.entClient.Client().Cart.Query().
		Where(
			cart.TenantIDEQ(tid),
			cart.UserIDEQ(uid),
		).
		Only(ctx)
	if cErr != nil || cartEnt == nil {
		return 0, nil
	}

	// 查购物车项（privacy 自动按 user_id 过滤）。
	items, qErr := s.entClient.Client().CartItem.Query().
		Where(cartitem.CartIDEQ(cartEnt.ID)).
		All(ctx)
	if qErr != nil {
		return 0, qErr
	}

	var total int64 = 0
	for _, ci := range items {
		skuId := ci.SkuID
		if skuId == nil || *skuId == 0 {
			continue
		}
		qty := ci.Quantity
		if qty == nil || *qty <= 0 {
			continue
		}

		// 价格行（sku_id + currency 唯一）。当前系统仅支持 CNY。
		priceEnt, pErr := s.entClient.Client().SkuPrice.Query().
			Where(
				skuprice.SkuIDEQ(*skuId),
				skuprice.CurrencyEQ("CNY"),
			).
			Only(ctx)
		if pErr != nil || priceEnt == nil || priceEnt.Amount == nil || *priceEnt.Amount == "" {
			continue
		}
		unitPrice, parseErr := strconv.ParseInt(*priceEnt.Amount, 10, 64)
		if parseErr != nil || unitPrice <= 0 {
			continue
		}
		total += unitPrice * int64(*qty)
	}

	return total, nil
}

// HandleCouponSweep asynq 周期性任务处理器（薄委托）。
// 由 server/asynq_server.go 注册，asynq worker 按 cron 周期触发。
// 注入 SystemViewer context 以通过 ent privacy 的 viewer 校验（跨租户/用户扫描），
// 与 HandleOrderTimeout 使用同一 SystemViewer 注入模式。
func (s *UserCouponService) HandleCouponSweep(taskType string, taskData *task.CouponExpireSweepTaskData) error {
	s.log.Infow(
		"msg", "HandleCouponSweep started",
		"task_type", taskType,
	)
	ctx := appViewer.NewSystemViewerContext(context.Background())
	return s.SweepExpiredCoupons(ctx)
}

// SweepExpiredCoupons 扫描全库 user_coupon，将关联模板已过期（valid_until < now）
// 且当前状态为 UNUSED 的券实例批量翻为 EXPIRED。
//
// 单事务内：查所有 coupon_template → 筛出 valid_until < now 的模板 ID 集合 →
// 对每个过期模板，batch update user_coupon Where(coupon_template_id, status=UNUSED) →
// SetStatus(EXPIRED)。与 ExpireOrderByTimeout 的批量更新模式一致。
//
// SystemViewer 上下文（由 HandleCouponSweep 注入）绕过 UserPrivacy/TenantPrivacy，
// 可跨用户/租户扫描。任一步失败整事务回滚。
func (s *UserCouponService) SweepExpiredCoupons(ctx context.Context) error {
	now := time.Now()

	tx, err := s.entClient.Client().Tx(ctx)
	if err != nil {
		s.log.Errorf("begin coupon sweep tx failed: %v", err)
		return fmt.Errorf("begin coupon sweep tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. 查所有模板（SystemViewer 放行，跨租户）。
	tmpls, qErr := tx.CouponTemplate.Query().All(ctx)
	if qErr != nil {
		s.log.Errorf("query coupon templates for sweep failed: %v", qErr)
		err = qErr
		return fmt.Errorf("query coupon templates for sweep failed: %w", qErr)
	}

	// 2. 筛出已过期模板（valid_until < now）。
	var expiredTmplIds []uint32
	for _, tmpl := range tmpls {
		if tmpl.ValidUntil != nil && tmpl.ValidUntil.Before(now) {
			expiredTmplIds = append(expiredTmplIds, tmpl.ID)
		}
	}

	if len(expiredTmplIds) == 0 {
		// 无过期模板，直接提交空事务。
		if cErr := tx.Commit(); cErr != nil {
			s.log.Errorf("commit coupon sweep tx (no-op) failed: %v", cErr)
			err = cErr
			return fmt.Errorf("commit coupon sweep tx failed: %w", cErr)
		}
		s.log.Infow("msg", "coupon sweep completed: no expired templates", "expired_templates", 0)
		return nil
	}

	// 3. 对每个过期模板，batch update 关联的 UNUSED 券为 EXPIRED。
	//    按 tmplId 逐个更新（避免 IN 大列表；过期模板数量通常很少）。
	expiredStatus := usercoupon.StatusUserCouponStatusExpired
	unusedStatus := usercoupon.StatusUserCouponStatusUnused
	totalSwept := 0
	for _, tmplId := range expiredTmplIds {
		affected, uErr := tx.UserCoupon.Update().
			Where(
				usercoupon.CouponTemplateIDEQ(tmplId),
				usercoupon.StatusEQ(unusedStatus),
			).
			SetStatus(expiredStatus).
			Save(ctx)
		if uErr != nil {
			s.log.Errorf("sweep user_coupons for template [%d] failed: %v", tmplId, uErr)
			err = uErr
			return fmt.Errorf("sweep user_coupons for template [%d] failed: %w", tmplId, uErr)
		}
		totalSwept += affected
	}

	// 4. 提交事务。
	if cErr := tx.Commit(); cErr != nil {
		s.log.Errorf("commit coupon sweep tx failed: %v", cErr)
		err = cErr
		return fmt.Errorf("commit coupon sweep tx failed: %w", cErr)
	}
	err = nil

	s.log.Infow(
		"msg", "coupon sweep completed",
		"expired_templates", len(expiredTmplIds),
		"swept_coupons", totalSwept,
	)
	return nil
}

func int64Ptr(v int64) *int64 { return &v }
