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
	"go-wind-shop/app/core/service/internal/data/ent/producttranslation"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductTranslationRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.ProductTranslation, ent.ProductTranslation]

	repository *entCrud.Repository[
		ent.ProductTranslationQuery, ent.ProductTranslationSelect,
		ent.ProductTranslationCreate, ent.ProductTranslationCreateBulk,
		ent.ProductTranslationUpdate, ent.ProductTranslationUpdateOne,
		ent.ProductTranslationDelete,
		predicate.ProductTranslation,
		catalogV1.ProductTranslation, ent.ProductTranslation,
	]
}

func NewProductTranslationRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *ProductTranslationRepo {
	repo := &ProductTranslationRepo{
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.ProductTranslation, ent.ProductTranslation](),
		log:       ctx.NewLoggerHelper("product-translation/repo/core-service"),
	}

	repo.init()

	return repo
}

func (r *ProductTranslationRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductTranslationQuery, ent.ProductTranslationSelect,
		ent.ProductTranslationCreate, ent.ProductTranslationCreateBulk,
		ent.ProductTranslationUpdate, ent.ProductTranslationUpdateOne,
		ent.ProductTranslationDelete,
		predicate.ProductTranslation,
		catalogV1.ProductTranslation, ent.ProductTranslation,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *ProductTranslationRepo) CleanTranslations(
	ctx context.Context,
	tx *ent.Tx,
	productID uint32,
) error {
	if _, err := tx.ProductTranslation.Delete().
		Where(
			producttranslation.ProductIDEQ(productID),
		).
		Exec(ctx); err != nil {
		r.log.Errorf("delete old product [%d] translations failed: %s", productID, err.Error())
		return catalogV1.ErrorInternalServerError("delete old product translations failed")
	}
	return nil
}

func (r *ProductTranslationRepo) ListTranslations(ctx context.Context, productID uint32, locale string, viewMask *fieldmaskpb.FieldMask) ([]*catalogV1.ProductTranslation, error) {
	builder := r.entClient.Client().ProductTranslation.Query().
		Where(
			producttranslation.ProductIDEQ(productID),
		)

	if len(locale) > 0 {
		builder.Where(
			producttranslation.LanguageCodeEQ(locale),
		)
	}

	if viewMask != nil {
		selectSelector, err := r.repository.BuildSelector(viewMask.GetPaths())
		if err != nil {
			r.log.Errorf("build product translation selector failed: %s", err.Error())
			return nil, catalogV1.ErrorInternalServerError("build product translation selector failed")
		}
		if selectSelector != nil {
			builder.Modify(selectSelector)
		}
	}

	entities, err := builder.
		All(ctx)
	if err != nil {
		r.log.Errorf("query product translations failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query product translations failed")
	}

	var dtos []*catalogV1.ProductTranslation
	for _, entity := range entities {
		dtos = append(dtos, r.mapper.ToDTO(entity))
	}

	return dtos, nil
}

