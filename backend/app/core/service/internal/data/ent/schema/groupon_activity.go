package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type GrouponActivity struct {
	ent.Schema
}

// 拼团活动：SKU 级成团价抢购，group_size 人支付成团，单团时限内未成团自动退款。
func (GrouponActivity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_groupon_activities",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("拼团活动表"),
	}
}

func (GrouponActivity) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("活动名称").NotEmpty(),
		field.Uint32("sku_id").Comment("拼团 SKU ID").Optional().Nillable(),
		field.Int64("group_price").Comment("成团价（分）").Default(0).Optional().Nillable(),
		field.Int32("group_size").Comment("成团人数").Default(2),
		field.Int32("group_valid_hours").Comment("单团成团时限（小时）").Default(24),
		field.Int32("per_user_limit").Comment("每人限购（活动窗口内）").Default(1),
		field.Time("valid_from").Comment("生效时间").Optional().Nillable(),
		field.Time("valid_until").Comment("失效时间").Optional().Nillable(),
		field.String("currency").Comment("币种").Default("CNY"),
		field.Enum("status").
			Comment("活动状态").
			NamedValues(
				"GrouponStatusActive", "ACTIVE",
				"GrouponStatusInactive", "INACTIVE",
			).
			Default("ACTIVE").
			Optional().
			Nillable(),
	}
}

func (GrouponActivity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (GrouponActivity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "sku_id", "status"),
	}
}
