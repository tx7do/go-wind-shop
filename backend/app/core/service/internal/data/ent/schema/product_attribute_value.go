package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type ProductAttributeValue struct {
	ent.Schema
}

func (ProductAttributeValue) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_product_attribute_values",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品属性值表"),
	}
}

func (ProductAttributeValue) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("attribute_id").
			Comment("所属属性ID").
			Optional().
			Nillable(),
	}
}

func (ProductAttributeValue) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
	}
}

func (ProductAttributeValue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("attribute_id"),
	}
}
