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
	"go-wind-shop/app/core/service/internal/data/ent/productattributevaluetranslation"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductAttributeValueTranslationRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.ProductAttributeValueTranslation, ent.ProductAttributeValueTranslation]

	repository *entCrud.Repository[
		ent.ProductAttributeValueTranslationQuery, ent.ProductAttributeValueTranslationSelect,
		ent.ProductAttributeValueTranslationCreate, ent.ProductAttributeValueTranslationCreateBulk,
		ent.ProductAttributeValueTranslationUpdate, ent.ProductAttributeValueTranslationUpdateOne,
		ent.ProductAttributeValueTranslationDelete,
		predicate.ProductAttributeValueTranslation,
		catalogV1.ProductAttributeValueTranslation, ent.ProductAttributeValueTranslation,
	]
}

func NewProductAttributeValueTranslationRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *ProductAttributeValueTranslationRepo {
	repo := &ProductAttributeValueTranslationRepo{
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.ProductAttributeValueTranslation, ent.ProductAttributeValueTranslation](),
		log:       ctx.NewLoggerHelper("product-attribute-value-translation/repo/core-service"),
	}

	repo.init()

	return repo
}

func (r *ProductAttributeValueTranslationRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductAttributeValueTranslationQuery, ent.ProductAttributeValueTranslationSelect,
		ent.ProductAttributeValueTranslationCreate, ent.ProductAttributeValueTranslationCreateBulk,
		ent.ProductAttributeValueTranslationUpdate, ent.ProductAttributeValueTranslationUpdateOne,
		ent.ProductAttributeValueTranslationDelete,
		predicate.ProductAttributeValueTranslation,
		catalogV1.ProductAttributeValueTranslation, ent.ProductAttributeValueTranslation,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *ProductAttributeValueTranslationRepo) CleanTranslations(
	ctx context.Context,
	tx *ent.Tx,
	valueID uint32,
) error {
	if _, err := tx.ProductAttributeValueTranslation.Delete().
		Where(
			productattributevaluetranslation.ValueIDEQ(valueID),
		).
		Exec(ctx); err != nil {
		r.log.Errorf("delete old product attribute value [%d] translations failed: %s", valueID, err.Error())
		return catalogV1.ErrorInternalServerError("delete old product attribute value translations failed")
	}
	return nil
}

