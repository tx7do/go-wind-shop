package task

const (
	// OrderSweepTaskType 订单兜底清扫周期任务：
	//  1. 扫描超过 TTL 仍 PENDING_PAYMENT 的订单，走 ExpireOrderByTimeout 兜底取消
	//     （覆盖延时任务因 Redis 清空/worker 宕机跨过 TTL 而丢失的场景）
	//  2. 扫描超过自动确认收货窗口仍 FULFILLED 的订单，推进 CLOSED 并通知买家
	OrderSweepTaskType = "order_sweep"
)

// OrderSweepTaskData 订单清扫任务的载荷（周期性任务，无实体 ID）。
type OrderSweepTaskData struct {
	Name string `json:"name"`
}

// GetName nil-safe 取名（与 proto 生成风格一致，供日志使用）。
func (d *OrderSweepTaskData) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}
