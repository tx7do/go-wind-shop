package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*Currency)(nil)

type Currency struct{ mixin.Schema }

func (Currency) Fields() []ent.Field {
	return []ent.Field{
		field.String("currency").
			Optional().
			Nillable().
			Default("CNY").
			Comment("币种（ISO 4217，当前仅支持CNY）"),
	}
}
