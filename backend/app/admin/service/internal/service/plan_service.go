package service

import (
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	"go-wind-shop/pkg/logx"

)

type PlanService struct {
	identityV1.UnimplementedPlanServiceServer

	log         *logx.Helper
	planClient  identityV1.PlanServiceClient
}

func NewPlanService(ctx *bootstrap.Context, planClient identityV1.PlanServiceClient) *PlanService {
	return &PlanService{
		log:        logx.New(ctx.NewLoggerHelper("plan/service/admin-service")),
		planClient: planClient,
	}
}

func (s *PlanService) List(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.ListPlanResponse, error) {
	return s.planClient.List(ctx, req)
}

func (s *PlanService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.CountPlanResponse, error) {
	return s.planClient.Count(ctx, req)
}

func (s *PlanService) Get(ctx context.Context, req *identityV1.GetPlanRequest) (*identityV1.Plan, error) {
	return s.planClient.Get(ctx, req)
}

func (s *PlanService) Create(ctx context.Context, req *identityV1.CreatePlanRequest) (*emptypb.Empty, error) {
	return s.planClient.Create(ctx, req)
}

func (s *PlanService) Update(ctx context.Context, req *identityV1.UpdatePlanRequest) (*emptypb.Empty, error) {
	return s.planClient.Update(ctx, req)
}

func (s *PlanService) Delete(ctx context.Context, req *identityV1.DeletePlanRequest) (*emptypb.Empty, error) {
	return s.planClient.Delete(ctx, req)
}
