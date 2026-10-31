package data

import (
	"go-wind-shop/pkg/logx"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	// 搜索后端为 OpenSearch（与 go-wind-cms 一致），沿用 elasticsearchCrud 别名减小 diff
	elasticsearchCrud "github.com/tx7do/go-crud/opensearch"
)

// ============================================================================
// OpenSearch 商品搜索 Repo —— products 索引的 OpenSearch 操作封装
//
// 与 go-wind-cms 的 search_repo.go 的关键差异：
//
//   1. 商品是全局共享目录（mall_products 无 TenantID mixin），文档无
//      tenant_id 字段。搜索过滤只有 status + language 两个 term，无 tenant_id。
//   2. OpenSearch wrapper（go-crud/opensearch）无 delete-by-query，只有
//      DeleteDocument(indexName, docID)。删除需枚举已索引语言逐个删，
//      docID = {product_id}_{language}。
//   3. 搜索用 wrapper 的 SearchWithHighlight（JSON query map + sourceFields
//      限制），功能等价于 CMS 的 raw OpenSearch DSL。
//
// 安全模型：
//   - 搜索 DSL 必带 bool.filter 的 term{status} + term{language}，调用方
//     无法覆盖这两个过滤条件
//   - status 在 service 层硬编码为 PRODUCT_STATUS_ACTIVE（前台只显示上架）
//   - WithSource 限制 ES 只回传 product_id / language / name
//   - 索引路径跳过非 ACTIVE 状态商品（不入索引）
// ============================================================================

const (
	productSearchIndexName     = "products"
	productSearchTemplateName  = "products_template"
	productSearchTemplatePrio  = 100
	maxProductSearchPageSize   = 50
	maxProductSearchResultFrom = 10000
)

// ProductDocument 是写入 ES products 索引的文档结构。
// 无 tenant_id（商品全局共享）。status 固定为 ACTIVE（非 ACTIVE 不入索引）。
type ProductDocument struct {
	ProductID         string `json:"product_id"`         // 必填，keyword
	Language          string `json:"language"`           // 必填，keyword，搜索强制 term 过滤
	Status            string `json:"status"`             // 必填，keyword，搜索强制 term 过滤（前台仅 ACTIVE）
	Name              string `json:"name"`               // text + smartcn，全文检索字段
	ShortDescription  string `json:"short_description"`  // text + smartcn，全文检索字段
	LongDescription   string `json:"long_description"`   // text + smartcn，全文检索字段
	ImageURL          string `json:"image_url"`          // keyword，locale 无关主图 URL（展示用，非检索字段）
}

// ProductSuggestion 搜索联想候选项（最小字段集）。
type ProductSuggestion struct {
	ProductId uint32 `json:"product_id"`
	Name      string `json:"name"`
	ImageUrl  string `json:"image_url"`
}

// ProductDocRef 是扫描索引得到的文档坐标（全量重索引的孤儿清扫用）。
type ProductDocRef struct {
	ProductID string
	Language  string
}

// ProductSearchHit 是搜索结果中的单条命中，只暴露最小字段集。
type ProductSearchHit struct {
	ProductID string
	Language string
	Name     string
	ImageURL string
	Score    float64
}

// ProductSearchResult 是搜索返回。
type ProductSearchResult struct {
	Total int
	Hits  []ProductSearchHit
}

type ProductSearchRepo struct {
	esClient *elasticsearchCrud.Client
	log *logx.Helper
}

func NewProductSearchRepo(ctx *bootstrap.Context, esClient *elasticsearchCrud.Client) *ProductSearchRepo {
	return &ProductSearchRepo{
		log:      logx.New(ctx.NewLoggerHelper("product-search/repo/core-service")),
		esClient: esClient,
	}
}

// EnsureIndexTemplate 幂等创建 products 索引模板。
// 模板绑定 smartcn 分词器到 name/short_description/long_description，
// 并定义 keyword 字段（product_id/language/status）。
func (r *ProductSearchRepo) EnsureIndexTemplate(ctx context.Context) error {
	if r.esClient == nil {
		return errors.New("opensearch client is nil")
	}

	properties := map[string]any{
		"product_id":        map[string]any{"type": "keyword"},
		"language":          map[string]any{"type": "keyword"},
		"status":            map[string]any{"type": "keyword"},
		"name":              map[string]any{"type": "text", "analyzer": "smartcn"},
		"short_description": map[string]any{"type": "text", "analyzer": "smartcn"},
		"long_description":  map[string]any{"type": "text", "analyzer": "smartcn"},
		"image_url":         map[string]any{"type": "keyword"},
	}

	templateBody := map[string]any{
		"index_patterns": []string{productSearchIndexName},
		"priority":       productSearchTemplatePrio,
		"template": map[string]any{
			"mappings": map[string]any{
				"dynamic":    false,
				"properties": properties,
			},
		},
	}

	bodyBytes, err := json.Marshal(templateBody)
	if err != nil {
		r.log.Errorf("marshal index template body failed: %v", err)
		return err
	}

	if err := r.esClient.CreateIndexTemplate(ctx, productSearchTemplateName, string(bodyBytes)); err != nil {
		r.log.Errorf("create index template failed: %v", err)
		return err
	}

	r.log.Infof("index template %s ensured for index %s", productSearchTemplateName, productSearchIndexName)
	return nil
}

