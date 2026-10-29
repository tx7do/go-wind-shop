package task

const (
	// GrouponSweepTaskType 拼团清扫周期任务：
	// 扫描 OPENING 团实时统计已支付成员数——达标翻 SUCCESS；过期未成团翻
	// FAILED 并对已支付成员订单自动退款关闭。
	GrouponSweepTaskType = "groupon_sweep"
)

// GrouponSweepTaskData 拼团清扫任务的载荷（周期性任务，无实体 ID）。
type GrouponSweepTaskData struct {
	Name string `json:"name"`
}

// GetName nil-safe 取名（与 proto 生成风格一致，供日志使用）。
func (d *GrouponSweepTaskData) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}
