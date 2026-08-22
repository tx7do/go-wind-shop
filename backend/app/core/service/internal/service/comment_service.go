package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"
)

type CommentService struct {
	commentV1.UnimplementedCommentServiceServer

	log *log.Helper

	repo *data.CommentRepo
}

func NewCommentService(ctx *bootstrap.Context, repo *data.CommentRepo) *CommentService {
	return &CommentService{
		log:  ctx.NewLoggerHelper("comment/service/core-service"),
		repo: repo,
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
