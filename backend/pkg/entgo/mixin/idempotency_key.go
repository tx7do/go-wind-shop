package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*IdempotencyKey)(nil)

type IdempotencyKey struct{ mixin.Schema }

func (IdempotencyKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("idempotency_key").
			Optional().
			Nillable().
			Comment("幂等键"),
	}
}
