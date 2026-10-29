package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type PaymentChannelConfig struct {
	ent.Schema
}

// 支付渠道配置：平台/租户维度的网关接入参数。凭据列存 AES-GCM 密文
// （服务层加密，ent 层无感知）；无用户行级隐私诉求，租户隔离走 TenantID mixin。
func (PaymentChannelConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_payment_channel_configs",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("支付渠道配置表"),
	}
}

func (PaymentChannelConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("渠道显示名").
			NotEmpty(),

		field.Enum("channel").
			Comment("渠道类型").
			NamedValues(
				"PaymentChannelTypeMock", "PAYMENT_CHANNEL_MOCK",
				"PaymentChannelTypeWechatNative", "PAYMENT_CHANNEL_WECHAT_NATIVE",
				"PaymentChannelTypeWechatH5", "PAYMENT_CHANNEL_WECHAT_H5",
				"PaymentChannelTypeAlipayPage", "PAYMENT_CHANNEL_ALIPAY_PAGE",
			).
			Default("PAYMENT_CHANNEL_MOCK").
			Optional().
			Nillable(),

		field.String("app_id").
			Comment("渠道应用ID（微信 appid / 支付宝 app_id）").
			Optional().
			Nillable(),

		field.String("mch_id").
			Comment("商户号（微信 mch_id；支付宝直连可空）").
			Optional().
			Nillable(),

		field.Bool("sandbox").
			Comment("沙箱模式（生产部署必须为 false）").
			Default(true),

		field.Bool("enabled").
			Comment("是否启用（仅启用的渠道出现在 C 端结算页）").
			// 安全默认：新渠道一律停用，运营显式开启
			Default(false),

		field.Text("credentials").
			Comment("网关凭据 JSON（AES-GCM 密文，enc: 前缀；读取路径不回填）").
			Optional().
			Nillable(),

		field.String("remark").
			Comment("备注").
			Optional().
			Nillable(),
	}
}

func (PaymentChannelConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

func (PaymentChannelConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "channel"),
	}
}
