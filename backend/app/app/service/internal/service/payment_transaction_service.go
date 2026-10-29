package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go-wind-shop/pkg/logx"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/emptypb"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"go-wind-shop/pkg/middleware/auth"
	"go-wind-shop/pkg/payment/gateway"
)

// envPaymentNotifyBaseURL 网关异步回调的公网基础地址环境变量。
// 部署方必须配置为 BFF 服务对支付渠道网关可达的公网地址
// （如 https://shop.example.com）；缺省回退本地开发地址。
const envPaymentNotifyBaseURL = "PAYMENT_NOTIFY_BASE_URL"

type PaymentTransactionService struct {
	appV1.PaymentTransactionServiceHTTPServer

	log *logx.Helper

	paymentTransactionServiceClient paymentV1.PaymentTransactionServiceClient
	channelConfigServiceClient      paymentV1.PaymentChannelConfigServiceClient
}

func NewPaymentTransactionService(
	ctx *bootstrap.Context,
	paymentTransactionServiceClient paymentV1.PaymentTransactionServiceClient,
	channelConfigServiceClient paymentV1.PaymentChannelConfigServiceClient,
) *PaymentTransactionService {
	return &PaymentTransactionService{
		log:                             logx.New(ctx.NewLoggerHelper("payment-transaction/service/app-service")),
		paymentTransactionServiceClient: paymentTransactionServiceClient,
		channelConfigServiceClient:      channelConfigServiceClient,
	}
}

func (s *PaymentTransactionService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentTransactionResponse, error) {
	return s.paymentTransactionServiceClient.List(ctx, req)
}

func (s *PaymentTransactionService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentTransactionResponse, error) {
	return s.paymentTransactionServiceClient.Count(ctx, req)
}

func (s *PaymentTransactionService) Get(ctx context.Context, req *paymentV1.GetPaymentTransactionRequest) (*paymentV1.PaymentTransaction, error) {
	return s.paymentTransactionServiceClient.Get(ctx, req)
}

func (s *PaymentTransactionService) Create(ctx context.Context, req *paymentV1.CreatePaymentTransactionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.paymentTransactionServiceClient.Create(ctx, req)
}

