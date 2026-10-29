package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type SeckillActivity struct {
	ent.Schema
}

// 秒杀活动：SKU 级限时限量抢购。sold_count 在下单事务内行锁扣减防超卖；
// per_user_limit 限每用户在活动窗口内的累计购买量。
func (SeckillActivity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_seckill_activities",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("秒杀活动表"),
	}
}

func (SeckillActivity) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("活动名称").
			NotEmpty(),

		field.Uint32("sku_id").
			Comment("秒杀 SKU ID").
			Optional().
			Nillable(),

		field.Int64("seckill_price").
			Comment("秒杀价（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("total_stock").
			Comment("秒杀总限量（件）").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("sold_count").
			Comment("已抢购件数").
			Default(0),
		field.Int32("per_user_limit").
			Comment("每人限购件数（活动窗口内累计）").
			Default(1),
		field.Int32("priority").
			Comment("优先级（同 SKU 并发活动取最高者）").
			Default(0),

		field.Time("valid_from").
			Comment("生效时间").
			Optional().
			Nillable(),
		field.Time("valid_until").
			Comment("失效时间").
			Optional().
			Nillable(),

		field.String("currency").
			Comment("币种").
			Default("CNY"),

		field.Enum("status").
			Comment("活动状态").
			NamedValues(
				"SeckillStatusActive", "ACTIVE",
				"SeckillStatusInactive", "INACTIVE",
			).
			Default("ACTIVE").
			Optional().
			Nillable(),
	}
}

func (SeckillActivity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (SeckillActivity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "sku_id", "status"),
	}
}
