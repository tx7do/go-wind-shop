package service

import (
	"context"
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
	"go-wind-shop/app/core/service/internal/data/ent/skuprice"

	entCrud "github.com/tx7do/go-crud/entgo"

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

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

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

func int64Ptr(v int64) *int64 { return &v }
