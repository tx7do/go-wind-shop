package service

import (
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	"go-wind-shop/pkg/logx"

)

type PlanModuleService struct {
	identityV1.UnimplementedPlanModuleServiceServer

	log            *logx.Helper
	planModuleClient identityV1.PlanModuleServiceClient
}

func NewPlanModuleService(ctx *bootstrap.Context, planModuleClient identityV1.PlanModuleServiceClient) *PlanModuleService {
	return &PlanModuleService{
		log:              logx.New(ctx.NewLoggerHelper("plan-module/service/admin-service")),
		planModuleClient: planModuleClient,
	}
}

func (s *PlanModuleService) List(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.ListPlanModuleResponse, error) {
	return s.planModuleClient.List(ctx, req)
}

func (s *PlanModuleService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.CountPlanModuleResponse, error) {
	return s.planModuleClient.Count(ctx, req)
}

func (s *PlanModuleService) Get(ctx context.Context, req *identityV1.GetPlanModuleRequest) (*identityV1.PlanModule, error) {
	return s.planModuleClient.Get(ctx, req)
}

func (s *PlanModuleService) Create(ctx context.Context, req *identityV1.CreatePlanModuleRequest) (*emptypb.Empty, error) {
	return s.planModuleClient.Create(ctx, req)
}

func (s *PlanModuleService) Update(ctx context.Context, req *identityV1.UpdatePlanModuleRequest) (*emptypb.Empty, error) {
	return s.planModuleClient.Update(ctx, req)
}

func (s *PlanModuleService) Delete(ctx context.Context, req *identityV1.DeletePlanModuleRequest) (*emptypb.Empty, error) {
	return s.planModuleClient.Delete(ctx, req)
}
