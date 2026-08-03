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
	"go-wind-shop/app/core/service/internal/data/ent/category"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type CategoryRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.Category, ent.Category]

	repository *entCrud.Repository[
		ent.CategoryQuery, ent.CategorySelect,
		ent.CategoryCreate, ent.CategoryCreateBulk,
		ent.CategoryUpdate, ent.CategoryUpdateOne,
		ent.CategoryDelete,
		predicate.Category,
		catalogV1.Category, ent.Category,
	]

	categoryTranslationRepo *CategoryTranslationRepo
}

func NewCategoryRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
	categoryTranslationRepo *CategoryTranslationRepo,
) *CategoryRepo {
	repo := &CategoryRepo{
		entClient:               entClient,
		log:                     ctx.NewLoggerHelper("category/repo/core-service"),
		mapper:                  mapper.NewCopierMapper[catalogV1.Category, ent.Category](),
		categoryTranslationRepo: categoryTranslationRepo,
	}

	repo.init()

	return repo
}

func (r *CategoryRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CategoryQuery, ent.CategorySelect,
		ent.CategoryCreate, ent.CategoryCreateBulk,
		ent.CategoryUpdate, ent.CategoryUpdateOne,
		ent.CategoryDelete,
		predicate.Category,
		catalogV1.Category, ent.Category,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *CategoryRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Category.Query()
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

func (r *CategoryRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListCategoryResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Category.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListCategoryResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListCategoryResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *CategoryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Category.Query().
		Where(category.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *CategoryRepo) Get(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.Category, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	build := r.entClient.Client().Category.Query()

	switch req.QueryBy.(type) {
	case *catalogV1.GetCategoryRequest_Id:
		build.Where(category.IDEQ(req.GetId()))

	case *catalogV1.GetCategoryRequest_Code:
		build.Where(category.PathEQ(req.GetCode()))

	default:
		return nil, catalogV1.ErrorBadRequest("invalid query field")
	}

	entity, err := build.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, catalogV1.ErrorNotFound("category not found")
		}

		r.log.Errorf("query category failed: %s", err.Error())

		return nil, catalogV1.ErrorInternalServerError("query category failed")
	}

	dto := r.mapper.ToDTO(entity)

	languages, err := r.categoryTranslationRepo.ListAvailedLanguages(ctx, dto.GetId())
	if err != nil {
		r.log.Errorf("query availed languages failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query availed languages failed")
	}
	dto.AvailableLanguages = languages

	if req.Locale == nil {
		translations, err := r.categoryTranslationRepo.ListTranslations(ctx, dto.GetId(), "", nil)
		if err != nil {
			r.log.Errorf("query translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("query translations failed")
		}
		dto.Translations = translations
	} else {
		translation, err := r.categoryTranslationRepo.GetTranslation(ctx, dto.GetId(), req.GetLocale())
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

func (r *CategoryRepo) Create(ctx context.Context, req *catalogV1.CreateCategoryRequest) (dto *catalogV1.Category, err error) {
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

	builder := tx.Category.Create().
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillablePath(req.Data.Path).
		SetNillableParentID(req.Data.ParentId).
		SetNillableDepth(req.Data.Depth).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	var entity *ent.Category
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert category failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("insert category failed")
	}

	if len(req.Data.Translations) > 0 {
		if err = r.categoryTranslationRepo.CleanTranslations(ctx, tx, entity.ID); err != nil {
			r.log.Errorf("clean translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("clean translations failed")
		}

		for i := range req.Data.Translations {
			req.Data.Translations[i].CategoryId = trans.Ptr(entity.ID)
		}

		if err = r.categoryTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *CategoryRepo) Update(ctx context.Context, req *catalogV1.UpdateCategoryRequest) (dto *catalogV1.Category, err error) {
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
			_, err = r.Create(ctx, &catalogV1.CreateCategoryRequest{Data: req.Data})
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
			req.Data.Translations[i].CategoryId = trans.Ptr(req.GetId())
		}

		if err = r.categoryTranslationRepo.BatchCreate(ctx, tx, req.Data.GetTranslations()); err != nil {
			r.log.Errorf("batch insert translations failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("batch insert translations failed")
		}
	}

	builder := tx.Category.UpdateOneID(req.GetId())
	result, err := r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.Category) {
			builder.
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillablePath(req.Data.Path).
				SetNillableParentID(req.Data.ParentId).
				SetNillableDepth(req.Data.Depth).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(category.FieldID, req.GetId()))
		},
	)

	return result, err
}

func (r *CategoryRepo) Delete(ctx context.Context, req *catalogV1.DeleteCategoryRequest) (err error) {
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

	// 删除类目数据
	if err = r.entClient.Client().Category.
		DeleteOneID(req.GetId()).
		Exec(ctx); err != nil {
		r.log.Errorf("delete one data failed: %s", err.Error())
	}

	// 删除关联翻译数据
	if err = r.categoryTranslationRepo.CleanTranslations(ctx, tx, req.GetId()); err != nil {
		r.log.Errorf("clean translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("clean translations failed")
	}

	return err
}

func (r *CategoryRepo) TranslationExists(ctx context.Context, categoryId uint32, languageCode string) (bool, error) {
	return r.categoryTranslationRepo.TranslationExists(ctx, categoryId, languageCode)
}

func (r *CategoryRepo) CreateTranslation(ctx context.Context, req *catalogV1.CreateCategoryTranslationRequest) (*catalogV1.CategoryTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetCategoryId() == 0 {
		return nil, catalogV1.ErrorBadRequest("category id is required")
	}

	return r.categoryTranslationRepo.CreateTranslation(ctx, req.Data)
}

func (r *CategoryRepo) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateCategoryTranslationRequest) (*catalogV1.CategoryTranslation, error) {
	if req == nil || req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if len(req.Data.GetLanguageCode()) == 0 {
		return nil, catalogV1.ErrorBadRequest("language code is required")
	}

	if req.Data.GetCategoryId() == 0 {
		return nil, catalogV1.ErrorBadRequest("category id is required")
	}

	if exist, err := r.TranslationExists(ctx, req.Data.GetCategoryId(), req.Data.GetLanguageCode()); err != nil {
		return nil, err
	} else if !exist {
		return nil, catalogV1.ErrorNotFound("translation not found")
	}

	return r.categoryTranslationRepo.UpdateTranslation(ctx, req.GetId(), req.Data, req.GetUpdateMask())
}

func (r *CategoryRepo) GetTranslation(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.CategoryTranslation, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	return r.categoryTranslationRepo.GetTranslation(ctx, req.GetId(), req.GetLocale())
}

func (r *CategoryRepo) ListTranslations(ctx context.Context, categoryId uint32) ([]*catalogV1.CategoryTranslation, error) {
	return r.categoryTranslationRepo.ListTranslations(ctx, categoryId, "", nil)
}

func (r *CategoryRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteCategoryTranslationRequest) error {
	return r.categoryTranslationRepo.DeleteTranslation(ctx, req)
}

func (r *CategoryRepo) CleanTranslations(ctx context.Context, tx *ent.Tx, categoryID uint32) error {
	return r.categoryTranslationRepo.CleanTranslations(ctx, tx, categoryID)
}
