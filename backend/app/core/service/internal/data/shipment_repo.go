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

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/shipment"

	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
)

type ShipmentRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[shippingV1.Shipment, ent.Shipment]

	statusConverter *mapper.EnumTypeConverter[shippingV1.Shipment_Status, shipment.Status]

	repository *entCrud.Repository[
		ent.ShipmentQuery, ent.ShipmentSelect,
		ent.ShipmentCreate, ent.ShipmentCreateBulk,
		ent.ShipmentUpdate, ent.ShipmentUpdateOne,
		ent.ShipmentDelete,
		predicate.Shipment,
		shippingV1.Shipment, ent.Shipment,
	]
}

func NewShipmentRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *ShipmentRepo {
	repo := &ShipmentRepo{
		log:             ctx.NewLoggerHelper("shipment/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[shippingV1.Shipment, ent.Shipment](),
		statusConverter: mapper.NewEnumTypeConverter[shippingV1.Shipment_Status, shipment.Status](shippingV1.Shipment_Status_name, shippingV1.Shipment_Status_value),
	}

	repo.init()

	return repo
}

func (r *ShipmentRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ShipmentQuery, ent.ShipmentSelect,
		ent.ShipmentCreate, ent.ShipmentCreateBulk,
		ent.ShipmentUpdate, ent.ShipmentUpdateOne,
		ent.ShipmentDelete,
		predicate.Shipment,
		shippingV1.Shipment, ent.Shipment,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *ShipmentRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().Shipment.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query shipment count failed: %s", err.Error())
		return 0, shippingV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *ShipmentRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShipmentResponse, error) {
	if req == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Shipment.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &shippingV1.ListShipmentResponse{Total: 0, Items: nil}, nil
	}

	return &shippingV1.ListShipmentResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *ShipmentRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Shipment.Query().
		Where(shipment.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, shippingV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ShipmentRepo) Get(ctx context.Context, req *shippingV1.GetShipmentRequest) (*shippingV1.Shipment, error) {
	if req == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Shipment.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *shippingV1.GetShipmentRequest_Id:
		whereCond = append(whereCond, shipment.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *ShipmentRepo) Create(ctx context.Context, req *shippingV1.CreateShipmentRequest) error {
	if req == nil || req.Data == nil {
		return shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Shipment.Create().
		SetNillableOrderID(req.Data.OrderId).
		SetNillableUserID(req.Data.UserId).
		SetNillableCarrier(req.Data.Carrier).
		SetNillableTrackingNumber(req.Data.TrackingNumber).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableTrackingEvents(req.Data.TrackingEvents).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert shipment failed: %s", err.Error())
		return shippingV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *ShipmentRepo) Update(ctx context.Context, req *shippingV1.UpdateShipmentRequest) error {
	if req == nil || req.Data == nil {
		return shippingV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return shippingV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &shippingV1.CreateShipmentRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Shipment.Update()

	// 状态前置条件：expected_status 非空时，把条件注入 SQL selector，
	// 仅当物流单当前 status ∈ expected_status 时 UPDATE 才命中（affected_rows=0 表示状态已被并发推进）。
	expectedStatusVals := make([]any, 0, len(req.GetExpectedStatus()))
	for _, st := range req.GetExpectedStatus() {
		st := st // 取地址前避免 range 变量复用
		if v := r.statusConverter.ToEntity(&st); v != nil {
			expectedStatusVals = append(expectedStatusVals, *v)
		}
	}

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *shippingV1.Shipment) {
			builder.
				SetNillableOrderID(req.Data.OrderId).
				SetNillableUserID(req.Data.UserId).
				SetNillableCarrier(req.Data.Carrier).
				SetNillableTrackingNumber(req.Data.TrackingNumber).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableTrackingEvents(req.Data.TrackingEvents).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(shipment.FieldID, req.GetId()))
			if len(expectedStatusVals) > 0 {
				s.Where(sql.In(shipment.FieldStatus, expectedStatusVals...))
			}
		},
	)
	if err != nil {
		// UpdateX 在 WHERE 条件不匹配（状态已被并发推进）时返回 NotFound。
		// 此时若传入了 expected_status，说明是状态前置条件失败，返回 Conflict 供调用方识别。
		if ent.IsNotFound(err) && len(expectedStatusVals) > 0 {
			r.log.Infof("update shipment [%d] precondition failed: status not in expected set", req.GetId())
			return shippingV1.ErrorConflict("shipment status precondition failed")
		}
		return err
	}

	return nil
}

func (r *ShipmentRepo) Delete(ctx context.Context, req *shippingV1.DeleteShipmentRequest) error {
	if req == nil {
		return shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Shipment.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(shipment.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete shipment failed: %s", err.Error())
		return shippingV1.ErrorInternalServerError("delete shipment failed")
	}

	return nil
}
