package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type SkuAttributeCombination struct {
	ent.Schema
}

func (SkuAttributeCombination) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_sku_attribute_combinations",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("SKU 属性组合关联表"),
	}
}

func (SkuAttributeCombination) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("sku_id").
			Comment("关联的 SKU ID").
			Optional().
			Nillable(),

		field.Uint32("attribute_id").
			Comment("属性ID").
			Optional().
			Nillable(),

		field.Uint32("attribute_value_id").
			Comment("属性值ID").
			Optional().
			Nillable(),
	}
}

func (SkuAttributeCombination) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (SkuAttributeCombination) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sku_id"),
		index.Fields("attribute_id"),
		index.Fields("attribute_value_id"),
	}
}
