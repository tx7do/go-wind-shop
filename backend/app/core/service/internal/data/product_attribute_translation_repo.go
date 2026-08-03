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
	"go-wind-shop/app/core/service/internal/data/ent/productattributetranslation"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductAttributeTranslationRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.ProductAttributeTranslation, ent.ProductAttributeTranslation]

	repository *entCrud.Repository[
		ent.ProductAttributeTranslationQuery, ent.ProductAttributeTranslationSelect,
		ent.ProductAttributeTranslationCreate, ent.ProductAttributeTranslationCreateBulk,
		ent.ProductAttributeTranslationUpdate, ent.ProductAttributeTranslationUpdateOne,
		ent.ProductAttributeTranslationDelete,
		predicate.ProductAttributeTranslation,
		catalogV1.ProductAttributeTranslation, ent.ProductAttributeTranslation,
	]
}

func NewProductAttributeTranslationRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *ProductAttributeTranslationRepo {
	repo := &ProductAttributeTranslationRepo{
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.ProductAttributeTranslation, ent.ProductAttributeTranslation](),
		log:       ctx.NewLoggerHelper("product-attribute-translation/repo/core-service"),
	}

	repo.init()

	return repo
}

func (r *ProductAttributeTranslationRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductAttributeTranslationQuery, ent.ProductAttributeTranslationSelect,
		ent.ProductAttributeTranslationCreate, ent.ProductAttributeTranslationCreateBulk,
		ent.ProductAttributeTranslationUpdate, ent.ProductAttributeTranslationUpdateOne,
		ent.ProductAttributeTranslationDelete,
		predicate.ProductAttributeTranslation,
		catalogV1.ProductAttributeTranslation, ent.ProductAttributeTranslation,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *ProductAttributeTranslationRepo) CleanTranslations(
	ctx context.Context,
	tx *ent.Tx,
	attributeID uint32,
) error {
	if _, err := tx.ProductAttributeTranslation.Delete().
		Where(
			productattributetranslation.AttributeIDEQ(attributeID),
		).
		Exec(ctx); err != nil {
		r.log.Errorf("delete old product attribute [%d] translations failed: %s", attributeID, err.Error())
		return catalogV1.ErrorInternalServerError("delete old product attribute translations failed")
	}
	return nil
}

