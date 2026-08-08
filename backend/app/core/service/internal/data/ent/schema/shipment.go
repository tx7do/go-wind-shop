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

type Shipment struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：买家只能查询 user_id = 自身 userID 的物流单，
// 防止同租户内越权查看他人物流信息。admin/系统视图放行。
func (Shipment) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (Shipment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_shipments",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("物流单表"),
	}
}

func (Shipment) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("order_id").
			Comment("关联的订单ID").
			Optional().
			Nillable(),

		field.Uint32("user_id").
			Comment("归属用户ID（与订单 user_id 一致，用于行级隔离）").
			Optional().
			Nillable(),

		field.String("carrier").
			Comment("承运商").
			Optional().
			Nillable(),

		field.String("tracking_number").
			Comment("运单号").
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("物流单状态").
			NamedValues(
				"ShipmentStatusPending", "PENDING",
				"ShipmentStatusShipped", "SHIPPED",
				"ShipmentStatusDelivered", "DELIVERED",
			).
			Optional().
			Nillable(),

		field.String("tracking_events").
			Comment("物流轨迹事件列表（JSON 字符串，结构 [{timestamp,location,description}]）").
			Optional().
			Nillable(),
	}
}

func (Shipment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (Shipment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("user_id"),
	}
}
