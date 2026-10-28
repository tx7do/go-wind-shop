package service

import (
	"strconv"
	"strings"
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/task"
)

type ProductService struct {
	catalogV1.UnimplementedProductServiceServer

	log *logx.Helper

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
		log:                  logx.New(ctx.NewLoggerHelper("product/service/core-service")),
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

// SearchProducts 商品全文搜索（OpenSearch）。
//
// 此前本 RPC 未实现（嵌入 Unimplemented），C端搜索链路一直是断的——本次随
// 搜索建议一并补上。安全口径与 ProductSearchService.Search 一致：status 硬编码
// ACTIVE、language 强制 term 过滤（缺省 zh-CN）。
func (s *ProductService) SearchProducts(ctx context.Context, req *catalogV1.SearchProductsRequest) (*catalogV1.SearchProductsResponse, error) {
	query := strings.TrimSpace(req.GetQuery())
	language := req.GetLanguage()
	if language == "" {
		language = "zh-CN"
	}

	result, err := s.productSearchService.Search(ctx, query, language, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("search products failed")
	}

	items := make([]*catalogV1.SearchProductHit, 0, len(result.Hits))
	for _, h := range result.Hits {
		pid, pErr := strconv.ParseUint(h.ProductID, 10, 32)
		if pErr != nil {
			continue
		}
		items = append(items, &catalogV1.SearchProductHit{
			ProductId: uint32(pid),
			Language:  h.Language,
			Name:      h.Name,
			ImageUrl:  h.ImageURL,
		})
	}

	return &catalogV1.SearchProductsResponse{
		Items: items,
		Total: int32(result.Total),
	}, nil
}

// SearchProductSuggestions 搜索建议（输入联想，前台免鉴权）。
func (s *ProductService) SearchProductSuggestions(ctx context.Context, req *catalogV1.SearchProductSuggestionsRequest) (*catalogV1.SearchProductSuggestionsResponse, error) {
	query := strings.TrimSpace(req.GetQuery())
	language := req.GetLanguage()
	if language == "" {
		language = "zh-CN"
	}

	suggestions, err := s.productSearchService.SearchSuggestions(ctx, query, language, int(req.GetSize()))
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("search suggestions failed")
	}

	items := make([]*catalogV1.SearchProductSuggestionsResponse_Suggestion, 0, len(suggestions))
	for _, sug := range suggestions {
		items = append(items, &catalogV1.SearchProductSuggestionsResponse_Suggestion{
			ProductId: sug.ProductId,
			Name:      sug.Name,
			ImageUrl:  sug.ImageUrl,
		})
	}

	return &catalogV1.SearchProductSuggestionsResponse{Items: items}, nil
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