func (r *ProductTranslationRepo) GetTranslation(ctx context.Context, productID uint32, languageCode string) (*catalogV1.ProductTranslation, error) {
	q := r.entClient.Client().ProductTranslation.Query().
		Where(
			producttranslation.ProductIDEQ(productID),
			producttranslation.LanguageCodeEQ(languageCode),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		r.log.Errorf("query translation by product id and language code failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query translation by product id and language code failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductTranslationRepo) newCreateBuilder(pt *ent.ProductTranslationClient, data *catalogV1.ProductTranslation) *ent.ProductTranslationCreate {
	now := time.Now()

	builder := pt.Create().
		SetNillableProductID(data.ProductId).
		SetNillableLanguageCode(data.LanguageCode).
		SetNillableName(data.Name).
		SetNillableSlug(data.Slug).
		SetNillableShortDescription(data.ShortDescription).
		SetNillableLongDescription(data.LongDescription).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	return builder
}

func (r *ProductTranslationRepo) BatchCreate(ctx context.Context, tx *ent.Tx, items []*catalogV1.ProductTranslation) error {
	if len(items) == 0 {
		return nil
	}

	builders := make([]*ent.ProductTranslationCreate, 0, len(items))
	for _, data := range items {
		builder := r.newCreateBuilder(tx.ProductTranslation, data)
		builders = append(builders, builder)
	}

	err := tx.ProductTranslation.CreateBulk(builders...).Exec(ctx)
	if err != nil {
		r.log.Errorf("batch create product translations failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("batch create product translations failed")
	}

	return nil
}

func (r *ProductTranslationRepo) CreateTranslation(ctx context.Context, data *catalogV1.ProductTranslation) (*catalogV1.ProductTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.newCreateBuilder(r.entClient.Client().ProductTranslation, data)

	var entity *ent.ProductTranslation
	var err error
	if entity, err = builder.Save(ctx); err != nil {
		r.log.Errorf("create product translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("create product translation failed")
	}

	return r.mapper.ToDTO(entity), nil
}

func (r *ProductTranslationRepo) UpdateTranslation(ctx context.Context, id uint32, data *catalogV1.ProductTranslation, updateMask *fieldmaskpb.FieldMask) (*catalogV1.ProductTranslation, error) {
	if data == nil {
		return nil, nil
	}

	builder := r.entClient.Client().ProductTranslation.UpdateOneID(id)

	dto, err := r.repository.UpdateOne(ctx, builder, data, updateMask,
		func(dto *catalogV1.ProductTranslation) {
			builder.
				SetNillableProductID(data.ProductId).
				SetNillableLanguageCode(data.LanguageCode).
				SetNillableName(data.Name).
				SetNillableSlug(data.Slug).
				SetNillableShortDescription(data.ShortDescription).
				SetNillableLongDescription(data.LongDescription).
				SetNillableUpdatedBy(data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(producttranslation.FieldID, id))
		},
	)
	if err != nil {
		r.log.Errorf("update product translation failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("update product translation failed")
	}

	return dto, nil
}

// TranslationExists checks if a translation exists for the given product ID and language code.
func (r *ProductTranslationRepo) TranslationExists(ctx context.Context, productId uint32, languageCode string) (bool, error) {
	c, err := r.entClient.Client().ProductTranslation.Query().
		Where(
			producttranslation.ProductIDEQ(productId),
			producttranslation.LanguageCodeEQ(languageCode),
		).
		Count(ctx)
	if err != nil {
		r.log.Errorf("count product translations by product id and language code failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("count product translations by product id and language code failed")
	}

	return c > 0, nil
}

// ListAvailedLanguages lists the language codes of all translations available for the given product ID.
func (r *ProductTranslationRepo) ListAvailedLanguages(ctx context.Context, productId uint32) ([]string, error) {
	entities, err := r.entClient.Client().ProductTranslation.Query().
		Where(
			producttranslation.ProductIDEQ(productId),
		).
		Select(producttranslation.FieldLanguageCode).
		Strings(ctx)
	if err != nil {
		r.log.Errorf("query available translation languages by product id failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query available translation languages by product id failed")
	}

	return entities, nil
}

func (r *ProductTranslationRepo) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductTranslationRequest) error {
	if req.QueryBy == nil {
		return catalogV1.ErrorBadRequest("invalid parameter: query_by is required")
	}

	switch req.QueryBy.(type) {
	case *catalogV1.DeleteProductTranslationRequest_Id:
		if req.GetId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: id must be greater than 0")
		}

	case *catalogV1.DeleteProductTranslationRequest_Identifier:
		if req.GetIdentifier() == nil {
			return catalogV1.ErrorBadRequest("invalid parameter: identifier is required")
		}
		if req.GetIdentifier().GetProductId() == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: product_id must be greater than 0")
		}
		if len(req.GetIdentifier().GetLanguageCode()) == 0 {
			return catalogV1.ErrorBadRequest("invalid parameter: language_code is required")
		}

	default:
		return catalogV1.ErrorBadRequest("invalid parameter: unsupported query_by type")
	}

	builder := r.entClient.Client().ProductTranslation.Delete()

	_, err := r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		switch req.QueryBy.(type) {
		case *catalogV1.DeleteProductTranslationRequest_Id:
			id := req.GetId()
			s.Where(sql.EQ(producttranslation.FieldID, id))

		case *catalogV1.DeleteProductTranslationRequest_Identifier:
			identifier := req.GetIdentifier()
			s.Where(
				sql.And(
					sql.EQ(producttranslation.FieldProductID, identifier.GetProductId()),
					sql.EQ(producttranslation.FieldLanguageCode, identifier.GetLanguageCode()),
				),
			)

		default:
			return
		}
	})
	if err != nil {
		r.log.Errorf("delete product translation failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete product translation failed")
	}

	return nil
}
