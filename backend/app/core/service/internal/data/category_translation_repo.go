package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/categorytranslation"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type CategoryTranslationRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.CategoryTranslation, ent.CategoryTranslation]

	repository *entCrud.Repository[
		ent.CategoryTranslationQuery, ent.CategoryTranslationSelect,
		ent.CategoryTranslationCreate, ent.CategoryTranslationCreateBulk,
		ent.CategoryTranslationUpdate, ent.CategoryTranslationUpdateOne,
		ent.CategoryTranslationDelete,
		predicate.CategoryTranslation,
		catalogV1.CategoryTranslation, ent.CategoryTranslation,
	]
}

func NewCategoryTranslationRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *CategoryTranslationRepo {
	repo := &CategoryTranslationRepo{
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.CategoryTranslation, ent.CategoryTranslation](),
		log:       ctx.NewLoggerHelper("category-translation/repo/core-service"),
	}

	repo.init()

	return repo
}

func (r *CategoryTranslationRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CategoryTranslationQuery, ent.CategoryTranslationSelect,
		ent.CategoryTranslationCreate, ent.CategoryTranslationCreateBulk,
		ent.CategoryTranslationUpdate, ent.CategoryTranslationUpdateOne,
		ent.CategoryTranslationDelete,
		predicate.CategoryTranslation,
		catalogV1.CategoryTranslation, ent.CategoryTranslation,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *CategoryTranslationRepo) CleanTranslations(
	ctx context.Context,
	tx *ent.Tx,
	categoryID uint32,
) error {
	if _, err := tx.CategoryTranslation.Delete().
		Where(
			categorytranslation.CategoryIDEQ(categoryID),
		).
		Exec(ctx); err != nil {
		r.log.Errorf("delete old category [%d] translations failed: %s", categoryID, err.Error())
		return catalogV1.ErrorInternalServerError("delete old category translations failed")
	}
	return nil
}

