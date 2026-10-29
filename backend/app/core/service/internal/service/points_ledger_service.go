package service

import (
	"context"
	"time"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"


	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/coupontemplate"
	"go-wind-shop/app/core/service/internal/data/ent/pointsledger"
	"go-wind-shop/app/core/service/internal/data/ent/usercoupon"


	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// PointsAwardRatePerYuan 消费入账比率：1 积分 / 1 元（totalAmount 分 → /100）。
const PointsAwardRatePerYuan = 1

type PointsLedgerService struct {
	catalogV1.UnimplementedPointsLedgerServiceServer

	log *logx.Helper

	repo                *data.PointsLedgerRepo
	entClient           *entCrud.EntClient[*ent.Client]
	couponTemplateRepo  *data.CouponTemplateRepo
	memberTierService   *MemberTierService
}

func NewPointsLedgerService(ctx *bootstrap.Context, repo *data.PointsLedgerRepo, entClient *entCrud.EntClient[*ent.Client], couponTemplateRepo *data.CouponTemplateRepo, memberTierService *MemberTierService) *PointsLedgerService {
	return &PointsLedgerService{
		log:                logx.New(ctx.NewLoggerHelper("points-ledger/service/core-service")),
		repo:               repo,
		entClient:          entClient,
		couponTemplateRepo: couponTemplateRepo,
		memberTierService:  memberTierService,
	}
}

// RegisterMemberTierService 延迟注入（构造环解法）。
func (s *PointsLedgerService) RegisterMemberTierService(svc *MemberTierService) {
	s.memberTierService = svc
}

func (s *PointsLedgerService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPointsLedgerResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *PointsLedgerService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountPointsLedgerResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}
	return &catalogV1.CountPointsLedgerResponse{Count: uint64(count)}, nil
}

func (s *PointsLedgerService) Get(ctx context.Context, req *catalogV1.GetPointsLedgerRequest) (*catalogV1.PointsLedger, error) {
	return s.repo.Get(ctx, req)
}

// GetUserPointsBalance 用户积分余额（delta 求和）。
func (s *PointsLedgerService) GetUserPointsBalance(ctx context.Context, req *catalogV1.GetUserPointsBalanceRequest) (*catalogV1.GetUserPointsBalanceResponse, error) {
	balance, err := s.userBalance(ctx, req.GetUserId())
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("query balance failed")
	}
	return &catalogV1.GetUserPointsBalanceResponse{Balance: balance}, nil
}

// checkInAwardPerDay 每日签到固定奖励积分数。
const checkInAwardPerDay = int64(5)

// CheckIn 每日签到（幂等：同日同用户仅一次）。奖励固定 +5 积分。
func (s *PointsLedgerService) CheckIn(ctx context.Context, req *catalogV1.CheckInRequest) (*catalogV1.CheckInResponse, error) {
	userId := req.GetUserId()
	if userId == 0 {
		return nil, catalogV1.ErrorBadRequest("user_id is required")
	}

	// 幂等：今日已签到 → 幂等返回
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	already, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.UserIDEQ(userId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeCheckIn),
			pointsledger.CreatedAtGTE(todayStart),
			pointsledger.CreatedAtLT(todayEnd),
		).
		Exist(ctx)
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("check-in query failed")
	}

	balance, bErr := s.userBalance(ctx, userId)
	if bErr != nil {
		return nil, catalogV1.ErrorInternalServerError("balance query failed")
	}

	if already {
		return &catalogV1.CheckInResponse{
			Awarded:          0,
			Balance:          balance,
			AlreadyCheckedIn: true,
		}, nil
	}

	award := checkInAwardPerDay
	balance += award

	cErr := s.entClient.Client().PointsLedger.Create().
		SetUserID(userId).
		SetDelta(award).
		SetBalanceAfter(balance).
		SetType(pointsledger.TypePointsTypeCheckIn).
		SetReason("每日签到").
		SetCreatedAt(now).
		Exec(ctx)
	if cErr != nil {
		return nil, catalogV1.ErrorInternalServerError("check-in write failed")
	}

	return &catalogV1.CheckInResponse{
		Awarded:          award,
		Balance:          balance,
		AlreadyCheckedIn: false,
	}, nil
}

