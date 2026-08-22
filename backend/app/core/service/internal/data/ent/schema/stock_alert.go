package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

// StockAlert 库存预警记录表。
//
// 由 StockAlertScannerService 周期任务扫描低库存 SKU 时落库（OPEN 态），
// admin 在低库存预警管理台查看并标记 RESOLVED。全局表（与 SKU 同模型——
// SKU 是全局共享目录，告警亦跨租户）。无 UserPrivacy/TenantPrivacy，访问
// 纯由 RBAC（sys:platform_admin）兜底。
type StockAlert struct {
	ent.Schema
}

func (StockAlert) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_stock_alerts",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("库存预警记录表"),
	}
}

func (StockAlert) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("sku_id").
			Comment("触发告警的SKU ID").
			Optional().
			Nillable(),

		field.Int32("stock_qty_at_trigger").
			Comment("检测时的库存数量").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("threshold").
			Comment("触发阈值").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("告警状态").
			NamedValues(
				"StockAlertStatusOpen", "OPEN",
				"StockAlertStatusResolved", "RESOLVED",
			).
			Optional().
			Nillable(),
	}
}

func (StockAlert) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (StockAlert) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sku_id"),
		index.Fields("status"),
	}
}