// IndexProduct 将一个商品的某个语言翻译 upsert 到 ES。
// 文档 id = {product_id}_{language}，同一商品的每种语言各一个 ES 文档。
func (r *ProductSearchRepo) IndexProduct(ctx context.Context, doc *ProductDocument) error {
	if r.esClient == nil {
		return errors.New("opensearch client is nil")
	}
	if doc == nil {
		return errors.New("nil product document")
	}
	if doc.ProductID == "" || doc.Language == "" {
		return errors.New("product document missing mandatory field (product_id/language)")
	}

	docID := doc.ProductID + "_" + doc.Language
	if err := r.esClient.InsertDocument(ctx, productSearchIndexName, docID, doc); err != nil {
		r.log.Errorf("index product document failed (product_id=%s lang=%s): %v", doc.ProductID, doc.Language, err)
		return err
	}
	return nil
}

// DeleteProduct 删除指定商品在 ES 中的所有已索引语言文档。
//
// ES wrapper 无 delete-by-query，只有 DeleteDocument(indexName, docID)。
// 故用传入的 languages 列表枚举，逐个按 docID = {product_id}_{language} 删除。
// languages 由 ProductRepo.ListAvailedLanguages 提供（DB 中该商品实际存在的翻译语言）。
func (r *ProductSearchRepo) DeleteProduct(ctx context.Context, productID uint32, languages []string) error {
	if r.esClient == nil {
		return errors.New("opensearch client is nil")
	}
	if productID == 0 {
		return errors.New("delete product requires non-zero productID")
	}

	pidStr := strconv.FormatUint(uint64(productID), 10)

	for _, lang := range languages {
		if lang == "" {
			continue
		}
		docID := pidStr + "_" + lang
		if err := r.esClient.DeleteDocument(ctx, productSearchIndexName, docID); err != nil {
			r.log.Errorf("delete product document failed (product_id=%s lang=%s): %v", pidStr, lang, err)
			// 继续尝试其他语言，不因单条失败中断
		}
	}

	r.log.Infof("deleted OpenSearch documents for product_id=%s (%d languages)", pidStr, len(languages))
	return nil
}

// SearchProducts 执行前台全文搜索。
//
// 安全保证：
//   - language 空 → 返回空
//   - status 空 → 返回空
//   - 查询 DSL 必带 bool.filter 的 term{status} + term{language}
//     调用方无法覆盖这两个过滤条件
//   - bool.must 的 multi_match 仅作用于 name/short_description/long_description
//   - sourceFields 限制 ES 只回传 product_id / language / name
func (r *ProductSearchRepo) SearchProducts(
	ctx context.Context,
	query string,
	language string,
	status string,
	page int,
	pageSize int,
) (*ProductSearchResult, error) {
	result := &ProductSearchResult{}

	if r.esClient == nil {
		return result, errors.New("opensearch client is nil")
	}

	// 强制不可绕过的状态/语言过滤
	if language == "" || status == "" {
		return result, nil
	}
	if query == "" {
		return result, nil
	}

	// 分页边界
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > maxProductSearchPageSize {
		pageSize = maxProductSearchPageSize
	}
	if page < 0 {
		page = 0
	}
	from := page * pageSize
	if from > maxProductSearchResultFrom {
		from = maxProductSearchResultFrom
	}

	// 构建 DSL：filter（访问控制，不参与评分）+ must（相关性评分）
	// 注意：商品无 tenant_id，故无 term{tenant_id} 过滤（与 CMS 的 posts 索引不同）
	dsl := map[string]any{
		"bool": map[string]any{
			"filter": []any{
				map[string]any{"term": map[string]any{"status": status}},
				map[string]any{"term": map[string]any{"language": language}},
			},
			"must": []any{
				map[string]any{
					"multi_match": map[string]any{
						"query":  query,
						"fields": []string{"name", "short_description", "long_description"},
					},
				},
			},
		},
	}

	// sourceFields 限制 ES 只回传最小字段集
	sourceFields := []string{"product_id", "language", "name", "image_url"}

	// 用 SearchWithHighlight 执行搜索（接受 JSON query map + sourceFields 限制）
	// highlight/sortBy 传 nil（前台搜索不启用高亮，排序按相关性评分）
	searchResult, err := r.esClient.SearchWithHighlight(
		ctx,
		productSearchIndexName,
		dsl,
		nil,
		sourceFields,
		nil,
		from,
		pageSize,
	)
	if err != nil {
		r.log.Errorf("search failed: %v", err)
		return result, err
	}

	hits := searchResult.Hits.Hits
	result.Total = searchResult.Hits.Total.Value
	result.Hits = make([]ProductSearchHit, 0, len(hits))

	for _, hit := range hits {
		// hit.Source 是 json.RawMessage，仅含 product_id/language/name/image_url（因 sourceFields 过滤）
		var src struct {
			ProductID string `json:"product_id"`
			Language  string `json:"language"`
			Name      string `json:"name"`
			ImageURL  string `json:"image_url"`
		}
		if err := json.Unmarshal(hit.Source, &src); err != nil {
			r.log.Warnf("unmarshal search hit source failed: %v", err)
			continue
		}
		result.Hits = append(result.Hits, ProductSearchHit{
			ProductID: src.ProductID,
			Language:  src.Language,
			Name:      src.Name,
			ImageURL:  src.ImageURL,
			Score:     float64(hit.Score),
		})
	}

	return result, nil
}