// GetCheckInStatus 签到状态（今日是否已签 + 连续签到天数）。
func (s *PointsLedgerService) GetCheckInStatus(ctx context.Context, req *catalogV1.GetCheckInStatusRequest) (*catalogV1.GetCheckInStatusResponse, error) {
	userId := req.GetUserId()
	if userId == 0 {
		return nil, catalogV1.ErrorBadRequest("user_id is required")
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 今日签到记录
	checkedToday, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.UserIDEQ(userId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeCheckIn),
			pointsledger.CreatedAtGTE(todayStart),
		).
		Exist(ctx)
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("check-in status query failed")
	}

	// 连续签到天数：从今天（或昨天）往前数有签到记录的连续天数
	consecutive := int32(0)
	day := todayStart
	if !checkedToday {
		day = day.AddDate(0, 0, -1)
	}
	for i := 0; i < 365; i++ {
		dayEnd := day.Add(24 * time.Hour)
		exist, eErr := s.entClient.Client().PointsLedger.Query().
			Where(
				pointsledger.UserIDEQ(userId),
				pointsledger.TypeEQ(pointsledger.TypePointsTypeCheckIn),
				pointsledger.CreatedAtGTE(day),
				pointsledger.CreatedAtLT(dayEnd),
			).
			Exist(ctx)
		if eErr != nil || !exist {
			break
		}
		consecutive++
		day = day.AddDate(0, 0, -1)
	}

	return &catalogV1.GetCheckInStatusResponse{
		CheckedInToday:  checkedToday,
		ConsecutiveDays: consecutive,
	}, nil
}

func (s *PointsLedgerService) userBalance(ctx context.Context, userId uint32) (int64, error) {
	var rows []struct{ Sum int64 }
	err := s.entClient.Client().PointsLedger.Query().
		Where(pointsledger.UserIDEQ(userId)).
		Aggregate(ent.As(ent.Sum(pointsledger.FieldDelta), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Sum, nil
}

// AwardForOrder 订单支付成功入账（幂等：同订单已有 EARN 台账则跳过）。
// best-effort 由调用方保证（积分失败不影响支付主流程，仅记日志）。
func (s *PointsLedgerService) AwardForOrder(ctx context.Context, userId uint32, orderId uint32, totalAmount int64) error {
	// 幂等：同订单已入账跳过
	exist, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeEarn),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	award := totalAmount / 100 * PointsAwardRatePerYuan
	// 会员等级积分倍率
	if s.memberTierService != nil {
		if m := s.memberTierService.GetPointsMultiplierForUser(ctx, userId); m > 1 {
			award *= int64(m)
		}
	}
	if award <= 0 {
		return nil
	}

	balance, err := s.userBalance(ctx, userId)
	if err != nil {
		return err
	}
	balance += award

	return s.entClient.Client().PointsLedger.Create().
		SetUserID(userId).
		SetDelta(award).
		SetBalanceAfter(balance).
		SetType(pointsledger.TypePointsTypeEarn).
		SetReason("订单消费入账").
		SetRefOrderID(orderId).
		SetCreatedAt(time.Now()).
		Exec(ctx)
}

// DeductForOrder 退款成功回冲（对订单已入账积分对冲，余额拖底为 0 不产生负积分）。
// 幂等：同订单已有 DEDUCT 台账则跳过。
func (s *PointsLedgerService) DeductForOrder(ctx context.Context, orderId uint32) error {
	// 幂等
	exist, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeDeduct),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	// 订单此前入账了多少
	var rows []struct{ Sum int64 }
	err = s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeEarn),
		).
		Aggregate(ent.As(ent.Sum(pointsledger.FieldDelta), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return err
	}
	var awarded int64
	if len(rows) > 0 {
		awarded = rows[0].Sum
	}
	if awarded <= 0 {
		return nil
	}

	// 找到订单归属用户（从 EARN 台账取）
	led, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeEarn),
		).
		First(ctx)
	if err != nil || led == nil || led.UserID == nil {
		return nil
	}
	userId := *led.UserID

	balance, err := s.userBalance(ctx, userId)
	if err != nil {
		return err
	}
	deduct := awarded
	if balance-deduct < 0 {
		deduct = balance // 拖底为 0
	}
	if deduct <= 0 {
		return nil
	}
	balance -= deduct

	return s.entClient.Client().PointsLedger.Create().
		SetUserID(userId).
		SetDelta(-deduct).
		SetBalanceAfter(balance).
		SetType(pointsledger.TypePointsTypeDeduct).
		SetReason("退款回冲").
		SetRefOrderID(orderId).
		SetCreatedAt(time.Now()).
		Exec(ctx)
}