func (r *CategoryTranslationRepo) ListTranslations(ctx context.Context, categoryID uint32, locale string, viewMask *fieldmaskpb.FieldMask) ([]*catalogV1.CategoryTranslation, error) {
	builder := r.entClient.Client().CategoryTranslation.Query().
		Where(
			categorytranslation.CategoryIDEQ(categoryID),
		)

	if len(locale) > 0 {
		builder.Where(
			categorytranslation.LanguageCodeEQ(locale),
		)
	}

	if viewMask != nil {
		selectSelector, err := r.repository.BuildSelector(viewMask.GetPaths())
		if err != nil {
			r.log.Errorf("build category translation selector failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("build category translation selector failed")
		}
		if selectSelector != nil {
			builder.Modify(selectSelector)
		}
	}

	entities, err := builder.
		All(ctx)
	if err != nil {
		r.log.Errorf("query category translations failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query category translations failed")
	}

	var dtos []*catalogV1.CategoryTranslation
	for _, entity := range entities {
		dtos = append(dtos, r.mapper.ToDTO(entity))
	}

	return dtos, nil
}

func (r *CategoryTranslationRepo) GetTranslation(ctx context.Context, categoryID uint32, languageCode string) (*catalogV1.CategoryTranslation, error) {
	q := r.entClient.Client().CategoryTranslation.Query().
		Where(
			categorytranslation.CategoryIDEQ(categoryID),
			categorytranslation.LanguageCodeEQ(languageCode),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		r.log.Errorf("query translation by category id and language code failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query translation by category id and language code failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *CategoryTranslationRepo) newCreateBuilder(ct *ent.CategoryTranslationClient, data *catalogV1.CategoryTranslation) *ent.CategoryTranslationCreate {
	now := time.Now()

	builder := ct.Create().
		SetNillableCategoryID(data.CategoryId).
		SetNillableLanguageCode(data.LanguageCode).
		SetNillableName(data.Name).
		SetNillableSlug(data.Slug).
		SetNillableDescription(data.Description).
		SetNillableFullPath(data.FullPath).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	return builder
}

func (r *CategoryTranslationRepo) BatchCreate(ctx context.Context, tx *ent.Tx, items []*catalogV1.CategoryTranslation) error {
	if len(items) == 0 {
		return nil
	}

	builders := make([]*ent.CategoryTranslationCreate, 0, len(items))
	for _, data := range items {
		builder := r.newCreateBuilder(tx.CategoryTranslation, data)
		builders = append(builders, builder)
	}

	err := tx.CategoryTranslation.CreateBulk(builders...).Exec(ctx)
	if err != nil {
		r.log.Errorf("batch create category translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("batch create category translations failed")
	}

	return nil
}

func (r *CategoryTranslationRepo) CreateTranslation(ctx context.Context, data *catalogV1.CategoryTranslation) (*catalogV1.CategoryTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.newCreateBuilder(r.entClient.Client().CategoryTranslation, data)

	var entity *ent.CategoryTranslation
	var err error
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("create category translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("create category translation failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *CategoryTranslationRepo) UpdateTranslation(ctx context.Context, id uint32, data *catalogV1.CategoryTranslation, updateMask *fieldmaskpb.FieldMask) (*catalogV1.CategoryTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.entClient.Client().CategoryTranslation.UpdateOneID(id)

	dto, err := r.repository.UpdateOne(ctx, builder, data, updateMask,
		func(dto *catalogV1.CategoryTranslation) {
			builder.
				SetNillableCategoryID(data.CategoryId).
				SetNillableLanguageCode(data.LanguageCode).
				SetNillableName(data.Name).
				SetNillableSlug(data.Slug).
				SetNillableDescription(data.Description).
				SetNillableFullPath(data.FullPath).
				SetNillableUpdatedBy(data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(categorytranslation.FieldID, id))
		},
	)
	if err != nil {
		r.log.Errorf("update category translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("update category translation failed")
	}

	return dto, nil
}

// TranslationExists checks if a translation exists for the given category ID and language code.
func (r *CategoryTranslationRepo) TranslationExists(ctx context.Context, categoryId uint32, languageCode string) (bool, error) {
	c, err := r.entClient.Client().CategoryTranslation.Query().
		Where(
			categorytranslation.CategoryIDEQ(categoryId),
			categorytranslation.LanguageCodeEQ(languageCode),
		).
		Count(ctx)
	if err != nil {
		r.log.Errorf("count category translations by category id and language code failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("count category translations by category id and language code failed")
	}

	return c > 0, nil
}

// ListAvailedLanguages lists the language codes of all translations available for the given category ID.
func (r *CategoryTranslationRepo) ListAvailedLanguages(ctx context.Context, categoryId uint32) ([]string, error) {
	entities, err := r.entClient.Client().CategoryTranslation.Query().
		Where(
			categorytranslation.CategoryIDEQ(categoryId),
		).
		Select(categorytranslation.FieldLanguageCode).
		Strings(ctx)
	if err != nil {
		r.log.Errorf("query available translation languages by category id failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query available translation languages by category id failed")
	}

	return entities, nil
}

func (r *CategoryTranslationRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteCategoryTranslationRequest) error {
	if req.QueryBy == nil {
		return catalogV1.ErrorBadRequest("invalid parameter: query_by is required")
	}

	switch req.QueryBy.(type) {
	case *catalogV1.DeleteCategoryTranslationRequest_Id:
		if req.GetId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: id must be greater than 0")
		}

	case *catalogV1.DeleteCategoryTranslationRequest_Identifier:
		if req.GetIdentifier() == nil {
			return catalogV1.ErrorBadRequest("invalid parameter: identifier is required")
		}
		if req.GetIdentifier().GetCategoryId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: category_id must be greater than 0")
		}
		if len(req.GetIdentifier().GetLanguageCode()) == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: language_code is required")
		}

	default:
		return catalogV1.ErrorBadRequest("invalid parameter: unsupported query_by type")
	}

	builder := r.entClient.Client().CategoryTranslation.Delete()

	_, err := r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		switch req.QueryBy.(type) {
		case *catalogV1.DeleteCategoryTranslationRequest_Id:
			id := req.GetId()
			s.Where(sql.EQ(categorytranslation.FieldID, id))

		case *catalogV1.DeleteCategoryTranslationRequest_Identifier:
			identifier := req.GetIdentifier()
			s.Where(
				sql.And(
					sql.EQ(categorytranslation.FieldCategoryID, identifier.GetCategoryId()),
					sql.EQ(categorytranslation.FieldLanguageCode, identifier.GetLanguageCode()),
				),
			)

		default:
			return
		}
	})
	if err != nil {
		r.log.Errorf("delete category translation failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete category translation failed")
	}

	return nil
}
