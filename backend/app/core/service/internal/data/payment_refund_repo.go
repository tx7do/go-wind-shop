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
	"go-wind-shop/app/core/service/internal/data/ent/paymentrefund"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

type PaymentRefundRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[paymentV1.PaymentRefund, ent.PaymentRefund]

	statusConverter *mapper.EnumTypeConverter[paymentV1.PaymentRefund_Status, paymentrefund.Status]

	repository *entCrud.Repository[
		ent.PaymentRefundQuery, ent.PaymentRefundSelect,
		ent.PaymentRefundCreate, ent.PaymentRefundCreateBulk,
		ent.PaymentRefundUpdate, ent.PaymentRefundUpdateOne,
		ent.PaymentRefundDelete,
		predicate.PaymentRefund,
		paymentV1.PaymentRefund, ent.PaymentRefund,
	]
}

func NewPaymentRefundRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *PaymentRefundRepo {
	repo := &PaymentRefundRepo{
		log:             ctx.NewLoggerHelper("payment-refund/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[paymentV1.PaymentRefund, ent.PaymentRefund](),
		statusConverter: mapper.NewEnumTypeConverter[paymentV1.PaymentRefund_Status, paymentrefund.Status](paymentV1.PaymentRefund_Status_name, paymentV1.PaymentRefund_Status_value),
	}

	repo.init()

	return repo
}

func (r *PaymentRefundRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.PaymentRefundQuery, ent.PaymentRefundSelect,
		ent.PaymentRefundCreate, ent.PaymentRefundCreateBulk,
		ent.PaymentRefundUpdate, ent.PaymentRefundUpdateOne,
		ent.PaymentRefundDelete,
		predicate.PaymentRefund,
		paymentV1.PaymentRefund, ent.PaymentRefund,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *PaymentRefundRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().PaymentRefund.Query()
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

func (r *PaymentRefundRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentRefundResponse, error) {
	if req == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentRefund.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &paymentV1.ListPaymentRefundResponse{Total: 0, Items: nil}, nil
	}

	return &paymentV1.ListPaymentRefundResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *PaymentRefundRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().PaymentRefund.Query().
		Where(paymentrefund.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, paymentV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *PaymentRefundRepo) Get(ctx context.Context, req *paymentV1.GetPaymentRefundRequest) (*paymentV1.PaymentRefund, error) {
	if req == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentRefund.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *paymentV1.GetPaymentRefundRequest_Id:
		whereCond = append(whereCond, paymentrefund.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *PaymentRefundRepo) Create(ctx context.Context, req *paymentV1.CreatePaymentRefundRequest) error {
	if req == nil || req.Data == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentRefund.Create().
		SetNillableTransactionID(req.Data.TransactionId).
		SetNillableAmount(req.Data.Amount).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableBusinessRefID(req.Data.BusinessRefId).
		SetNillableIdempotencyKey(req.Data.IdempotencyKey).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert payment refund failed: %s", err.Error())
		return paymentV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *PaymentRefundRepo) Update(ctx context.Context, req *paymentV1.UpdatePaymentRefundRequest) error {
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
			createReq := &paymentV1.CreatePaymentRefundRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().PaymentRefund.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *paymentV1.PaymentRefund) {
			builder.
				SetNillableTransactionID(req.Data.TransactionId).
				SetNillableAmount(req.Data.Amount).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableBusinessRefID(req.Data.BusinessRefId).
				SetNillableIdempotencyKey(req.Data.IdempotencyKey).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(paymentrefund.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *PaymentRefundRepo) Delete(ctx context.Context, req *paymentV1.DeletePaymentRefundRequest) error {
	if req == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().PaymentRefund.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(paymentrefund.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete payment refund failed: %s", err.Error())
		return paymentV1.ErrorInternalServerError("delete payment refund failed")
	}

	return nil
}
