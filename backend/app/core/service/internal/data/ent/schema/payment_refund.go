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

type PaymentRefund struct {
	ent.Schema
}

func (PaymentRefund) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_payment_refunds",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("支付退款表"),
	}
}

func (PaymentRefund) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("transaction_id").
			Comment("关联的支付流水ID").
			Optional().
			Nillable(),

		field.Int64("amount").
			Comment("退款金额（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("退款状态").
			NamedValues(
				"PaymentRefundStatusPending", "PENDING",
				"PaymentRefundStatusSucceeded", "SUCCEEDED",
				"PaymentRefundStatusFailed", "FAILED",
			).
			Optional().
			Nillable(),
	}
}

func (PaymentRefund) Mixin() []ent.Mixin {
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

func (PaymentRefund) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("transaction_id"),
		// 幂等键唯一索引：与 order/payment_transaction 一致，同租户内
		// idempotency_key 唯一，防止重放导致重复退款记录。
		index.Fields("tenant_id", "idempotency_key").Unique(),
	}
}
