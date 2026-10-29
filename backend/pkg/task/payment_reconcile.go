package task

const (
	// PaymentReconcileTaskType 支付对账周期任务：
	// 扫描陈旧 PENDING 支付流水（超 N 分钟未确认的网关渠道流水），主动查询
	// 渠道侧订单状态——回调丢失/延迟时自愈：渠道已支付则补确认推进订单，
	// 渠道已关闭则收敛流水状态。模拟渠道与记账型渠道（BALANCE/OFFLINE）无
	// 网关语义，自然跳过。
	PaymentReconcileTaskType = "payment_reconcile"
)

// PaymentReconcileTaskData 支付对账任务的载荷（周期性任务，无实体 ID）。
type PaymentReconcileTaskData struct {
	Name string `json:"name"`
}

// GetName nil-safe 取名（与 proto 生成风格一致，供日志使用）。
func (d *PaymentReconcileTaskData) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}
