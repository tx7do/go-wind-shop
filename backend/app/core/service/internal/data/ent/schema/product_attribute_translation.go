package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type ProductAttributeTranslation struct {
	ent.Schema
}

func (ProductAttributeTranslation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_product_attribute_translations",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品属性翻译表"),
	}
}

func (ProductAttributeTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("attribute_id").
			Comment("关联的属性ID").
			Optional().
			Nillable(),

		field.String("language_code").
			Comment("语言代码").
			Optional().
			Nillable(),

		field.String("name").
			Comment("属性名称").
			Optional().
			Nillable(),
	}
}

func (ProductAttributeTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (ProductAttributeTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("attribute_id", "language_code").Unique(),
		index.Fields("attribute_id"),
		index.Fields("language_code"),
	}
}
