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
	"go-wind-shop/app/core/service/internal/data/ent/brand"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type BrandRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.Brand, ent.Brand]

	repository *entCrud.Repository[
		ent.BrandQuery, ent.BrandSelect,
		ent.BrandCreate, ent.BrandCreateBulk,
		ent.BrandUpdate, ent.BrandUpdateOne,
		ent.BrandDelete,
		predicate.Brand,
		catalogV1.Brand, ent.Brand,
	]

	brandTranslationRepo *BrandTranslationRepo
}

func NewBrandRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
	brandTranslationRepo *BrandTranslationRepo,
) *BrandRepo {
	repo := &BrandRepo{
		entClient:            entClient,
		log:                  ctx.NewLoggerHelper("brand/repo/core-service"),
		mapper:               mapper.NewCopierMapper[catalogV1.Brand, ent.Brand](),
		brandTranslationRepo: brandTranslationRepo,
	}

	repo.init()

	return repo
}

func (r *BrandRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.BrandQuery, ent.BrandSelect,
		ent.BrandCreate, ent.BrandCreateBulk,
		ent.BrandUpdate, ent.BrandUpdateOne,
		ent.BrandDelete,
		predicate.Brand,
		catalogV1.Brand, ent.Brand,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *BrandRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Brand.Query()
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

func (r *BrandRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListBrandResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Brand.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListBrandResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListBrandResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *BrandRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Brand.Query().
		Where(brand.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *BrandRepo) Get(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.Brand, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	build := r.entClient.Client().Brand.Query()

	switch req.QueryBy.(type) {
	case *catalogV1.GetBrandRequest_Id:
		build.Where(brand.IDEQ(req.GetId()))

	default:
		return nil, catalogV1.ErrorBadRequest("invalid query field")
	}

	entity, err := build.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, catalogV1.ErrorNotFound("brand not found")
		}

		r.log.Errorf("query brand failed: %s", err.Error())

		return nil, catalogV1.ErrorInternalServerError("query brand failed")
	}

	dto := r.mapper.ToDTO(entity)

	languages, err := r.brandTranslationRepo.ListAvailedLanguages(ctx, dto.GetId())
	if err != nil {
		r.log.Errorf("query availed languages failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query availed languages failed")
	}
	dto.AvailableLanguages = languages

	if req.Locale == nil {
		translations, err := r.brandTranslationRepo.ListTranslations(ctx, dto.GetId(), "", nil)
		if err != nil {
			r.log.Errorf("query translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("query translations failed")
		}
		dto.Translations = translations
	} else {
		translation, err := r.brandTranslationRepo.GetTranslation(ctx, dto.GetId(), req.GetLocale())
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

func (r *BrandRepo) Create(ctx context.Context, req *catalogV1.CreateBrandRequest) (dto *catalogV1.Brand, err error) {
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

	builder := tx.Brand.Create().
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableLogoURL(req.Data.LogoUrl).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	var entity *ent.Brand
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert brand failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("insert brand failed")
	}

	if len(req.Data.Translations) > 0 {
		if err = r.brandTranslationRepo.CleanTranslations(ctx, tx, entity.ID); err != nil {
			r.log.Errorf("clean translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("clean translations failed")
		}

		for i := range req.Data.Translations {
			req.Data.Translations[i].BrandId = trans.Ptr(entity.ID)
		}

		if err = r.brandTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *BrandRepo) Update(ctx context.Context, req *catalogV1.UpdateBrandRequest) (dto *catalogV1.Brand, err error) {
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
			_, err = r.Create(ctx, &catalogV1.CreateBrandRequest{Data: req.Data})
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
			req.Data.Translations[i].BrandId = trans.Ptr(req.GetId())
		}

		if err = r.brandTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	builder := tx.Brand.UpdateOneID(req.GetId())
	result, err := r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.Brand) {
			builder.
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableLogoURL(req.Data.LogoUrl).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(brand.FieldID, req.GetId()))
		},
	)

	return result, err
}

func (r *BrandRepo) Delete(ctx context.Context, req *catalogV1.DeleteBrandRequest) (err error) {
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

	// 删除品牌数据
	if err = r.entClient.Client().Brand.
		DeleteOneID(req.GetId()).
		Exec(ctx); err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
	}

	// 删除关联翻译数据
	if err = r.brandTranslationRepo.CleanTranslations(ctx, tx, req.GetId()); err != nil {
		r.log.Errorf("clean translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("clean translations failed")
	}

	return err
}

func (r *BrandRepo) TranslationExists(ctx context.Context, brandId uint32, languageCode string) (bool, error) {
	return r.brandTranslationRepo.TranslationExists(ctx, brandId, languageCode)
}

func (r *BrandRepo) CreateTranslation(ctx context.Context, req *catalogV1.CreateBrandTranslationRequest) (*catalogV1.BrandTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetBrandId() == 0 {
		return nil, catalogV1.ErrorBadRequest("brand id is required")
	}

	return r.brandTranslationRepo.CreateTranslation(ctx, req.Data)
}

func (r *BrandRepo) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateBrandTranslationRequest) (*catalogV1.BrandTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetBrandId() == 0 {
		return nil, catalogV1.ErrorBadRequest("brand id is required")
	}

	if exist, err := r.TranslationExists(ctx, req.Data.GetBrandId(), req.Data.GetLanguageCode()); err != nil {
		return nil, err
	} else if !exist {
		return nil, catalogV1.ErrorNotFound("translation not found")
	}

	return r.brandTranslationRepo.UpdateTranslation(ctx, req.GetId(), req.Data, req.GetUpdateMask())
}

func (r *BrandRepo) GetTranslation(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.BrandTranslation, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	return r.brandTranslationRepo.GetTranslation(ctx, req.GetId(), req.GetLocale())
}

func (r *BrandRepo) ListTranslations(ctx context.Context, brandId uint32) ([]*catalogV1.BrandTranslation, error) {
	return r.brandTranslationRepo.ListTranslations(ctx, brandId, "", nil)
}

func (r *BrandRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteBrandTranslationRequest) error {
	return r.brandTranslationRepo.DeleteTranslation(ctx, req)
}

func (r *BrandRepo) CleanTranslations(ctx context.Context, tx *ent.Tx, brandID uint32) error {
	return r.brandTranslationRepo.CleanTranslations(ctx, tx, brandID)
}
