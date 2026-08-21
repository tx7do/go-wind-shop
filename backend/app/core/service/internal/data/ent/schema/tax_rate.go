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

// TaxRate 税率规则表。
// 按目的地区域（国家代码）+ 币种配置税率百分比，租户级运营配置。
// 下单时 OrderService.Create 按 (viewer tenant_id, shipping_region, currency) 查询
// 当前适用规则，计算 tax_amount = (originalAmount - discount) * tax_rate / 100（整数 floor）。
// 无规则时 tax_amount = 0，不阻塞下单。
type TaxRate struct {
	ent.Schema
}

func (TaxRate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_tax_rates",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("税率规则表（按地区+币种配置）"),
	}
}

func (TaxRate) Fields() []ent.Field {
	return []ent.Field{
		field.String("region").
			Comment("目的地区域（ISO 3166 国家代码）").
			Optional().
			Nillable(),

		field.Int32("tax_rate").
			Comment("税率百分比（0-100，如 20 表示 20% VAT）").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("规则状态").
			NamedValues(
				"TaxRateStatusActive", "ACTIVE",
				"TaxRateStatusInactive", "INACTIVE",
			).
			Optional().
			Nillable(),
	}
}

func (TaxRate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
	}
}

func (TaxRate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "region", "currency").Unique(),
	}
}
