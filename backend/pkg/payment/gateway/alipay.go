package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/smartwalle/alipay/v3"

	"go-wind-shop/pkg/payment"
)

// AlipayPageGateway 支付宝电脑网站支付（page pay，直连商户）。
// sandbox=true 走支付宝开放平台沙箱网关（沙箱账号验收，不出真实资金）。
type AlipayPageGateway struct {
	client *alipay.Client
}

func NewAlipayPageGateway(appID string, sandbox bool, credentialsJSON string) (*AlipayPageGateway, error) {
	cred, err := payment.ParseAlipayCredentials(credentialsJSON)
	if err != nil {
		return nil, err
	}

	// production=false 即沙箱网关
	client, err := alipay.New(appID, cred.PrivateKey, !sandbox)
	if err != nil {
		return nil, fmt.Errorf("create alipay client failed: %w", err)
	}
	if err = client.LoadAliPayPublicKey(cred.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("load alipay public key failed: %w", err)
	}

	return &AlipayPageGateway{client: client}, nil
}

func (g *AlipayPageGateway) CreatePayment(ctx context.Context, req PayRequest) (*PayResult, error) {
	payURL, err := g.client.TradePagePay(alipay.TradePagePay{
		Trade: alipay.Trade{
			NotifyURL:   req.NotifyURL,
			Subject:     req.Subject,
			OutTradeNo:  req.OutTradeNo,
			TotalAmount: CentsToYuanString(req.AmountCents),
			ProductCode: "FAST_INSTANT_TRADE_PAY",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("alipay page pay build failed: %w", err)
	}

	return &PayResult{RedirectURL: payURL.String()}, nil
}

// ParseNotify 验签并解析异步通知（只有 TRADE_SUCCESS 才算支付成功）。
func (g *AlipayPageGateway) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("parse alipay notify form failed: %w", err)
	}

	n, err := g.client.DecodeNotification(ctx, r.Form)
	if err != nil {
		// 验签失败/参数缺失：一律拒绝，调用方应返回失败应答让网关重试
		return nil, fmt.Errorf("verify alipay notify failed: %w", err)
	}

	paid, err := YuanStringToCents(n.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("parse alipay notify amount failed: %w", err)
	}

	return &NotifyResult{
		OutTradeNo:     n.OutTradeNo,
		GatewayTradeNo: n.TradeNo,
		PaidCents:      paid,
		Success:        n.TradeStatus == alipay.TradeStatusSuccess,
	}, nil
}

// CloseOrder 关闭未支付的支付宝交易（订单取消/超时后让旧收银台链接失效）。
// 注意：alipay.trade.close 对「已支付」交易会失败——该情形由支付回调确认
// 路径兜底，此处错误仅由调用方记录。
func (g *AlipayPageGateway) CloseOrder(ctx context.Context, outTradeNo string) error {
	rsp, err := g.client.TradeClose(ctx, alipay.TradeClose{
		OutTradeNo: outTradeNo,
	})
	if err != nil {
		return fmt.Errorf("alipay close order [%s] failed: %w", outTradeNo, err)
	}
	if rsp != nil && !rsp.Error.IsSuccess() {
		return fmt.Errorf("alipay close order [%s] rejected: %s", outTradeNo, rsp.Error.Error())
	}
	return nil
}

// Refund 发起支付宝退款（alipay.trade.refund，同步受理）。
// 同一 out_request_no 重复提交渠道幂等处理（受理成功但 FundChange=N）。
func (g *AlipayPageGateway) Refund(ctx context.Context, req RefundRequest) error {
	rsp, err := g.client.TradeRefund(ctx, alipay.TradeRefund{
		OutTradeNo:   req.OutTradeNo,
		OutRequestNo: req.OutRequestNo,
		RefundAmount: CentsToYuanString(req.RefundCents),
		RefundReason: req.Reason,
	})
	if err != nil {
		return fmt.Errorf("alipay refund [%s/%s] failed: %w", req.OutTradeNo, req.OutRequestNo, err)
	}
	if rsp != nil && !rsp.Error.IsSuccess() {
		return fmt.Errorf("alipay refund [%s/%s] rejected: %s", req.OutTradeNo, req.OutRequestNo, rsp.Error.Error())
	}
	return nil
}

// QueryOrder 主动查询支付宝订单状态（对账兜底；alipay.trade.query 同步查询）。
func (g *AlipayPageGateway) QueryOrder(ctx context.Context, outTradeNo string) (*OrderQueryResult, error) {
	rsp, err := g.client.TradeQuery(ctx, alipay.TradeQuery{
		OutTradeNo: outTradeNo,
	})
	if err != nil {
		return nil, fmt.Errorf("alipay query order [%s] failed: %w", outTradeNo, err)
	}
	if rsp != nil && !rsp.Error.IsSuccess() {
		// 交易不存在 = 从未支付过，按未支付状态返回（非错误）
		return &OrderQueryResult{TradeState: "NOT_FOUND"}, nil
	}
	if rsp == nil {
		return nil, fmt.Errorf("alipay query order [%s] returned empty response", outTradeNo)
	}

	paidCents, pErr := YuanStringToCents(rsp.TotalAmount)
	if pErr != nil {
		paidCents = 0
	}
	return &OrderQueryResult{
		GatewayTradeNo: rsp.TradeNo,
		TradeState:     string(rsp.TradeStatus),
		PaidCents:      paidCents,
		Success:        rsp.TradeStatus == alipay.TradeStatusSuccess,
	}, nil
}