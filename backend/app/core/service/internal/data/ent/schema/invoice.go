package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"

	appmixin "go-wind-shop/pkg/entgo/mixin"
	appPrivacy "go-wind-shop/pkg/entgo/privacy"
)

type Invoice struct {
	ent.Schema
}

// Policy 注入 UserPrivacy：普通用户只能查询 user_id = 自身 userID 的发票记录。
// 系统/平台视图放行，以便 admin 管理开票。防同租户内越权看/改他人发票。
// 与 payment_refund / user_coupon 同模式。
func (Invoice) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (Invoice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_invoices",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("发票表"),
	}
}

func (Invoice) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("order_id").
			Comment("关联的订单ID").
			Optional().
			Nillable(),

		field.Uint32("user_id").
			Comment("归属用户ID（与订单 user_id 一致，用于行级隔离）").
			Optional().
			Nillable(),

		field.String("invoice_number").
			Comment("发票号").
			Optional().
			Nillable(),

		field.Enum("invoice_type").
			Comment("发票类型").
			NamedValues(
				"InvoiceInvoiceTypeVatGeneral", "VAT_GENERAL",
				"InvoiceInvoiceTypeVatSpecial", "VAT_SPECIAL",
				"InvoiceInvoiceTypeElectronic", "ELECTRONIC",
			).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("开票状态").
			NamedValues(
				"InvoiceStatusPending", "PENDING",
				"InvoiceStatusIssued", "ISSUED",
				"InvoiceStatusCancelled", "CANCELLED",
			).
			Optional().
			Nillable(),

		field.String("buyer_name").
			Comment("购方名称").
			Optional().
			Nillable(),

		field.String("buyer_tax_id").
			Comment("购方税号").
			Optional().
			Nillable(),

		field.String("buyer_address").
			Comment("购方地址").
			Optional().
			Nillable(),

		field.String("buyer_phone").
			Comment("购方电话").
			Optional().
			Nillable(),

		field.Int64("amount").
			Comment("开票金额（最小货币单位，分）").
			Default(0).
			Optional().
			Nillable(),
	}
}

func (Invoice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		appmixin.Currency{},
	}
}

func (Invoice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("order_id"),
		index.Fields("user_id"),
	}
}
