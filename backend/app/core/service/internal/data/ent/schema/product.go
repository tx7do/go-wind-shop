package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type Product struct {
	ent.Schema
}

func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_products",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品表"),
	}
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Comment("商品状态").
			NamedValues(
				"ProductStatusDraft", "PRODUCT_STATUS_DRAFT",
				"ProductStatusActive", "PRODUCT_STATUS_ACTIVE",
				"ProductStatusInactive", "PRODUCT_STATUS_INACTIVE",
			).
			Optional().
			Nillable(),

		field.Uint32("category_id").
			Comment("所属类目ID").
			Optional().
			Nillable(),

		field.Uint32("brand_id").
			Comment("所属品牌ID").
			Optional().
			Nillable(),

		field.String("image_url").
			Comment("商品主图资源 URL（locale 无关）").
			Optional().
			Nillable(),
	}
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
	}
}

func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category_id"),
		index.Fields("brand_id"),
		index.Fields("status"),
	}
}
