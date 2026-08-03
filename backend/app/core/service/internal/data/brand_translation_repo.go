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
	"go-wind-shop/app/core/service/internal/data/ent/brandtranslation"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type BrandTranslationRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.BrandTranslation, ent.BrandTranslation]

	repository *entCrud.Repository[
		ent.BrandTranslationQuery, ent.BrandTranslationSelect,
		ent.BrandTranslationCreate, ent.BrandTranslationCreateBulk,
		ent.BrandTranslationUpdate, ent.BrandTranslationUpdateOne,
		ent.BrandTranslationDelete,
		predicate.BrandTranslation,
		catalogV1.BrandTranslation, ent.BrandTranslation,
	]
}

func NewBrandTranslationRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *BrandTranslationRepo {
	repo := &BrandTranslationRepo{
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.BrandTranslation, ent.BrandTranslation](),
		log:       ctx.NewLoggerHelper("brand-translation/repo/core-service"),
	}

	repo.init()

	return repo
}

func (r *BrandTranslationRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.BrandTranslationQuery, ent.BrandTranslationSelect,
		ent.BrandTranslationCreate, ent.BrandTranslationCreateBulk,
		ent.BrandTranslationUpdate, ent.BrandTranslationUpdateOne,
		ent.BrandTranslationDelete,
		predicate.BrandTranslation,
		catalogV1.BrandTranslation, ent.BrandTranslation,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *BrandTranslationRepo) CleanTranslations(
	ctx context.Context,
	tx *ent.Tx,
	brandID uint32,
) error {
	if _, err := tx.BrandTranslation.Delete().
		Where(
			brandtranslation.BrandIDEQ(brandID),
		).
		Exec(ctx); err != nil {
		r.log.Errorf("delete old brand [%d] translations failed: %s", brandID, err.Error())
		return catalogV1.ErrorInternalServerError("delete old brand translations failed")
	}
	return nil
}

