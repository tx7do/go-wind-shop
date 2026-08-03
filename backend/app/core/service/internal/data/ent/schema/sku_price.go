package schema

import (
	appmixin "go-wind-shop/pkg/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type SkuPrice struct {
	ent.Schema
}

func (SkuPrice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_sku_prices",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("SKU 多币种价格表"),
	}
}

func (SkuPrice) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("sku_id").
			Comment("关联的 SKU ID").
			Optional().
			Nillable(),

		field.String("amount").
			Comment("该币种下的价格金额（字符串表示，避免浮点精度问题）").
			Optional().
			Nillable(),
	}
}

func (SkuPrice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		appmixin.Currency{},
	}
}

func (SkuPrice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sku_id", "currency").Unique(),
		index.Fields("sku_id"),
	}
}
