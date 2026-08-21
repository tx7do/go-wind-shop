package service

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-shop/app/core/service/internal/data"
	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"

	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
)

// StockAlertThreshold 触发库存预警的阈值（stock_qty 低于此值即告警）。
const StockAlertThreshold int32 = 10

// StockAlertService 库存预警周期任务服务。
// 扫描 stock_qty <= 阈值的 SKU，向全体运营人员发送站内消息通知补货。
type StockAlertService struct {
	log        *log.Helper
	skuRepo    *data.SkuRepo
	messageSvc *InternalMessageService
}

func NewStockAlertService(
	ctx *bootstrap.Context,
	skuRepo *data.SkuRepo,
	messageSvc *InternalMessageService,
) *StockAlertService {
	return &StockAlertService{
		log:        ctx.NewLoggerHelper("stock-alert/service/core-service"),
		skuRepo:    skuRepo,
		messageSvc: messageSvc,
	}
}

// HandleStockAlert 是 asynq "stock_alert" 周期任务的 worker handler。
// 注入 SystemViewer 跨租户扫描 SKU，随后向运营全员发送站内预警消息。
func (s *StockAlertService) HandleStockAlert(taskType string, taskData *task.StockAlertTaskData) error {
	s.log.Infow(
		"msg", "HandleStockAlert started",
		"task_type", taskType,
	)
	ctx := appViewer.NewSystemViewerContext(context.Background())
	return s.ScanLowStockAndNotify(ctx)
}

// ScanLowStockAndNotify 扫描低库存 SKU 并发送预警通知。
// 1. 查询 stock_qty <= 阈值 的 SKU 列表
// 2. 若有低库存项，向全体运营发送站内消息（含 SKU ID 与当前库存数）
// 3. 无低库存则跳过
func (s *StockAlertService) ScanLowStockAndNotify(ctx context.Context) error {
	items, err := s.skuRepo.ListLowStock(ctx, StockAlertThreshold)
	if err != nil {
		s.log.Errorf("scan low stock failed: %v", err)
		return err
	}

	if len(items) == 0 {
		s.log.Infow("msg", "stock alert scan completed: no low-stock items", "threshold", StockAlertThreshold)
		return nil
	}

	// 构造预警消息内容
	content := "检测到 " + strconv.Itoa(len(items)) + " 个 SKU 库存低于阈值(" +
		strconv.FormatInt(int64(StockAlertThreshold), 10) + ")，请及时补货。明细：\n"
	for _, item := range items {
		content += "SKU ID=" + strconv.FormatUint(uint64(item.SkuID), 10) +
			" 当前库存=" + strconv.FormatInt(int64(item.StockQty), 10) + "\n"
	}

	s.log.Warnw(
		"msg", "low stock detected, sending alert",
		"low_stock_count", len(items),
	)

	// 向全体运营发送站内预警消息（target_all=true）。
	// Type 默认 NOTIFICATION（枚举零值），SendUserId=0 表示系统发送。
	_, sendErr := s.messageSvc.SendMessage(ctx, &internalMessageV1.SendMessageRequest{
		Title:      trans.Ptr("库存预警通知"),
		Content:    content,
		TargetAll:  trans.Ptr(true),
		SendUserId: trans.Ptr(uint32(0)),
	})
	if sendErr != nil {
		s.log.Errorf("send stock alert internal message failed: %v", sendErr)
		return sendErr
	}

	s.log.Infow("msg", "stock alert notification sent", "recipients", "all", "low_stock_count", len(items))
	return nil
}
