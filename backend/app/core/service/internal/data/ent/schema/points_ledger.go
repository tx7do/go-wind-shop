package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"

	appPrivacy "go-wind-shop/pkg/entgo/privacy"
)

type PointsLedger struct {
	ent.Schema
}

// 积分台账：append-only。余额 = 用户全部 delta 求和；订单入账幂等由
// (tenant_id, ref_order_id, type=EARN) 查重保证。
func (PointsLedger) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_points_ledgers",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("积分台账表"),
	}
}

// Policy 注入 UserPrivacy：普通用户只能读本人台账。
func (PointsLedger) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (PointsLedger) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("用户ID").
			Optional().
			Nillable(),

		field.Int64("delta").
			Comment("积分变动（±）").
			Default(0),

		field.Int64("balance_after").
			Comment("变动后余额快照").
			Default(0),

		field.Enum("type").
			Comment("类型").
			NamedValues(
				"PointsTypeEarn", "POINTS_TYPE_EARN",
				"PointsTypeDeduct", "POINTS_TYPE_DEDUCT",
			).
			Default("POINTS_TYPE_EARN").
			Optional().
			Nillable(),

		field.String("reason").
			Comment("事由").
			Default(""),


		field.Uint32("ref_order_id").
			Comment("关联订单ID").
			Optional().
			Nillable(),
	}
}

func (PointsLedger) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (PointsLedger) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_id"),
		index.Fields("ref_order_id"),
	}
}
