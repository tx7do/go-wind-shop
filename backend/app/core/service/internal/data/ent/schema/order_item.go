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

type OrderItem struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：订单项随所属订单按用户隔离。
// user_id 在 Create 时由 UserPrivacy.EvalMutation 强制覆盖为 viewer（即所属 Order
// 的 user_id，因 Order 也受 UserPrivacy 约束、归属同一 viewer），List/Update/Delete
// 注入 user_id=viewer 过滤。防枚举他人 order_id 越权查/改他人订单项（与 CartItem
// 同模式）。系统/平台视图放行。
func (OrderItem) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (OrderItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_order_items",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("订单项表"),
	}
}

func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("order_id").
			Comment("关联的订单ID").
			Optional().
			Nillable(),

		field.Uint32("user_id").
			Comment("用户ID（随所属订单归属，由隐私层强制）").
			Optional().
			Nillable(),

		field.Uint32("sku_id").
			Comment("关联的 SKU ID").
			Optional().
			Nillable(),

		field.String("sku_snapshot").
			Comment("下单时 SKU 快照（JSON）").
			Optional().
			Nillable(),

		field.Int32("quantity").
			Comment("数量").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("unit_price").
			Comment("单价（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),

		field.Int64("subtotal").
			Comment("小计金额（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (OrderItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
	}
}
