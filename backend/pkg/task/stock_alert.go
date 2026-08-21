package task

const (
	StockAlertTaskType = "stock_alert"
)

// StockAlertTaskData 库存预警周期任务的载荷（周期性任务，无实体 ID）。
// 扫描全库 SKU，对 stock_qty 低于阈值的 SKU 生成站内消息通知运营补货。
type StockAlertTaskData struct {
	Name string `json:"name"`
}
