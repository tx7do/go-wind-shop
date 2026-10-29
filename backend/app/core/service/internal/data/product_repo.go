package data

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"go-wind-shop/pkg/logx"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"
	"github.com/tx7do/go-utils/trans"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/product"
	productTranslation "go-wind-shop/app/core/service/internal/data/ent/producttranslation"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.Product, ent.Product]

	repository *entCrud.Repository[
		ent.ProductQuery, ent.ProductSelect,
		ent.ProductCreate, ent.ProductCreateBulk,
		ent.ProductUpdate, ent.ProductUpdateOne,
		ent.ProductDelete,
		predicate.Product,
		catalogV1.Product, ent.Product,
	]

	statusConverter *mapper.EnumTypeConverter[catalogV1.Product_ProductStatus, product.Status]

	productTranslationRepo *ProductTranslationRepo
}

func NewProductRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
	productTranslationRepo *ProductTranslationRepo,
) *ProductRepo {
	repo := &ProductRepo{
		entClient:   entClient,
		log:         logx.New(ctx.NewLoggerHelper("product/repo/core-service")),
		mapper:      mapper.NewCopierMapper[catalogV1.Product, ent.Product](),
		statusConverter: mapper.NewEnumTypeConverter[catalogV1.Product_ProductStatus, product.Status](
			catalogV1.Product_ProductStatus_name, catalogV1.Product_ProductStatus_value,
		),
		productTranslationRepo: productTranslationRepo,
	}

	repo.init()

	return repo
}

func (r *ProductRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductQuery, ent.ProductSelect,
		ent.ProductCreate, ent.ProductCreateBulk,
		ent.ProductUpdate, ent.ProductUpdateOne,
		ent.ProductDelete,
		predicate.Product,
		catalogV1.Product, ent.Product,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *ProductRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Product.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *ProductRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Product.Query()

	// 商品名搜索：商品名落在关联翻译表 mall_product_translations，不在 products 表，
	// 通用 DSL 会把 name 当作 products.name 列处理而报 SQL 错误。因此这里在交给
	// 通用分页框架前，先把名称条件从 query JSON 里剥离出来，改走翻译表：
	// 先查出命中的 product_id 集合，再以 id IN (...) 过滤主查询。
	//
	// 性能：LIKE '%kw%' 带前导通配符，普通 B-tree 索引无效，会全表扫描
	// mall_product_translations。数据量增大时需 DBA 按实际数据库加专用索引：
	//   PostgreSQL → CREATE EXTENSION pg_trgm; CREATE INDEX ... USING GIN (name gin_trgm_ops);
	//   MySQL      → FULLTEXT 索引（注意中文需 ngram 分词）。
	var nameFilter string
	var nameExact bool
	if contains, equals, ferr := ExtractTranslationFilters(req, "name"); ferr != nil {
		return nil, catalogV1.ErrorBadRequest("invalid query")
	} else if kw := contains["name"]; kw != "" {
		nameFilter, nameExact = kw, false
	} else if kw := equals["name"]; kw != "" {
		nameFilter, nameExact = kw, true
	}
	if kw := extractAndStripKeyword(req); kw != "" && nameFilter == "" {
		// 兼容老的裸关键字字段 keyword/q，语义等同名称模糊搜索。
		nameFilter, nameExact = kw, false
	}
	if nameFilter != "" {
		if found, terr := r.filterByTranslationName(ctx, builder, nameFilter, !nameExact); terr != nil {
			return nil, terr
		} else if !found {
			return &catalogV1.ListProductResponse{Total: 0, Items: nil}, nil
		}
	}

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListProductResponse{Total: 0, Items: nil}, nil
	}

	// 填充每个返回项的 availableLanguages 与 translations（与 Get 同逻辑）。
	for _, item := range ret.Items {
		if err := r.populateTranslations(ctx, item, ""); err != nil {
			return nil, err
		}
	}

	return &catalogV1.ListProductResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// extractAndStripKeyword 从分页请求的 query JSON 中取出裸关键字字段