func (r *BrandTranslationRepo) ListTranslations(ctx context.Context, brandID uint32, locale string, viewMask *fieldmaskpb.FieldMask) ([]*catalogV1.BrandTranslation, error) {
	builder := r.entClient.Client().BrandTranslation.Query().
		Where(
			brandtranslation.BrandIDEQ(brandID),
		)

	if len(locale) > 0 {
		builder.Where(
			brandtranslation.LanguageCodeEQ(locale),
		)
	}

	if viewMask != nil {
		selectSelector, err := r.repository.BuildSelector(viewMask.GetPaths())
		if err != nil {
			r.log.Errorf("build brand translation selector failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("build brand translation selector failed")
		}
		if selectSelector != nil {
			builder.Modify(selectSelector)
		}
	}

	entities, err := builder.
		All(ctx)
	if err != nil {
		r.log.Errorf("query brand translations failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query brand translations failed")
	}

	var dtos []*catalogV1.BrandTranslation
	for _, entity := range entities {
		dtos = append(dtos, r.mapper.ToDTO(entity))
	}

	return dtos, nil
}

func (r *BrandTranslationRepo) GetTranslation(ctx context.Context, brandID uint32, languageCode string) (*catalogV1.BrandTranslation, error) {
	q := r.entClient.Client().BrandTranslation.Query().
		Where(
			brandtranslation.BrandIDEQ(brandID),
			brandtranslation.LanguageCodeEQ(languageCode),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		r.log.Errorf("query translation by brand id and language code failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query translation by brand id and language code failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *BrandTranslationRepo) newCreateBuilder(bt *ent.BrandTranslationClient, data *catalogV1.BrandTranslation) *ent.BrandTranslationCreate {
	now := time.Now()

	builder := bt.Create().
		SetNillableBrandID(data.BrandId).
		SetNillableLanguageCode(data.LanguageCode).
		SetNillableName(data.Name).
		SetNillableSlug(data.Slug).
		SetNillableDescription(data.Description).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	return builder
}

func (r *BrandTranslationRepo) BatchCreate(ctx context.Context, tx *ent.Tx, items []*catalogV1.BrandTranslation) error {
	if len(items) == 0 {
		return nil
	}

	builders := make([]*ent.BrandTranslationCreate, 0, len(items))
	for _, data := range items {
		builder := r.newCreateBuilder(tx.BrandTranslation, data)
		builders = append(builders, builder)
	}

	err := tx.BrandTranslation.CreateBulk(builders...).Exec(ctx)
	if err != nil {
		r.log.Errorf("batch create brand translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("batch create brand translations failed")
	}

	return nil
}

func (r *BrandTranslationRepo) CreateTranslation(ctx context.Context, data *catalogV1.BrandTranslation) (*catalogV1.BrandTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.newCreateBuilder(r.entClient.Client().BrandTranslation, data)

	var entity *ent.BrandTranslation
	var err error
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("create brand translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("create brand translation failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *BrandTranslationRepo) UpdateTranslation(ctx context.Context, id uint32, data *catalogV1.BrandTranslation, updateMask *fieldmaskpb.FieldMask) (*catalogV1.BrandTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.entClient.Client().BrandTranslation.UpdateOneID(id)

	dto, err := r.repository.UpdateOne(ctx, builder, data, updateMask,
		func(dto *catalogV1.BrandTranslation) {
			builder.
				SetNillableBrandID(data.BrandId).
				SetNillableLanguageCode(data.LanguageCode).
				SetNillableName(data.Name).
				SetNillableSlug(data.Slug).
				SetNillableDescription(data.Description).
				SetNillableUpdatedBy(data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(brandtranslation.FieldID, id))
		},
	)
	if err != nil {
		r.log.Errorf("update brand translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("update brand translation failed")
	}

	return dto, nil
}

// TranslationExists checks if a translation exists for the given brand ID and language code.
func (r *BrandTranslationRepo) TranslationExists(ctx context.Context, brandId uint32, languageCode string) (bool, error) {
	c, err := r.entClient.Client().BrandTranslation.Query().
		Where(
			brandtranslation.BrandIDEQ(brandId),
			brandtranslation.LanguageCodeEQ(languageCode),
		).
		Count(ctx)
	if err != nil {
		r.log.Errorf("count brand translations by brand id and language code failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("count brand translations by brand id and language code failed")
	}

	return c > 0, nil
}

// ListAvailedLanguages lists the language codes of all translations available for the given brand ID.
func (r *BrandTranslationRepo) ListAvailedLanguages(ctx context.Context, brandId uint32) ([]string, error) {
	entities, err := r.entClient.Client().BrandTranslation.Query().
		Where(
			brandtranslation.BrandIDEQ(brandId),
		).
		Select(brandtranslation.FieldLanguageCode).
		Strings(ctx)
	if err != nil {
		r.log.Errorf("query available translation languages by brand id failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query available translation languages by brand id failed")
	}

	return entities, nil
}

func (r *BrandTranslationRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteBrandTranslationRequest) error {
	if req.QueryBy == nil {
		return catalogV1.ErrorBadRequest("invalid parameter: query_by is required")
	}

	switch req.QueryBy.(type) {
	case *catalogV1.DeleteBrandTranslationRequest_Id:
		if req.GetId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: id must be greater than 0")
		}

	case *catalogV1.DeleteBrandTranslationRequest_Identifier:
		if req.GetIdentifier() == nil {
			return catalogV1.ErrorBadRequest("invalid parameter: identifier is required")
		}
		if req.GetIdentifier().GetBrandId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: brand_id must be greater than 0")
		}
		if len(req.GetIdentifier().GetLanguageCode()) == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: language_code is required")
		}

	default:
		return catalogV1.ErrorBadRequest("invalid parameter: unsupported query_by type")
	}

	builder := r.entClient.Client().BrandTranslation.Delete()

	_, err := r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		switch req.QueryBy.(type) {
		case *catalogV1.DeleteBrandTranslationRequest_Id:
			id := req.GetId()
			s.Where(sql.EQ(brandtranslation.FieldID, id))

		case *catalogV1.DeleteBrandTranslationRequest_Identifier:
			identifier := req.GetIdentifier()
			s.Where(
				sql.And(
					sql.EQ(brandtranslation.FieldBrandID, identifier.GetBrandId()),
					sql.EQ(brandtranslation.FieldLanguageCode, identifier.GetLanguageCode()),
				),
			)

		default:
			return
		}
	})
	if err != nil {
		r.log.Errorf("delete brand translation failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete brand translation failed")
	}

	return nil
}
