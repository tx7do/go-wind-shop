package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type BrandService struct {
	adminV1.BrandServiceHTTPServer

	log *log.Helper

	brandServiceClient catalogV1.BrandServiceClient
}

func NewBrandService(
	ctx *bootstrap.Context,
	brandServiceClient catalogV1.BrandServiceClient,
) *BrandService {
	return &BrandService{
		log:                ctx.NewLoggerHelper("brand/service/admin-service"),
		brandServiceClient: brandServiceClient,
	}
}

func (s *BrandService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListBrandResponse, error) {
	return s.brandServiceClient.List(ctx, req)
}

func (s *BrandService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountBrandResponse, error) {
	return s.brandServiceClient.Count(ctx, req)
}

func (s *BrandService) Get(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.Brand, error) {
	return s.brandServiceClient.Get(ctx, req)
}

func (s *BrandService) Create(ctx context.Context, req *catalogV1.CreateBrandRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	for i := range req.Data.Translations {
		req.Data.Translations[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.brandServiceClient.Create(ctx, req)
}

func (s *BrandService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateBrandsRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
		for j := range req.Items[i].Translations {
			req.Items[i].Translations[j].CreatedBy = trans.Ptr(operator.UserId)
		}
	}

	return s.brandServiceClient.BatchCreate(ctx, req)
}

func (s *BrandService) Update(ctx context.Context, req *catalogV1.UpdateBrandRequest) (*emptypb.Empty, error) {
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

	for i := range req.Data.Translations {
		req.Data.Translations[i].UpdatedBy = trans.Ptr(operator.UserId)
	}

	return s.brandServiceClient.Update(ctx, req)
}

func (s *BrandService) Delete(ctx context.Context, req *catalogV1.DeleteBrandRequest) (*emptypb.Empty, error) {
	return s.brandServiceClient.Delete(ctx, req)
}

func (s *BrandService) TranslationExists(ctx context.Context, req *catalogV1.BrandTranslationExistsRequest) (*catalogV1.BrandTranslationExistsResponse, error) {
	return s.brandServiceClient.TranslationExists(ctx, req)
}

func (s *BrandService) GetTranslation(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.BrandTranslation, error) {
	return s.brandServiceClient.GetTranslation(ctx, req)
}

func (s *BrandService) CreateTranslation(ctx context.Context, req *catalogV1.CreateBrandTranslationRequest) (*catalogV1.BrandTranslation, error) {
	return s.brandServiceClient.CreateTranslation(ctx, req)
}

func (s *BrandService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateBrandTranslationRequest) (*catalogV1.BrandTranslation, error) {
	return s.brandServiceClient.UpdateTranslation(ctx, req)
}

func (s *BrandService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteBrandTranslationRequest) (*emptypb.Empty, error) {
	return s.brandServiceClient.DeleteTranslation(ctx, req)
}