// （keyword / q，旧版前端字段名），并将其从 query 中移除——这些键同样不是
// products 表的列，交给通用 DSL 会报错。返回空串表示没有这些字段。
//
// 仅当某 key 的值是有效非空字符串时才提取并从 query 删除该 key；
// 非字符串值（如 {keyword:123}）不提取、不删除，保留原样交调用方/DSL 处理，
// 避免静默吞掉过滤条件污染剩余 query。
func extractAndStripKeyword(req *paginationV1.PagingRequest) string {
	if req == nil || req.GetQuery() == "" {
		return ""
	}
	raw := req.GetQuery()
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return ""
	}
	var keyword string
	for _, key := range []string{"keyword", "q"} {
		if v, ok := m[key]; ok {
			s, ok2 := v.(string)
			if !ok2 {
				continue // 非字符串值，不处理，保留原样
			}
			if strings.TrimSpace(s) == "" {
				continue // 空值跳过
			}
			keyword = strings.TrimSpace(s)
			delete(m, key)
		}
	}
	if keyword == "" {
		return ""
	}
	// 剩余字段写回 query；若全部被剥离则置空 query，清空 oneof 过滤。
	if len(m) == 0 {
		req.FilteringType = nil
	} else if newJSON, err := json.Marshal(m); err == nil {
		req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newJSON)}
	}
	return keyword
}

// filterByTranslationName 把名称条件改写为翻译表过滤（主表无 name 列）。
// 返回 found=false 表示没有任何翻译命中，调用方可直接返回空结果。
func (r *ProductRepo) filterByTranslationName(ctx context.Context, builder *ent.ProductQuery, keyword string, contains bool) (found bool, err error) {
	q := r.entClient.Client().ProductTranslation.Query()
	if contains {
		q = q.Where(productTranslation.NameContains(keyword))
	} else {
		q = q.Where(productTranslation.NameEQ(keyword))
	}
	rows, err := q.Select(productTranslation.FieldProductID).Ints(ctx)
	if err != nil {
		r.log.Errorf("query product ids by translation name failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query translations failed")
	}
	if len(rows) == 0 {
		return false, nil
	}
	ids := make([]uint32, 0, len(rows))
	for _, n := range rows {
		ids = append(ids, uint32(n))
	}
	builder.Where(product.IDIn(ids...))
	return true, nil
}

func (r *ProductRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Product.Query().
		Where(product.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ProductRepo) Get(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.Product, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	build := r.entClient.Client().Product.Query()

	switch req.QueryBy.(type) {
	case *catalogV1.GetProductRequest_Id:
		build.Where(product.IDEQ(req.GetId()))

	default:
		return nil, catalogV1.ErrorBadRequest("invalid query field")
	}

	entity, err := build.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, catalogV1.ErrorNotFound("product not found")
		}

		r.log.Errorf("query product failed: %s", err.Error())

		return nil, catalogV1.ErrorInternalServerError("query product failed")
	}

	dto := r.mapper.ToDTO(entity)

	if err := r.populateTranslations(ctx, dto, req.GetLocale()); err != nil {
		return nil, err
	}

	return dto, nil
}

