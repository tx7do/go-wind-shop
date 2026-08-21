package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductService struct {
	appV1.ProductServiceHTTPServer

	log *log.Helper

	productServiceClient catalogV1.ProductServiceClient
}

func NewProductService(
	ctx *bootstrap.Context,
	productServiceClient catalogV1.ProductServiceClient,
) *ProductService {
	return &ProductService{
		log:                  ctx.NewLoggerHelper("product/service/app-service"),
		productServiceClient: productServiceClient,
	}
}

func (s *ProductService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductResponse, error) {
	// 店面只展示上架商品。强制注入 status=PRODUCT_STATUS_ACTIVE 过滤，
	// 无论客户端传入何种 status，都覆盖为 ACTIVE，确保 DRAFT/INACTIVE
	// 不对 shopper 可见。
	merged := injectStatusFilter(req.GetQuery(), "PRODUCT_STATUS_ACTIVE")
	req.FilteringType = &paginationV1.PagingRequest_Query{Query: merged}
	return s.productServiceClient.List(ctx, req)
}

func (s *ProductService) Get(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.Product, error) {
	resp, err := s.productServiceClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	// 店面不可见非上架商品。
	if resp.GetStatus() != catalogV1.Product_PRODUCT_STATUS_ACTIVE {
		return nil, catalogV1.ErrorNotFound("product not found")
	}
	return resp, nil
}

// SearchProducts 商品全文搜索（Elasticsearch）。
//
// 透传到 core ProductSearchService.Search，该 service 硬编码 status=ACTIVE，
// 客户端无法覆盖。商品无租户隔离（mall_products 无 TenantID mixin），故无
// tenant_id 过滤。结果只回传 product_id/language/name（最小字段集）。
func (s *ProductService) SearchProducts(ctx context.Context, req *catalogV1.SearchProductsRequest) (*catalogV1.SearchProductsResponse, error) {
	return s.productServiceClient.SearchProducts(ctx, req)
}

// injectStatusFilter 将 status 过滤条件合并进现有 query JSON 字符串。
// 始终覆盖 status 字段为给定值。空输入则新建只含 status 的对象。
func injectStatusFilter(existing, status string) string {
	parsed := map[string]any{}
	if existing != "" {
		_ = json.Unmarshal([]byte(existing), &parsed)
	}
	parsed["status"] = status
	out, err := json.Marshal(parsed)
	if err != nil {
		return ""
	}
	return string(out)
}
