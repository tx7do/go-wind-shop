package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"

	appmixin "go-wind-shop/pkg/entgo/mixin"
	appPrivacy "go-wind-shop/pkg/entgo/privacy"
)

type PaymentTransaction struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：普通用户只能查询/变更 user_id = 自身 userID
// 的支付流水。系统/平台视图放行。防同租户内越权看/改他人支付记录。
func (PaymentTransaction) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (PaymentTransaction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_payment_transactions",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("支付流水表"),
	}
}

func (PaymentTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("order_id").
			Comment("关联的订单ID").
			Optional().
			Nillable(),

		field.Uint32("user_id").
			Comment("用户ID").
			Optional().
			Nillable(),

		field.Int64("amount").
			Comment("支付金额（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("支付状态").
			NamedValues(
				"PaymentTransactionStatusPending", "PENDING",
				"PaymentTransactionStatusSucceeded", "SUCCEEDED",
				"PaymentTransactionStatusFailed", "FAILED",
				"PaymentTransactionStatusRefunded", "REFUNDED",
			).
			Optional().
			Nillable(),
	}
}

func (PaymentTransaction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
		appmixin.BusinessRefId{},
		appmixin.IdempotencyKey{},
		appmixin.PaymentMethod{},
		appmixin.BusinessType{},
	}
}

func (PaymentTransaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		// 幂等键唯一索引：与 order 一致，同租户内 idempotency_key 唯一，
		// 防止重放导致重复支付流水。
		index.Fields("tenant_id", "idempotency_key").Unique(),
	}
}
