package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type CategoryService struct {
	appV1.CategoryServiceHTTPServer

	log *log.Helper

	categoryServiceClient catalogV1.CategoryServiceClient
}

func NewCategoryService(
	ctx *bootstrap.Context,
	categoryServiceClient catalogV1.CategoryServiceClient,
) *CategoryService {
	return &CategoryService{
		log:                   ctx.NewLoggerHelper("category/service/app-service"),
		categoryServiceClient: categoryServiceClient,
	}
}

func (s *CategoryService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListCategoryResponse, error) {
	return s.categoryServiceClient.List(ctx, req)
}

func (s *CategoryService) Get(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.Category, error) {
	return s.categoryServiceClient.Get(ctx, req)
}
