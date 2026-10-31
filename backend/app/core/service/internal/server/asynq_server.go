package server

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-transport/transport/asynq"

	"go-wind-shop/app/core/service/internal/service"

	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"
)

// NewAsynqServer creates a new asynq server.
func NewAsynqServer(ctx *bootstrap.Context, taskService *service.TaskService, orderService *service.OrderService, userCouponService *service.UserCouponService, stockAlertService *service.StockAlertScannerService, productSearchService *service.ProductSearchService, paymentTransactionService *service.PaymentTransactionService, grouponService *service.GrouponService) *asynq.Server {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Asynq == nil {
		return nil
	}

	srv := asynq.NewServer(
		asynq.WithCodec(cfg.Server.Asynq.GetCodec()),
		asynq.WithRedisURI(cfg.Server.Asynq.GetUri()),
		asynq.WithLocation(cfg.Server.Asynq.GetLocation()),
		asynq.WithGracefullyShutdown(cfg.Server.Asynq.GetEnableGracefullyShutdown()),
		asynq.WithShutdownTimeout(cfg.Server.Asynq.GetShutdownTimeout().AsDuration()),
	)

	taskService.RegisterTaskScheduler(srv)
	orderService.RegisterTaskScheduler(srv)

	var err error

	// 注册任务
	if err = asynq.RegisterSubscriber(srv, task.BackupTaskType, taskService.AsyncBackup); err != nil {
		log.Error(err)
	}
	// 注册订单超时延时任务处理器（N 分钟后触发 ExpireOrderByTimeout，含幂等校验与库存释放）
	if err = asynq.RegisterSubscriber(srv, task.OrderTimeoutTaskType, orderService.HandleOrderTimeout); err != nil {
		log.Error(err)
	}
	// 注册订单兜底清扫周期任务处理器（扫描超时未支付订单兜底取消 + 发货超窗自动确认收货）
	if err = asynq.RegisterSubscriber(srv, task.OrderSweepTaskType, orderService.HandleOrderSweep); err != nil {
		log.Error(err)
	}
	// 注册优惠券过期清扫周期任务处理器（按 cron 周期触发 SweepExpiredCoupons，
	// 扫描全库过期券批量翻 EXPIRED）。仿 BackupTask 的 periodic 模式。
	if err = asynq.RegisterSubscriber(srv, task.CouponExpireSweepTaskType, userCouponService.HandleCouponSweep); err != nil {
		log.Error(err)
	}
	// 注册库存预警周期任务处理器（按 cron 周期触发 ScanLowStockAndNotify，
	// 扫描 stock_qty <= 阈值的 SKU 并向运营全员发送站内预警）。
	if err = asynq.RegisterSubscriber(srv, task.StockAlertTaskType, stockAlertService.HandleStockAlert); err != nil {
		log.Error(err)
	}
	// 注册商品搜索重索引任务处理器（商品/翻译变更后同步 ES 索引）。
	if err = asynq.RegisterSubscriber(srv, task.SearchReindexTaskType, productSearchService.ReindexProduct); err != nil {
		log.Error(err)
	}
	// 注册商品搜索周期全量重索引任务处理器（以 DB 为准双向收敛 ES 索引：
	// 补漏 upsert 种子数据/丢失任务 + 清扫孤儿文档，兜底单条任务覆盖不到的场景）。
	if err = asynq.RegisterSubscriber(srv, task.SearchReindexAllTaskType, productSearchService.HandleReindexAll); err != nil {
		log.Error(err)
	}
	// 注册支付对账周期任务处理器（扫描陈旧 PENDING 流水主动查渠道自愈：
	// 回调丢失则补确认推进订单；渠道已关闭则收敛流水状态）。
	if err = asynq.RegisterSubscriber(srv, task.PaymentReconcileTaskType, paymentTransactionService.HandlePaymentReconcile); err != nil {
		log.Error(err)
	}

	// 启动所有的任务
	_, _ = taskService.StartAllTask(appViewer.NewSystemViewerContext(ctx.Context()), nil)

	return srv
}
