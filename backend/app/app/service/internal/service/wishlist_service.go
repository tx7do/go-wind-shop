package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	wishlistV1 "go-wind-shop/api/gen/go/wishlist/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// WishlistService 收藏夹前台 BFF 服务（裁剪版：仅 List/Create/Delete）。
// 仿 CartItemService 的透传 + 身份注入范式：
//   - List/Delete：纯透传到 core WishlistService（core 侧 UserPrivacy 按 viewer user_id 行级隔离）。
//   - Create：强制覆盖 tenantId/createdBy 为当前登录用户，防伪造越权；
//     user_id 由 core 隐私层在 Create 时强制覆盖为 viewer，客户端不可伪造。
type WishlistService struct {
	appV1.WishlistServiceHTTPServer

	log *log.Helper

	wishlistServiceClient wishlistV1.WishlistServiceClient
}

func NewWishlistService(
	ctx *bootstrap.Context,
	wishlistServiceClient wishlistV1.WishlistServiceClient,
) *WishlistService {
	return &WishlistService{
		log:                   ctx.NewLoggerHelper("wishlist/service/app-service"),
		wishlistServiceClient: wishlistServiceClient,
	}
}

func (s *WishlistService) List(ctx context.Context, req *paginationV1.PagingRequest) (*wishlistV1.ListWishlistResponse, error) {
	return s.wishlistServiceClient.List(ctx, req)
}

func (s *WishlistService) Create(ctx context.Context, req *wishlistV1.CreateWishlistRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 强制覆盖 tenantId 为当前登录用户的租户，防伪造越权。
	// user_id 不在此设置——由 core UserPrivacy 隐私层强制覆盖为 viewer。
	req.Data.TenantId = trans.Ptr(operator.GetTenantId())
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.wishlistServiceClient.Create(ctx, req)
}

func (s *WishlistService) Delete(ctx context.Context, req *wishlistV1.DeleteWishlistRequest) (*emptypb.Empty, error) {
	return s.wishlistServiceClient.Delete(ctx, req)
}
