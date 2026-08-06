package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/redis/go-redis/v9"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"

	"go-wind-shop/pkg/email"
	"go-wind-shop/pkg/middleware/auth"
	"go-wind-shop/pkg/resetcode"
)

type AuthenticationService struct {
	appV1.AuthenticationServiceHTTPServer

	authenticationServiceClient authenticationV1.AuthenticationServiceClient
	userServiceClient           identityV1.UserServiceClient
	userCredentialServiceClient authenticationV1.UserCredentialServiceClient

	codeStore *resetcode.Store
	emailer   *email.Sender

	log *log.Helper
}

func NewAuthenticationService(
	ctx *bootstrap.Context,
	authenticationServiceClient authenticationV1.AuthenticationServiceClient,
	userServiceClient identityV1.UserServiceClient,
	userCredentialServiceClient authenticationV1.UserCredentialServiceClient,
	rdb *redis.Client,
) *AuthenticationService {
	helper := ctx.NewLoggerHelper("authn/service/app-service")
	return &AuthenticationService{
		log:                         helper,
		authenticationServiceClient: authenticationServiceClient,
		userServiceClient:           userServiceClient,
		userCredentialServiceClient: userCredentialServiceClient,
		codeStore:                   resetcode.New(rdb),
		// SMTP 配置从环境变量读取；未配置时 Sender 自动降级为日志输出验证码。
		emailer: email.NewSender(email.LoadConfig(), helper),
	}
}

// Login 登陆
func (s *AuthenticationService) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	if req == nil {
		return nil, authenticationV1.ErrorBadRequest("invalid request")
	}

	req.ClientType = trans.Ptr(authenticationV1.ClientType_app)

	if req.GetGrantType() == authenticationV1.GrantType_refresh_token {
		operator, err := auth.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		req.Jti = operator.Jti
		req.UserId = trans.Ptr(operator.GetUserId())
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
		ClientType: authenticationV1.ClientType_app,
		UserId:     operator.GetUserId(),
	})
}

// SendResetCode 生成 6 位验证码存 Redis 并发邮件。
// 安全要点：无论邮箱是否存在，都返回脱敏预览（防枚举探测）；仅在邮箱存在时才真正发送。
// 未配置 SMTP 时降级为日志输出验证码，便于开发环境跑通流程。
func (s *AuthenticationService) SendResetCode(ctx context.Context, req *appV1.SendResetCodeRequest) (*appV1.SendResetCodeResponse, error) {
	if req == nil {
		return nil, appV1.ErrorBadRequest("invalid request")
	}
	emailAddr := strings.TrimSpace(req.GetEmail())
	if emailAddr == "" {
		return nil, appV1.ErrorBadRequest("email is required")
	}

	preview := email.MaskEmail(emailAddr)
	resp := &appV1.SendResetCodeResponse{
		EmailPreview: preview,
		ExpiresIn:    uint32(s.codeStore.TTL() / time.Second),
	}

	// 先校验该邮箱是否对应用户，不存在则静默返回脱敏预览（防邮箱枚举）。
	username, ok := s.lookupUsernameByEmail(ctx, emailAddr)
	if !ok {
		s.log.Infof("send reset code skipped: no user bound to email %s", preview)
		return resp, nil
	}

	code, err := s.codeStore.Generate(ctx, emailAddr)
	if err != nil {
		s.log.Errorf("generate reset code failed: %s", err.Error())
		return nil, appV1.ErrorInternalServerError("send reset code failed")
	}

	body := strings.NewReplacer(
		"{code}", code,
		"{minutes}", "10",
	).Replace(resetCodeEmailTemplate)

	if err := s.emailer.Send(emailAddr, resetCodeEmailSubject, body); err != nil {
		s.log.Errorf("send reset email failed: %s", err.Error())
		return nil, appV1.ErrorInternalServerError("send reset code failed")
	}

	// 降级模式下日志已含验证码；真实模式下记录用户名便于排查（不含验证码）。
	s.log.Infof("reset code sent to %s (user=%s)", preview, username)

	return resp, nil
}

