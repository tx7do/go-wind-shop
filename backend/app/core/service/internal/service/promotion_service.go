package service

import (
	"go-wind-shop/pkg/logx"
	"context"
	"strconv"
	"time"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-crud/viewer"
	"google.golang.org/protobuf/types/known/emptypb"
	"github.com/tx7do/go-utils/trans"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/promotion"
	"go-wind-shop/app/core/service/internal/data/ent/sku"
	"go-wind-shop/app/core/service/internal/data/ent/skuprice"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type PromotionService struct {
	catalogV1.UnimplementedPromotionServiceServer

	log *logx.Helper

	repo *data.PromotionRepo

	entClient *entCrud.EntClient[*ent.Client]
}

func NewPromotionService(ctx *bootstrap.Context, repo *data.PromotionRepo, entClient *entCrud.EntClient[*ent.Client]) *PromotionService {
	return &PromotionService{
		log:       logx.New(ctx.NewLoggerHelper("promotion/service/core-service")),
		repo:      repo,
		entClient: entClient,
	}
}

func (s *PromotionService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPromotionResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *PromotionService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountPromotionResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &catalogV1.CountPromotionResponse{
		Count: uint64(count),
	}, nil
}

func (s *PromotionService) Get(ctx context.Context, req *catalogV1.GetPromotionRequest) (*catalogV1.Promotion, error) {
	return s.repo.Get(ctx, req)
}

func (s *PromotionService) Create(ctx context.Context, req *catalogV1.CreatePromotionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.validatePromotion(req.Data); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PromotionService) Update(ctx context.Context, req *catalogV1.UpdatePromotionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.validatePromotion(req.Data); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PromotionService) Delete(ctx context.Context, req *catalogV1.DeletePromotionRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// validatePromotion 活动参数校验：专价活动必须指定商品与正数专价；
// 满减活动必须门槛/让利合法且让利 ≤ 门槛；窗口必须起止有序。
func (s *PromotionService) validatePromotion(data *catalogV1.Promotion) error {
	pt := data.GetPromoType()
	if pt == catalogV1.Promotion_SPECIAL_PRICE {
		if data.GetProductId() == 0 {
			return catalogV1.ErrorBadRequest("special-price promotion requires product_id")
		}
		if data.GetSpecialPrice() <= 0 {
			return catalogV1.ErrorBadRequest("special_price must be positive")
		}
	}
	if pt == catalogV1.Promotion_FULL_REDUCTION {
		if data.GetThresholdAmount() <= 0 || data.GetReductionAmount() <= 0 {
			return catalogV1.ErrorBadRequest("full-reduction requires positive threshold and reduction")
		}
		if data.GetReductionAmount() > data.GetThresholdAmount() {
			return catalogV1.ErrorBadRequest("reduction_amount must not exceed threshold_amount")
		}
	}
	if data.ValidFrom != nil && data.ValidUntil != nil {
		from := data.ValidFrom.AsTime()
		until := data.ValidUntil.AsTime()
		if !until.After(from) {
			return catalogV1.ErrorBadRequest("valid_until must be after valid_from")
		}
	}
	return nil
}

// ListActivePromotions 当前生效中的活动（C端商品页角标/活动列表）。
// 租户隔离由 TenantPrivacy 按 viewer 自动注入。
func (s *PromotionService) ListActivePromotions(ctx context.Context, req *catalogV1.ListActivePromotionsRequest) (*catalogV1.ListPromotionResponse, error) {
	resp, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	// 可选商品过滤（服务端兜底；C端一般自行过滤）。
	if req != nil && len(req.GetProductIds()) > 0 {
		want := make(map[uint32]bool, len(req.GetProductIds()))
		for _, pid := range req.GetProductIds() {
			want[pid] = true
		}
		filtered := make([]*catalogV1.Promotion, 0, len(resp.Items))
		for _, item := range resp.Items {
			if item.GetProductId() == 0 || want[item.GetProductId()] {
				filtered = append(filtered, item)
			}
		}
		resp.Items = filtered
		resp.Total = uint64(len(filtered))
	}
	return resp, nil
}

// QuoteOrderPromotions 促销试算：按服务端价格与专价活动重算折前总额并计算满减。
// 不持锁、不落库；最终金额以下单事务内校验为准。
// 租户取 viewer（app BFF 转发 operator 元数据 → TenantPrivacy 自动过滤；
// 这里显式读取 tenantId 用于 sku/promotion 查询，与隐私层双重保险）。
func (s *PromotionService) QuoteOrderPromotions(ctx context.Context, req *catalogV1.QuoteOrderPromotionsRequest) (*catalogV1.QuoteOrderPromotionsResponse, error) {
	if req == nil || len(req.GetItems()) == 0 {
		return nil, catalogV1.ErrorBadRequest("items is required")
	}

	tenantId := uint32(0)
	if vc, exist := viewer.FromContext(ctx); exist && vc != nil {
		tenantId = uint32(vc.TenantID())
	}
	currency := req.GetCurrency()
	if currency == "" {
		currency = "CNY"
	}
	now := time.Now()

	db := s.entClient.Client()

	var originalAmount int64
	for _, item := range req.GetItems() {
		skuId := item.GetSkuId()
		qty := item.GetQuantity()
		if skuId == 0 || qty <= 0 {
			continue
		}

		skuEnt, err := db.Sku.Query().
			Where(sku.IDEQ(skuId)).
			Only(ctx)
		if err != nil || skuEnt == nil {
			continue
		}

		priceEnt, err := db.SkuPrice.Query().
			Where(
				skuprice.SkuIDEQ(skuId),
				skuprice.CurrencyEQ(currency),
			).
			Only(ctx)
		if err != nil || priceEnt == nil || priceEnt.Amount == nil {
			continue
		}
		unitPrice, perr := strconv.ParseInt(*priceEnt.Amount, 10, 64)
		if perr != nil || unitPrice <= 0 {
			continue
		}

		// 专价覆盖（与下单事务同语义：取 priority 最高者）。
		var productId uint32
		if skuEnt.ProductID != nil {
			productId = *skuEnt.ProductID
		}
		if sp, ok := s.activeSpecialPrice(ctx, db, tenantId, productId, currency, now); ok && sp < unitPrice {
			unitPrice = sp
		}

		originalAmount += unitPrice * int64(qty)
	}

	// 满减让利。
	promoDiscount, matched := s.computeFullReduction(ctx, db, tenantId, currency, originalAmount, now)

	return &catalogV1.QuoteOrderPromotionsResponse{
		OriginalAmount:    originalAmount,
		PromoDiscount:     promoDiscount,
		MatchedPromotions: matched,
	}, nil
}

// activeSpecialPrice 查询商品当前生效的专价（priority 最高者）。
// 与 order_service.Create 的下单事务同语义，抽到这里供试算复用。
func (s *PromotionService) activeSpecialPrice(ctx context.Context, client *ent.Client, tenantId, productId uint32, currency string, now time.Time) (int64, bool) {
	q := client.Promotion.Query().
		Where(
			promotion.StatusEQ(promotion.StatusPromotionStatusActive),
			promotion.PromoTypeEQ(promotion.PromoTypePromotionTypeSpecialPrice),
			promotion.ValidFromLTE(now),
			promotion.ValidUntilGTE(now),
		)
	if tenantId != 0 {
		q = q.Where(promotion.TenantIDEQ(tenantId))
	}
	if productId != 0 {
		q = q.Where(promotion.ProductIDEQ(productId))
	}
	ents, err := q.All(ctx)
	if err != nil || len(ents) == 0 {
		return 0, false
	}
	// priority 最大者优先（并列时取先建者，按 ID 升序稳定）。
	var best *ent.Promotion
	for i := range ents {
		p := ents[i]
		if p.SpecialPrice == nil || *p.SpecialPrice <= 0 {
			continue
		}
		if p.Currency != nil && *p.Currency != currency {
			continue
		}
		if best == nil || promoPriority(p) > promoPriority(best) ||
			(promoPriority(p) == promoPriority(best) && p.ID < best.ID) {
			best = p
		}
	}
	if best == nil {
		return 0, false
	}
	return *best.SpecialPrice, true
}

func promoPriority(p *ent.Promotion) int32 {
	if p.Priority == nil {
		return 0
	}
	return *p.Priority
}

// computeFullReduction 计算满减让利：窗口内命中的店铺级 FULL_REDUCTION 中取让利最大者（不叠加）。
// 返回让利额与命中的活动（供前端展示）。
func (s *PromotionService) computeFullReduction(ctx context.Context, client *ent.Client, tenantId uint32, currency string, originalAmount int64, now time.Time) (int64, []*catalogV1.Promotion) {
	if originalAmount <= 0 {
		return 0, nil
	}
	q := client.Promotion.Query().
		Where(
			promotion.StatusEQ(promotion.StatusPromotionStatusActive),
			promotion.PromoTypeEQ(promotion.PromoTypePromotionTypeFullReduction),
			promotion.ValidFromLTE(now),
			promotion.ValidUntilGTE(now),
		)
	if tenantId != 0 {
		q = q.Where(promotion.TenantIDEQ(tenantId))
	}
	ents, err := q.All(ctx)
	if err != nil {
		return 0, nil
	}

	var best int64
	var bestEnt *ent.Promotion
	for _, p := range ents {
		if p.Currency != nil && *p.Currency != currency {
			continue
		}
		th := int64(0)
		if p.ThresholdAmount != nil {
			th = *p.ThresholdAmount
		}
		red := int64(0)
		if p.ReductionAmount != nil {
			red = *p.ReductionAmount
		}
		if originalAmount >= th && red > best {
			best = red
			bestEnt = p
		}
	}
	if best <= 0 || bestEnt == nil {
		return 0, nil
	}
	if best > originalAmount {
		best = originalAmount
	}
	dto := &catalogV1.Promotion{
		Id:              trans.Ptr(bestEnt.ID),
		Name:            bestEnt.Name,
		ThresholdAmount: bestEnt.ThresholdAmount,
		ReductionAmount: bestEnt.ReductionAmount,
	}
	return best, []*catalogV1.Promotion{dto}
}
