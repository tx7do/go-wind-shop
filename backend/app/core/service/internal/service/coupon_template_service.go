package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"
	"github.com/tx7do/go-crud/viewer"

	"go-wind-shop/app/core/service/internal/data"

	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
)

type CouponTemplateService struct {
	couponV1.UnimplementedCouponTemplateServiceServer

	log *log.Helper

	repo *data.CouponTemplateRepo
}

func NewCouponTemplateService(ctx *bootstrap.Context, repo *data.CouponTemplateRepo) *CouponTemplateService {
	return &CouponTemplateService{
		log:  ctx.NewLoggerHelper("coupon-template/service/core-service"),
		repo: repo,
	}
}

// List 按 viewer 类型分流：
//   - 平台视图（admin）：repo.List，返回全量模板（运营审核需看全部）。
//   - 普通用户视图或匿名（app 领券中心）：repo.ListClaimable，仅 claimable=true AND status=ACTIVE，
//     并在 service 层后过滤有效窗口（valid_from/until 时间判断非 ent 谓词）。
//
// 匿名可读由 TenantPrivacy 对无 viewer 放行保证（与商品 List 同机制）。
func (s *CouponTemplateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListCouponTemplateResponse, error) {
	vc, exist := viewer.FromContext(ctx)
	if exist && vc.IsPlatformContext() {
		return s.repo.List(ctx, req)
	}

	// app 路径：claimable + active + 窗口后过滤。
	resp, err := s.repo.ListClaimable(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Items) == 0 {
		return resp, nil
	}
	now := time.Now()
	filtered := make([]*couponV1.CouponTemplate, 0, len(resp.Items))
	for _, item := range resp.Items {
		if couponApplicableNow(item, now) {
			filtered = append(filtered, item)
		}
	}
	resp.Items = filtered
	resp.Total = uint64(len(filtered))
	return resp, nil
}

func (s *CouponTemplateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.CountCouponTemplateResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &couponV1.CountCouponTemplateResponse{
		Count: uint64(count),
	}, nil
}

func (s *CouponTemplateService) Get(ctx context.Context, req *couponV1.GetCouponTemplateRequest) (*couponV1.CouponTemplate, error) {
	return s.repo.Get(ctx, req)
}

func (s *CouponTemplateService) Create(ctx context.Context, req *couponV1.CreateCouponTemplateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CouponTemplateService) Update(ctx context.Context, req *couponV1.UpdateCouponTemplateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CouponTemplateService) Delete(ctx context.Context, req *couponV1.DeleteCouponTemplateRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
