package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-shop/app/core/service/internal/data"

	interactionV1 "go-wind-shop/api/gen/go/interaction/service/v1"
)

// InteractionService 是评论点赞计数内聚子系统的服务端入口。
//
// 安全契约：
//   - viewer 用户身份强制从鉴权上下文（viewer context）提取，绝不接受请求体
//     中的 user_id。未登录调用一律 401。
//   - 跨租户隔离、防重复、幂等性由 InteractionRepo 保证。
type InteractionService struct {
	interactionV1.UnimplementedInteractionServiceServer

	interactionRepo *data.InteractionRepo
	log             *log.Helper
}

func NewInteractionService(
	ctx *bootstrap.Context,
	interactionRepo *data.InteractionRepo,
) *InteractionService {
	return &InteractionService{
		log:             ctx.NewLoggerHelper("interaction/service/core-service"),
		interactionRepo: interactionRepo,
	}
}

// Like 点赞评论。viewer 身份从鉴权上下文提取，请求体的 user_id 被忽略。
func (s *InteractionService) Like(ctx context.Context, req *interactionV1.LikeRequest) (*interactionV1.LikeResponse, error) {
	viewerUserID, ok := viewerUserIDFromContext(ctx)
	if !ok {
		return nil, interactionV1.ErrorUnauthorized("login required")
	}
	liked, likeCount, err := s.interactionRepo.Like(ctx, viewerUserID, req.GetTargetType(), req.GetTargetId())
	if err != nil {
		return nil, err
	}
	return &interactionV1.LikeResponse{Liked: liked, LikeCount: likeCount}, nil
}

// Unlike 取消点赞评论。
func (s *InteractionService) Unlike(ctx context.Context, req *interactionV1.LikeRequest) (*interactionV1.LikeResponse, error) {
	viewerUserID, ok := viewerUserIDFromContext(ctx)
	if !ok {
		return nil, interactionV1.ErrorUnauthorized("login required")
	}
	liked, likeCount, err := s.interactionRepo.Unlike(ctx, viewerUserID, req.GetTargetType(), req.GetTargetId())
	if err != nil {
		return nil, err
	}
	return &interactionV1.LikeResponse{Liked: liked, LikeCount: likeCount}, nil
}

// GetInteractionStatus 批量查询当前 viewer 对指定评论的点赞状态。
func (s *InteractionService) GetInteractionStatus(ctx context.Context, req *interactionV1.GetInteractionStatusRequest) (*interactionV1.GetInteractionStatusResponse, error) {
	viewerUserID, ok := viewerUserIDFromContext(ctx)
	if !ok {
		return nil, interactionV1.ErrorUnauthorized("login required")
	}
	statuses, err := s.interactionRepo.GetInteractionStatus(ctx, viewerUserID, req.GetTargetType(), req.GetTargetIds())
	if err != nil {
		return nil, err
	}
	return &interactionV1.GetInteractionStatusResponse{Statuses: statuses}, nil
}

// GetCounts 批量查询指定评论的计数（如点赞数）。
//
// 计数查询本身不依赖 viewer 用户身份，仅按 tenant 隔离。登录用户由 repo 层
// 从 viewer context 提取 tenant_id 过滤；未登录（SystemViewer，tenant_id==0）
// 由 repo 层返回空结果，避免跨租户泄漏的同时不抛 401 打断公开页渲染。
// Like/Unlike 等写操作仍强制登录。
func (s *InteractionService) GetCounts(ctx context.Context, req *interactionV1.GetCountsRequest) (*interactionV1.GetCountsResponse, error) {
	counts, err := s.interactionRepo.GetCounts(ctx, req.GetTargetType(), req.GetTargetIds(), req.GetMetrics())
	if err != nil {
		return nil, err
	}
	return &interactionV1.GetCountsResponse{Counts: counts}, nil
}
