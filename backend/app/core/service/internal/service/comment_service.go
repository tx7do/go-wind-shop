package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"
)

type CommentService struct {
	commentV1.UnimplementedCommentServiceServer

	log *logx.Helper

	repo       *data.CommentRepo
	configRepo *data.ConfigRepo
	aiDeps     *CommentAIModerationDeps
}

func NewCommentService(
	ctx *bootstrap.Context,
	repo *data.CommentRepo,
	configRepo *data.ConfigRepo,
	aiDeps *CommentAIModerationDeps,
) *CommentService {
	return &CommentService{
		log:        logx.New(ctx.NewLoggerHelper("comment/service/core-service")),
		repo:       repo,
		configRepo: configRepo,
		aiDeps:     aiDeps,
	}
}

func (s *CommentService) List(ctx context.Context, req *paginationV1.PagingRequest) (*commentV1.ListCommentResponse, error) {
	return s.repo.List(ctx, req, true)
}

func (s *CommentService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*commentV1.CountCommentResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &commentV1.CountCommentResponse{
		Count: uint64(count),
	}, nil
}

func (s *CommentService) Get(ctx context.Context, req *commentV1.GetCommentRequest) (*commentV1.Comment, error) {
	return s.repo.Get(ctx, req)
}

func (s *CommentService) Create(ctx context.Context, req *commentV1.CreateCommentRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, commentV1.ErrorBadRequest("invalid parameter")
	}

	// AI 反垃圾闸门（系统参数开关，默认关）：判为垃圾即拒绝创建。
	// AI 链路任何错误都 fail-open（放行）——可用性优先，垃圾漏网可由人工审核兜底。
	if err := s.moderateCommentAI(ctx, req.GetData().GetContent()); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CommentService) Update(ctx context.Context, req *commentV1.UpdateCommentRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, commentV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CommentService) Delete(ctx context.Context, req *commentV1.DeleteCommentRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *CommentService) GetProductRating(ctx context.Context, req *commentV1.GetProductRatingRequest) (*commentV1.ProductRatingSummary, error) {
	return s.repo.GetProductRating(ctx, req.GetProductId())
}
