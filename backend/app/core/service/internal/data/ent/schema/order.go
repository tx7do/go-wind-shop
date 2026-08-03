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

type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_orders",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("订单表"),
	}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("下单用户ID").
			Optional().
			Nillable(),

		field.Int64("total_amount").
			Comment("订单总金额（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("订单状态").
			NamedValues(
				"OrderStatusPendingPayment", "PENDING_PAYMENT",
				"OrderStatusPaid", "PAID",
				"OrderStatusCancelled", "CANCELLED",
				"OrderStatusFulfilled", "FULFILLED",
				"OrderStatusClosed", "CLOSED",
			).
			Optional().
			Nillable(),

		field.String("recipient_name").
			Comment("收件人姓名").
			Optional().
			Nillable(),

		field.String("recipient_phone").
			Comment("收件人电话").
			Optional().
			Nillable(),

		field.String("shipping_address").
			Comment("收货地址（结构化文本）").
			Optional().
			Nillable(),
	}
}

func (Order) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
		appmixin.BusinessRefId{},
		appmixin.IdempotencyKey{},
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "idempotency_key").Unique(),
		index.Fields("user_id"),
		index.Fields("status"),
	}
}
