package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/task"
)

type ProductService struct {
	catalogV1.UnimplementedProductServiceServer

	log *log.Helper

	repo                 *data.ProductRepo
	productSearchService *ProductSearchService
	taskService          *TaskService
}

func NewProductService(
	ctx *bootstrap.Context,
	repo *data.ProductRepo,
	productSearchService *ProductSearchService,
	taskService *TaskService,
) *ProductService {
	return &ProductService{
		log:                  ctx.NewLoggerHelper("product/service/core-service"),
		repo:                repo,
		productSearchService: productSearchService,
		taskService:          taskService,
	}
}

func (s *ProductService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *ProductService) Get(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.Product, error) {
	return s.repo.Get(ctx, req)
}

func (s *ProductService) Create(ctx context.Context, req *catalogV1.CreateProductRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	dto, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// 双写钩子：商品创建后入队 reindex（best-effort，失败仅日志）
	s.enqueueProductReindex(ctx, dto.GetId(), "index")

	return &emptypb.Empty{}, nil
}

func (s *ProductService) Update(ctx context.Context, req *catalogV1.UpdateProductRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	dto, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}

	// 双写钩子：商品更新后入队 reindex（best-effort，失败仅日志）
	s.enqueueProductReindex(ctx, dto.GetId(), "index")

	return &emptypb.Empty{}, nil
}

func (s *ProductService) Delete(ctx context.Context, req *catalogV1.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}

	// 双写钩子：商品删除后入队 reindex（delete op，清理 ES 文档）
	s.enqueueProductReindex(ctx, req.GetId(), "delete")

	return &emptypb.Empty{}, nil
}

func (s *ProductService) TranslationExists(ctx context.Context, req *catalogV1.ProductTranslationExistsRequest) (*catalogV1.ProductTranslationExistsResponse, error) {
	exists, err := s.repo.TranslationExists(ctx, req.GetProductId(), req.GetLanguageCode())
	if err != nil {
		return nil, err
	}

	return &catalogV1.ProductTranslationExistsResponse{
		Exists: exists,
	}, nil
}

func (s *ProductService) GetTranslation(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.ProductTranslation, error) {
	return s.repo.GetTranslation(ctx, req)
}

func (s *ProductService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	dto, err := s.repo.CreateTranslation(ctx, req)
	if err != nil {
		return nil, err
	}

	// 双写钩子：翻译创建后入队所属商品的 reindex（index op）
	// dto 是新建的翻译记录，其 ProductId 指向所属商品
	if dto != nil {
		s.enqueueProductReindex(ctx, dto.GetProductId(), "index")
	}

	return dto, nil
}

func (s *ProductService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	dto, err := s.repo.UpdateTranslation(ctx, req)
	if err != nil {
		return nil, err
	}

	// 双写钩子：翻译更新后入队所属商品的 reindex（index op）
	if dto != nil {
		s.enqueueProductReindex(ctx, dto.GetProductId(), "index")
	}

	return dto, nil
}

func (s *ProductService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductTranslationRequest) (*emptypb.Empty, error) {
	err := s.repo.DeleteTranslation(ctx, req)
	if err != nil {
		return nil, err
	}

	// 双写钩子：翻译删除后入队所属商品的 reindex（index op，worker 从 DB 取最新状态
	// 决定是更新还是删除 ES 文档）。仅在用 Identifier 分支时可提取 product_id；
	// Id 分支无 product_id，跳过（ES 文档由周期 ReindexAll 最终修复）。
	if identifier := req.GetIdentifier(); identifier != nil {
		s.enqueueProductReindex(ctx, identifier.GetProductId(), "index")
	}

	return &emptypb.Empty{}, nil
}

// enqueueProductReindex 是双写钩子的统一入口。
// 商品/翻译的 Create/Update/Delete 在 DB 事务提交后调用，入队 search.reindex 任务。
// best-effort：入队失败仅日志，不阻塞主流程。失败由 asynq 自动重试或周期
// ReindexAll 修复。
func (s *ProductService) enqueueProductReindex(ctx context.Context, productID uint32, op string) {
	if productID == 0 {
		return
	}
	payload := &task.SearchReindexPayload{
		Entity: "product",
		ID:     productID,
		Op:     op,
	}
	if err := s.taskService.EnqueueSearchReindex(payload); err != nil {
		s.log.Errorf("enqueue product reindex failed (product_id=%d op=%s): %v", productID, op, err)
	}
}
