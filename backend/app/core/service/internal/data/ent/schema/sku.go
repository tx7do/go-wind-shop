package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type Sku struct {
	ent.Schema
}

func (Sku) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_skus",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("SKU 表"),
	}
}

func (Sku) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("product_id").
			Comment("所属商品ID").
			Optional().
			Nillable(),

		field.String("sku_code").
			Comment("SKU 唯一编码").
			Optional().
			Nillable(),

		field.Int32("stock_qty").
			Comment("库存数量").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (Sku) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (Sku) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("sku_code").Unique(),
	}
}
