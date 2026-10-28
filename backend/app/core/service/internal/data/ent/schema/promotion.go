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

type Promotion struct {
	ent.Schema
}

// Promotion 不注入 UserPrivacy：促销活动是租户级运营配置，非用户私有数据。
// 租户隔离由 TenantID mixin 自动注入的 TenantPrivacy 提供（platform/system 视图放行）。

func (Promotion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_promotions",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("促销活动表"),
	}
}

func (Promotion) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("活动名称").
			Optional().
			Nillable(),

		field.Enum("promo_type").
			Comment("促销类型").
			NamedValues(
				"PromotionTypeSpecialPrice", "SPECIAL_PRICE",
				"PromotionTypeFullReduction", "FULL_REDUCTION",
			).
			Optional().
			Nillable(),

		field.Uint32("product_id").
			Comment("生效商品ID（仅 SPECIAL_PRICE；0 表示店铺级活动）").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("special_price").
			Comment("专价（最小货币单位，分；promo_type=SPECIAL_PRICE 时生效）").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("threshold_amount").
			Comment("满减门槛（最小货币单位，分；promo_type=FULL_REDUCTION 时生效）").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("reduction_amount").
			Comment("满减让利（最小货币单位，分；promo_type=FULL_REDUCTION 时生效）").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("priority").
			Comment("优先级（同商品同类型多活动并发时取最大者）").
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

		field.Enum("status").
			Comment("活动状态").
			NamedValues(
				"PromotionStatusActive", "ACTIVE",
				"PromotionStatusInactive", "INACTIVE",
			).
			Optional().
			Nillable(),
	}
}

func (Promotion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
	}
}

func (Promotion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("tenant_id", "status"),
		index.Fields("product_id"),
	}
}
