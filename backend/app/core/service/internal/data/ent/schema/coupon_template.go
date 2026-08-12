package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"

	appmixin "go-wind-shop/pkg/entgo/mixin"
)

type CouponTemplate struct {
	ent.Schema
}

// CouponTemplate 不注入 UserPrivacy：模板是租户级运营配置，非用户私有数据。
// 租户隔离由 TenantID mixin 自动注入的 TenantPrivacy 提供（platform/system 视图放行）。
// 仅 admin proto 暴露 CRUD，app proto 不含模板服务。

func (CouponTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_coupon_templates",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("优惠券模板表"),
	}
}

func (CouponTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("discount_type").
			Comment("抵扣类型").
			NamedValues(
				"CouponTemplateDiscountTypeFixedAmount", "FIXED_AMOUNT",
				"CouponTemplateDiscountTypePercentage", "PERCENTAGE",
			).
			Optional().
			Nillable(),

		field.Int64("discount_value").
			Comment("固定金额抵扣值（最小货币单位，分；discount_type=FIXED_AMOUNT 时生效）").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("discount_percentage").
			Comment("百分比抵扣值（0-100；discount_type=PERCENTAGE 时生效）").
			Default(0).
			Optional().
			Nillable(),

		field.Time("valid_from").
			Comment("有效期起始时间").
			Optional().
			Nillable(),

		field.Time("valid_until").
			Comment("有效期截止时间").
			Optional().
			Nillable(),

		field.Int32("max_redemptions").
			Comment("全局核销上限（0=不限量；>0 为该模板累计可核销次数）").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("redeemed_count").
			Comment("已核销次数（核销自增、返还自减，行锁保护）").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("模板状态").
			NamedValues(
				"CouponTemplateStatusActive", "ACTIVE",
				"CouponTemplateStatusInactive", "INACTIVE",
			).
			Optional().
			Nillable(),
	}
}

func (CouponTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
	}
}

func (CouponTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
	}
}
