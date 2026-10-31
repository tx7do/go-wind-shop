package service

import (
	"context"
	"strconv"

	"go-wind-shop/pkg/logx"
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
	log *logx.Helper
	productSearchRepo  *data.ProductSearchRepo
	productRepo        *data.ProductRepo
}

func NewProductSearchService(
	ctx *bootstrap.Context,
	productSearchRepo *data.ProductSearchRepo,
	productRepo *data.ProductRepo,
) *ProductSearchService {
	return &ProductSearchService{
		log:               logx.New(ctx.NewLoggerHelper("product-search/service/core-service")),
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
		return s.reindexProductByID(ctx, payload.ID)

	default:
		s.log.Warnf("reindex product: unknown op %s", payload.Op)
		return nil
	}
}

// reindexProductByID 从 DB 取商品最新数据 upsert 到 ES（"index" op 的实体）。
// docs 为空（商品不存在/非 ACTIVE/无有效翻译）时改为清理 ES 残留文档。
func (s *ProductSearchService) reindexProductByID(ctx context.Context, productID uint32) error {
	docs, err := s.productRepo.GetReindexDocuments(ctx, productID)
	if err != nil {
		s.log.Errorf("reindex get documents for product %d failed: %v", productID, err)
		return err
	}

	// docs 为空表示 product 不存在/非 ACTIVE
	// 此时若 ES 中有残留文档（如状态从 ACTIVE 变 DRAFT/INACTIVE），需删除
	if len(docs) == 0 {
		languages, langErr := s.productRepo.GetAvailableLanguages(ctx, productID)
		if langErr != nil {
			s.log.Errorf("reindex cleanup product %d: get languages failed: %v", productID, langErr)
			return langErr
		}
		if len(languages) == 0 {
			return nil
		}
		if err := s.productSearchRepo.DeleteProduct(ctx, productID, languages); err != nil {
			s.log.Errorf("reindex cleanup product %d failed: %v", productID, err)
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
}

// HandleReindexAll asynq 周期全量重索引任务处理器（薄委托），仿 HandleCouponSweep。
// 由 server/asynq_server.go 注册，asynq worker 按 cron 周期触发。
// 注入 SystemViewer context 以跨状态读 DB（与 ReindexProduct 同一注入模式）。
func (s *ProductSearchService) HandleReindexAll(taskType string, taskData *task.SearchReindexAllTaskData) error {
	s.log.Infow(
		"msg", "HandleReindexAll started",
		"task_type", taskType,
	)
	ctx := appViewer.NewSystemViewerContext(context.Background())
	return s.ReindexAll(ctx)
}

// ReindexAll 周期全量重索引：以 DB 为准双向收敛 ES 索引（兜底修复路径）。
//
// 正向：全部 ACTIVE 商品逐个 upsert——修复 SQL 直插的种子数据、入队失败/丢失的
// search.reindex 单条任务等场景导致的"DB 有、索引无"。
// 反向：清扫 ES 中不在 DB 期望集合（ListActiveProductLanguagePairs，与
// GetReindexDocuments 同口径）的孤儿文档——修复按 Id 删翻译等提取不到
// product_id 而漏发任务的场景导致的"索引有、DB 无"。
//
// 容错：单个商品 upsert 失败不中断整体（记日志继续，漏的下轮补）；孤儿清扫
// 失败同样仅日志。孤儿判定只依赖 DB 期望集合，与 upsert 成败无关（upsert 失败
// 的商品其文档本就在期望集合内，不会被误删）。仅期望集合扫描（DB）失败才返回
// error 交由 asynq 重试。
func (s *ProductSearchService) ReindexAll(ctx context.Context) error {
	pairs, err := s.productRepo.ListActiveProductLanguagePairs(ctx)
	if err != nil {
		return err
	}

	// 正向：按商品维度逐个 upsert（去重，商品多语言共享一次 Get 查询）
	pidSet := make(map[uint32]struct{}, len(pairs))
	for _, p := range pairs {
		pidSet[p.ProductID] = struct{}{}
	}

	var indexed, failed int
	for pid := range pidSet {
		if err := s.reindexProductByID(ctx, pid); err != nil {
			failed++
			s.log.Errorf("reindex all: index product %d failed: %v", pid, err)
			continue
		}
		indexed++
	}

	// 反向：孤儿清扫
	if cleanErr := s.cleanupOrphanDocuments(ctx, pairs); cleanErr != nil {
		s.log.Errorf("reindex all: cleanup orphan documents failed: %v", cleanErr)
	}

	s.log.Infof("reindex all done: active_products=%d indexed=%d failed=%d",
		len(pidSet), indexed, failed)
	return nil
}

// cleanupOrphanDocuments 删除 ES 中不在 DB 期望集合内的孤儿文档。
// 孤儿按 product 分组后复用 DeleteProduct 的枚举删除。
func (s *ProductSearchService) cleanupOrphanDocuments(ctx context.Context, expected []data.ProductLanguagePair) error {
	expectedSet := make(map[string]struct{}, len(expected))
	for _, p := range expected {
		docID := strconv.FormatUint(uint64(p.ProductID), 10) + "_" + p.Language
		expectedSet[docID] = struct{}{}
	}

	refs, err := s.productSearchRepo.ListAllProductDocRefs(ctx)
	if err != nil {
		return err
	}

	orphans := make(map[uint32][]string)
	for _, ref := range refs {
		docID := ref.ProductID + "_" + ref.Language
		if _, ok := expectedSet[docID]; ok {
			continue
		}
		// product_id 非数字的脏数据不碰（保守，人工排查）
		pid, pErr := strconv.ParseUint(ref.ProductID, 10, 32)
		if pErr != nil {
			s.log.Warnf("cleanup orphans: skip malformed doc product_id=%q lang=%q", ref.ProductID, ref.Language)
			continue
		}
		orphans[uint32(pid)] = append(orphans[uint32(pid)], ref.Language)
	}

	for pid, langs := range orphans {
		if err := s.productSearchRepo.DeleteProduct(ctx, pid, langs); err != nil {
			s.log.Errorf("cleanup orphans: delete product %d langs=%v failed: %v", pid, langs, err)
			continue
		}
		s.log.Infof("cleanup orphans: deleted docs for product_id=%d langs=%v", pid, langs)
	}
	return nil
}

// SearchSuggestions 搜索联想（输入联想，前台免鉴权）。
//
// 安全：status 硬编码 PRODUCT_STATUS_ACTIVE；language 缺省 zh-CN；size 缺省 8、上限 20。
func (s *ProductSearchService) SearchSuggestions(
	ctx context.Context,
	query string,
	language string,
	size int,
) ([]data.ProductSuggestion, error) {
	const status = "PRODUCT_STATUS_ACTIVE"
	if language == "" {
		language = "zh-CN"
	}
	if size <= 0 {
		size = 8
	}
	return s.productSearchRepo.SuggestProductNames(ctx, query, language, status, size)
}