// ResetPassword 校验验证码后重置密码。
// 流程：校验 Redis 验证码（一次性）→ email 反查 username → 调核心 ResetCredential
// 以 USERNAME 凭证重置（注册时仅创建 USERNAME 凭证）。
func (s *AuthenticationService) ResetPassword(ctx context.Context, req *appV1.ResetPasswordRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, appV1.ErrorBadRequest("invalid request")
	}
	emailAddr := strings.TrimSpace(req.GetEmail())
	code := strings.TrimSpace(req.GetCode())
	newPwd := req.GetNewPassword()
	if emailAddr == "" || code == "" || newPwd == "" {
		return nil, appV1.ErrorBadRequest("email, code and new password are required")
	}

	// 校验验证码（正确则删除，错误则累加尝试次数；超限则作废验证码）。
	ok, err := s.codeStore.Verify(ctx, emailAddr, code)
	if err != nil {
		if errors.Is(err, resetcode.ErrTooManyAttempts) {
			// 尝试次数耗尽，验证码已作废，要求重新获取。
			return nil, appV1.ErrorTooManyRequests("too many failed attempts, please request a new code")
		}
		s.log.Errorf("verify reset code failed: %s", err.Error())
		return nil, appV1.ErrorInternalServerError("reset password failed")
	}
	if !ok {
		return nil, appV1.ErrorBadRequest("invalid or expired verification code")
	}

	// 反查用户名。验证码已消耗，理论上邮箱存在（发码时校验过），此处防御性再查。
	username, found := s.lookupUsernameByEmail(ctx, emailAddr)
	if !found {
		return nil, appV1.ErrorNotFound("user not found")
	}

	// 调核心重置 USERNAME 凭证。新密码的 needDecrypt 与登录/注册约定一致。
	if _, err := s.userCredentialServiceClient.ResetCredential(ctx, &authenticationV1.ResetCredentialRequest{
		IdentityType:  authenticationV1.UserCredential_USERNAME,
		Identifier:    username,
		NewCredential: newPwd,
		NeedDecrypt:   req.GetNeedDecrypt(),
	}); err != nil {
		// 凭证不存在 / 校验失败等，透传核心层错误。
		return nil, err
	}

	s.log.Infof("password reset successfully for user %s", username)
	return &emptypb.Empty{}, nil
}

// lookupUsernameByEmail 按邮箱查询用户，返回其用户名。
// user 表的 email 列存在，通用 DSL 走 raw column 过滤可命中。
// 找不到或异常时返回 ("", false)。
func (s *AuthenticationService) lookupUsernameByEmail(ctx context.Context, emailAddr string) (string, bool) {
	resp, err := s.userServiceClient.List(ctx, listUsersByEmailRequest(emailAddr))
	if err != nil {
		s.log.Errorf("lookup user by email failed: %s", err.Error())
		return "", false
	}
	if resp == nil || len(resp.GetItems()) == 0 {
		return "", false
	}
	// 取首个匹配。email 在用户维度应唯一。
	u := resp.GetItems()[0]
	if u == nil || u.GetUsername() == "" {
		return "", false
	}
	return u.GetUsername(), true
}

// resetCodeEmailTemplate 验证码邮件正文模板。
var resetCodeEmailSubject = "密码重置验证码 / Password Reset Code"

const resetCodeEmailTemplate = `你正在重置 GoWind Shop 账户密码。

验证码：{code}
有效期：{minutes} 分钟

如非本人操作，请忽略此邮件，你的密码不会被更改。

—————

You are resetting your GoWind Shop account password.

Verification code: {code}
Expires in: {minutes} minutes

If this was not you, please ignore this email — your password will not be changed.
`

// listUsersByEmailRequest 构造按 email 过滤用户的分页请求。
// user 表有 email 列，通用 DSL 走 raw column 即可命中；仅需 1 条用于取 username。
//
// 租户范围：app 前台登录不传 tenantCode（平台/单租户模式），后端 Login 据此将
// 凭证查询限定在 tenantId=0。找回密码必须与之一致，故 query 固定带 tenantId=0，
// 避免 (tenant_id,email) 复合唯一索引下跨租户命中同名 email 的其它用户。
func listUsersByEmailRequest(emailAddr string) *paginationV1.PagingRequest {
	queryJSON := fmt.Sprintf(`{"email":%q,"tenantId":0}`, emailAddr)
	return &paginationV1.PagingRequest{
		Page:          trans.Ptr(uint32(1)),
		PageSize:      trans.Ptr(uint32(1)),
		FilteringType: &paginationV1.PagingRequest_Query{Query: queryJSON},
	}
}