// CreatePaymentIntent 创建支付意图（C 端入口）：
// 登录态注入 user_id → core 落 PENDING 流水（订单归属/待支付态/幂等校验在 core）
// → BFF 持解密凭据调渠道网关下单 → 返回二维码/跳转链接。
// 金额取服务端订单 total_amount，不信任客户端传值。
func (s *PaymentTransactionService) CreatePaymentIntent(ctx context.Context, req *paymentV1.CreatePaymentIntentRequest) (*paymentV1.CreatePaymentIntentAppResponse, error) {
	if req == nil || req.GetOrderId() == 0 || req.GetChannelId() == 0 {
		return nil, appV1.ErrorBadRequest("orderId and channelId are required")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.UserId = operator.UserId

	resp, err := s.paymentTransactionServiceClient.CreatePaymentIntent(ctx, req)
	if err != nil {
		return nil, err
	}

	channel := resp.GetChannel()
	gw, err := gateway.New(channel.GetChannel(), channel.GetAppId(), channel.GetMchId(), channel.GetSandbox(), channel.GetCredentials())
	if err != nil {
		s.log.Errorf("build payment gateway for channel [%d] failed: %v", channel.GetId(), err)
		return nil, appV1.ErrorInternalServerError("payment gateway unavailable")
	}

	pr, err := gw.CreatePayment(ctx, gateway.PayRequest{
		OutTradeNo:  strconv.FormatUint(uint64(resp.GetTransaction().GetId()), 10),
		AmountCents: resp.GetTransaction().GetAmount(),
		Subject:     resp.GetSubject(),
		NotifyURL:   s.notifyURL(channel.GetChannel()),
		ClientIp:    clientIPFromContext(ctx),
	})
	if err != nil {
		s.log.Errorf("gateway create payment for tx [%d] failed: %v", resp.GetTransaction().GetId(), err)
		return nil, appV1.ErrorInternalServerError("gateway create payment failed")
	}

	out := &paymentV1.CreatePaymentIntentAppResponse{
		TransactionId: resp.GetTransaction().GetId(),
		Status:        resp.GetTransaction().GetStatus(),
		Amount:        resp.GetTransaction().GetAmount(),
		Currency:      resp.GetTransaction().GetCurrency(),
	}
	switch channel.GetChannel() {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK:
		out.PayType = paymentV1.PayType_PAY_TYPE_MOCK
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		out.PayType = paymentV1.PayType_PAY_TYPE_QR
		out.QrCode = pr.QRCode
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_H5:
		out.PayType = paymentV1.PayType_PAY_TYPE_REDIRECT
		out.PayUrl = pr.RedirectURL
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		out.PayType = paymentV1.PayType_PAY_TYPE_REDIRECT
		out.PayUrl = pr.RedirectURL
	default:
		return nil, appV1.ErrorBadRequest("unsupported payment channel")
	}

	return out, nil
}

// MockConfirmPayment 模拟渠道确认支付（开发/演示）。
// 仅对 MOCK 渠道的流水放行——真实渠道流水绝无此入口（否则等于免费购买后门）。
func (s *PaymentTransactionService) MockConfirmPayment(ctx context.Context, req *paymentV1.MockConfirmPaymentRequest) (*paymentV1.MockConfirmPaymentResponse, error) {
	if req == nil || req.GetTransactionId() == 0 {
		return nil, appV1.ErrorBadRequest("transactionId is required")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get 走 core UserPrivacy：用户只能查到自己的流水
	tx, err := s.paymentTransactionServiceClient.Get(ctx, &paymentV1.GetPaymentTransactionRequest{
		QueryBy: &paymentV1.GetPaymentTransactionRequest_Id{Id: req.GetTransactionId()},
	})
	if err != nil || tx == nil {
		return nil, appV1.ErrorBadRequest("payment transaction not found")
	}
	if tx.GetPaymentMethod() != paymentV1.PaymentMethod_MOCK {
		return nil, appV1.ErrorForbidden("mock confirm is only allowed for mock channel transactions")
	}

	resp, err := s.paymentTransactionServiceClient.ConfirmPayment(ctx, &paymentV1.ConfirmPaymentRequest{
		TransactionId:  req.GetTransactionId(),
		GatewayTradeNo: fmt.Sprintf("MOCK-%d", operator.UserId),
		PaidAmount:     tx.GetAmount(),
	})
	if err != nil {
		return nil, err
	}

	return &paymentV1.MockConfirmPaymentResponse{Status: resp.GetStatus()}, nil
}

// HandleNotify 处理渠道网关异步回调（裸路由，无登录态，凭验签保护）。
// 应答口径：验签失败/确认失败 → 让网关重试；其余（含非成功态通知、幂等重放）
// → 应答成功让网关停止重试。
func (s *PaymentTransactionService) HandleNotify(ctx context.Context, channelSlug string, r *http.Request) (int, string) {
	channelType, ok := channelTypeBySlug(channelSlug)
	if !ok {
		return http.StatusNotFound, "unknown channel"
	}

	cfg, err := s.channelConfigServiceClient.ResolveEnabledChannel(ctx, &paymentV1.ResolveEnabledChannelRequest{
		Channel: channelType,
	})
	if err != nil {
		s.log.Errorf("resolve enabled channel [%s] for notify failed: %v", channelSlug, err)
		return http.StatusInternalServerError, "resolve channel failed"
	}

	gw, err := gateway.New(cfg.GetChannel(), cfg.GetAppId(), cfg.GetMchId(), cfg.GetSandbox(), cfg.GetCredentials())
	if err != nil {
		s.log.Errorf("build gateway for notify channel [%s] failed: %v", channelSlug, err)
		return http.StatusInternalServerError, "gateway unavailable"
	}

	nr, err := gw.ParseNotify(ctx, r)
	if err != nil {
		// 验签失败：拒绝（4xx/5xx 让网关重试；伪造请求被挡在验签外）
		s.log.Warnf("notify verification failed for channel [%s]: %v", channelSlug, err)
		return http.StatusBadRequest, "verification failed"
	}

	if !nr.Success {
		// 非支付成功态通知（如交易关闭）：应答成功即可，无需业务处理
		s.log.Infof("notify for tx [%s] on channel [%s] is not success, ack", nr.OutTradeNo, channelSlug)
		return s.ackNotify(channelType)
	}

	txId, err := strconv.ParseUint(nr.OutTradeNo, 10, 32)
	if err != nil {
		s.log.Errorf("notify out_trade_no [%s] is not a transaction id", nr.OutTradeNo)
		return s.ackNotify(channelType)
	}

	if _, err = s.paymentTransactionServiceClient.ConfirmPayment(ctx, &paymentV1.ConfirmPaymentRequest{
		TransactionId:  uint32(txId),
		GatewayTradeNo: nr.GatewayTradeNo,
		PaidAmount:     nr.PaidCents,
	}); err != nil {
		// 金额不符/状态冲突：应答失败让网关重试，同时保留 PENDING 供人工排查
		s.log.Errorf("confirm payment from notify failed: tx=%s channel=%s err=%v", nr.OutTradeNo, channelSlug, err)
		return s.failNotify(channelType, "confirm failed")
	}

	s.log.Infof("payment notify confirmed: tx=%d channel=%s gateway_no=%s amount=%d", txId, channelSlug, nr.GatewayTradeNo, nr.PaidCents)
	return s.ackNotify(channelType)
}

// ackNotify 渠道网关的成功应答格式（支付宝 "success" 明文；微信 JSON SUCCESS）。
func (s *PaymentTransactionService) ackNotify(channel paymentV1.PaymentChannelType) (int, string) {
	switch channel {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		return http.StatusOK, "success"
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		return http.StatusOK, `{"code":"SUCCESS","message":"成功"}`
	default:
		return http.StatusOK, "ok"
	}
}

// failNotify 渠道网关的失败应答格式（让网关按退避策略重试）。
func (s *PaymentTransactionService) failNotify(channel paymentV1.PaymentChannelType, reason string) (int, string) {
	switch channel {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		return http.StatusInternalServerError, "failure"
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		return http.StatusInternalServerError, `{"code":"FAIL","message":"` + reason + `"}`
	default:
		return http.StatusInternalServerError, reason
	}
}

// clientIPFromContext 从 kratos transport 取买家客户端 IP（微信 H5 支付必传）。
// 取 X-Forwarded-For / X-Real-IP 头，缺省回环地址（本机/内网部署形态）。
func clientIPFromContext(ctx context.Context) string {
	if req, ok := khttp.RequestFromServerContext(ctx); ok && req != nil {
		if v := req.Header.Get("X-Forwarded-For"); v != "" {
			if idx := strings.Index(v, ","); idx > 0 {
				v = v[:idx]
			}
			return strings.TrimSpace(v)
		}
		if v := req.Header.Get("X-Real-Ip"); v != "" {
			return v
		}
		if req.RemoteAddr != "" {
			if idx := strings.LastIndex(req.RemoteAddr, ":"); idx > 0 {
				return strings.Trim(req.RemoteAddr[:idx], "[]")
			}
		}
	}
	return "127.0.0.1"
}

func (s *PaymentTransactionService) notifyURL(channel paymentV1.PaymentChannelType) string {
	base := strings.TrimRight(os.Getenv(envPaymentNotifyBaseURL), "/")
	if base == "" {
		base = "http://localhost:10700"
	}
	return base + "/app/v1/payments/notify/" + channelSlug(channel)
}

func channelSlug(channel paymentV1.PaymentChannelType) string {
	switch channel {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		return "alipay"
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		return "wechat"
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_H5:
		return "wechat-h5"
	default:
		return "mock"
	}
}

func channelTypeBySlug(slug string) (paymentV1.PaymentChannelType, bool) {
	switch slug {
	case "alipay":
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE, true
	case "wechat":
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE, true
	case "wechat-h5":
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_H5, true
	default:
		// mock 渠道无服务端回调：明确拒绝而不是静默 404
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_UNSPECIFIED, false
	}
}