// ListAllProductDocRefs 扫描 products 索引，返回全部文档的 (product_id, language)。
//
// 供 ReindexAll 做孤儿清扫：ES 中存在而 DB 期望集合中没有的文档即为孤儿。
// wrapper 无 scroll/search_after，用 match_all + from/size 封顶（与搜索分页同一上限），
// 超出上限的文档本轮不参与清扫（宁可漏删不可误删，下轮继续）。
func (r *ProductSearchRepo) ListAllProductDocRefs(ctx context.Context) ([]ProductDocRef, error) {
	if r.esClient == nil {
		return nil, errors.New("opensearch client is nil")
	}

	dsl := map[string]any{"match_all": map[string]any{}}
	searchResult, err := r.esClient.SearchWithHighlight(
		ctx,
		productSearchIndexName,
		dsl,
		nil,
		[]string{"product_id", "language"},
		nil,
		0,
		maxProductSearchResultFrom,
	)
	if err != nil {
		r.log.Errorf("scan all product docs failed: %v", err)
		return nil, err
	}

	if total := searchResult.Hits.Total.Value; total > maxProductSearchResultFrom {
		r.log.Warnf("products index has %d docs, scan capped at %d, orphan cleanup partial this round",
			total, maxProductSearchResultFrom)
	}

	refs := make([]ProductDocRef, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var src struct {
			ProductID string `json:"product_id"`
			Language  string `json:"language"`
		}
		if err := json.Unmarshal(hit.Source, &src); err != nil {
			r.log.Warnf("unmarshal product doc ref failed: %v", err)
			continue
		}
		refs = append(refs, ProductDocRef{ProductID: src.ProductID, Language: src.Language})
	}
	return refs, nil
}

// SuggestProductNames 搜索联想：按商品名前缀短语匹配返回候选。
//
// DSL：filter（status/language 强制过滤，不参与评分）+ should（match_phrase_prefix
// 对 name 做前缀短语匹配，最低匹配 1 条）。与 cms 的 SuggestPostTitles 同模式。
// sourceFields 仅回传 product_id/name/image_url 最小字段集。
func (r *ProductSearchRepo) SuggestProductNames(
	ctx context.Context,
	query string,
	language string,
	status string,
	size int,
) ([]ProductSuggestion, error) {
	suggestions := make([]ProductSuggestion, 0, size)

	if r.esClient == nil {
		return suggestions, errors.New("opensearch client is nil")
	}
	if language == "" || status == "" || query == "" {
		return suggestions, nil
	}
	if size <= 0 {
		size = 8
	}
	if size > 20 {
		size = 20
	}

	dsl := map[string]any{
		"bool": map[string]any{
			"filter": []any{
				map[string]any{"term": map[string]any{"status": status}},
				map[string]any{"term": map[string]any{"language": language}},
			},
			"must": []any{
				map[string]any{
					"match_phrase_prefix": map[string]any{
						"name": map[string]any{"query": query},
					},
				},
			},
		},
	}

	sourceFields := []string{"product_id", "name", "image_url"}

	searchResult, err := r.esClient.SearchWithHighlight(
		ctx,
		productSearchIndexName,
		dsl,
		nil,
		sourceFields,
		nil,
		0,
		size,
	)
	if err != nil {
		r.log.Errorf("suggest search failed: %v", err)
		return suggestions, err
	}

	for _, hit := range searchResult.Hits.Hits {
		// product_id 在索引中是 keyword 字符串（与 SearchProducts 同一解析口径）
		var src struct {
			ProductID string `json:"product_id"`
			Name      string `json:"name"`
			ImageURL  string `json:"image_url"`
		}
		if err := json.Unmarshal(hit.Source, &src); err != nil {
			r.log.Warnf("unmarshal suggest hit failed: %v", err)
			continue
		}
		pid, pErr := strconv.ParseUint(src.ProductID, 10, 32)
		if pErr != nil {
			r.log.Warnf("parse suggest product_id [%s] failed: %v", src.ProductID, pErr)
			continue
		}
		suggestions = append(suggestions, ProductSuggestion{
			ProductId: uint32(pid),
			Name:      src.Name,
			ImageUrl:  src.ImageURL,
		})
	}

	return suggestions, nil
}
