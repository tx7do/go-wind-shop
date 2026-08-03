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

type CategoryService struct {
	adminV1.CategoryServiceHTTPServer

	log *log.Helper

	categoryServiceClient catalogV1.CategoryServiceClient
}

func NewCategoryService(
	ctx *bootstrap.Context,
	categoryServiceClient catalogV1.CategoryServiceClient,
) *CategoryService {
	return &CategoryService{
		log:                   ctx.NewLoggerHelper("category/service/admin-service"),
		categoryServiceClient: categoryServiceClient,
	}
}

func (s *CategoryService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListCategoryResponse, error) {
	return s.categoryServiceClient.List(ctx, req)
}

func (s *CategoryService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountCategoryResponse, error) {
	return s.categoryServiceClient.Count(ctx, req)
}

func (s *CategoryService) Get(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.Category, error) {
	return s.categoryServiceClient.Get(ctx, req)
}

func (s *CategoryService) Create(ctx context.Context, req *catalogV1.CreateCategoryRequest) (*emptypb.Empty, error) {
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

	return s.categoryServiceClient.Create(ctx, req)
}

func (s *CategoryService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateCategoriesRequest) (*emptypb.Empty, error) {
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

	return s.categoryServiceClient.BatchCreate(ctx, req)
}

func (s *CategoryService) Update(ctx context.Context, req *catalogV1.UpdateCategoryRequest) (*emptypb.Empty, error) {
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

	return s.categoryServiceClient.Update(ctx, req)
}

func (s *CategoryService) Delete(ctx context.Context, req *catalogV1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	return s.categoryServiceClient.Delete(ctx, req)
}

func (s *CategoryService) TranslationExists(ctx context.Context, req *catalogV1.CategoryTranslationExistsRequest) (*catalogV1.CategoryTranslationExistsResponse, error) {
	return s.categoryServiceClient.TranslationExists(ctx, req)
}

func (s *CategoryService) GetTranslation(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.CategoryTranslation, error) {
	return s.categoryServiceClient.GetTranslation(ctx, req)
}

func (s *CategoryService) CreateTranslation(ctx context.Context, req *catalogV1.CreateCategoryTranslationRequest) (*catalogV1.CategoryTranslation, error) {
	return s.categoryServiceClient.CreateTranslation(ctx, req)
}

func (s *CategoryService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateCategoryTranslationRequest) (*catalogV1.CategoryTranslation, error) {
	return s.categoryServiceClient.UpdateTranslation(ctx, req)
}

func (s *CategoryService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteCategoryTranslationRequest) (*emptypb.Empty, error) {
	return s.categoryServiceClient.DeleteTranslation(ctx, req)
}
