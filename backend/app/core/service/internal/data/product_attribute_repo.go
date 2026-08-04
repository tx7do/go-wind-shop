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
	"go-wind-shop/app/core/service/internal/data/ent/productattribute"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductAttributeRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.ProductAttribute, ent.ProductAttribute]

	repository *entCrud.Repository[
		ent.ProductAttributeQuery, ent.ProductAttributeSelect,
		ent.ProductAttributeCreate, ent.ProductAttributeCreateBulk,
		ent.ProductAttributeUpdate, ent.ProductAttributeUpdateOne,
		ent.ProductAttributeDelete,
		predicate.ProductAttribute,
		catalogV1.ProductAttribute, ent.ProductAttribute,
	]

	productAttributeTranslationRepo *ProductAttributeTranslationRepo
}

func NewProductAttributeRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
	productAttributeTranslationRepo *ProductAttributeTranslationRepo,
) *ProductAttributeRepo {
	repo := &ProductAttributeRepo{
		entClient:                       entClient,
		log:                             ctx.NewLoggerHelper("product-attribute/repo/core-service"),
		mapper:                          mapper.NewCopierMapper[catalogV1.ProductAttribute, ent.ProductAttribute](),
		productAttributeTranslationRepo: productAttributeTranslationRepo,
	}

	repo.init()

	return repo
}

func (r *ProductAttributeRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductAttributeQuery, ent.ProductAttributeSelect,
		ent.ProductAttributeCreate, ent.ProductAttributeCreateBulk,
		ent.ProductAttributeUpdate, ent.ProductAttributeUpdateOne,
		ent.ProductAttributeDelete,
		predicate.ProductAttribute,
		catalogV1.ProductAttribute, ent.ProductAttribute,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *ProductAttributeRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().ProductAttribute.Query()
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

func (r *ProductAttributeRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ProductAttribute.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListProductAttributeResponse{Total: 0, Items: nil}, nil
	}

	// 填充每个返回项的 availableLanguages 与 translations（与 Get 同逻辑）。
	for _, item := range ret.Items {
		if err := r.populateTranslations(ctx, item, ""); err != nil {
			return nil, err
		}
	}

	return &catalogV1.ListProductAttributeResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *ProductAttributeRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().ProductAttribute.Query().
		Where(productattribute.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ProductAttributeRepo) Get(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttribute, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	build := r.entClient.Client().ProductAttribute.Query()

	switch req.QueryBy.(type) {
	case *catalogV1.GetProductAttributeRequest_Id:
		build.Where(productattribute.IDEQ(req.GetId()))

	default:
		return nil, catalogV1.ErrorBadRequest("invalid query field")
	}

	entity, err := build.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, catalogV1.ErrorNotFound("product attribute not found")
		}

		r.log.Errorf("query product attribute failed: %s", err.Error())

		return nil, catalogV1.ErrorInternalServerError("query product attribute failed")
	}

	dto := r.mapper.ToDTO(entity)

	if err := r.populateTranslations(ctx, dto, req.GetLocale()); err != nil {
		return nil, err
	}

	return dto, nil
}

