package service

import (
	"context"
	"fmt"
	"os"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/emptypb"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"

	"go-wind-shop/pkg/crypto"
	"go-wind-shop/pkg/logx"

	"go-wind-shop/app/core/service/internal/data"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
	"go-wind-shop/pkg/payment"
)

// envPaymentCredentialKey 渠道凭据加密密钥的环境变量名。
// 部署方必须自行提供（任意强随机串，SHA-256 派生 AES-256 密钥）；
// 未提供时非 MOCK 渠道的凭据写入会被拒绝（模拟渠道不受影响）。
const envPaymentCredentialKey = "PAYMENT_CREDENTIAL_KEY"

// PaymentChannelConfigService 支付渠道配置服务（admin 管理 + C 端启用列表）。
//
// 合规约束：凭据仅在服务端流转——入口（Create/Update）AES-GCM 加密后落库，
// 出口（Get/List）永远剥离 credentials 明文/密文，只回 credentials_configured
// 布尔；唯一例外是服务间 CreatePaymentIntent 内网下发（core→BFF gRPC）。
type PaymentChannelConfigService struct {
	paymentV1.UnimplementedPaymentChannelConfigServiceServer

	log  *logx.Helper
	repo *data.PaymentChannelConfigRepo
}

func NewPaymentChannelConfigService(
	ctx *bootstrap.Context,
	repo *data.PaymentChannelConfigRepo,
) *PaymentChannelConfigService {
	return &PaymentChannelConfigService{
		log:  logx.New(ctx.NewLoggerHelper("payment-channel-config/service/core-service")),
		repo: repo,
	}
}

func (s *PaymentChannelConfigService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentChannelConfigResponse, error) {
	ret, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, cfg := range ret.GetItems() {
		stripCredentials(cfg)
	}
	return ret, nil
}

func (s *PaymentChannelConfigService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentChannelConfigResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &paymentV1.CountPaymentChannelConfigResponse{
		Count: uint64(count),
	}, nil
}

func (s *PaymentChannelConfigService) Get(ctx context.Context, req *paymentV1.GetPaymentChannelConfigRequest) (*paymentV1.PaymentChannelConfig, error) {
	dto, err := s.repo.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	stripCredentials(dto)
	return dto, nil
}

func (s *PaymentChannelConfigService) ListEnabledChannels(ctx context.Context, req *paymentV1.ListEnabledChannelsRequest) (*paymentV1.ListEnabledChannelsResponse, error) {
	return s.repo.ListEnabled(ctx)
}

// ResolveEnabledChannel 按渠道类型解析一条启用渠道并解密凭据。
// 供 BFF 网关回调验签使用（服务间内网 gRPC，不经任何 REST 出口）。
// 同类型多条启用渠道时取 id 最大（最近配置）的一条。
func (s *PaymentChannelConfigService) ResolveEnabledChannel(ctx context.Context, req *paymentV1.ResolveEnabledChannelRequest) (*paymentV1.PaymentChannelConfig, error) {
	if req == nil || req.GetChannel() == paymentV1.PaymentChannelType_PAYMENT_CHANNEL_UNSPECIFIED {
		return nil, paymentV1.ErrorBadRequest("channel is required")
	}

	channelLabel, ok := paymentV1.PaymentChannelType_name[int32(req.GetChannel())]
	if !ok {
		return nil, paymentV1.ErrorBadRequest("unknown channel type")
	}

	ret, err := s.repo.List(ctx, &paginationV1.PagingRequest{
		NoPaging: trans.Ptr(true),
		FilteringType: &paginationV1.PagingRequest_Query{
			Query: `{"channel":"` + channelLabel + `","enabled":true}`,
		},
		Sorting: []*paginationV1.Sorting{
			{Field: "id", Direction: paginationV1.Sorting_DESC},
		},
	})
	if err != nil {
		return nil, err
	}
	if ret == nil || len(ret.GetItems()) == 0 {
		return nil, paymentV1.ErrorNotFound("no enabled payment channel of type %s", channelLabel)
	}

	cfg := ret.Items[0]

	raw := cfg.GetCredentials()
	if req.GetChannel() != paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK {
		enc, eErr := newCredentialEncryptor()
		if eErr != nil {
			return nil, paymentV1.ErrorInternalServerError("%s", eErr.Error())
		}
		raw, eErr = enc.Decrypt(raw)
		if eErr != nil {
			s.log.Errorf("decrypt credentials for channel type [%s] failed: %v", channelLabel, eErr)
			return nil, paymentV1.ErrorInternalServerError("decrypt credentials failed")
		}
	}
	cfg.Credentials = trans.Ptr(raw)

	return cfg, nil
}

