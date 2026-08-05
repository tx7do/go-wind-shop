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

type ShippingAddress struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：用户只能查询/变更 user_id = 自身 userID 的收货地址，
// 防止同租户内越权读取/篡改他人地址簿。
func (ShippingAddress) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (ShippingAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "user_shipping_addresses",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("用户收货地址表"),
	}
}

func (ShippingAddress) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("所属用户ID").
			Optional().
			Nillable(),

		field.String("recipient_name").
			Comment("收件人姓名").
			Optional().
			Nillable(),

		field.String("recipient_phone").
			Comment("收件人电话").
			Optional().
			Nillable(),

		field.String("region").
			Comment("省/市/区（结构化文本）").
			Optional().
			Nillable(),

		field.String("detail_address").
			Comment("详细地址").
			Optional().
			Nillable(),

		field.String("postal_code").
			Comment("邮政编码").
			Optional().
			Nillable(),

		field.String("tag").
			Comment("地址标签（如：家、公司、学校）").
			Optional().
			Nillable(),

		field.Bool("is_default").
			Comment("是否默认地址").
			Default(false).
			Optional().
			Nillable(),
	}
}

func (ShippingAddress) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (ShippingAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