func (r *ProductAttributeTranslationRepo) ListTranslations(ctx context.Context, attributeID uint32, locale string, viewMask *fieldmaskpb.FieldMask) ([]*catalogV1.ProductAttributeTranslation, error) {
	builder := r.entClient.Client().ProductAttributeTranslation.Query().
		Where(
			productattributetranslation.AttributeIDEQ(attributeID),
		)

	if len(locale) > 0 {
		builder.Where(
			productattributetranslation.LanguageCodeEQ(locale),
		)
	}

	if viewMask != nil {
		selectSelector, err := r.repository.BuildSelector(viewMask.GetPaths())
		if err != nil {
			r.log.Errorf("build product attribute translation selector failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("build product attribute translation selector failed")
		}
		if selectSelector != nil {
			builder.Modify(selectSelector)
		}
	}

	entities, err := builder.
		All(ctx)
	if err != nil {
		r.log.Errorf("query product attribute translations failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query product attribute translations failed")
	}

	var dtos []*catalogV1.ProductAttributeTranslation
	for _, entity := range entities {
		dtos = append(dtos, r.mapper.ToDTO(entity))
	}

	return dtos, nil
}

func (r *ProductAttributeTranslationRepo) GetTranslation(ctx context.Context, attributeID uint32, languageCode string) (*catalogV1.ProductAttributeTranslation, error) {
	q := r.entClient.Client().ProductAttributeTranslation.Query().
		Where(
			productattributetranslation.AttributeIDEQ(attributeID),
			productattributetranslation.LanguageCodeEQ(languageCode),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		r.log.Errorf("query translation by product attribute id and language code failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query translation by product attribute id and language code failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductAttributeTranslationRepo) newCreateBuilder(pt *ent.ProductAttributeTranslationClient, data *catalogV1.ProductAttributeTranslation) *ent.ProductAttributeTranslationCreate {
	now := time.Now()

	builder := pt.Create().
		SetNillableAttributeID(data.AttributeId).
		SetNillableLanguageCode(data.LanguageCode).
		SetNillableName(data.Name).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	return builder
}

func (r *ProductAttributeTranslationRepo) BatchCreate(ctx context.Context, tx *ent.Tx, items []*catalogV1.ProductAttributeTranslation) error {
	if len(items) == 0 {
		return nil
	}

	builders := make([]*ent.ProductAttributeTranslationCreate, 0, len(items))
	for _, data := range items {
		builder := r.newCreateBuilder(tx.ProductAttributeTranslation, data)
		builders = append(builders, builder)
	}

	err := tx.ProductAttributeTranslation.CreateBulk(builders...).Exec(ctx)
	if err != nil {
		r.log.Errorf("batch create product attribute translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("batch create product attribute translations failed")
	}

	return nil
}

func (r *ProductAttributeTranslationRepo) CreateTranslation(ctx context.Context, data *catalogV1.ProductAttributeTranslation) (*catalogV1.ProductAttributeTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.newCreateBuilder(r.entClient.Client().ProductAttributeTranslation, data)

	var entity *ent.ProductAttributeTranslation
	var err error
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("create product attribute translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("create product attribute translation failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductAttributeTranslationRepo) UpdateTranslation(ctx context.Context, id uint32, data *catalogV1.ProductAttributeTranslation, updateMask *fieldmaskpb.FieldMask) (*catalogV1.ProductAttributeTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.entClient.Client().ProductAttributeTranslation.UpdateOneID(id)

	dto, err := r.repository.UpdateOne(ctx, builder, data, updateMask,
		func(dto *catalogV1.ProductAttributeTranslation) {
			builder.
				SetNillableAttributeID(data.AttributeId).
				SetNillableLanguageCode(data.LanguageCode).
				SetNillableName(data.Name).
				SetNillableUpdatedBy(data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(productattributetranslation.FieldID, id))
		},
	)
	if err != nil {
		r.log.Errorf("update product attribute translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("update product attribute translation failed")
	}

	return dto, nil
}

// TranslationExists checks if a translation exists for the given product attribute ID and language code.
func (r *ProductAttributeTranslationRepo) TranslationExists(ctx context.Context, attributeId uint32, languageCode string) (bool, error) {
	c, err := r.entClient.Client().ProductAttributeTranslation.Query().
		Where(
			productattributetranslation.AttributeIDEQ(attributeId),
			productattributetranslation.LanguageCodeEQ(languageCode),
		).
		Count(ctx)
	if err != nil {
		r.log.Errorf("count product attribute translations by attribute id and language code failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("count product attribute translations by attribute id and language code failed")
	}

	return c > 0, nil
}

// ListAvailedLanguages lists the language codes of all translations available for the given product attribute ID.
func (r *ProductAttributeTranslationRepo) ListAvailedLanguages(ctx context.Context, attributeId uint32) ([]string, error) {
	entities, err := r.entClient.Client().ProductAttributeTranslation.Query().
		Where(
			productattributetranslation.AttributeIDEQ(attributeId),
		).
		Select(productattributetranslation.FieldLanguageCode).
		Strings(ctx)
	if err != nil {
		r.log.Errorf("query available translation languages by product attribute id failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query available translation languages by product attribute id failed")
	}

	return entities, nil
}

func (r *ProductAttributeTranslationRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeTranslationRequest) error {
	if req.QueryBy == nil {
		return catalogV1.ErrorBadRequest("invalid parameter: query_by is required")
	}

	switch req.QueryBy.(type) {
	case *catalogV1.DeleteProductAttributeTranslationRequest_Id:
		if req.GetId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: id must be greater than 0")
		}

	case *catalogV1.DeleteProductAttributeTranslationRequest_Identifier:
		if req.GetIdentifier() == nil {
			return catalogV1.ErrorBadRequest("invalid parameter: identifier is required")
		}
		if req.GetIdentifier().GetAttributeId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: attribute_id must be greater than 0")
		}
		if len(req.GetIdentifier().GetLanguageCode()) == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: language_code is required")
		}

	default:
		return catalogV1.ErrorBadRequest("invalid parameter: unsupported query_by type")
	}

	builder := r.entClient.Client().ProductAttributeTranslation.Delete()

	_, err := r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		switch req.QueryBy.(type) {
		case *catalogV1.DeleteProductAttributeTranslationRequest_Id:
			id := req.GetId()
			s.Where(sql.EQ(productattributetranslation.FieldID, id))

		case *catalogV1.DeleteProductAttributeTranslationRequest_Identifier:
			identifier := req.GetIdentifier()
			s.Where(
				sql.And(
					sql.EQ(productattributetranslation.FieldAttributeID, identifier.GetAttributeId()),
					sql.EQ(productattributetranslation.FieldLanguageCode, identifier.GetLanguageCode()),
				),
			)

		default:
			return
		}
	})
	if err != nil {
		r.log.Errorf("delete product attribute translation failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete product attribute translation failed")
	}

	return nil
}
