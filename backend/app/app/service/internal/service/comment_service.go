package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// CommentService 商品评论前台服务（app BFF，REST → gRPC 转发）。
//
// 安全双闸：
//  1. List/Get 在 BFF 层过滤 STATUS_APPROVED——未审核/拒绝/垃圾评论对前台不可见；
//  2. Update/Delete 调 ensureCommentOwner——仅作者可改/删自己的评论（IDOR 防护）。
//
// Create 注入 CreatedBy（防作者伪造）；用户隔离由 core UserPrivacy 行级隔离兜底。
type CommentService struct {
	appV1.CommentServiceHTTPServer

	log *log.Helper

	commentClient commentV1.CommentServiceClient
}

func NewCommentService(
	ctx *bootstrap.Context,
	commentClient commentV1.CommentServiceClient,
) *CommentService {
	return &CommentService{
		log:           ctx.NewLoggerHelper("comment/service/app-service"),
		commentClient: commentClient,
	}
}

func (s *CommentService) List(ctx context.Context, req *paginationV1.PagingRequest) (*commentV1.ListCommentResponse, error) {
	resp, err := s.commentClient.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// 前台仅展示已审核通过的评论，未审核/拒绝/垃圾一律过滤。
	if resp != nil {
		filtered := make([]*commentV1.Comment, 0, len(resp.GetItems()))
		for _, c := range resp.GetItems() {
			if c.GetStatus() == commentV1.Comment_STATUS_APPROVED {
				filtered = append(filtered, c)
			}
		}
		resp.Items = filtered
		resp.Total = uint64(len(filtered))
	}

	return resp, nil
}

func (s *CommentService) Get(ctx context.Context, req *commentV1.GetCommentRequest) (*commentV1.Comment, error) {
	resp, err := s.commentClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	// 非 APPROVED 评论对前台不可见，返回 NotFound 不泄露存在性。
	if resp == nil || resp.GetStatus() != commentV1.Comment_STATUS_APPROVED {
		return nil, commentV1.ErrorNotFound("comment not found")
	}

	return resp, nil
}

func (s *CommentService) Create(ctx context.Context, req *commentV1.CreateCommentRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 强制注入作者为当前登录用户，防客户端伪造 CreatedBy。
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.commentClient.Create(ctx, req)
}

func (s *CommentService) Update(ctx context.Context, req *commentV1.UpdateCommentRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// IDOR 防护：仅作者可改自己的评论。
	if err := s.ensureCommentOwner(ctx, req.GetId(), operator.GetUserId()); err != nil {
		return nil, err
	}

	req.Data.Id = trans.Ptr(req.GetId())
	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	return s.commentClient.Update(ctx, req)
}

func (s *CommentService) Delete(ctx context.Context, req *commentV1.DeleteCommentRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// IDOR 防护：仅作者可删自己的评论。
	if err := s.ensureCommentOwner(ctx, req.GetId(), operator.GetUserId()); err != nil {
		return nil, err
	}

	return s.commentClient.Delete(ctx, req)
}

// ensureCommentOwner 校验指定评论的 created_by 是否等于操作者 userID，不等返回 Forbidden。
func (s *CommentService) ensureCommentOwner(ctx context.Context, commentID uint32, userID uint32) error {
	resp, err := s.commentClient.Get(ctx, &commentV1.GetCommentRequest{
		QueryBy: &commentV1.GetCommentRequest_Id{Id: commentID},
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.GetCreatedBy() != userID {
		return commentV1.ErrorForbidden("you can only modify your own comments")
	}
	return nil
}
