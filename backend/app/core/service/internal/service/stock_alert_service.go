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

// StockAlertScannerService 库存预警周期任务服务。
// 扫描 stock_qty <= 阈值的 SKU，向全体运营人员发送站内消息通知补货，
// 并向 stock_alert 表落库 OPEN 态告警记录（去重：同 SKU 仅保留一条 OPEN）。
type StockAlertScannerService struct {
	log            *log.Helper
	skuRepo        *data.SkuRepo
	messageSvc     *InternalMessageService
	stockAlertRepo *data.StockAlertRepo
}

func NewStockAlertScannerService(
	ctx *bootstrap.Context,
	skuRepo *data.SkuRepo,
	messageSvc *InternalMessageService,
	stockAlertRepo *data.StockAlertRepo,
) *StockAlertScannerService {
	return &StockAlertScannerService{
		log:            ctx.NewLoggerHelper("stock-alert-scanner/service/core-service"),
		skuRepo:        skuRepo,
		messageSvc:     messageSvc,
		stockAlertRepo: stockAlertRepo,
	}
}

// HandleStockAlert 是 asynq "stock_alert" 周期任务的 worker handler。
// 注入 SystemViewer 跨租户扫描 SKU，随后向运营全员发送站内预警消息并落库告警记录。
func (s *StockAlertScannerService) HandleStockAlert(taskType string, taskData *task.StockAlertTaskData) error {
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
func (s *StockAlertScannerService) ScanLowStockAndNotify(ctx context.Context) error {
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

	// 落库 OPEN 态告警记录（去重：同 SKU 仅保留一条 OPEN）。
	// 单线程 cron 无 TOCTOU 风险。任一记录插入失败仅记日志，不阻断发信。
	for _, item := range items {
		if err := s.stockAlertRepo.CreateAlertIfNotOpen(ctx, item.SkuID, item.StockQty, StockAlertThreshold); err != nil {
			s.log.Errorf("persist stock alert record failed for sku %d: %v", item.SkuID, err)
		}
	}

	s.log.Infow("msg", "stock alert notification sent", "recipients", "all", "low_stock_count", len(items))
	return nil
}
