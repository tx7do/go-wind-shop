package server

import (
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-bootstrap/transport/sse"

	sseServer "github.com/tx7do/kratos-transport/transport/sse"

	"go-wind-shop/app/admin/service/internal/service"
)

// NewSseServer creates a new SSE server.
func NewSseServer(
	ctx *bootstrap.Context,
	internalMessageService *service.InternalMessageService,
	aiChatService *service.AiChatService,
) *sseServer.Server {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Sse == nil {
		return nil
	}

	srv := sse.NewSseServer(cfg.Server.Sse,
		sseServer.WithSubscriberFunction(internalMessageService.HandleSubscribe),
		sseServer.WithAuthorizeFunc(internalMessageService.HandleAuthorize),
	)

	internalMessageService.RegisterInternalMessagePublisher(srv)

	// AI chat chunk/tool 帧经同一 SSE 网关推送（此前漏接线，admin 对话一直
	// 运行在无流式的降级模式）。
	aiChatService.RegisterAiChatPublisher(srv)

	//srv.CreateStream("test")

	return srv
}
