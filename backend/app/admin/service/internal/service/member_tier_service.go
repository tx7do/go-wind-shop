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

type MemberTierService struct {
	adminV1.MemberTierServiceHTTPServer
	log *logx.Helper
	memberTierServiceClient catalogV1.MemberTierServiceClient
}

func NewMemberTierService(ctx *bootstrap.Context, memberTierServiceClient catalogV1.MemberTierServiceClient) *MemberTierService {
	return &MemberTierService{
		log:                      logx.New(ctx.NewLoggerHelper("member-tier/service/admin-service")),
		memberTierServiceClient:  memberTierServiceClient,
	}
}

func (s *MemberTierService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListMemberTierResponse, error) {
	return s.memberTierServiceClient.List(ctx, req)
}
func (s *MemberTierService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountMemberTierResponse, error) {
	return s.memberTierServiceClient.Count(ctx, req)
}
func (s *MemberTierService) Get(ctx context.Context, req *catalogV1.GetMemberTierRequest) (*catalogV1.MemberTier, error) {
	return s.memberTierServiceClient.Get(ctx, req)
}
func (s *MemberTierService) Create(ctx context.Context, req *catalogV1.CreateMemberTierRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return s.memberTierServiceClient.Create(ctx, req)
}
func (s *MemberTierService) Update(ctx context.Context, req *catalogV1.UpdateMemberTierRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return nil, adminV1.ErrorBadRequest("id is required")
	}
	return s.memberTierServiceClient.Update(ctx, req)
}
func (s *MemberTierService) Delete(ctx context.Context, req *catalogV1.DeleteMemberTierRequest) (*emptypb.Empty, error) {
	return s.memberTierServiceClient.Delete(ctx, req)
}
