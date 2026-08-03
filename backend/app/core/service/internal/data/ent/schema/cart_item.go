package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type CartItem struct {
	ent.Schema
}

func (CartItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_cart_items",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("购物车项表"),
	}
}

func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("cart_id").
			Comment("关联的购物车ID").
			Optional().
			Nillable(),

		field.Uint32("sku_id").
			Comment("关联的 SKU ID").
			Optional().
			Nillable(),

		field.Int32("quantity").
			Comment("数量").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (CartItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id"),
	}
}