// populateTranslations 为单个 Product DTO 填充 availableLanguages 与
// translations。locale 为空串时回填全部语言，否则只回填该语言的单条
// 翻译。List 与 Get 共用此逻辑。
func (r *ProductRepo) populateTranslations(
	ctx context.Context,
	dto *catalogV1.Product,
	locale string,
) error {
	languages, err := r.productTranslationRepo.ListAvailedLanguages(ctx, dto.GetId())
	if err != nil {
		r.log.Errorf("query availed languages failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("query availed languages failed")
	}
	dto.AvailableLanguages = languages

	if locale == "" {
		translations, err := r.productTranslationRepo.ListTranslations(ctx, dto.GetId(), "", nil)
		if err != nil {
			r.log.Errorf("query translations failed: %s", err.Error())
			return catalogV1.ErrorInternalServerError("query translations failed")
		}
		dto.Translations = translations
	} else {
		translation, err := r.productTranslationRepo.GetTranslation(ctx, dto.GetId(), locale)
		if err != nil {
			r.log.Errorf("query translation failed: %s", err.Error())
			return catalogV1.ErrorInternalServerError("query translation failed")
		}
		if translation != nil {
			dto.Translations = append(dto.Translations, translation)
		}
	}

	return nil
}

func (r *ProductRepo) Create(ctx context.Context, req *catalogV1.CreateProductRequest) (dto *catalogV1.Product, err error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.Translations) == 0 {
		return nil, catalogV1.ErrorBadRequest("at least one translation is required")
	}

	var tx *ent.Tx
	tx, err = r.entClient.Client().Tx(ctx)
	if err != nil {
		r.log.Errorf("start transaction failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("start transaction failed")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.log.Errorf("transaction rollback failed: %s", rollbackErr.Error())
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			r.log.Errorf("transaction commit failed: %s", commitErr.Error())
			err = catalogV1.ErrorInternalServerError("transaction commit failed")
		}
	}()

	builder := tx.Product.Create().
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCategoryID(req.Data.CategoryId).
		SetNillableBrandID(req.Data.BrandId).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	var entity *ent.Product
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert product failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("insert product failed")
	}

	if len(req.Data.Translations) > 0 {
		if err = r.productTranslationRepo.CleanTranslations(ctx, tx, entity.ID); err != nil {
			r.log.Errorf("clean translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("clean translations failed")
		}

		for i := range req.Data.Translations {
			req.Data.Translations[i].ProductId = trans.Ptr(entity.ID)
		}

		if err = r.productTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductRepo) Update(ctx context.Context, req *catalogV1.UpdateProductRequest) (dto *catalogV1.Product, err error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return nil, err
		}
		if !exist {
			req.Data.CreatedBy = req.Data.UpdatedBy
			req.Data.UpdatedBy = nil
			_, err = r.Create(ctx, &catalogV1.CreateProductRequest{Data: req.Data})
			return nil, err
		}
	}

	var tx *ent.Tx
	tx, err = r.entClient.Client().Tx(ctx)
	if err != nil {
		r.log.Errorf("start transaction failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("start transaction failed")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.log.Errorf("transaction rollback failed: %s", rollbackErr.Error())
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			r.log.Errorf("transaction commit failed: %s", commitErr.Error())
			err = catalogV1.ErrorInternalServerError("transaction commit failed")
		}
	}()

	if len(req.Data.Translations) > 0 {
		for i := range req.Data.Translations {
			req.Data.Translations[i].ProductId = trans.Ptr(req.GetId())
		}

		// 与 Create 同策略：先清空旧翻译再批量插入（翻译表唯一键，直接插入会撞键）。
		if err = r.productTranslationRepo.CleanTranslations(ctx, tx, req.GetId()); err != nil {
			r.log.Errorf("clean translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("clean translations failed")
		}

		if err = r.productTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	builder := tx.Product.UpdateOneID(req.GetId())
	result, err := r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.Product) {
			builder.
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCategoryID(req.Data.CategoryId).
				SetNillableBrandID(req.Data.BrandId).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(product.FieldID, req.GetId()))
		},
	)

	return result, err
}

func (r *ProductRepo) Delete(ctx context.Context, req *catalogV1.DeleteProductRequest) (err error) {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	var tx *ent.Tx
	tx, err = r.entClient.Client().Tx(ctx)
	if err != nil {
		r.log.Errorf("start transaction failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("start transaction failed")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.log.Errorf("transaction rollback failed: %s", rollbackErr.Error())
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			r.log.Errorf("transaction commit failed: %s", commitErr.Error())
			err = catalogV1.ErrorInternalServerError("transaction commit failed")
		}
	}()

	// 删除商品数据
	if err = r.entClient.Client().Product.
		DeleteOneID(req.GetId()).
		Exec(ctx); err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
	}

	// 删除关联翻译数据
	if err = r.productTranslationRepo.CleanTranslations(ctx, tx, req.GetId()); err != nil {
		r.log.Errorf("clean translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("clean translations failed")
	}

	return err
}

func (r *ProductRepo) TranslationExists(ctx context.Context, productId uint32, languageCode string) (bool, error) {
	return r.productTranslationRepo.TranslationExists(ctx, productId, languageCode)
}

func (r *ProductRepo) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetProductId() == 0 {
		return nil, catalogV1.ErrorBadRequest("product id is required")
	}

	return r.productTranslationRepo.CreateTranslation(ctx, req.Data)
}

func (r *ProductRepo) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetProductId() == 0 {
		return nil, catalogV1.ErrorBadRequest("product id is required")
	}

	if exist, err := r.TranslationExists(ctx, req.Data.GetProductId(), req.Data.GetLanguageCode()); err != nil {
		return nil, err
	} else if !exist {
		return nil, catalogV1.ErrorNotFound("translation not found")
	}

	return r.productTranslationRepo.UpdateTranslation(ctx, req.GetId(), req.Data, req.GetUpdateMask())
}

func (r *ProductRepo) GetTranslation(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.ProductTranslation, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	return r.productTranslationRepo.GetTranslation(ctx, req.GetId(), req.GetLocale())
}

func (r *ProductRepo) ListTranslations(ctx context.Context, productId uint32) ([]*catalogV1.ProductTranslation, error) {
	return r.productTranslationRepo.ListTranslations(ctx, productId, "", nil)
}

