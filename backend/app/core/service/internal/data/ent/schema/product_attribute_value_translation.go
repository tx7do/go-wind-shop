package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type ProductAttributeValueTranslation struct {
	ent.Schema
}

func (ProductAttributeValueTranslation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_product_attribute_value_translations",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品属性值翻译表"),
	}
}

func (ProductAttributeValueTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("value_id").
			Comment("关联的属性值ID").
			Optional().
			Nillable(),

		field.String("language_code").
			Comment("语言代码").
			Optional().
			Nillable(),

		field.String("display_name").
			Comment("属性值显示名").
			Optional().
			Nillable(),
	}
}

func (ProductAttributeValueTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (ProductAttributeValueTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("value_id", "language_code").Unique(),
		index.Fields("value_id"),
		index.Fields("language_code"),
	}
}
