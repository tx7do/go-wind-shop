package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type MemberTier struct {
	ent.Schema
}

// 会员等级：按累计已支付金额自动分层。level 越大越高，min_spend 为门槛，
// points_multiplier 积分入账倍率。用户等级 = 累计消费满足的最高等级。
func (MemberTier) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_member_tiers",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("会员等级表"),
	}
}

func (MemberTier) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("等级名称").NotEmpty(),
		field.Int32("level").Comment("等级（数值越大越高）").Default(1),
		field.Int64("min_spend").Comment("累计消费门槛（分）").Default(0),
		field.Int32("points_multiplier").Comment("积分倍率").Default(1),
		field.String("badge_icon").Comment("等级徽章图标").Optional().Nillable(),
		field.String("currency").Comment("币种").Default("CNY"),
		field.Enum("status").
			Comment("状态").
			NamedValues(
				"MemberTierStatusActive", "ACTIVE",
				"MemberTierStatusInactive", "INACTIVE",
			).
			Default("ACTIVE").
			Optional().
			Nillable(),
	}
}

func (MemberTier) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (MemberTier) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "level"),
	}
}
