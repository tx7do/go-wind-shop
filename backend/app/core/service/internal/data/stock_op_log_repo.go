package data

import (
	"go-wind-shop/pkg/logx"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/stockoplog"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// StockOpLogRepo 库存操作流水（append-only 审计，仅 List/Count/Get）。
type StockOpLogRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.StockOpLog, ent.StockOpLog]

	reasonConverter *mapper.EnumTypeConverter[catalogV1.StockOpLog_Reason, stockoplog.Reason]

	repository *entCrud.Repository[
		ent.StockOpLogQuery, ent.StockOpLogSelect,
		ent.StockOpLogCreate, ent.StockOpLogCreateBulk,
		ent.StockOpLogUpdate, ent.StockOpLogUpdateOne,
		ent.StockOpLogDelete,
		predicate.StockOpLog,
		catalogV1.StockOpLog, ent.StockOpLog,
	]
}

func NewStockOpLogRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *StockOpLogRepo {
	repo := &StockOpLogRepo{
		log:             logx.New(ctx.NewLoggerHelper("stock-op-log/repo/core-service")),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[catalogV1.StockOpLog, ent.StockOpLog](),
		reasonConverter: mapper.NewEnumTypeConverter[catalogV1.StockOpLog_Reason, stockoplog.Reason](catalogV1.StockOpLog_Reason_name, catalogV1.StockOpLog_Reason_value),
	}

	repo.init()

	return repo
}

func (r *StockOpLogRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.StockOpLogQuery, ent.StockOpLogSelect,
		ent.StockOpLogCreate, ent.StockOpLogCreateBulk,
		ent.StockOpLogUpdate, ent.StockOpLogUpdateOne,
		ent.StockOpLogDelete,
		predicate.StockOpLog,
		catalogV1.StockOpLog, ent.StockOpLog,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.reasonConverter.NewConverterPair())
}

func (r *StockOpLogRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().StockOpLog.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query stock op log count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *StockOpLogRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListStockOpLogResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().StockOpLog.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListStockOpLogResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListStockOpLogResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *StockOpLogRepo) Get(ctx context.Context, req *catalogV1.GetStockOpLogRequest) (*catalogV1.StockOpLog, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().StockOpLog.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *catalogV1.GetStockOpLogRequest_Id:
		whereCond = append(whereCond, stockoplog.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

// CreateInTx 在外部事务内追加一条流水（append-only；仅供库存变更点调用）。
// createdAt 取当前时间。
func (r *StockOpLogRepo) CreateInTx(
	ctx context.Context,
	tx *ent.Tx,
	skuId, productId uint32,
	delta, stockBefore, stockAfter int32,
	reason catalogV1.StockOpLog_Reason,
	refId uint32,
	operator uint32,
	remark string,
) error {
	builder := tx.StockOpLog.Create().
		SetNillableSkuID(&skuId).
		SetNillableProductID(&productId).
		SetNillableDelta(&delta).
		SetNillableStockBefore(&stockBefore).
		SetNillableStockAfter(&stockAfter).
		SetNillableReason(r.reasonConverter.ToEntity(&reason)).
		SetNillableRefID(&refId).
		SetNillableCreatedBy(&operator).
		SetCreatedAt(time.Now())
	if remark != "" {
		builder.SetNillableRemark(&remark)
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert stock op log failed: %s", err.Error())
		return err
	}
	return nil
}
