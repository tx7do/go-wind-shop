package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type OrderItem struct {
	ent.Schema
}

func (OrderItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_order_items",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("订单项表"),
	}
}

func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("order_id").
			Comment("关联的订单ID").
			Optional().
			Nillable(),

		field.Uint32("sku_id").
			Comment("关联的 SKU ID").
			Optional().
			Nillable(),

		field.String("sku_snapshot").
			Comment("下单时 SKU 快照（JSON）").
			Optional().
			Nillable(),

		field.Int32("quantity").
			Comment("数量").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("unit_price").
			Comment("单价（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("subtotal").
			Comment("小计金额（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (OrderItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
	}
}
