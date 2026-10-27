package service

import (
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	scriptV1 "go-wind-shop/api/gen/go/script/service/v1"
	"go-wind-shop/pkg/logx"

)

type ScriptLogService struct {
	scriptV1.UnimplementedScriptLogServiceServer

	log                    *logx.Helper
	scriptLogServiceClient scriptV1.ScriptLogServiceClient
}

func NewScriptLogService(ctx *bootstrap.Context, scriptLogServiceClient scriptV1.ScriptLogServiceClient) *ScriptLogService {
	return &ScriptLogService{
		log:                    logx.New(ctx.NewLoggerHelper("script-log/service/admin-service")),
		scriptLogServiceClient: scriptLogServiceClient,
	}
}

func (s *ScriptLogService) List(ctx context.Context, req *paginationV1.PagingRequest) (*scriptV1.ListScriptLogsResponse, error) {
	return s.scriptLogServiceClient.List(ctx, req)
}

func (s *ScriptLogService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*scriptV1.CountScriptLogsResponse, error) {
	return s.scriptLogServiceClient.Count(ctx, req)
}

func (s *ScriptLogService) Purge(ctx context.Context, req *scriptV1.PurgeScriptLogsRequest) (*scriptV1.PurgeScriptLogsResponse, error) {
	return s.scriptLogServiceClient.Purge(ctx, req)
}
