package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/tx7do/go-crud/entgo/mixin"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

func (Category) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_categories",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品类目表"),
	}
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("depth").
			Comment("类目层级深度").
			Default(0).
			Optional().
			Nillable(),

		field.String("image_url").
			Comment("类目图片资源 URL（locale 无关）").
			Optional().
			Nillable(),
	}
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
		mixin.TreePath{},
		mixin.Tree[Category]{},
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{}
}
