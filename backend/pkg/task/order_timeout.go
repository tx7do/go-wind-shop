package task

import "fmt"

const (
	OrderTimeoutTaskType = "order_timeout"
)

// OrderTimeoutTaskData 订单超时检查任务的载荷
type OrderTimeoutTaskData struct {
	OrderId  uint32 `json:"order_id"`  // 订单ID
	TaskId   string `json:"task_id,omitempty"`
	TraceId  string `json:"trace_id,omitempty"`
}

// CreateOrderTimeoutTaskID creates a unique task ID for an order timeout task.
func CreateOrderTimeoutTaskID(orderId uint32) string {
	return fmt.Sprintf("%s:%d",
		OrderTimeoutTaskType, orderId,
	)
}
