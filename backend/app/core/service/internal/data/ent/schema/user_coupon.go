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

type UserCoupon struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：普通用户只能查询/变更 user_id = 自身 userID 的券实例。
// 系统/平台视图（如超时取消/退款返还钩子）放行，以便跨用户反查 redeemed_order_id
// 进行券返还。防同租户内越权看/改他人券。与 order/order_item/payment_refund 同模式。
func (UserCoupon) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (UserCoupon) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_user_coupons",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("用户优惠券实例表"),
	}
}

func (UserCoupon) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("归属用户ID（由隐私层强制）").
			Optional().
			Nillable(),

		field.Uint32("coupon_template_id").
			Comment("关联的优惠券模板ID").
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("券实例状态").
			NamedValues(
				"UserCouponStatusUnused", "UNUSED",
				"UserCouponStatusUsed", "USED",
				"UserCouponStatusExpired", "EXPIRED",
			).
			Optional().
			Nillable(),

		field.Time("redeemed_at").
			Comment("核销时间（status=USED 时填写，返还时清空）").
			Optional().
			Nillable(),

		field.Uint32("redeemed_order_id").
			Comment("核销时关联的订单ID（返还钩子据此反查；UNUSED 时为空）").
			Optional().
			Nillable(),

		field.Int64("applied_discount_amount").
			Comment("实际抵扣额（最小货币单位，分；券侧审计，返还时清空）").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (UserCoupon) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (UserCoupon) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("coupon_template_id"),
		index.Fields("redeemed_order_id"),
		// 复合索引：支撑 per-user 限用 count 查询
		// (user_id, coupon_template_id, status)。
		// 用于 redeemCouponInTx 事务内的行锁 count（防 TOCTOU）与
		// Quote 的只读预览 count，使 count 走 index range scan 而非全表扫描。
		index.Fields("user_id", "coupon_template_id", "status"),
	}
}