// populateTranslations 为单个 ProductAttribute DTO 填充 availableLanguages
// 与 translations。locale 为空串时回填全部语言，否则只回填该语言的单条
// 翻译。List 与 Get 共用此逻辑。
func (r *ProductAttributeRepo) populateTranslations(
	ctx context.Context,
	dto *catalogV1.ProductAttribute,
	locale string,
) error {
	languages, err := r.productAttributeTranslationRepo.ListAvailedLanguages(ctx, dto.GetId())
	if err != nil {
		r.log.Errorf("query availed languages failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("query availed languages failed")
	}
	dto.AvailableLanguages = languages

	if locale == "" {
		translations, err := r.productAttributeTranslationRepo.ListTranslations(ctx, dto.GetId(), "", nil)
		if err != nil {
			r.log.Errorf("query translations failed: %s", err.Error())
			return catalogV1.ErrorInternalServerError("query translations failed")
		}
		dto.Translations = translations
	} else {
		translation, err := r.productAttributeTranslationRepo.GetTranslation(ctx, dto.GetId(), locale)
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

func (r *ProductAttributeRepo) Create(ctx context.Context, req *catalogV1.CreateProductAttributeRequest) (dto *catalogV1.ProductAttribute, err error) {
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

	builder := tx.ProductAttribute.Create().
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	var entity *ent.ProductAttribute
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert product attribute failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("insert product attribute failed")
	}

	if len(req.Data.Translations) > 0 {
		if err = r.productAttributeTranslationRepo.CleanTranslations(ctx, tx, entity.ID); err != nil {
			r.log.Errorf("clean translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("clean translations failed")
		}

		for i := range req.Data.Translations {
			req.Data.Translations[i].AttributeId = trans.Ptr(entity.ID)
		}

		if err = r.productAttributeTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductAttributeRepo) Update(ctx context.Context, req *catalogV1.UpdateProductAttributeRequest) (dto *catalogV1.ProductAttribute, err error) {
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
			_, err = r.Create(ctx, &catalogV1.CreateProductAttributeRequest{Data: req.Data})
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
			req.Data.Translations[i].AttributeId = trans.Ptr(req.GetId())
		}

		if err = r.productAttributeTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	builder := tx.ProductAttribute.UpdateOneID(req.GetId())
	result, err := r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.ProductAttribute) {
			builder.
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(productattribute.FieldID, req.GetId()))
		},
	)

	return result, err
}

func (r *ProductAttributeRepo) Delete(ctx context.Context, req *catalogV1.DeleteProductAttributeRequest) (err error) {
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

	// 删除商品属性数据
	if err = r.entClient.Client().ProductAttribute.
		DeleteOneID(req.GetId()).
		Exec(ctx); err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
	}

	// 删除关联翻译数据
	if err = r.productAttributeTranslationRepo.CleanTranslations(ctx, tx, req.GetId()); err != nil {
		r.log.Errorf("clean translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("clean translations failed")
	}

	return err
}

func (r *ProductAttributeRepo) TranslationExists(ctx context.Context, attributeId uint32, languageCode string) (bool, error) {
	return r.productAttributeTranslationRepo.TranslationExists(ctx, attributeId, languageCode)
}

func (r *ProductAttributeRepo) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductAttributeTranslationRequest) (*catalogV1.ProductAttributeTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetAttributeId() == 0 {
		return nil, catalogV1.ErrorBadRequest("attribute id is required")
	}

	return r.productAttributeTranslationRepo.CreateTranslation(ctx, req.Data)
}

func (r *ProductAttributeRepo) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductAttributeTranslationRequest) (*catalogV1.ProductAttributeTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetAttributeId() == 0 {
		return nil, catalogV1.ErrorBadRequest("attribute id is required")
	}

	if exist, err := r.TranslationExists(ctx, req.Data.GetAttributeId(), req.Data.GetLanguageCode()); err != nil {
		return nil, err
	} else if !exist {
		return nil, catalogV1.ErrorNotFound("translation not found")
	}

	return r.productAttributeTranslationRepo.UpdateTranslation(ctx, req.GetId(), req.Data, req.GetUpdateMask())
}

func (r *ProductAttributeRepo) GetTranslation(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttributeTranslation, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	return r.productAttributeTranslationRepo.GetTranslation(ctx, req.GetId(), req.GetLocale())
}

func (r *ProductAttributeRepo) ListTranslations(ctx context.Context, attributeId uint32) ([]*catalogV1.ProductAttributeTranslation, error) {
	return r.productAttributeTranslationRepo.ListTranslations(ctx, attributeId, "", nil)
}

func (r *ProductAttributeRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeTranslationRequest) error {
	return r.productAttributeTranslationRepo.DeleteTranslation(ctx, req)
}

func (r *ProductAttributeRepo) CleanTranslations(ctx context.Context, tx *ent.Tx, attributeID uint32) error {
	return r.productAttributeTranslationRepo.CleanTranslations(ctx, tx, attributeID)
}
