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
