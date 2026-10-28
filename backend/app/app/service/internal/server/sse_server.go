package server

import (
	"go-wind-shop/pkg/logx"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-bootstrap/transport/sse"

	sseServer "github.com/tx7do/kratos-transport/transport/sse"

	"go-wind-shop/app/app/service/internal/service"
)

// NewSseServer creates a new SSE server.
//
// C端 SSE 网关：流 ID = accessToken（C端以 `?stream=<token>&token=<token>` 订阅），
// AI 导购的 chunk/tool 帧经本网关按用户隔离推送（AiChatService.publishToUserStreams）。
func NewSseServer(
	ctx *bootstrap.Context,
	aiChatService *service.AiChatService,
) *sseServer.Server {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Sse == nil {
		return nil
	}

	l := logx.New(ctx.NewLoggerHelper("sse-server/app-service"))

	opts := []sseServer.ServerOption{
		sseServer.WithSubscriberFunction(func(streamID sseServer.StreamID, sub *sseServer.Subscriber) {
			//l.Infof("SSE: [%s]", sub.URL)
			l.Infof("subscriber [%s] connected", streamID)
		}),
	}
	if aiChatService != nil {
		// AI 导购 chunk/tool 帧经本网关推送；连接鉴权走 accessToken 校验。
		opts = append(opts, sseServer.WithAuthorizeFunc(aiChatService.HandleAuthorize))
	}
	srv := sse.NewSseServer(cfg.Server.Sse, opts...)

	if aiChatService != nil {
		aiChatService.RegisterAiChatPublisher(srv)
	}

	return srv
}
