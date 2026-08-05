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

type CartItem struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：购物车项随所属购物车按用户隔离。
// user_id 在 Create 时由 UserPrivacy 强制覆盖为 viewer（即所属 cart 的 user_id，
// 因 cart 也受 UserPrivacy 约束、归属同一 viewer），List/Update/Delete 注入
// user_id=viewer 过滤。防同租户内枚举他人 cart_id 越权查/改他人购物车项。
// 系统/平台视图放行。
func (CartItem) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (CartItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_cart_items",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("购物车项表"),
	}
}

func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("cart_id").
			Comment("关联的购物车ID").
			Optional().
			Nillable(),

		field.Uint32("user_id").
			Comment("用户ID（随所属购物车归属，由隐私层强制）").
			Optional().
			Nillable(),

		field.Uint32("sku_id").
			Comment("关联的 SKU ID").
			Optional().
			Nillable(),

		field.Int32("quantity").
			Comment("数量").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (CartItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id"),
	}
}
