package service

import (
	"context"
	"time"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/promotion"
)

// 本文件是促销计价在下单事务内的执行体。
//
// 与 PromotionService.QuoteOrderPromotions 试算保持同一语义：
//   - SPECIAL_PRICE：按 (tenant, product, 窗口内, priority 最大) 覆盖 SKU 单价，
//     仅当专价低于现价时生效（不抬价）；订单项快照 unit_price 记录生效后价格；
//   - FULL_REDUCTION：店铺级满 X 减 Y，窗口内命中者取减让最大的一条（不叠加），
//     门槛按折前总额判定，最终与"券后余额"取 min 保证应付额 ≥ 0。
//
// order.discount_amount 语义自此扩展为「优惠抵扣合计（优惠券 + 促销）」审计值。

// activeSpecialPriceInTx 下单事务内的专价查询（复用与试算一致的选取规则）。
func activeSpecialPriceInTx(
	ctx context.Context,
	tx *ent.Tx,
	tenantId, productId uint32,
	currency string,
	now time.Time,
) (int64, bool) {
	if productId == 0 {
		return 0, false
	}
	ents, err := tx.Promotion.Query().
		Where(
			promotion.StatusEQ(promotion.StatusPromotionStatusActive),
			promotion.PromoTypeEQ(promotion.PromoTypePromotionTypeSpecialPrice),
			promotion.ProductIDEQ(productId),
			promotion.ValidFromLTE(now),
			promotion.ValidUntilGTE(now),
		).
		All(ctx)
	if err != nil || len(ents) == 0 {
		return 0, false
	}

	var best *ent.Promotion
	for i := range ents {
		p := ents[i]
		if p.SpecialPrice == nil || *p.SpecialPrice <= 0 {
			continue
		}
		if p.Currency != nil && *p.Currency != currency {
			continue
		}
		if tenantId != 0 && p.TenantID != nil && *p.TenantID != tenantId {
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

// computeFullReductionInTx 下单事务内的满减让利计算。
// 返回让利额（0 = 未命中）。
func computeFullReductionInTx(
	ctx context.Context,
	tx *ent.Tx,
	tenantId uint32,
	currency string,
	originalAmount int64,
) int64 {
	if originalAmount <= 0 {
		return 0
	}
	q := tx.Promotion.Query().
		Where(
			promotion.StatusEQ(promotion.StatusPromotionStatusActive),
			promotion.PromoTypeEQ(promotion.PromoTypePromotionTypeFullReduction),
			promotion.ValidFromLTE(time.Now()),
			promotion.ValidUntilGTE(time.Now()),
		)
	if tenantId != 0 {
		q = q.Where(promotion.TenantIDEQ(tenantId))
	}
	ents, err := q.All(ctx)
	if err != nil {
		return 0
	}

	var best int64
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
		}
	}
	if best > originalAmount {
		best = originalAmount
	}
	return best
}
