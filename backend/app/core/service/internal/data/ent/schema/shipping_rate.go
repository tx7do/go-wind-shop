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

// ShippingRate 运费模板表。
// 按目的地区域（国家代码）+ 币种配置运费规则，租户级运营配置。
// 下单时 OrderService.Create 按 (viewer tenant_id, shipping_region, currency) 查询
// 当前适用规则，计算 shipping_fee = base_fee + per_unit_fee * item_count。
// 无规则时 shipping_fee = 0，不阻塞下单。
type ShippingRate struct {
	ent.Schema
}

func (ShippingRate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_shipping_rates",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("运费模板表（按地区+币种配置）"),
	}
}

func (ShippingRate) Fields() []ent.Field {
	return []ent.Field{
		field.String("region").
			Comment("目的地区域（ISO 3166 国家代码）").
			Optional().
			Nillable(),

		field.Int64("base_fee").
			Comment("基础运费（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("per_unit_fee").
			Comment("按订单项数加收的运费（最小货币单位，分/件）").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("模板状态").
			NamedValues(
				"ShippingRateStatusActive", "ACTIVE",
				"ShippingRateStatusInactive", "INACTIVE",
			).
			Optional().
			Nillable(),
	}
}

func (ShippingRate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
	}
}

func (ShippingRate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "region", "currency").Unique(),
	}
}
