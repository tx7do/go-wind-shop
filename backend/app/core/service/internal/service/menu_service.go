package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	permissionV1 "go-wind-shop/api/gen/go/permission/service/v1"

	"go-wind-shop/pkg/constants"
	appViewer "go-wind-shop/pkg/entgo/viewer"
)

type MenuService struct {
	permissionV1.UnimplementedMenuServiceServer

	log *logx.Helper

	menuRepo *data.MenuRepo
}

func NewMenuService(ctx *bootstrap.Context, menuRepo *data.MenuRepo) *MenuService {
	svc := &MenuService{
		log:      logx.New(ctx.NewLoggerHelper("menu/service/core-service")),
		menuRepo: menuRepo,
	}

	svc.init()

	return svc
}

func (s *MenuService) init() {
	ctx := appViewer.NewSystemViewerContext(context.Background())
	if count, _ := s.menuRepo.Count(ctx, nil); count == 0 {
		_ = s.createDefaultMenus(ctx)
	}
}

func (s *MenuService) List(ctx context.Context, req *paginationV1.PagingRequest) (*permissionV1.ListMenuResponse, error) {
	ret, err := s.menuRepo.List(ctx, req, false)
	if err != nil {

		return nil, err
	}

	return ret, nil
}

func (s *MenuService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*permissionV1.CountMenuResponse, error) {
	count, err := s.menuRepo.Count(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &permissionV1.CountMenuResponse{
		Count: uint64(count),
	}, nil
}

func (s *MenuService) Get(ctx context.Context, req *permissionV1.GetMenuRequest) (*permissionV1.Menu, error) {
	ret, err := s.menuRepo.Get(ctx, req)
	if err != nil {

		return nil, err
	}

	return ret, nil
}

func (s *MenuService) Create(ctx context.Context, req *permissionV1.CreateMenuRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, permissionV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.menuRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *MenuService) Update(ctx context.Context, req *permissionV1.UpdateMenuRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, permissionV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.menuRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *MenuService) Delete(ctx context.Context, req *permissionV1.DeleteMenuRequest) (*emptypb.Empty, error) {
	if err := s.menuRepo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *MenuService) createDefaultMenus(ctx context.Context) error {
	for _, m := range constants.DefaultMenus {
		if err := s.menuRepo.Create(ctx, &permissionV1.CreateMenuRequest{Data: m}); err != nil {
			s.log.Errorf("create default menu err: %v", err)
			return err
		}
	}
	return nil
}

// SyncMenus 事务化菜单同步（REPLACE 全量重建 / MERGE 增量合并）。
func (s *MenuService) SyncMenus(ctx context.Context, req *permissionV1.SyncMenusRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, permissionV1.ErrorBadRequest("invalid parameter")
	}
	// BFF gRPC 转发链路无 auth 中间件注入操作者，created_by/updated_by 留空（menu repo 字段可空）。
	count, err := s.menuRepo.SyncMenus(ctx, req.GetItems(), req.GetMode(), 0)
	if err != nil {
		return nil, err
	}
	s.log.Infof("sync menus success, mode: %s, total: %d", req.GetMode().String(), count)
	return &emptypb.Empty{}, nil
}
