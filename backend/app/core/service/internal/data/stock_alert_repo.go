package data

import (
	"context"
	"encoding/json"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/stockalert"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type StockAlertRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper         *mapper.CopierMapper[catalogV1.StockAlert, ent.StockAlert]
	statusConverter *mapper.EnumTypeConverter[catalogV1.StockAlert_Status, stockalert.Status]

	repository *entCrud.Repository[
		ent.StockAlertQuery, ent.StockAlertSelect,
		ent.StockAlertCreate, ent.StockAlertCreateBulk,
		ent.StockAlertUpdate, ent.StockAlertUpdateOne,
		ent.StockAlertDelete,
		predicate.StockAlert,
		catalogV1.StockAlert, ent.StockAlert,
	]
}

func NewStockAlertRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *StockAlertRepo {
	repo := &StockAlertRepo{
		log:             ctx.NewLoggerHelper("stock-alert/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[catalogV1.StockAlert, ent.StockAlert](),
		statusConverter: mapper.NewEnumTypeConverter[catalogV1.StockAlert_Status, stockalert.Status](catalogV1.StockAlert_Status_name, catalogV1.StockAlert_Status_value),
	}

	repo.init()

	return repo
}

func (r *StockAlertRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.StockAlertQuery, ent.StockAlertSelect,
		ent.StockAlertCreate, ent.StockAlertCreateBulk,
		ent.StockAlertUpdate, ent.StockAlertUpdateOne,
		ent.StockAlertDelete,
		predicate.StockAlert,
		catalogV1.StockAlert, ent.StockAlert,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *StockAlertRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListStockAlertResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().StockAlert.Query()

	// status 枚举过滤兜底：go-crud 的通用分页转换器不原生支持枚举 EQ，
	// 故在此显式解析 req.Query 中的 status 键，构造 stockalert.StatusEQ 谓词，
	// 并从 JSON 中移除 status 键后重序列化回 req.Query（让通用层处理其余字段）。
	var whereCond []func(s *sql.Selector)
	if raw := req.GetQuery(); raw != "" {
		var qm map[string]any
		if jErr := json.Unmarshal([]byte(raw), &qm); jErr == nil {
			if sv, ok := qm["status"]; ok {
				if sStr, ok2 := sv.(string); ok2 {
					if v, ok3 := catalogV1.StockAlert_Status_value[sStr]; ok3 {
						entStatus := stockalert.Status(sStr)
						_ = v
						whereCond = append(whereCond, stockalert.StatusEQ(entStatus))
					}
				}
			}
			delete(qm, "status")
			if len(qm) == 0 {
				req.FilteringType = &paginationV1.PagingRequest_Query{Query: ""}
			} else if newJSON, mErr := json.Marshal(qm); mErr == nil {
				req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newJSON)}
			}
		}
	}

	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListStockAlertResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListStockAlertResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *StockAlertRepo) Get(ctx context.Context, req *catalogV1.GetStockAlertRequest) (*catalogV1.StockAlert, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().StockAlert.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *catalogV1.GetStockAlertRequest_Id:
		whereCond = append(whereCond, stockalert.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *StockAlertRepo) Update(ctx context.Context, req *catalogV1.UpdateStockAlertRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return catalogV1.ErrorBadRequest("id is required")
	}

	builder := r.entClient.Client().Debug().StockAlert.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.StockAlert) {
			builder.
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(stockalert.FieldID, req.GetId()))
		},
	)

	return err
}

// CreateAlertIfNotOpen 供 StockAlertScannerService 周期任务调用：
// 仅当该 SKU 不存在 OPEN 态告警时插入一条 OPEN 记录（去重，单线程 cron 无
// TOCTOU 风险）。stock_qty_at_trigger/threshold 为检测时的快照值。created_by
// 强制 0（系统任务），created_at 为当前时间。
func (r *StockAlertRepo) CreateAlertIfNotOpen(ctx context.Context, skuId uint32, stockQty int32, threshold int32) error {
	exists, err := r.entClient.Client().StockAlert.Query().
		Where(
			stockalert.SkuIDEQ(skuId),
			stockalert.StatusEQ(stockalert.StatusStockAlertStatusOpen),
		).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query stock alert existence failed: %s", err.Error())
		return err
	}
	if exists {
		return nil
	}

	builder := r.entClient.Client().StockAlert.Create().
		SetSkuID(skuId).
		SetStockQtyAtTrigger(stockQty).
		SetThreshold(threshold).
		SetStatus(stockalert.StatusStockAlertStatusOpen).
		SetCreatedBy(0).
		SetCreatedAt(time.Now())

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert stock alert failed: %s", err.Error())
		return err
	}

	return nil
}
