package data

import (
	"context"
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

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListProductResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListProductResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
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

	languages, err := r.productTranslationRepo.ListAvailedLanguages(ctx, dto.GetId())
	if err != nil {
		r.log.Errorf("query availed languages failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query availed languages failed")
	}
	dto.AvailableLanguages = languages

	if req.Locale == nil {
		translations, err := r.productTranslationRepo.ListTranslations(ctx, dto.GetId(), "", nil)
		if err != nil {
			r.log.Errorf("query translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("query translations failed")
		}
		dto.Translations = translations
	} else {
		translation, err := r.productTranslationRepo.GetTranslation(ctx, dto.GetId(), req.GetLocale())
		if err != nil {
			r.log.Errorf("query translation failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("query translation failed")
		}
		if translation != nil {
			dto.Translations = append(dto.Translations, translation)
		}
	}

	return dto, nil
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
