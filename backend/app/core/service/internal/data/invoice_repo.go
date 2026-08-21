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
	"go-wind-shop/app/core/service/internal/data/ent/invoice"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	invoiceV1 "go-wind-shop/api/gen/go/invoice/service/v1"
)

type InvoiceRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[invoiceV1.Invoice, ent.Invoice]

	statusConverter      *mapper.EnumTypeConverter[invoiceV1.Invoice_Status, invoice.Status]
	invoiceTypeConverter *mapper.EnumTypeConverter[invoiceV1.Invoice_InvoiceType, invoice.InvoiceType]

	repository *entCrud.Repository[
		ent.InvoiceQuery, ent.InvoiceSelect,
		ent.InvoiceCreate, ent.InvoiceCreateBulk,
		ent.InvoiceUpdate, ent.InvoiceUpdateOne,
		ent.InvoiceDelete,
		predicate.Invoice,
		invoiceV1.Invoice, ent.Invoice,
	]
}

func NewInvoiceRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *InvoiceRepo {
	repo := &InvoiceRepo{
		log:                   ctx.NewLoggerHelper("invoice/repo/core-service"),
		entClient:             entClient,
		mapper:                mapper.NewCopierMapper[invoiceV1.Invoice, ent.Invoice](),
		statusConverter:       mapper.NewEnumTypeConverter[invoiceV1.Invoice_Status, invoice.Status](invoiceV1.Invoice_Status_name, invoiceV1.Invoice_Status_value),
		invoiceTypeConverter:  mapper.NewEnumTypeConverter[invoiceV1.Invoice_InvoiceType, invoice.InvoiceType](invoiceV1.Invoice_InvoiceType_name, invoiceV1.Invoice_InvoiceType_value),
	}

	repo.init()

	return repo
}

func (r *InvoiceRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.InvoiceQuery, ent.InvoiceSelect,
		ent.InvoiceCreate, ent.InvoiceCreateBulk,
		ent.InvoiceUpdate, ent.InvoiceUpdateOne,
		ent.InvoiceDelete,
		predicate.Invoice,
		invoiceV1.Invoice, ent.Invoice,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
	r.mapper.AppendConverters(r.invoiceTypeConverter.NewConverterPair())
}

func (r *InvoiceRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().Invoice.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query invoice count failed: %s", err.Error())
		return 0, invoiceV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *InvoiceRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*invoiceV1.ListInvoiceResponse, error) {
	if req == nil {
		return nil, invoiceV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Invoice.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &invoiceV1.ListInvoiceResponse{Total: 0, Items: nil}, nil
	}

	return &invoiceV1.ListInvoiceResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *InvoiceRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Invoice.Query().
		Where(invoice.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, invoiceV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *InvoiceRepo) Get(ctx context.Context, req *invoiceV1.GetInvoiceRequest) (*invoiceV1.Invoice, error) {
	if req == nil {
		return nil, invoiceV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Invoice.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *invoiceV1.GetInvoiceRequest_Id:
		whereCond = append(whereCond, invoice.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *InvoiceRepo) Create(ctx context.Context, req *invoiceV1.CreateInvoiceRequest) error {
	if req == nil || req.Data == nil {
		return invoiceV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Invoice.Create().
		SetNillableOrderID(req.Data.OrderId).
		SetNillableUserID(req.Data.UserId).
		SetNillableInvoiceNumber(req.Data.InvoiceNumber).
		SetNillableInvoiceType(r.invoiceTypeConverter.ToEntity(req.Data.InvoiceType)).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableBuyerName(req.Data.BuyerName).
		SetNillableBuyerTaxID(req.Data.BuyerTaxId).
		SetNillableBuyerAddress(req.Data.BuyerAddress).
		SetNillableBuyerPhone(req.Data.BuyerPhone).
		SetNillableAmount(req.Data.Amount).
		SetNillableCurrency(req.Data.Currency).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert invoice failed: %s", err.Error())
		return invoiceV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *InvoiceRepo) Update(ctx context.Context, req *invoiceV1.UpdateInvoiceRequest) error {
	if req == nil || req.Data == nil {
		return invoiceV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return invoiceV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &invoiceV1.CreateInvoiceRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Invoice.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *invoiceV1.Invoice) {
			builder.
				SetNillableOrderID(req.Data.OrderId).
				SetNillableUserID(req.Data.UserId).
				SetNillableInvoiceNumber(req.Data.InvoiceNumber).
				SetNillableInvoiceType(r.invoiceTypeConverter.ToEntity(req.Data.InvoiceType)).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableBuyerName(req.Data.BuyerName).
				SetNillableBuyerTaxID(req.Data.BuyerTaxId).
				SetNillableBuyerAddress(req.Data.BuyerAddress).
				SetNillableBuyerPhone(req.Data.BuyerPhone).
				SetNillableAmount(req.Data.Amount).
				SetNillableCurrency(req.Data.Currency).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(invoice.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *InvoiceRepo) Delete(ctx context.Context, req *invoiceV1.DeleteInvoiceRequest) error {
	if req == nil {
		return invoiceV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Invoice.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(invoice.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete invoice failed: %s", err.Error())
		return invoiceV1.ErrorInternalServerError("delete invoice failed")
	}

	return nil
}
