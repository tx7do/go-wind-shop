package service

import (
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	"go-wind-shop/pkg/logx"

)

type PlanQuotaService struct {
	identityV1.UnimplementedPlanQuotaServiceServer

	log            *logx.Helper
	planQuotaClient  identityV1.PlanQuotaServiceClient
}

func NewPlanQuotaService(ctx *bootstrap.Context, planQuotaClient identityV1.PlanQuotaServiceClient) *PlanQuotaService {
	return &PlanQuotaService{
		log:             logx.New(ctx.NewLoggerHelper("plan-quota/service/admin-service")),
		planQuotaClient: planQuotaClient,
	}
}

func (s *PlanQuotaService) List(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.ListPlanQuotaResponse, error) {
	return s.planQuotaClient.List(ctx, req)
}

func (s *PlanQuotaService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.CountPlanQuotaResponse, error) {
	return s.planQuotaClient.Count(ctx, req)
}

func (s *PlanQuotaService) Get(ctx context.Context, req *identityV1.GetPlanQuotaRequest) (*identityV1.PlanQuota, error) {
	return s.planQuotaClient.Get(ctx, req)
}

func (s *PlanQuotaService) Create(ctx context.Context, req *identityV1.CreatePlanQuotaRequest) (*emptypb.Empty, error) {
	return s.planQuotaClient.Create(ctx, req)
}

func (s *PlanQuotaService) Update(ctx context.Context, req *identityV1.UpdatePlanQuotaRequest) (*emptypb.Empty, error) {
	return s.planQuotaClient.Update(ctx, req)
}

func (s *PlanQuotaService) Delete(ctx context.Context, req *identityV1.DeletePlanQuotaRequest) (*emptypb.Empty, error) {
	return s.planQuotaClient.Delete(ctx, req)
}
