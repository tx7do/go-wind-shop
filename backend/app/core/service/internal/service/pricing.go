package service

import (
	"time"

	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/coupontemplate"
)

// discountParams 把优惠券模板中影响抵扣计算的字段抽出来，供 computeDiscount 使用。
// 这样 computeDiscount 既能接受 proto DTO（Quote 试算路径），也能接受 ent 实体
// （Create 核销路径），避免重复实现两套抵扣算法——单一真相源。
type discountParams struct {
	discountType       couponV1.CouponTemplate_DiscountType
	discountValue      int64
	discountPercentage int32
}

// discountParamsFromEntity 从 ent CouponTemplate 实体提取抵扣参数。
// 仅用于 Create 事务内 ForUpdate 锁定的模板行——把 ent 枚举映射回 proto 枚举，
// 供 computeDiscount 使用。映射是 1:1 的字符串对齐（FIXED_AMOUNT/PERCENTAGE）。
func discountParamsFromEntity(tmpl *ent.CouponTemplate) discountParams {
	if tmpl == nil || tmpl.DiscountType == nil {
		return discountParams{discountType: couponV1.CouponTemplate_DISCOUNT_TYPE_UNSPECIFIED}
	}
	switch *tmpl.DiscountType {
	case coupontemplate.DiscountTypeCouponTemplateDiscountTypeFixedAmount:
		return discountParams{
			discountType:  couponV1.CouponTemplate_FIXED_AMOUNT,
			discountValue: derefInt64(tmpl.DiscountValue),
		}
	case coupontemplate.DiscountTypeCouponTemplateDiscountTypePercentage:
		return discountParams{
			discountType:       couponV1.CouponTemplate_PERCENTAGE,
			discountPercentage: derefInt32(tmpl.DiscountPercentage),
		}
	default:
		return discountParams{discountType: couponV1.CouponTemplate_DISCOUNT_TYPE_UNSPECIFIED}
	}
}

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// computeDiscount 是优惠券抵扣的单一真相源，下单核销（Create 事务内）与试算（Quote）
// 共用本函数，确保预览金额与最终入账金额算法一致。
//
// 入参：
//   - preDiscountTotal: 折前总额（各订单项 subtotal 之和，最小货币单位，分，≥0）
//   - params: 优惠券模板的抵扣参数
//
// 返回：
//   - discount: 抵扣额（最小货币单位，分，≥0）
//   - applicable: 该券在当前参数下是否产生有效抵扣
//
// 抵扣规则（与 coupon_error.proto 注释对齐，整数 floor，禁止 round）：
//   - FIXED_AMOUNT: discount = min(params.discount_value, preDiscountTotal)
//     （面额超过总额时仅抵至 0，不产生负数）
//   - PERCENTAGE: discount = (preDiscountTotal * params.discount_percentage) / 100
//     （Go 对非负操作数的整数除法即向下取整；例如 199 分 × 15% = 29.85 → 29 分）
//
// 调用方在拿到 discount 后须再施加 finalTotal = max(0, preDiscount_total - discount)。
// 本函数本身不修改任何状态、不触达 DB。
//
// 注意：保持 floor 语义至关重要——若后续改为 round/half-up，会导致预览与入账偏差、
// 对账金额漂移。修改此处须同步更新 coupon_error.proto 注释。
func computeDiscount(preDiscountTotal int64, params discountParams) (discount int64, applicable bool) {
	if preDiscountTotal <= 0 {
		return 0, false
	}

	switch params.discountType {
	case couponV1.CouponTemplate_FIXED_AMOUNT:
		face := params.discountValue
		if face <= 0 {
			return 0, false
		}
		if face >= preDiscountTotal {
			// 面额不低于总额：抵至 0，不产生负数。
			return preDiscountTotal, true
		}
		return face, true

	case couponV1.CouponTemplate_PERCENTAGE:
		pct := params.discountPercentage
		if pct <= 0 {
			return 0, false
		}
		if pct >= 100 {
			// 百分比 ≥100：全额抵扣。
			return preDiscountTotal, true
		}
		// 整数 floor：Go 对非负 int64 的 / 即向下取整。
		return (preDiscountTotal * int64(pct)) / 100, true

	default:
		return 0, false
	}
}

// couponApplicableNowEntity 校验 ent 模板实体在给定时刻是否处于有效窗口内。
// valid_from ≤ now ≤ valid_until 才视为有效；任一端缺失（nil）则该侧不约束。
// 用于 Create 事务内 ForUpdate 锁定的模板行。
func couponApplicableNowEntity(tmpl *ent.CouponTemplate, now time.Time) bool {
	if tmpl == nil {
		return false
	}
	if tmpl.ValidFrom != nil && now.Before(*tmpl.ValidFrom) {
		return false
	}
	if tmpl.ValidUntil != nil && now.After(*tmpl.ValidUntil) {
		return false
	}
	return true
}

// couponApplicableNow 校验 proto DTO 模板在给定时刻是否处于有效窗口内。
// 用于 Quote 试算路径（只读快照）。valid_from ≤ now ≤ valid_until 才视为有效；
// 任一端缺失（nil）则该侧不约束。
func couponApplicableNow(tmpl *couponV1.CouponTemplate, now time.Time) bool {
	if tmpl == nil {
		return false
	}
	from := tmpl.GetValidFrom()
	until := tmpl.GetValidUntil()
	if from != nil && now.Before(from.AsTime()) {
		return false
	}
	if until != nil && now.After(until.AsTime()) {
		return false
	}
	return true
}
