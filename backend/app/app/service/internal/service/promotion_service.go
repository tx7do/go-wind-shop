package service

import (
	"go-wind-shop/pkg/logx"
	"context"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// PromotionService 促销活动（app BFF，C端只读 + 试算）。
// 租户隔离由 core 侧 TenantPrivacy 按 viewer 自动注入；试算的租户同样取 viewer。
type PromotionService struct {
	appV1.PromotionServiceHTTPServer

	log *logx.Helper

	promotionServiceClient catalogV1.PromotionServiceClient
}

func NewPromotionService(
	ctx *bootstrap.Context,
	promotionServiceClient catalogV1.PromotionServiceClient,
) *PromotionService {
	return &PromotionService{
		log:                    logx.New(ctx.NewLoggerHelper("promotion/app-service")),
		promotionServiceClient: promotionServiceClient,
	}
}

// ListActivePromotions 当前生效中的活动列表（匿名可读，TenantPrivacy 对无 viewer 放行）。
func (s *PromotionService) ListActivePromotions(ctx context.Context, req *catalogV1.ListActivePromotionsRequest) (*catalogV1.ListPromotionResponse, error) {
	return s.promotionServiceClient.ListActivePromotions(ctx, req)
}

// QuoteOrderPromotions 促销试算（结算页展示满减；登录态要求与结算一致）。
func (s *PromotionService) QuoteOrderPromotions(ctx context.Context, req *catalogV1.QuoteOrderPromotionsRequest) (*catalogV1.QuoteOrderPromotionsResponse, error) {
	if _, err := auth.FromContext(ctx); err != nil {
		return nil, err
	}
	return s.promotionServiceClient.QuoteOrderPromotions(ctx, req)
}
