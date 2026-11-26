package service

import (
	"context"
	"strings"

	"github.com/tx7do/go-utils/captcha"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"go-wind-shop/pkg/logx"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"

	"go-wind-shop/app/admin/service/internal/data"
	"go-wind-shop/pkg/middleware/auth"
	"go-wind-shop/pkg/netutil"
)

// 验证码相关请求头（H5：登录强制验证码，通过 header 传递以避免改动 proto 与三套前端生成代码）。
const (
	headerCaptchaID    = "X-Captcha-Id"
	headerCaptchaValue = "X-Captcha-Value"
)

// CaptchaEnabled 控制登录是否启用验证码校验（渐进式闸门的总开关）。
// 开发/无 Redis 等环境可改为 false 跳过验证码校验，避免登录被验证码闸门阻断。
const CaptchaEnabled = true

// captchaRequiredThreshold 渐进式验证码阈值（对齐上游基座）：同一 ip+用户名
// 的登录失败计数达到该值后，后续登录强制验证码。计数由登录失败自增
// （rateLimiter.CheckAndIncr）维护、登录成功 Reset，滑动窗口 loginFailKeyTTL
// 过期自然清零——阈值取 1 即「错一次下一次就要验证码」，与锁定闸门（阈值 5）
// 之间留出验证码阶段。
const captchaRequiredThreshold = 1

type AuthenticationService struct {
	adminV1.AuthenticationServiceHTTPServer

	log *logx.Helper

	authenticationServiceClient authenticationV1.AuthenticationServiceClient

	captchaClient *captcha.Captcha
	rateLimiter   *data.LoginRateLimiter
}

func NewAuthenticationService(
	ctx *bootstrap.Context,
	authenticationServiceClient authenticationV1.AuthenticationServiceClient,
	captchaClient *captcha.Captcha,
	rateLimiter *data.LoginRateLimiter,
) *AuthenticationService {
	return &AuthenticationService{
		log:                         logx.NewFromLogger(ctx.GetLogger().With("module", "user/service/admin-service")),
		authenticationServiceClient: authenticationServiceClient,
		captchaClient:               captchaClient,
		rateLimiter:                 rateLimiter,
	}
}

// verifyLoginCaptcha 渐进式校验登录请求携带的验证码。
// 验证码 id/value 通过 HTTP Header（X-Captcha-Id / X-Captcha-Value）传递。
// 仅当同一 ip+用户名的登录失败计数达到 captchaRequiredThreshold 时强制——
// 计数由登录失败自增维护、登录成功 Reset，未达阈值时不校验也不消费验证码。
// captchaClient.Verify 已是 verify-and-delete 单次有效语义。
// 注意：refresh_token / client_credentials 等非密码授权不走此校验（仅 password 授权调用）。
func (s *AuthenticationService) verifyLoginCaptcha(ctx context.Context, clientIP, username string) error {
	if !CaptchaEnabled {
		// 验证码开关关闭，跳过校验
		return nil
	}
	if s.captchaClient == nil {
		// captcha 未配置时 fail-open（仅记录告警），避免影响登录基本功能
		return nil
	}
	// 渐进式闸门：失败计数读的是限流器的只读快照；Redis 不可用时返回 0，
	// 与限流闸门的 fail-open 语义一致（限流是防御增强，不应故障时堵死登录）。
	if s.rateLimiter == nil || s.rateLimiter.FailureAttempts(ctx, clientIP, username) < captchaRequiredThreshold {
		return nil
	}
	header := netutil.HeaderFromContext(ctx)
	if header == nil {
		return adminV1.ErrorCaptchaRequired("captcha required")
	}
	captchaID := strings.TrimSpace(header.Get(headerCaptchaID))
	captchaValue := strings.TrimSpace(header.Get(headerCaptchaValue))
	if captchaID == "" || captchaValue == "" {
		return adminV1.ErrorCaptchaRequired("captcha required")
	}
	ok, err := s.captchaClient.Verify(ctx, captchaID, captchaValue)
	if err != nil {
		s.log.Errorf("verify captcha failed: %s", err.Error())
		// P3 加固：校验异常也销毁该验证码（图穷匕见——不可信状态下让单码
		// 不再可用，缩小在线试猜窗口）。
		_ = s.captchaClient.Delete(ctx, captchaID)
		return adminV1.ErrorCaptchaInvalid("invalid captcha")
	}
	if !ok {
		// P3 加固：校验失败（含不匹配/过期不存在）销毁验证码——失败后单码
		// 仍可用会在 10 分钟 TTL 内形成在线试猜窗口。
		_ = s.captchaClient.Delete(ctx, captchaID)
		return adminV1.ErrorCaptchaInvalid("invalid captcha")
	}
	return nil
}