// ExchangePointsForCoupon 积分兑换优惠券（事务内：扣积分 + 发券）。
// 校验：模板可兑换（points_price > 0）、余额充足、模板未领完。
func (s *PointsLedgerService) ExchangePointsForCoupon(ctx context.Context, req *catalogV1.ExchangePointsForCouponRequest) (*catalogV1.ExchangePointsForCouponResponse, error) {
	userId := req.GetUserId()
	templateId := req.GetCouponTemplateId()
	if userId == 0 || templateId == 0 {
		return nil, catalogV1.ErrorBadRequest("user_id and coupon_template_id are required")
	}

	// 事务
	tx, txErr := s.entClient.Client().Tx(ctx)
	if txErr != nil {
		return nil, catalogV1.ErrorInternalServerError("begin tx failed")
	}

	// 查模板（行锁防并发兑换超发）
	tmpl, tErr := tx.CouponTemplate.Query().
		Where(coupontemplate.IDEQ(templateId)).
		ForUpdate().
		Only(ctx)
	if tErr != nil || tmpl == nil {
		tx.Rollback()
		return nil, catalogV1.ErrorBadRequest("coupon template not found")
	}
	if tmpl.PointsPrice <= 0 {
		tx.Rollback()
		return nil, catalogV1.ErrorBadRequest("coupon is not exchangeable with points")
	}
	pointsPrice := tmpl.PointsPrice

	// 领取量校验
	if tmpl.MaxRedemptions != nil && tmpl.RedeemedCount != nil && *tmpl.RedeemedCount >= *tmpl.MaxRedemptions {
		tx.Rollback()
		return nil, catalogV1.ErrorBadRequest("coupon template fully redeemed")
	}

	// 余额校验（事务内查）
	var balRows []struct{ Sum int64 }
	bErr := tx.PointsLedger.Query().
		Where(pointsledger.UserIDEQ(userId)).
		Aggregate(ent.As(ent.Sum(pointsledger.FieldDelta), "sum")).
		Scan(ctx, &balRows)
	if bErr != nil {
		tx.Rollback()
		return nil, catalogV1.ErrorInternalServerError("balance query failed")
	}
	balance := int64(0)
	if len(balRows) > 0 {
		balance = balRows[0].Sum
	}
	if balance < pointsPrice {
		tx.Rollback()
		return nil, catalogV1.ErrorBadRequest("insufficient points (need %d, have %d)", pointsPrice, balance)
	}

	// 扣积分
	newBalance := balance - pointsPrice
	if cErr := tx.PointsLedger.Create().
		SetUserID(userId).
		SetDelta(-pointsPrice).
		SetBalanceAfter(newBalance).
		SetType(pointsledger.TypePointsTypeDeduct).
		SetReason("积分兑换优惠券").
		SetCreatedAt(time.Now()).
		Exec(ctx); cErr != nil {
		tx.Rollback()
		return nil, catalogV1.ErrorInternalServerError("deduct points failed")
	}

	// 发券
	ucStatus := usercoupon.StatusUserCouponStatusUnused
	uc, uErr := tx.UserCoupon.Create().
		SetUserID(userId).
		SetCouponTemplateID(templateId).
		SetStatus(ucStatus).
		SetCreatedAt(time.Now()).
		Save(ctx)
	if uErr != nil {
		tx.Rollback()
		return nil, catalogV1.ErrorInternalServerError("issue coupon failed")
	}

	// 模板领取计数++
	newCount := int32(1)
	if tmpl.RedeemedCount != nil {
		newCount = *tmpl.RedeemedCount + 1
	}
	if _, rErr := tx.CouponTemplate.UpdateOneID(templateId).
		SetRedeemedCount(newCount).
		Save(ctx); rErr != nil {
		tx.Rollback()
		return nil, catalogV1.ErrorInternalServerError("update redeemed count failed")
	}

	if cErr := tx.Commit(); cErr != nil {
		return nil, catalogV1.ErrorInternalServerError("commit exchange failed")
	}

	s.log.Infof("points exchanged for coupon: user=%d points=%d template=%d", userId, pointsPrice, templateId)

	return &catalogV1.ExchangePointsForCouponResponse{
		CouponId:       uc.ID,
		PointsDeducted: pointsPrice,
		Balance:        newBalance,
	}, nil
}
