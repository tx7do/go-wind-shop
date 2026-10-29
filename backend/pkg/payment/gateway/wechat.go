package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"

	"go-wind-shop/pkg/payment"
)

// WechatNativeGateway 微信支付 Native（扫码支付，直连商户，APIv3）。
//
// 注意：微信支付 V3 无沙箱环境——sandbox=true 时拒绝创建（防止误把生产商户号
// 当沙箱刷）；验收用真实商户号小额支付。
type WechatNativeGateway struct {
	appID    string
	mchID    string
	apiV3Key string
	client   *core.Client
}

func NewWechatNativeGateway(appID, mchID string, sandbox bool, credentialsJSON string) (*WechatNativeGateway, error) {
	if sandbox {
		return nil, fmt.Errorf("wechat pay v3 has no sandbox environment: disable sandbox mode and verify with small real payments")
	}
	cred, err := payment.ParseWechatCredentials(credentialsJSON)
	if err != nil {
		return nil, err
	}

	privKey, err := utils.LoadPrivateKey(cred.PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("load wechat merchant private key failed: %w", err)
	}

	ctx := context.Background()
	client, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(mchID, cred.MchCertSerialNo, privKey, cred.APIv3Key),
	)
	if err != nil {
		return nil, fmt.Errorf("create wechat pay client failed: %w", err)
	}

	return &WechatNativeGateway{appID: appID, mchID: mchID, apiV3Key: cred.APIv3Key, client: client}, nil
}

func (g *WechatNativeGateway) CreatePayment(ctx context.Context, req PayRequest) (*PayResult, error) {
	svc := native.NativeApiService{Client: g.client}
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(g.appID),
		Mchid:       core.String(g.mchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount:      &native.Amount{Total: core.Int64(req.AmountCents)},
	})
	if err != nil {
		return nil, fmt.Errorf("wechat native prepay failed: %w", err)
	}
	if resp == nil || resp.CodeUrl == nil {
		return nil, fmt.Errorf("wechat native prepay returned empty code_url")
	}

	return &PayResult{QRCode: *resp.CodeUrl}, nil
}

// ParseNotify 验签并解析支付结果通知（平台证书链自动维护）。
func (g *WechatNativeGateway) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	certVisitor := downloader.MgrInstance().GetCertificateVisitor(g.mchID)
	verifier := verifiers.NewSHA256WithRSAVerifier(certVisitor)
	handler := notify.NewNotifyHandler(g.apiV3Key, verifier)

	transaction := new(payments.Transaction)
	if _, err := handler.ParseNotifyRequest(ctx, r, transaction); err != nil {
		return nil, fmt.Errorf("parse wechat notify failed: %w", err)
	}

	outTradeNo := deref(transaction.OutTradeNo)
	if outTradeNo == "" {
		return nil, fmt.Errorf("wechat notify missing out_trade_no")
	}

	var paidCents int64
	if transaction.Amount != nil && transaction.Amount.Total != nil {
		paidCents = *transaction.Amount.Total
	}

	return &NotifyResult{
		OutTradeNo:     outTradeNo,
		GatewayTradeNo: deref(transaction.TransactionId),
		PaidCents:      paidCents,
		Success:        deref(transaction.TradeState) == "SUCCESS",
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