func (s *AuthenticationService) GenerateCaptcha(ctx context.Context, _ *emptypb.Empty) (*authenticationV1.GenerateCaptchaResponse, error) {
	captchaId, captchaImage, answer, err := s.captchaClient.Generate()
	if err != nil {
		s.log.Errorf("generate captcha failed: %s", err.Error())
		return nil, authenticationV1.ErrorInternalServerError("generate captcha failed")
	}

	// Generate() 只生成验证码但不落盘，必须手动 Save 到 Redis，否则 Verify 时查不到。
	if err = s.captchaClient.Save(ctx, captchaId, answer); err != nil {
		s.log.Errorf("save captcha failed: %s", err.Error())
		return nil, authenticationV1.ErrorInternalServerError("save captcha failed")
	}

	return &authenticationV1.GenerateCaptchaResponse{
		CaptchaId:   captchaId,
		ImageBase64: captchaImage,
	}, nil
}

func (s *AuthenticationService) VerifyCaptcha(ctx context.Context, req *authenticationV1.VerifyCaptchaRequest) (*authenticationV1.VerifyCaptchaResponse, error) {
	ok, err := s.captchaClient.Verify(ctx, req.GetCaptchaId(), req.GetUserInput())
	if err != nil {
		s.log.Errorf("verify captcha failed: %s", err.Error())
		return nil, authenticationV1.ErrorInternalServerError("verify captcha failed")
	}

	return &authenticationV1.VerifyCaptchaResponse{
		Valid: ok,
	}, nil
}

// Login 登录
func (s *AuthenticationService) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid request")
	}

	req.ClientType = trans.Ptr(authenticationV1.ClientType_admin)

	if req.GetGrantType() == authenticationV1.GrantType_refresh_token {
		operator, err := auth.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		req.Jti = operator.Jti
		req.UserId = trans.Ptr(operator.GetUserId())
	} else if req.GetGrantType() == authenticationV1.GrantType_password {
		// ===== H5 闸门：渐进式验证码（仅密码授权；通过 HTTP Header 传递，避免改动 proto/前端生成代码）=====
		// 同一 ip+用户名失败计数达阈值后强制：缺失返回 CAPTCHA_REQUIRED
		//（登录页按此 reason 点亮验证码字段）、校验不通过返回 CAPTCHA_INVALID。
		clientIP := netutil.ClientIPFromContext(ctx)
		username := strings.TrimSpace(req.GetUsername())
		if err := s.verifyLoginCaptcha(ctx, clientIP, username); err != nil {
			return nil, err
		}

		resp, err := s.authenticationServiceClient.Login(ctx, req)
		if err != nil {
			// H5：凭证类失败（用户不存在/冻结/密码错——core 侧已统一归一为
			// INVALID_PASSWORD）自增失败计数，驱动渐进式验证码与锁定闸门。
			// 租户无效、策略封锁等其余失败不计——它们不是密码试猜信号。
			if authenticationV1.IsInvalidPassword(err) {
				if _, _, _, cerr := s.rateLimiter.CheckAndIncr(ctx, clientIP, username); cerr != nil {
					s.log.Errorf("login rate limiter incr failed: %s", cerr.Error())
				}
			}
			return nil, err
		}
		// H5：登录成功清零失败计数。MFA 分支（返回 operation_id、未签发 token）
		// 不清零——认证尚未完成，由 MFA 挑战验证通过后的 Reset 收口。
		if resp.GetMfaOperationId() == "" {
			s.rateLimiter.Reset(ctx, clientIP, username)
		}
		return resp, nil
	}

	return s.authenticationServiceClient.Login(ctx, req)
}

// Logout 登出
func (s *AuthenticationService) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.authenticationServiceClient.Logout(ctx, &authenticationV1.LogoutRequest{
		ClientType: authenticationV1.ClientType_admin,
		UserId:     operator.GetUserId(),
	})
}

// RefreshToken 刷新令牌。
// 本端点在鉴权白名单内：access token 已过期恰恰是触发刷新的前提，不能再要求
// 有效 AT（也不能从 operator 上下文取身份）。user_id/jti/refresh_token 三个
// 身份键全部由前端从持久化会话带回。RT 是不透明随机串，真正鉴权由 core 的
// VerifyRefreshToken 做值比对；user_id/jti 只是查找键，伪造它们拿不到有效 RT。
func (s *AuthenticationService) RefreshToken(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid request")
	}

	req.ClientType = trans.Ptr(authenticationV1.ClientType_admin)

	if req.GetUserId() == 0 || req.GetJti() == "" || req.GetRefreshToken() == "" {
		return nil, authenticationV1.ErrorBadRequest("user_id, jti and refresh_token are required")
	}

	return s.authenticationServiceClient.RefreshToken(ctx, req)
}

func (s *AuthenticationService) WhoAmI(ctx context.Context, _ *emptypb.Empty) (*authenticationV1.WhoAmIResponse, error) {
	// 获取操作人信息
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &authenticationV1.WhoAmIResponse{
		UserId:   operator.GetUserId(),
		Username: operator.GetUsername(),
	}, nil
}
