package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type Brand struct {
	ent.Schema
}

func (Brand) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_brands",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("品牌表"),
	}
}

func (Brand) Fields() []ent.Field {
	return []ent.Field{
		field.String("logo_url").
			Comment("品牌 Logo 资源 URL（locale 无关）").
			Optional().
			Nillable(),
	}
}

func (Brand) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
	}
}

func (Brand) Indexes() []ent.Index {
	return []ent.Index{}
}
