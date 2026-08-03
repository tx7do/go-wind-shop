package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*BusinessType)(nil)

type BusinessType struct{ mixin.Schema }

func (BusinessType) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("business_type").
			NamedValues(
				"BusinessTypeRecharge", "BUSINESS_TYPE_RECHARGE",
				"BusinessTypeConsume", "BUSINESS_TYPE_CONSUME",
				"BusinessTypeRefund", "BUSINESS_TYPE_REFUND",
				"BusinessTypeGift", "BUSINESS_TYPE_GIFT",
				"BusinessTypeExpire", "BUSINESS_TYPE_EXPIRE",
				"BusinessTypeAdjust", "BUSINESS_TYPE_ADJUST",
				"BusinessTypeTransfer", "BUSINESS_TYPE_TRANSFER",
			).
			Optional().
			Nillable().
			Comment("业务类型"),
	}
}