func (r *ProductRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductTranslationRequest) error {
	return r.productTranslationRepo.DeleteTranslation(ctx, req)
}

func (r *ProductRepo) CleanTranslations(ctx context.Context, tx *ent.Tx, productID uint32) error {
	return r.productTranslationRepo.CleanTranslations(ctx, tx, productID)
}

// ============================================================================
// Elasticsearch 重索引辅助方法
//
// 仅供 ProductSearchService.ReindexProduct 调用。该方法的 ctx 由
// ProductSearchService 注入 SystemViewer（跨状态读 DB 的特权上下文）。
//
// 与 go-wind-cms 的差异：商品是全局共享目录（无 TenantID mixin），故
// ProductReindexDocument 无 tenant_id 字段，ES 文档也无 tenant_id。
// ============================================================================

// ProductReindexDocument 是 ProductRepo 返回给 ProductSearchService 的单条
// ES 文档数据。无 tenant_id（商品全局共享）。
type ProductReindexDocument struct {
	ProductID        uint32
	Language         string
	Status           string
	Name             string
	ShortDescription string
	LongDescription  string
	ImageURL         string
}

// GetReindexDocuments 取指定商品及其所有翻译，组装成 ES 文档数据。
//
// 安全：
//   - 跳过非 ACTIVE 状态的商品（不入索引）
//   - 跳过 name 和 descriptions 均空的翻译（避免索引无意义文档）
//   - 调用方须以 SystemViewer ctx 调用，方能跨状态读取
func (r *ProductRepo) GetReindexDocuments(ctx context.Context, productID uint32) ([]ProductReindexDocument, error) {
	if productID == 0 {
		return nil, catalogV1.ErrorBadRequest("invalid product id")
	}

	// 用 Get 加载商品 DTO（含全部翻译，locale=""）
	req := &catalogV1.GetProductRequest{
		QueryBy: &catalogV1.GetProductRequest_Id{Id: productID},
		Locale:  nil,
	}
	dto, err := r.Get(ctx, req)
	if err != nil {
		if catalogV1.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("query product for reindex failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query product for reindex failed")
	}

	// 仅索引 ACTIVE 状态
	if dto.GetStatus() != catalogV1.Product_PRODUCT_STATUS_ACTIVE {
		return nil, nil
	}

	// 取所有翻译
	translations := dto.GetTranslations()
	if len(translations) == 0 {
		return nil, nil
	}

	docs := make([]ProductReindexDocument, 0, len(translations))
	for _, tr := range translations {
		if tr == nil {
			continue
		}
		// 跳过 name 和 descriptions 均空的翻译
		if tr.GetName() == "" && tr.GetShortDescription() == "" && tr.GetLongDescription() == "" {
			continue
		}
		docs = append(docs, ProductReindexDocument{
			ProductID:        productID,
			Language:         tr.GetLanguageCode(),
			Status:           "PRODUCT_STATUS_ACTIVE",
			Name:             tr.GetName(),
			ShortDescription: tr.GetShortDescription(),
			LongDescription:  tr.GetLongDescription(),
			ImageURL:         dto.GetImageUrl(),
		})
	}

	return docs, nil
}

// ListActiveProductIDs 列出所有 ACTIVE 状态商品的 ID。
// 供 ProductSearchService.ReindexAll 周期全量重索引使用。
func (r *ProductRepo) ListActiveProductIDs(ctx context.Context) ([]uint32, error) {
	entities, err := r.entClient.Client().Product.Query().
		Select(product.FieldID).
		All(ctx)
	if err != nil {
		r.log.Errorf("list product ids for reindex failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("list product ids for reindex failed")
	}

	ids := make([]uint32, 0, len(entities))
	for _, e := range entities {
		if e == nil {
			continue
		}
		// 检查状态是否为 ACTIVE
		if e.Status != nil && string(*e.Status) == "PRODUCT_STATUS_ACTIVE" {
			ids = append(ids, e.ID)
		}
	}
	return ids, nil
}

// GetAvailableLanguages 返回指定商品在 DB 中实际存在的翻译语言列表。
// 供 ProductSearchService.ReindexProduct 的 delete 操作枚举删除使用。
func (r *ProductRepo) GetAvailableLanguages(ctx context.Context, productID uint32) ([]string, error) {
	return r.productTranslationRepo.ListAvailedLanguages(ctx, productID)
}
