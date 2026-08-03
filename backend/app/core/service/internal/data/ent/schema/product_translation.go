package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

type ProductTranslation struct {
	ent.Schema
}

func (ProductTranslation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_product_translations",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品翻译表"),
	}
}

func (ProductTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("product_id").
			Comment("关联的商品ID").
			Optional().
			Nillable(),

		field.String("language_code").
			Comment("语言代码").
			Optional().
			Nillable(),

		field.String("name").
			Comment("商品名称").
			Optional().
			Nillable(),

		field.String("slug").
			Comment("商品别名").
			Optional().
			Nillable(),

		field.String("short_description").
			Comment("商品简短描述").
			Optional().
			Nillable(),

		field.String("long_description").
			Comment("商品长描述").
			Optional().
			Nillable(),
	}
}

func (ProductTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (ProductTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "language_code").Unique(),
		index.Fields("product_id"),
		index.Fields("language_code"),
		index.Fields("slug"),
	}
}
