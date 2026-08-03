package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*BusinessRefId)(nil)

type BusinessRefId struct{ mixin.Schema }

func (BusinessRefId) Fields() []ent.Field {
	return []ent.Field{
		field.String("business_ref_id").
			Optional().
			Nillable().
			Comment("业务单号"),
	}
}
