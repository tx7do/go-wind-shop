package service

import (
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	scriptV1 "go-wind-shop/api/gen/go/script/service/v1"
	"go-wind-shop/pkg/logx"
)

type ScriptService struct {
	scriptV1.UnimplementedScriptServiceServer

	log               *logx.Helper
	scriptServiceClient scriptV1.ScriptServiceClient
}

func NewScriptService(ctx *bootstrap.Context, scriptServiceClient scriptV1.ScriptServiceClient) *ScriptService {
	return &ScriptService{
		log:                 logx.New(ctx.NewLoggerHelper("script/service/admin-service")),
		scriptServiceClient: scriptServiceClient,
	}
}

func (s *ScriptService) List(ctx context.Context, req *paginationV1.PagingRequest) (*scriptV1.ListScriptsResponse, error) {
	return s.scriptServiceClient.List(ctx, req)
}

func (s *ScriptService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*scriptV1.CountScriptsResponse, error) {
	return s.scriptServiceClient.Count(ctx, req)
}

func (s *ScriptService) Get(ctx context.Context, req *scriptV1.GetScriptRequest) (*scriptV1.Script, error) {
	return s.scriptServiceClient.Get(ctx, req)
}

func (s *ScriptService) Create(ctx context.Context, req *scriptV1.CreateScriptRequest) (*emptypb.Empty, error) {
	return s.scriptServiceClient.Create(ctx, req)
}

func (s *ScriptService) Update(ctx context.Context, req *scriptV1.UpdateScriptRequest) (*emptypb.Empty, error) {
	return s.scriptServiceClient.Update(ctx, req)
}

func (s *ScriptService) Delete(ctx context.Context, req *scriptV1.DeleteScriptRequest) (*emptypb.Empty, error) {
	return s.scriptServiceClient.Delete(ctx, req)
}

func (s *ScriptService) TestRun(ctx context.Context, req *scriptV1.TestRunScriptRequest) (*scriptV1.TestRunScriptResponse, error) {
	return s.scriptServiceClient.TestRun(ctx, req)
}

func (s *ScriptService) ListHookPoints(ctx context.Context, req *emptypb.Empty) (*scriptV1.ListHookPointsResponse, error) {
	return s.scriptServiceClient.ListHookPoints(ctx, req)
}
