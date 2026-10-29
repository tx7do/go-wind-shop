package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// GrouponService 拼团活动管理（admin BFF：活动全 CRUD + 团只读转发）。
type GrouponService struct {
	adminV1.GrouponActivityServiceHTTPServer

	log *logx.Helper

	grouponActivityServiceClient catalogV1.GrouponActivityServiceClient
}

func NewGrouponService(
	ctx *bootstrap.Context,
	grouponActivityServiceClient catalogV1.GrouponActivityServiceClient,
) *GrouponService {
	return &GrouponService{
		log:                          logx.New(ctx.NewLoggerHelper("groupon/service/admin-service")),
		grouponActivityServiceClient: grouponActivityServiceClient,
	}
}

func (s *GrouponService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListGrouponActivityResponse, error) {
	return s.grouponActivityServiceClient.List(ctx, req)
}

func (s *GrouponService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountGrouponActivityResponse, error) {
	return s.grouponActivityServiceClient.Count(ctx, req)
}

func (s *GrouponService) Get(ctx context.Context, req *catalogV1.GetGrouponActivityRequest) (*catalogV1.GrouponActivity, error) {
	return s.grouponActivityServiceClient.Get(ctx, req)
}

func (s *GrouponService) Create(ctx context.Context, req *catalogV1.CreateGrouponActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return s.grouponActivityServiceClient.Create(ctx, req)
}

func (s *GrouponService) Update(ctx context.Context, req *catalogV1.UpdateGrouponActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return nil, adminV1.ErrorBadRequest("id is required")
	}
	return s.grouponActivityServiceClient.Update(ctx, req)
}

func (s *GrouponService) Delete(ctx context.Context, req *catalogV1.DeleteGrouponActivityRequest) (*emptypb.Empty, error) {
	return s.grouponActivityServiceClient.Delete(ctx, req)
}

func (s *GrouponService) ListGroups(ctx context.Context, req *catalogV1.ListGrouponGroupsRequest) (*catalogV1.ListGrouponGroupResponse, error) {
	return s.grouponActivityServiceClient.ListGroups(ctx, req)
}
