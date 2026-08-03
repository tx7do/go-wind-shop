package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

// CategoryTranslation holds the schema definition for the CategoryTranslation entity.
type CategoryTranslation struct {
	ent.Schema
}

func (CategoryTranslation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_category_translations",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品类目翻译表"),
	}
}

func (CategoryTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("category_id").
			Comment("关联的类目ID").
			Optional().
			Nillable(),

		field.String("language_code").
			Comment("语言代码").
			Optional().
			Nillable(),

		field.String("name").
			Comment("类目名称").
			Optional().
			Nillable(),

		field.String("slug").
			Comment("类目别名").
			Optional().
			Nillable(),

		field.String("description").
			Comment("类目描述").
			Optional().
			Nillable(),

		field.String("full_path").
			Comment("完整路径").
			Optional().
			Nillable(),
	}
}

func (CategoryTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (CategoryTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category_id", "language_code").Unique(),
		index.Fields("category_id"),
		index.Fields("language_code"),
		index.Fields("slug"),
	}
}
