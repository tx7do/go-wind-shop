package service

import (
	"context"
	"fmt"

	"github.com/tx7do/go-utils/trans"

	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
)

// 本文件是订单状态变更 → 买家站内通知的生产者缝。
//
// 链路：订单状态机推进点（Update / ExpireOrderByTimeout / ConfirmReceipt / 扫描任务 /
// 退款成功）→ InternalMessageService.SendMessage（定向收件人）→ 落站内信收件行 →
// Notifier 出站缝（INTERNAL 渠道确认送达；EMAIL 渠道按通知规则路由）。
// 全程 best-effort：通知失败仅记日志，不回滚订单状态。

// notifyOrderStatus 按目标状态发送标准文案通知（best-effort）。
func (s *OrderService) notifyOrderStatus(ctx context.Context, orderId uint32, status orderV1.Order_Status) {
	title, content := orderStatusNotificationText(status, orderId)
	if title == "" {
		return
	}
	s.notifyOrderCustom(ctx, orderId, title, content)
}

// notifyOrderCustom 发送自定义文案的订单通知（best-effort）。
func (s *OrderService) notifyOrderCustom(ctx context.Context, orderId uint32, title, content string) {
	if s.internalMessageService == nil {
		return
	}

	orderEnt, err := s.orderRepo.Get(ctx, &orderV1.GetOrderRequest{
		QueryBy: &orderV1.GetOrderRequest_Id{Id: orderId},
	})
	if err != nil || orderEnt == nil {
		s.log.Warnf("notify order status: load order [%d] failed: %v", orderId, err)
		return
	}
	uid := orderEnt.GetUserId()
	if uid == 0 {
		return
	}

	if _, sErr := s.internalMessageService.SendMessage(ctx, &internalMessageV1.SendMessageRequest{
		Title:           trans.Ptr(title),
		Content:         content,
		RecipientUserId: trans.Ptr(uid),
		SendUserId:      trans.Ptr(uint32(0)),
	}); sErr != nil {
		s.log.Warnf("notify order status: send message for order [%d] to user [%d] failed: %v", orderId, uid, sErr)
	}
}

// notifyOrderRefundSuccess 退款成功通知（best-effort）。金额为最小货币单位（分）。
func (s *OrderService) notifyOrderRefundSuccess(ctx context.Context, orderId uint32, amount int64) {
	content := fmt.Sprintf("您的订单 #%d 的退款已处理成功，退款金额 %.2f 元，请注意查收。", orderId, float64(amount)/100)
	s.notifyOrderCustom(ctx, orderId, "退款成功通知", content)
}

// orderStatusNotificationText 状态变更的标准通知文案。
// 返回空 title 表示该状态不发通知（CLOSED 由确认收货/自动确认路径发自定义文案，
// 避免通用"订单已关闭"与专属文案重复打扰）。
func orderStatusNotificationText(status orderV1.Order_Status, orderId uint32) (string, string) {
	switch status {
	case orderV1.Order_PAID:
		return "订单支付成功", fmt.Sprintf("您的订单 #%d 已支付成功，我们将尽快为您安排发货。", orderId)
	case orderV1.Order_FULFILLED:
		return "订单已发货", fmt.Sprintf("您的订单 #%d 已发货，请注意查收物流信息。", orderId)
	case orderV1.Order_CANCELLED:
		return "订单已取消", fmt.Sprintf("您的订单 #%d 已取消。", orderId)
	default:
		return "", ""
	}
}
