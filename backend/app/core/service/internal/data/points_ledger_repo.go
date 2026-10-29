package data

import (
	"go-wind-shop/pkg/logx"
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/pointsledger"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type PointsLedgerRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.PointsLedger, ent.PointsLedger]

	typeConverter *mapper.EnumTypeConverter[catalogV1.PointsLedgerType, pointsledger.Type]

	repository *entCrud.Repository[
		ent.PointsLedgerQuery, ent.PointsLedgerSelect,
		ent.PointsLedgerCreate, ent.PointsLedgerCreateBulk,
		ent.PointsLedgerUpdate, ent.PointsLedgerUpdateOne,
		ent.PointsLedgerDelete,
		predicate.PointsLedger,
		catalogV1.PointsLedger, ent.PointsLedger,
	]
}

func NewPointsLedgerRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *PointsLedgerRepo {
	repo := &PointsLedgerRepo{
		log:           logx.New(ctx.NewLoggerHelper("points-ledger/repo/core-service")),
		entClient:     entClient,
		mapper:        mapper.NewCopierMapper[catalogV1.PointsLedger, ent.PointsLedger](),
		typeConverter: mapper.NewEnumTypeConverter[catalogV1.PointsLedgerType, pointsledger.Type](catalogV1.PointsLedgerType_name, catalogV1.PointsLedgerType_value),
	}

	repo.init()

	return repo
}

func (r *PointsLedgerRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.PointsLedgerQuery, ent.PointsLedgerSelect,
		ent.PointsLedgerCreate, ent.PointsLedgerCreateBulk,
		ent.PointsLedgerUpdate, ent.PointsLedgerUpdateOne,
		ent.PointsLedgerDelete,
		predicate.PointsLedger,
		catalogV1.PointsLedger, ent.PointsLedger,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.typeConverter.NewConverterPair())
}

func (r *PointsLedgerRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().PointsLedger.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query points ledger count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *PointsLedgerRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPointsLedgerResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PointsLedger.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListPointsLedgerResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListPointsLedgerResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *PointsLedgerRepo) Get(ctx context.Context, req *catalogV1.GetPointsLedgerRequest) (*catalogV1.PointsLedger, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PointsLedger.Query()

	var whereCond []func(s *sql.Selector)
	whereCond = append(whereCond, pointsledger.IDEQ(req.GetId()))

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}
