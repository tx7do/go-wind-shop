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
	"go-wind-shop/app/core/service/internal/data/ent/paymenttransaction"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

type PaymentTransactionRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[paymentV1.PaymentTransaction, ent.PaymentTransaction]

	statusConverter         *mapper.EnumTypeConverter[paymentV1.PaymentTransaction_Status, paymenttransaction.Status]
	paymentMethodConverter  *mapper.EnumTypeConverter[paymentV1.PaymentMethod, paymenttransaction.PaymentMethod]
	businessTypeConverter   *mapper.EnumTypeConverter[paymentV1.BusinessType, paymenttransaction.BusinessType]

	repository *entCrud.Repository[
		ent.PaymentTransactionQuery, ent.PaymentTransactionSelect,
		ent.PaymentTransactionCreate, ent.PaymentTransactionCreateBulk,
		ent.PaymentTransactionUpdate, ent.PaymentTransactionUpdateOne,
		ent.PaymentTransactionDelete,
		predicate.PaymentTransaction,
		paymentV1.PaymentTransaction, ent.PaymentTransaction,
	]
}

func NewPaymentTransactionRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *PaymentTransactionRepo {
	repo := &PaymentTransactionRepo{
		log:                     ctx.NewLoggerHelper("payment-transaction/repo/core-service"),
		entClient:               entClient,
		mapper:                  mapper.NewCopierMapper[paymentV1.PaymentTransaction, ent.PaymentTransaction](),
		statusConverter:         mapper.NewEnumTypeConverter[paymentV1.PaymentTransaction_Status, paymenttransaction.Status](paymentV1.PaymentTransaction_Status_name, paymentV1.PaymentTransaction_Status_value),
		paymentMethodConverter:  mapper.NewEnumTypeConverter[paymentV1.PaymentMethod, paymenttransaction.PaymentMethod](paymentV1.PaymentMethod_name, paymentV1.PaymentMethod_value),
		businessTypeConverter:   mapper.NewEnumTypeConverter[paymentV1.BusinessType, paymenttransaction.BusinessType](paymentV1.BusinessType_name, paymentV1.BusinessType_value),
	}

	repo.init()

	return repo
}

func (r *PaymentTransactionRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.PaymentTransactionQuery, ent.PaymentTransactionSelect,
		ent.PaymentTransactionCreate, ent.PaymentTransactionCreateBulk,
		ent.PaymentTransactionUpdate, ent.PaymentTransactionUpdateOne,
		ent.PaymentTransactionDelete,
		predicate.PaymentTransaction,
		paymentV1.PaymentTransaction, ent.PaymentTransaction,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
	r.mapper.AppendConverters(r.paymentMethodConverter.NewConverterPair())
	r.mapper.AppendConverters(r.businessTypeConverter.NewConverterPair())
}

func (r *PaymentTransactionRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().PaymentTransaction.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, paymentV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *PaymentTransactionRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentTransactionResponse, error) {
	if req == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentTransaction.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &paymentV1.ListPaymentTransactionResponse{Total: 0, Items: nil}, nil
	}

	return &paymentV1.ListPaymentTransactionResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *PaymentTransactionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().PaymentTransaction.Query().
		Where(paymenttransaction.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, paymentV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *PaymentTransactionRepo) Get(ctx context.Context, req *paymentV1.GetPaymentTransactionRequest) (*paymentV1.PaymentTransaction, error) {
	if req == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentTransaction.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *paymentV1.GetPaymentTransactionRequest_Id:
		whereCond = append(whereCond, paymenttransaction.IDEQ(req.GetId()))

	case *paymentV1.GetPaymentTransactionRequest_BusinessRefId:
		// business_ref_id 非全局唯一，必须配合 tenant_id 过滤，防止跨租户命中。
		tenantId := req.GetTenantId()
		whereCond = append(whereCond, func(s *sql.Selector) {
			s.Where(
				sql.And(
					sql.EQ(paymenttransaction.FieldBusinessRefID, req.GetBusinessRefId()),
					sql.EQ(paymenttransaction.FieldTenantID, tenantId),
				),
			)
		})
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *PaymentTransactionRepo) Create(ctx context.Context, req *paymentV1.CreatePaymentTransactionRequest) error {
	if req == nil || req.Data == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentTransaction.Create().
		SetNillableOrderID(req.Data.OrderId).
		SetNillableUserID(req.Data.UserId).
		SetNillableAmount(req.Data.Amount).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableBusinessRefID(req.Data.BusinessRefId).
		SetNillableIdempotencyKey(req.Data.IdempotencyKey).
		SetNillablePaymentMethod(r.paymentMethodConverter.ToEntity(req.Data.PaymentMethod)).
		SetNillableBusinessType(r.businessTypeConverter.ToEntity(req.Data.BusinessType)).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert payment transaction failed: %s", err.Error())
		return paymentV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *PaymentTransactionRepo) Update(ctx context.Context, req *paymentV1.UpdatePaymentTransactionRequest) error {
	if req == nil || req.Data == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &paymentV1.CreatePaymentTransactionRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().PaymentTransaction.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *paymentV1.PaymentTransaction) {
			builder.
				SetNillableOrderID(req.Data.OrderId).
				SetNillableUserID(req.Data.UserId).
				SetNillableAmount(req.Data.Amount).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableBusinessRefID(req.Data.BusinessRefId).
				SetNillableIdempotencyKey(req.Data.IdempotencyKey).
				SetNillablePaymentMethod(r.paymentMethodConverter.ToEntity(req.Data.PaymentMethod)).
				SetNillableBusinessType(r.businessTypeConverter.ToEntity(req.Data.BusinessType)).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(paymenttransaction.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *PaymentTransactionRepo) Delete(ctx context.Context, req *paymentV1.DeletePaymentTransactionRequest) error {
	if req == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().PaymentTransaction.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(paymenttransaction.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete payment transaction failed: %s", err.Error())
		return paymentV1.ErrorInternalServerError("delete payment transaction failed")
	}

	return nil
}
