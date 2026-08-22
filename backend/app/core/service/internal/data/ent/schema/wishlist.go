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

type Wishlist struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：普通用户只能查询/变更 user_id = 自身 userID 的收藏项。
// user_id 在 Create 时由 UserPrivacy 强制覆盖为 viewer，List/Update/Delete 注入
// user_id=viewer 过滤。防同租户内枚举他人收藏。系统/平台视图放行。
func (Wishlist) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (Wishlist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_wishlist",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("收藏夹表"),
	}
}

func (Wishlist) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("用户ID（由隐私层强制）").
			Optional().
			Nillable(),

		field.Uint32("product_id").
			Comment("关联的商品ID").
			Optional().
			Nillable(),
	}
}

func (Wishlist) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (Wishlist) Indexes() []ent.Index {
	return []ent.Index{
		// 每用户每商品仅允许一条收藏记录。
		index.Fields("tenant_id", "user_id", "product_id").Unique(),
	}
}
