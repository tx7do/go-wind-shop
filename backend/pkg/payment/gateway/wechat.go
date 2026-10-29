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
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"go-wind-shop/pkg/payment"
)

// wechatPayKind 微信支付下单形式：Native 扫码 / H5 移动端跳转。
// 二者共用商户凭据、回调验签、关单、退款与查单，仅下单 API 不同。
type wechatPayKind int

const (
	wechatPayKindNative wechatPayKind = iota
	wechatPayKindH5
)

// WechatGateway 微信支付直连商户（APIv3）。
//
// 注意：微信支付 V3 无沙箱环境——sandbox=true 时拒绝创建（防止误把生产商户号
// 当沙箱刷）；验收用真实商户号小额支付。
type WechatGateway struct {
	base wechatBase
	kind wechatPayKind
}

// wechatBase 持有商户凭据与 API 客户端，承载 Native/H5 共用的回调/关单/退款/查单。
type wechatBase struct {
	appID    string
	mchID    string
	apiV3Key string
	client   *core.Client
}

func NewWechatGateway(appID, mchID string, sandbox bool, credentialsJSON string, kind wechatPayKind) (*WechatGateway, error) {
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

	return &WechatGateway{
		base: wechatBase{appID: appID, mchID: mchID, apiV3Key: cred.APIv3Key, client: client},
		kind: kind,
	}, nil
}

func (g *WechatGateway) CreatePayment(ctx context.Context, req PayRequest) (*PayResult, error) {
	if g.kind == wechatPayKindH5 {
		return g.createPaymentH5(ctx, req)
	}
	return g.createPaymentNative(ctx, req)
}

// createPaymentNative Native 扫码支付：返回 code_url 供前端渲染二维码。
func (g *WechatGateway) createPaymentNative(ctx context.Context, req PayRequest) (*PayResult, error) {
	svc := native.NativeApiService{Client: g.base.client}
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(g.base.appID),
		Mchid:       core.String(g.base.mchID),
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

// createPaymentH5 H5 支付（移动端浏览器）：返回 h5_url 供整页跳转。
// SceneInfo.PayerClientIp 为渠道必填字段。
func (g *WechatGateway) createPaymentH5(ctx context.Context, req PayRequest) (*PayResult, error) {
	clientIp := req.ClientIp
	if clientIp == "" {
		clientIp = "127.0.0.1"
	}
	svc := h5.H5ApiService{Client: g.base.client}
	resp, _, err := svc.Prepay(ctx, h5.PrepayRequest{
		Appid:       core.String(g.base.appID),
		Mchid:       core.String(g.base.mchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount:      &h5.Amount{Total: core.Int64(req.AmountCents)},
		SceneInfo: &h5.SceneInfo{
			PayerClientIp: core.String(clientIp),
			H5Info: &h5.H5Info{
				Type:    core.String("Wap"),
				AppUrl:  core.String(req.NotifyURL),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("wechat h5 prepay failed: %w", err)
	}
	if resp == nil || resp.H5Url == nil {
		return nil, fmt.Errorf("wechat h5 prepay returned empty h5_url")
	}

	return &PayResult{RedirectURL: *resp.H5Url}, nil
}

// ParseNotify 验签并解析支付结果通知（平台证书链自动维护）。
func (g *WechatGateway) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	certVisitor := downloader.MgrInstance().GetCertificateVisitor(g.base.mchID)
	verifier := verifiers.NewSHA256WithRSAVerifier(certVisitor)
	handler := notify.NewNotifyHandler(g.base.apiV3Key, verifier)

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

// CloseOrder 关闭未支付的微信预支付单（订单取消/超时后让旧二维码失效）。
// 渠道侧已支付时关单失败——该情形由支付回调确认路径兜底，此处错误仅记录。
func (g *WechatGateway) CloseOrder(ctx context.Context, outTradeNo string) error {
	svc := native.NativeApiService{Client: g.base.client}
	_, err := svc.CloseOrder(ctx, native.CloseOrderRequest{
		OutTradeNo: core.String(outTradeNo),
		Mchid:      core.String(g.base.mchID),
	})
	if err != nil {
		return fmt.Errorf("wechat close order [%s] failed: %w", outTradeNo, err)
	}
	return nil
}

// Refund 发起微信退款。渠道受理（含 PROCESSING 处理中）即成功——到账由
// 微信侧异步保证；拒绝时返回错误，调用方不得推进退款状态。
func (g *WechatGateway) Refund(ctx context.Context, req RefundRequest) error {
	svc := refunddomestic.RefundsApiService{Client: g.base.client}
	_, _, err := svc.Create(ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(req.OutTradeNo),
		OutRefundNo: core.String(req.OutRequestNo),
		Reason:      core.String(req.Reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(req.RefundCents),
			Total:    core.Int64(req.TotalCents),
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		return fmt.Errorf("wechat refund [%s/%s] failed: %w", req.OutTradeNo, req.OutRequestNo, err)
	}
	return nil
}

// QueryOrder 主动查询渠道侧订单状态（对账兜底）。
func (g *WechatGateway) QueryOrder(ctx context.Context, outTradeNo string) (*OrderQueryResult, error) {
	svc := native.NativeApiService{Client: g.base.client}
	tx, _, err := svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(outTradeNo),
		Mchid:      core.String(g.base.mchID),
	})
	if err != nil {
		// 订单不存在（从未支付过）是常态而非错误，交由调用方按未支付处理
		return &OrderQueryResult{TradeState: "NOT_FOUND"}, nil
	}

	if tx == nil {
		return &OrderQueryResult{TradeState: "NOT_FOUND"}, nil
	}
	var paidCents int64
	if tx.Amount != nil && tx.Amount.Total != nil {
		paidCents = *tx.Amount.Total
	}
	state := deref(tx.TradeState)
	return &OrderQueryResult{
		GatewayTradeNo: deref(tx.TransactionId),
		TradeState:     state,
		PaidCents:      paidCents,
		Success:        state == "SUCCESS",
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// QueryRefund 查询微信退款单状态（按商户退款单号）。
func (g *WechatGateway) QueryRefund(ctx context.Context, outRefundNo string) (*RefundQueryResult, error) {
	svc := refunddomestic.RefundsApiService{Client: g.base.client}
	resp, _, err := svc.QueryByOutRefundNo(ctx, refunddomestic.QueryByOutRefundNoRequest{
		OutRefundNo: core.String(outRefundNo),
	})
	if err != nil {
		return nil, fmt.Errorf("wechat query refund [%s] failed: %w", outRefundNo, err)
	}
	if resp == nil || resp.Status == nil {
		return &RefundQueryResult{Status: "NOT_FOUND"}, nil
	}
	return &RefundQueryResult{Status: string(*resp.Status)}, nil
}
