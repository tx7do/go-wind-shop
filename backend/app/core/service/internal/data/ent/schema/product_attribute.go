package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"github.com/tx7do/go-crud/entgo/mixin"
)

type ProductAttribute struct {
	ent.Schema
}

func (ProductAttribute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_product_attributes",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品属性表"),
	}
}

func (ProductAttribute) Fields() []ent.Field {
	return []ent.Field{}
}

func (ProductAttribute) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
	}
}

func (ProductAttribute) Indexes() []ent.Index {
	return []ent.Index{}
}