func (r *ProductAttributeValueTranslationRepo) ListTranslations(ctx context.Context, valueID uint32, locale string, viewMask *fieldmaskpb.FieldMask) ([]*catalogV1.ProductAttributeValueTranslation, error) {
	builder := r.entClient.Client().ProductAttributeValueTranslation.Query().
		Where(
			productattributevaluetranslation.ValueIDEQ(valueID),
		)

	if len(locale) > 0 {
		builder.Where(
			productattributevaluetranslation.LanguageCodeEQ(locale),
		)
	}

	if viewMask != nil {
		selectSelector, err := r.repository.BuildSelector(viewMask.GetPaths())
		if err != nil {
			r.log.Errorf("build product attribute value translation selector failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("build product attribute value translation selector failed")
		}
		if selectSelector != nil {
			builder.Modify(selectSelector)
		}
	}

	entities, err := builder.
		All(ctx)
	if err != nil {
		r.log.Errorf("query product attribute value translations failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query product attribute value translations failed")
	}

	var dtos []*catalogV1.ProductAttributeValueTranslation
	for _, entity := range entities {
		dtos = append(dtos, r.mapper.ToDTO(entity))
	}

	return dtos, nil
}

func (r *ProductAttributeValueTranslationRepo) GetTranslation(ctx context.Context, valueID uint32, languageCode string) (*catalogV1.ProductAttributeValueTranslation, error) {
	q := r.entClient.Client().ProductAttributeValueTranslation.Query().
		Where(
			productattributevaluetranslation.ValueIDEQ(valueID),
			productattributevaluetranslation.LanguageCodeEQ(languageCode),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		r.log.Errorf("query translation by product attribute value id and language code failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query translation by product attribute value id and language code failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductAttributeValueTranslationRepo) newCreateBuilder(pt *ent.ProductAttributeValueTranslationClient, data *catalogV1.ProductAttributeValueTranslation) *ent.ProductAttributeValueTranslationCreate {
	now := time.Now()

	builder := pt.Create().
		SetNillableValueID(data.ValueId).
		SetNillableLanguageCode(data.LanguageCode).
		SetNillableDisplayName(data.DisplayName).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	return builder
}

func (r *ProductAttributeValueTranslationRepo) BatchCreate(ctx context.Context, tx *ent.Tx, items []*catalogV1.ProductAttributeValueTranslation) error {
	if len(items) == 0 {
		return nil
	}

	builders := make([]*ent.ProductAttributeValueTranslationCreate, 0, len(items))
	for _, data := range items {
		builder := r.newCreateBuilder(tx.ProductAttributeValueTranslation, data)
		builders = append(builders, builder)
	}

	err := tx.ProductAttributeValueTranslation.CreateBulk(builders...).Exec(ctx)
	if err != nil {
		r.log.Errorf("batch create product attribute value translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("batch create product attribute value translations failed")
	}

	return nil
}

func (r *ProductAttributeValueTranslationRepo) CreateTranslation(ctx context.Context, data *catalogV1.ProductAttributeValueTranslation) (*catalogV1.ProductAttributeValueTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.newCreateBuilder(r.entClient.Client().ProductAttributeValueTranslation, data)

	var entity *ent.ProductAttributeValueTranslation
	var err error
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("create product attribute value translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("create product attribute value translation failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductAttributeValueTranslationRepo) UpdateTranslation(ctx context.Context, id uint32, data *catalogV1.ProductAttributeValueTranslation, updateMask *fieldmaskpb.FieldMask) (*catalogV1.ProductAttributeValueTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.entClient.Client().ProductAttributeValueTranslation.UpdateOneID(id)

	dto, err := r.repository.UpdateOne(ctx, builder, data, updateMask,
		func(dto *catalogV1.ProductAttributeValueTranslation) {
			builder.
				SetNillableValueID(data.ValueId).
				SetNillableLanguageCode(data.LanguageCode).
				SetNillableDisplayName(data.DisplayName).
				SetNillableUpdatedBy(data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(productattributevaluetranslation.FieldID, id))
		},
	)
	if err != nil {
		r.log.Errorf("update product attribute value translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("update product attribute value translation failed")
	}

	return dto, nil
}

// TranslationExists checks if a translation exists for the given product attribute value ID and language code.
func (r *ProductAttributeValueTranslationRepo) TranslationExists(ctx context.Context, valueId uint32, languageCode string) (bool, error) {
	c, err := r.entClient.Client().ProductAttributeValueTranslation.Query().
		Where(
			productattributevaluetranslation.ValueIDEQ(valueId),
			productattributevaluetranslation.LanguageCodeEQ(languageCode),
		).
		Count(ctx)
	if err != nil {
		r.log.Errorf("count product attribute value translations by value id and language code failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("count product attribute value translations by value id and language code failed")
	}

	return c > 0, nil
}

// ListAvailedLanguages lists the language codes of all translations available for the given product attribute value ID.
func (r *ProductAttributeValueTranslationRepo) ListAvailedLanguages(ctx context.Context, valueId uint32) ([]string, error) {
	entities, err := r.entClient.Client().ProductAttributeValueTranslation.Query().
		Where(
			productattributevaluetranslation.ValueIDEQ(valueId),
		).
		Select(productattributevaluetranslation.FieldLanguageCode).
		Strings(ctx)
	if err != nil {
		r.log.Errorf("query available translation languages by product attribute value id failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query available translation languages by product attribute value id failed")
	}

	return entities, nil
}

func (r *ProductAttributeValueTranslationRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeValueTranslationRequest) error {
	if req.QueryBy == nil {
		return catalogV1.ErrorBadRequest("invalid parameter: query_by is required")
	}

	switch req.QueryBy.(type) {
	case *catalogV1.DeleteProductAttributeValueTranslationRequest_Id:
		if req.GetId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: id must be greater than 0")
		}

	case *catalogV1.DeleteProductAttributeValueTranslationRequest_Identifier:
		if req.GetIdentifier() == nil {
			return catalogV1.ErrorBadRequest("invalid parameter: identifier is required")
		}
		if req.GetIdentifier().GetValueId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: value_id must be greater than 0")
		}
		if len(req.GetIdentifier().GetLanguageCode()) == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: language_code is required")
		}

	default:
		return catalogV1.ErrorBadRequest("invalid parameter: unsupported query_by type")
	}

	builder := r.entClient.Client().ProductAttributeValueTranslation.Delete()

	_, err := r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		switch req.QueryBy.(type) {
		case *catalogV1.DeleteProductAttributeValueTranslationRequest_Id:
			id := req.GetId()
			s.Where(sql.EQ(productattributevaluetranslation.FieldID, id))

		case *catalogV1.DeleteProductAttributeValueTranslationRequest_Identifier:
			identifier := req.GetIdentifier()
			s.Where(
				sql.And(
					sql.EQ(productattributevaluetranslation.FieldValueID, identifier.GetValueId()),
					sql.EQ(productattributevaluetranslation.FieldLanguageCode, identifier.GetLanguageCode()),
				),
			)

		default:
			return
		}
	})
	if err != nil {
		r.log.Errorf("delete product attribute value translation failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete product attribute value translation failed")
	}

	return nil
}
