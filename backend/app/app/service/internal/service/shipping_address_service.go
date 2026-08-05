package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	addressV1 "go-wind-shop/api/gen/go/address/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// ShippingAddressService 收货地址前台服务（BFF 网关，REST → gRPC 转发）。
// 读操作纯转发；写操作强制注入当前登录用户的 userId/tenantId，客户端无法伪造归属，
// 配合核心层的 UserPrivacy 策略实现"只能操作自己的地址"。
type ShippingAddressService struct {
	appV1.ShippingAddressServiceHTTPServer

	log *log.Helper

	shippingAddressServiceClient addressV1.ShippingAddressServiceClient
}

func NewShippingAddressService(
	ctx *bootstrap.Context,
	shippingAddressServiceClient addressV1.ShippingAddressServiceClient,
) *ShippingAddressService {
	return &ShippingAddressService{
		log:                          ctx.NewLoggerHelper("shipping-address/service/app-service"),
		shippingAddressServiceClient: shippingAddressServiceClient,
	}
}

func (s *ShippingAddressService) List(ctx context.Context, req *paginationV1.PagingRequest) (*addressV1.ListShippingAddressResponse, error) {
	return s.shippingAddressServiceClient.List(ctx, req)
}

func (s *ShippingAddressService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*addressV1.CountShippingAddressResponse, error) {
	return s.shippingAddressServiceClient.Count(ctx, req)
}

func (s *ShippingAddressService) Get(ctx context.Context, req *addressV1.GetShippingAddressRequest) (*addressV1.ShippingAddress, error) {
	return s.shippingAddressServiceClient.Get(ctx, req)
}

func (s *ShippingAddressService) Create(ctx context.Context, req *addressV1.CreateShippingAddressRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 强制覆盖为当前登录用户，防止客户端伪造 userId/tenantId 把地址挂到他人名下。
	req.Data.UserId = trans.Ptr(operator.GetUserId())
	req.Data.TenantId = trans.Ptr(operator.GetTenantId())
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.shippingAddressServiceClient.Create(ctx, req)
}

func (s *ShippingAddressService) Update(ctx context.Context, req *addressV1.UpdateShippingAddressRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 锁定归属：userId/tenantId 不可被客户端篡改，地址始终归属创建者。
	// UserPrivacy 策略保证只能更新自身地址，否则 affected_rows=0。
	req.Data.UserId = trans.Ptr(operator.GetUserId())
	req.Data.TenantId = trans.Ptr(operator.GetTenantId())
	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	return s.shippingAddressServiceClient.Update(ctx, req)
}

func (s *ShippingAddressService) Delete(ctx context.Context, req *addressV1.DeleteShippingAddressRequest) (*emptypb.Empty, error) {
	return s.shippingAddressServiceClient.Delete(ctx, req)
}
