package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type GrouponGroup struct {
	ent.Schema
}

// 拼团：以订单为成员载体（order.groupon_group_id）。paid_count 达 target_size
// 即 SUCCESS；expire_at 到期未成团 → FAILED（groupon_sweep 自动退款关闭成员订单）。
func (GrouponGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_groupon_groups",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("拼团表"),
	}
}

func (GrouponGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("activity_id").Comment("活动ID").Optional().Nillable(),
		field.Uint32("leader_order_id").Comment("团长订单ID").Optional().Nillable(),
		field.Int32("member_count").Comment("已参团人数（含未支付）").Default(1),
		field.Int32("paid_count").Comment("已支付人数").Default(0),
		field.Int32("target_size").Comment("成团所需人数").Default(2),
		field.Time("expire_at").Comment("成团截止时间").Optional().Nillable(),
		field.Enum("status").
			Comment("团状态").
			NamedValues(
				"GrouponGroupStatusOpening", "OPENING",
				"GrouponGroupStatusSuccess", "SUCCESS",
				"GrouponGroupStatusFailed", "FAILED",
			).
			Default("OPENING").
			Optional().
			Nillable(),
	}
}

func (GrouponGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (GrouponGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("activity_id", "status"),
		index.Fields("status", "expire_at"),
	}
}
