package data

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
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
	log       *log.Helper

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
		log:         ctx.NewLoggerHelper("product/repo/core-service"),
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
	// 通用分页框架前，先把 name 关键字从 query JSON 里剥离出来，改用子查询
	// products.id IN (SELECT product_id FROM translations WHERE name ILIKE %q%)
	// 实现（跨任意语言命中）。剥离后剩余字段（categoryId/brandId 等）仍走 DSL。
	if nameKeyword := extractAndStripNameKeyword(req); nameKeyword != "" {
		like := "%" + strings.ToLower(nameKeyword) + "%"
		builder.Modify(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"\""+product.FieldID+"\" IN (SELECT \""+productTranslation.FieldProductID+
					"\" FROM \""+productTranslation.Table+
					"\" WHERE LOWER(\""+productTranslation.FieldName+"\") LIKE ?)",
				like,
			))
		})
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

// extractAndStripNameKeyword 从分页请求的 query JSON 中取出名称搜索关键字
// （兼容 name / keyword / q 三个前端字段名），并将其从 query 中移除，
// 避免通用 DSL 再以 products.name 列处理而报错。无 query 或不含这些字段时返回空串。
func extractAndStripNameKeyword(req *paginationV1.PagingRequest) string {
	if req == nil || req.GetQuery() == "" {
		return ""
	}
	raw := req.GetQuery()
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return ""
	}
	var keyword string
	for _, key := range []string{"name", "keyword", "q"} {
		if v, ok := m[key]; ok {
			if s, ok2 := v.(string); ok2 && strings.TrimSpace(s) != "" {
				keyword = strings.TrimSpace(s)
			}
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
