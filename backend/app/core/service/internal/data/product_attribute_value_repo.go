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
	"go-wind-shop/app/core/service/internal/data/ent/productattributevalue"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductAttributeValueRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.ProductAttributeValue, ent.ProductAttributeValue]

	repository *entCrud.Repository[
		ent.ProductAttributeValueQuery, ent.ProductAttributeValueSelect,
		ent.ProductAttributeValueCreate, ent.ProductAttributeValueCreateBulk,
		ent.ProductAttributeValueUpdate, ent.ProductAttributeValueUpdateOne,
		ent.ProductAttributeValueDelete,
		predicate.ProductAttributeValue,
		catalogV1.ProductAttributeValue, ent.ProductAttributeValue,
	]

	productAttributeValueTranslationRepo *ProductAttributeValueTranslationRepo
}

func NewProductAttributeValueRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
	productAttributeValueTranslationRepo *ProductAttributeValueTranslationRepo,
) *ProductAttributeValueRepo {
	repo := &ProductAttributeValueRepo{
		entClient:                            entClient,
		log:                                  ctx.NewLoggerHelper("product-attribute-value/repo/core-service"),
		mapper:                               mapper.NewCopierMapper[catalogV1.ProductAttributeValue, ent.ProductAttributeValue](),
		productAttributeValueTranslationRepo: productAttributeValueTranslationRepo,
	}

	repo.init()

	return repo
}

func (r *ProductAttributeValueRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductAttributeValueQuery, ent.ProductAttributeValueSelect,
		ent.ProductAttributeValueCreate, ent.ProductAttributeValueCreateBulk,
		ent.ProductAttributeValueUpdate, ent.ProductAttributeValueUpdateOne,
		ent.ProductAttributeValueDelete,
		predicate.ProductAttributeValue,
		catalogV1.ProductAttributeValue, ent.ProductAttributeValue,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *ProductAttributeValueRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().ProductAttributeValue.Query()
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

func (r *ProductAttributeValueRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeValueResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ProductAttributeValue.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListProductAttributeValueResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListProductAttributeValueResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *ProductAttributeValueRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().ProductAttributeValue.Query().
		Where(productattributevalue.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ProductAttributeValueRepo) Get(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValue, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	build := r.entClient.Client().ProductAttributeValue.Query()

	switch req.QueryBy.(type) {
	case *catalogV1.GetProductAttributeValueRequest_Id:
		build.Where(productattributevalue.IDEQ(req.GetId()))

	default:
		return nil, catalogV1.ErrorBadRequest("invalid query field")
	}

	entity, err := build.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, catalogV1.ErrorNotFound("product attribute value not found")
		}

		r.log.Errorf("query product attribute value failed: %s", err.Error())

		return nil, catalogV1.ErrorInternalServerError("query product attribute value failed")
	}

	dto := r.mapper.ToDTO(entity)

	languages, err := r.productAttributeValueTranslationRepo.ListAvailedLanguages(ctx, dto.GetId())
	if err != nil {
		r.log.Errorf("query availed languages failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query availed languages failed")
	}
	dto.AvailableLanguages = languages

	if req.Locale == nil {
		translations, err := r.productAttributeValueTranslationRepo.ListTranslations(ctx, dto.GetId(), "", nil)
		if err != nil {
			r.log.Errorf("query translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("query translations failed")
		}
		dto.Translations = translations
	} else {
		translation, err := r.productAttributeValueTranslationRepo.GetTranslation(ctx, dto.GetId(), req.GetLocale())
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

func (r *ProductAttributeValueRepo) Create(ctx context.Context, req *catalogV1.CreateProductAttributeValueRequest) (dto *catalogV1.ProductAttributeValue, err error) {
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

	builder := tx.ProductAttributeValue.Create().
		SetNillableAttributeID(req.Data.AttributeId).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	var entity *ent.ProductAttributeValue
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert product attribute value failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("insert product attribute value failed")
	}

	if len(req.Data.Translations) > 0 {
		if err = r.productAttributeValueTranslationRepo.CleanTranslations(ctx, tx, entity.ID); err != nil {
			r.log.Errorf("clean translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("clean translations failed")
		}

		for i := range req.Data.Translations {
			req.Data.Translations[i].ValueId = trans.Ptr(entity.ID)
		}

		if err = r.productAttributeValueTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductAttributeValueRepo) Update(ctx context.Context, req *catalogV1.UpdateProductAttributeValueRequest) (dto *catalogV1.ProductAttributeValue, err error) {
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
			_, err = r.Create(ctx, &catalogV1.CreateProductAttributeValueRequest{Data: req.Data})
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
			req.Data.Translations[i].ValueId = trans.Ptr(req.GetId())
		}

		if err = r.productAttributeValueTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	builder := tx.ProductAttributeValue.UpdateOneID(req.GetId())
	result, err := r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.ProductAttributeValue) {
			builder.
				SetNillableAttributeID(req.Data.AttributeId).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(productattributevalue.FieldID, req.GetId()))
		},
	)

	return result, err
}

func (r *ProductAttributeValueRepo) Delete(ctx context.Context, req *catalogV1.DeleteProductAttributeValueRequest) (err error) {
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

	// 删除商品属性值数据
	if err = r.entClient.Client().ProductAttributeValue.
		DeleteOneID(req.GetId()).
		Exec(ctx); err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
	}

	// 删除关联翻译数据
	if err = r.productAttributeValueTranslationRepo.CleanTranslations(ctx, tx, req.GetId()); err != nil {
		r.log.Errorf("clean translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("clean translations failed")
	}

	return err
}

func (r *ProductAttributeValueRepo) TranslationExists(ctx context.Context, valueId uint32, languageCode string) (bool, error) {
	return r.productAttributeValueTranslationRepo.TranslationExists(ctx, valueId, languageCode)
}

func (r *ProductAttributeValueRepo) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductAttributeValueTranslationRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetValueId() == 0 {
		return nil, catalogV1.ErrorBadRequest("value id is required")
	}

	return r.productAttributeValueTranslationRepo.CreateTranslation(ctx, req.Data)
}

func (r *ProductAttributeValueRepo) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductAttributeValueTranslationRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetValueId() == 0 {
		return nil, catalogV1.ErrorBadRequest("value id is required")
	}

	if exist, err := r.TranslationExists(ctx, req.Data.GetValueId(), req.Data.GetLanguageCode()); err != nil {
		return nil, err
	} else if !exist {
		return nil, catalogV1.ErrorNotFound("translation not found")
	}

	return r.productAttributeValueTranslationRepo.UpdateTranslation(ctx, req.GetId(), req.Data, req.GetUpdateMask())
}

func (r *ProductAttributeValueRepo) GetTranslation(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	return r.productAttributeValueTranslationRepo.GetTranslation(ctx, req.GetId(), req.GetLocale())
}

func (r *ProductAttributeValueRepo) ListTranslations(ctx context.Context, valueId uint32) ([]*catalogV1.ProductAttributeValueTranslation, error) {
	return r.productAttributeValueTranslationRepo.ListTranslations(ctx, valueId, "", nil)
}

func (r *ProductAttributeValueRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeValueTranslationRequest) error {
	return r.productAttributeValueTranslationRepo.DeleteTranslation(ctx, req)
}

func (r *ProductAttributeValueRepo) CleanTranslations(ctx context.Context, tx *ent.Tx, valueID uint32) error {
	return r.productAttributeValueTranslationRepo.CleanTranslations(ctx, tx, valueID)
}
