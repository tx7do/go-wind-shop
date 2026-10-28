package service

import (
	"go-wind-shop/pkg/logx"
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// PromotionService 促销活动管理（admin BFF，REST → gRPC 转发）。
type PromotionService struct {
	adminV1.PromotionServiceHTTPServer

	log *logx.Helper

	promotionServiceClient catalogV1.PromotionServiceClient
}

func NewPromotionService(
	ctx *bootstrap.Context,
	promotionServiceClient catalogV1.PromotionServiceClient,
) *PromotionService {
	return &PromotionService{
		log:                    logx.New(ctx.NewLoggerHelper("promotion/service/admin-service")),
		promotionServiceClient: promotionServiceClient,
	}
}

func (s *PromotionService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPromotionResponse, error) {
	return s.promotionServiceClient.List(ctx, req)
}

func (s *PromotionService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountPromotionResponse, error) {
	return s.promotionServiceClient.Count(ctx, req)
}

func (s *PromotionService) Get(ctx context.Context, req *catalogV1.GetPromotionRequest) (*catalogV1.Promotion, error) {
	return s.promotionServiceClient.Get(ctx, req)
}

func (s *PromotionService) Create(ctx context.Context, req *catalogV1.CreatePromotionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.promotionServiceClient.Create(ctx, req)
}

func (s *PromotionService) Update(ctx context.Context, req *catalogV1.UpdatePromotionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.Id = trans.Ptr(req.GetId())

	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	return s.promotionServiceClient.Update(ctx, req)
}

func (s *PromotionService) Delete(ctx context.Context, req *catalogV1.DeletePromotionRequest) (*emptypb.Empty, error) {
	return s.promotionServiceClient.Delete(ctx, req)
}
