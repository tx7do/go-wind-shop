package service

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-shop/app/core/service/internal/data"

	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"
)

// ============================================================================
// ProductSearchService —— Elasticsearch 商品搜索与重索引的业务编排层
//
// 职责：
//   - Search：前台全文搜索入口，强制 status=ACTIVE + language 过滤，不接受 bypass
//   - ReindexProduct：asynq worker handler，从 DB 取数据写入/删除 ES
//
// 依赖：
//   - productSearchRepo：ES 操作（强制隔离）
//   - productRepo：reindex 时读 DB
//
// 与 go-wind-cms SearchService 的差异：
//   - 商品无 TenantID mixin（全局共享目录），故无租户提取、无 tenant_id 过滤
//   - status 在 service 层硬编码为 PRODUCT_STATUS_ACTIVE（客户端不可覆盖）
//   - 删除用枚举语言逐个 DeleteDocument（ES wrapper 无 delete-by-query）
//
// 安全模型详见 data/product_search_repo.go 顶部注释。核心差异：
//   - 搜索路径：无租户上下文要求，但 status 强制 ACTIVE、language 强制 term 过滤
//   - reindex 路径：注入 SystemViewer 跨状态读 DB
// ============================================================================

type ProductSearchService struct {
	log                *log.Helper
	productSearchRepo  *data.ProductSearchRepo
	productRepo        *data.ProductRepo
}

func NewProductSearchService(
	ctx *bootstrap.Context,
	productSearchRepo *data.ProductSearchRepo,
	productRepo *data.ProductRepo,
) *ProductSearchService {
	return &ProductSearchService{
		log:               ctx.NewLoggerHelper("product-search/service/core-service"),
		productSearchRepo: productSearchRepo,
		productRepo:       productRepo,
	}
}

// Search 前台全文搜索。
//
// 安全：
//   - status 硬编码为 PRODUCT_STATUS_ACTIVE，调用方无法覆盖
//   - language 由调用方传，但 SearchRepo 内部强制 term 过滤，不可绕过
//   - 商品无租户隔离，故无 tenant_id 过滤
func (s *ProductSearchService) Search(
	ctx context.Context,
	query string,
	language string,
	page int,
	pageSize int,
) (*data.ProductSearchResult, error) {
	// status 硬编码，客户端不可覆盖
	const status = "PRODUCT_STATUS_ACTIVE"

	return s.productSearchRepo.SearchProducts(ctx, query, language, status, page, pageSize)
}

// ReindexProduct 是 asynq "search.reindex" 任务的 worker handler。
//
// 签名遵循 (taskType string, payload *T) error 模式（参考 TaskService.AsyncBackup）。
//
// 安全：
//   - 注入 SystemViewer 跨状态读 DB（特权）
//   - ES 文档无 tenant_id（商品全局共享）
//   - 非 ACTIVE / 空翻译 → 跳过或删除
func (s *ProductSearchService) ReindexProduct(taskType string, payload *task.SearchReindexPayload) error {
	if payload == nil {
		s.log.Warnf("reindex product: nil payload")
		return nil
	}

	s.log.Infof("reindex product: entity=%s id=%d op=%s",
		payload.Entity, payload.ID, payload.Op)

	if payload.Entity != "product" {
		s.log.Warnf("reindex product: unsupported entity %s", payload.Entity)
		return nil
	}

	// 注入 SystemViewer：跨状态读 DB 的特权上下文
	ctx := appViewer.NewSystemViewerContext(context.Background())

	switch payload.Op {
	case "delete":
		// product 被删除（软删/硬删/状态变更），按已索引语言枚举删 ES 文档
		languages, err := s.productRepo.GetAvailableLanguages(ctx, payload.ID)
		if err != nil {
			s.log.Errorf("reindex delete product %d: get languages failed: %v", payload.ID, err)
			return err
		}
		if len(languages) == 0 {
			s.log.Infof("reindex delete product %d: no languages to delete", payload.ID)
			return nil
		}
		if err := s.productSearchRepo.DeleteProduct(ctx, payload.ID, languages); err != nil {
			s.log.Errorf("reindex delete product %d failed: %v", payload.ID, err)
			return err
		}
		return nil

	case "index":
		// product 创建/更新，从 DB 取最新数据写入 ES
		docs, err := s.productRepo.GetReindexDocuments(ctx, payload.ID)
		if err != nil {
			s.log.Errorf("reindex get documents for product %d failed: %v", payload.ID, err)
			return err
		}

		// docs 为空表示 product 不存在/非 ACTIVE
		// 此时若 ES 中有残留文档（如状态从 ACTIVE 变 DRAFT/INACTIVE），需删除
		if len(docs) == 0 {
			languages, langErr := s.productRepo.GetAvailableLanguages(ctx, payload.ID)
			if langErr != nil {
				s.log.Errorf("reindex cleanup product %d: get languages failed: %v", payload.ID, langErr)
				return langErr
			}
			if len(languages) == 0 {
				return nil
			}
			if err := s.productSearchRepo.DeleteProduct(ctx, payload.ID, languages); err != nil {
				s.log.Errorf("reindex cleanup product %d failed: %v", payload.ID, err)
				return err
			}
			return nil
		}

		// 确保 products 索引模板存在（幂等）
		if err := s.productSearchRepo.EnsureIndexTemplate(ctx); err != nil {
			s.log.Errorf("ensure index template failed: %v", err)
			return err
		}

		// 逐条 upsert ES 文档（每个语言一个文档）
		for i := range docs {
			doc := &data.ProductDocument{
				ProductID:        strconv.FormatUint(uint64(docs[i].ProductID), 10),
				Language:         docs[i].Language,
				Status:           docs[i].Status,
				Name:             docs[i].Name,
				ShortDescription: docs[i].ShortDescription,
				LongDescription:  docs[i].LongDescription,
				ImageURL:         docs[i].ImageURL,
			}
			if err := s.productSearchRepo.IndexProduct(ctx, doc); err != nil {
				s.log.Errorf("index product document %d lang=%s failed: %v",
					docs[i].ProductID, docs[i].Language, err)
				return err
			}
		}
		return nil

	default:
		s.log.Warnf("reindex product: unknown op %s", payload.Op)
		return nil
	}
}
