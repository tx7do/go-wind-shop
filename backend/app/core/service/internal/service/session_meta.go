package service

import (
	"context"
	"time"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"
	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"
	"go-wind-shop/pkg/netutil"

	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	"go-wind-shop/app/core/service/internal/data"
)

// recordSessionMeta 令牌对签发成功后记录会话元数据（在线会话列表展示用）。
// 登录时间取当前时间。最佳努力：记录失败仅记日志，不阻断登录主流程。
func recordSessionMeta(
	ctx context.Context,
	log *bLogger.Helper,
	authenticator *data.Authenticator,
	clientType authenticationV1.ClientType,
	tokenPayload *authenticationV1.UserTokenPayload,
) {
	recordSessionMetaAt(ctx, log, authenticator, clientType, tokenPayload, 0)
}

// recordSessionMetaAt 同 recordSessionMeta，但允许显式指定登录时间（unix 秒）。
// 刷新轮换时传入旧会话的登录时间，使在线列表展示的"登录时间"不随刷新重置。
func recordSessionMetaAt(
	ctx context.Context,
	log *bLogger.Helper,
	authenticator *data.Authenticator,
	clientType authenticationV1.ClientType,
	tokenPayload *authenticationV1.UserTokenPayload,
	loginAt int64,
) {
	if authenticator == nil || tokenPayload == nil || tokenPayload.GetJti() == "" {
		return
	}
	if loginAt <= 0 {
		loginAt = time.Now().Unix()
	}

	// User-Agent 经 HTTP header 透传；非 HTTP 传输（如任务/SSE 内部调用）取不到时留空
	ua := ""
	if header := netutil.HeaderFromContext(ctx); header != nil {
		ua = header.Get("User-Agent")
	}

	meta := &data.SessionMeta{
		Username:  tokenPayload.GetUsername(),
		TenantId:  tokenPayload.GetTenantId(),
		Ip:        clientIPFromGRPCContext(ctx),
		UserAgent: ua,
		DeviceId:  tokenPayload.GetDeviceId(),
		LoginAt:   loginAt,
	}

authenticator.SaveSessionMeta(
		ctx,
		clientType,
		tokenPayload.GetUserId(),
		tokenPayload.GetJti(),
		meta,
		authenticator.GetRefreshTokenExpires(clientType),
	)
}
func recordUserLastLogin(ctx context.Context, log *bLogger.Helper, userRepo data.UserRepo, userID uint32, clientIP string) {
	if userID == 0 || userRepo == nil {
		return
	}
	mask, err := fieldmaskpb.New(&identityV1.User{}, "last_login_at", "last_login_ip")
	if err != nil {
		log.Errorf(ctx, "build last-login fieldmask failed: %s", err.Error())
		return
	}
	uerr := userRepo.Update(ctx, &identityV1.UpdateUserRequest{
		Id: userID,
		Data: &identityV1.User{
			LastLoginAt: timeutil.TimeToTimestamppb(trans.Ptr(time.Now())),
			LastLoginIp: trans.Ptr(clientIP),
		},
		UpdateMask: mask,
	})
	if uerr != nil {
		log.Errorf(ctx, "record last login for user [%d] failed: %s", userID, uerr.Error())
	}
}