func (s *PaymentChannelConfigService) Create(ctx context.Context, req *paymentV1.CreatePaymentChannelConfigRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.validateAndEncrypt(ctx, req.Data); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentChannelConfigService) Update(ctx context.Context, req *paymentV1.UpdatePaymentChannelConfigRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return nil, paymentV1.ErrorBadRequest("id is required")
	}

	// credentials 仅在本次提交携带非空值时才改写（更新走 update_mask 白名单），
	// 空值/缺省一律保持库中原密文——管理端「不改密钥」的编辑路径天然安全。
	if req.Data.GetCredentials() != "" {
		if err := s.validateAndEncrypt(ctx, req.Data); err != nil {
			return nil, err
		}
		if req.GetUpdateMask() != nil {
			found := false
			for _, p := range req.GetUpdateMask().GetPaths() {
				if p == "credentials" {
					found = true
					break
				}
			}
			if !found {
				req.GetUpdateMask().Paths = append(req.GetUpdateMask().GetPaths(), "credentials")
			}
		}
	} else {
		req.Data.Credentials = nil
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentChannelConfigService) Delete(ctx context.Context, req *paymentV1.DeletePaymentChannelConfigRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetDecryptedCredentials 供服务间调用：读取渠道配置并解密凭据 JSON。
// 仅 CreatePaymentIntent（core→BFF gRPC 内网下发）使用，不经过任何 REST 出口。
func (s *PaymentChannelConfigService) GetDecryptedCredentials(ctx context.Context, channelId uint32) (*paymentV1.PaymentChannelConfig, string, error) {
	cfg, err := s.repo.Get(ctx, &paymentV1.GetPaymentChannelConfigRequest{
		Id: channelId,
	})
	if err != nil {
		return nil, "", err
	}
	if !cfg.GetEnabled() {
		return nil, "", paymentV1.ErrorBadRequest("payment channel [%d] is not enabled", channelId)
	}

	raw := cfg.GetCredentials()
	if cfg.GetChannel() != paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK {
		enc, eErr := newCredentialEncryptor()
		if eErr != nil {
			return nil, "", paymentV1.ErrorInternalServerError("%s", eErr.Error())
		}
		raw, eErr = enc.Decrypt(raw)
		if eErr != nil {
			s.log.Errorf("decrypt credentials for channel [%d] failed: %v", channelId, eErr)
			return nil, "", paymentV1.ErrorInternalServerError("decrypt credentials failed")
		}
	}

	return cfg, raw, nil
}

// validateAndEncrypt 校验渠道参数与凭据结构，并把明文凭据原位替换为密文。
func (s *PaymentChannelConfigService) validateAndEncrypt(ctx context.Context, data *paymentV1.PaymentChannelConfig) error {
	channel := data.GetChannel()
	if data.GetName() == "" {
		return paymentV1.ErrorBadRequest("channel name is required")
	}

	raw := data.GetCredentials()
	switch channel {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK:
		// 模拟渠道无凭据；防误存敏感内容
		data.Credentials = nil
		return nil

	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		if data.GetAppId() == "" || data.GetMchId() == "" {
			return paymentV1.ErrorBadRequest("wechat native channel requires app_id and mch_id")
		}
		if _, pErr := payment.ParseWechatCredentials(raw); pErr != nil {
			return paymentV1.ErrorBadRequest("%v", pErr.Error())
		}

	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		if data.GetAppId() == "" {
			return paymentV1.ErrorBadRequest("alipay page channel requires app_id")
		}
		if _, pErr := payment.ParseAlipayCredentials(raw); pErr != nil {
			return paymentV1.ErrorBadRequest("%v", pErr.Error())
		}

	default:
		return paymentV1.ErrorBadRequest("unsupported payment channel type")
	}

	enc, eErr := newCredentialEncryptor()
	if eErr != nil {
		return paymentV1.ErrorInternalServerError("%s", eErr.Error())
	}
	cipher, eErr := enc.Encrypt(raw)
	if eErr != nil {
		s.log.Errorf("encrypt payment credentials failed: %v", eErr)
		return paymentV1.ErrorInternalServerError("encrypt credentials failed")
	}
	data.Credentials = trans.Ptr(cipher)

	return nil
}

// stripCredentials 读路径出口加固：剥凭据（密文也不出服务），回填配置状态位。
func stripCredentials(cfg *paymentV1.PaymentChannelConfig) {
	if cfg == nil {
		return
	}
	configured := cfg.GetCredentials() != ""
	cfg.Credentials = nil
	cfg.CredentialsConfigured = trans.Ptr(configured)
}

// newCredentialEncryptor 从环境变量构造凭据加密器（SHA-256 派生 AES-256-GCM）。
func newCredentialEncryptor() (*crypto.Encryptor, error) {
	key := os.Getenv(envPaymentCredentialKey)
	if key == "" {
		return nil, fmt.Errorf("payment credential encryption key not configured (set %s)", envPaymentCredentialKey)
	}
	return crypto.NewEncryptor(key)
}
