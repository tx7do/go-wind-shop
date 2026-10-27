package channel

import (
	"context"
	"strconv"

	notificationV1 "go-wind-shop/api/gen/go/notification/service/v1"
)

// InternalChannelSender 站内信渠道（shop 特有实现）。
//
// 与 go-wind-cms 的 InternalMessageSender 语义差异：cms 由本 sender 落收件行，
// 而 shop 的发送编排（admin BFF SendMessage → 逐收件人 RecipientService/Create）
// 在调用通知域之前就已落好收件行并完成实时推送——本 sender 在投递链路里只做
// "站内行已落库"的送达确认，把台账记为 SENT；不重复写收件行，也不重复推送。
type InternalChannelSender struct{}

func NewInternalChannelSender() *InternalChannelSender {
	return &InternalChannelSender{}
}

func (s *InternalChannelSender) Channel() notificationV1.Channel {
	return notificationV1.Channel_INTERNAL
}

// Send 确认站内送达。收件行由发送编排负责（见上），这里校验收件人即可。
func (s *InternalChannelSender) Send(_ context.Context, req *SendRequest) (*SendReceipt, error) {
	if req == nil {
		return nil, ErrChannelNotConfigured
	}
	if req.RecipientUserID == 0 {
		// INTERNAL 渠道必须有收件人；Target 只是它的字符串形式
		if _, err := strconv.ParseUint(req.Target, 10, 32); err != nil {
			return nil, ErrChannelNotConfigured
		}
	}
	return &SendReceipt{}, nil
}
