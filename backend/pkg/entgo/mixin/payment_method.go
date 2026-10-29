package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*PaymentMethod)(nil)

type PaymentMethod struct{ mixin.Schema }

func (PaymentMethod) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("payment_method").
			NamedValues(
				"Balance", "BALANCE",
				"Credit", "CREDIT",
				"Wechat", "WECHAT",
				"Alipay", "ALIPAY",
				"Unionpay", "UNIONPAY",
				"Stripe", "STRIPE",
				"Paypal", "PAYPAL",
				"Mock", "MOCK",
				"Mixed", "MIXED",
				"Offline", "OFFLINE",
			).
			Optional().
			Nillable().
			Comment("支付方式"),
	}
}
