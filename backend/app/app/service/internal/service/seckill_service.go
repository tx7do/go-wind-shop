package service

import (
	"context"

	"go-wind-shop/pkg/logx"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// SeckillService 秒杀活动（app BFF，C端只读）。
// 租户隔离由 core 侧 TenantPrivacy 按 viewer 自动注入。
type SeckillService struct {
	appV1.SeckillServiceHTTPServer

	log *logx.Helper

	seckillActivityServiceClient catalogV1.SeckillActivityServiceClient
}

func NewSeckillService(
	ctx *bootstrap.Context,
	seckillActivityServiceClient catalogV1.SeckillActivityServiceClient,
) *SeckillService {
	return &SeckillService{
		log:                          logx.New(ctx.NewLoggerHelper("seckill/app-service")),
		seckillActivityServiceClient: seckillActivityServiceClient,
	}
}

// ListActiveSeckills 当前生效中的秒杀列表（匿名可读，商品页秒杀价角标）。
func (s *SeckillService) ListActiveSeckills(ctx context.Context, req *catalogV1.ListActiveSeckillsRequest) (*catalogV1.ListSeckillActivityResponse, error) {
	return s.seckillActivityServiceClient.ListActiveSeckills(ctx, req)
}
