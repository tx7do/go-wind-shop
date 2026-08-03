package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type BrandTranslation struct {
	ent.Schema
}

func (BrandTranslation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_brand_translations",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("品牌翻译表"),
	}
}

func (BrandTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("brand_id").
			Comment("关联的品牌ID").
			Optional().
			Nillable(),

		field.String("language_code").
			Comment("语言代码").
			Optional().
			Nillable(),

		field.String("name").
			Comment("品牌名称").
			Optional().
			Nillable(),

		field.String("slug").
			Comment("品牌别名").
			Optional().
			Nillable(),

		field.String("description").
			Comment("品牌描述").
			Optional().
			Nillable(),
	}
}

func (BrandTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (BrandTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("brand_id", "language_code").Unique(),
		index.Fields("brand_id"),
		index.Fields("language_code"),
		index.Fields("slug"),
	}
}
