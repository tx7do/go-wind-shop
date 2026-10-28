package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type StockOpLog struct {
	ent.Schema
}

// StockOpLog 库存操作流水（append-only 审计）。
//
// 只写不改：每次 SKU stock_qty 变更（下单扣减/取消回补/退款回补/手动调整）
// 在同一事务内追加一条流水，记录变动前后值与原因，供运营追溯对账。
// 不提供 Update/Delete RPC；与 SKU 同为平台级数据（无租户隔离）。

func (StockOpLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_stock_op_logs",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("库存操作流水表"),
	}
}

func (StockOpLog) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("sku_id").
			Comment("SKU ID").
			Optional().
			Nillable(),

		field.Uint32("product_id").
			Comment("所属商品ID（冗余，便于按商品聚合查询）").
			Optional().
			Nillable(),

		field.Int32("delta").
			Comment("变动量（正=回补/入库，负=扣减/出库）").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("stock_before").
			Comment("变动前库存").
			Default(0).
			Optional().
			Nillable(),

		field.Int32("stock_after").
			Comment("变动后库存").
			Default(0).
			Optional().
			Nillable(),

		field.Enum("reason").
			Comment("变动原因").
			NamedValues(
				"StockOpReasonOrderDeduct", "ORDER_DEDUCT",
				"StockOpReasonCancelRestore", "CANCEL_RESTORE",
				"StockOpReasonRefundRestore", "REFUND_RESTORE",
				"StockOpReasonManualAdjust", "MANUAL_ADJUST",
			).
			Optional().
			Nillable(),

		field.Uint32("ref_id").
			Comment("关联业务ID（订单ID等，手动调整为 0）").
			Default(0).
			Optional().
			Nillable(),

		field.String("remark").
			Comment("备注（手动调整时填写）").
			Optional().
			Nillable(),
	}
}

func (StockOpLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

func (StockOpLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sku_id", "created_at"),
		index.Fields("product_id"),
		index.Fields("reason"),
	}
}
