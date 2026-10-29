package data

import (
	"go-wind-shop/pkg/logx"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/paymentchannelconfig"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

type PaymentChannelConfigRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[paymentV1.PaymentChannelConfig, ent.PaymentChannelConfig]

	channelConverter *mapper.EnumTypeConverter[paymentV1.PaymentChannelType, paymentchannelconfig.Channel]

	repository *entCrud.Repository[
		ent.PaymentChannelConfigQuery, ent.PaymentChannelConfigSelect,
		ent.PaymentChannelConfigCreate, ent.PaymentChannelConfigCreateBulk,
		ent.PaymentChannelConfigUpdate, ent.PaymentChannelConfigUpdateOne,
		ent.PaymentChannelConfigDelete,
		predicate.PaymentChannelConfig,
		paymentV1.PaymentChannelConfig, ent.PaymentChannelConfig,
	]
}

func NewPaymentChannelConfigRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *PaymentChannelConfigRepo {
	repo := &PaymentChannelConfigRepo{
		log:              logx.New(ctx.NewLoggerHelper("payment-channel-config/repo/core-service")),
		entClient:        entClient,
		mapper:           mapper.NewCopierMapper[paymentV1.PaymentChannelConfig, ent.PaymentChannelConfig](),
		channelConverter: mapper.NewEnumTypeConverter[paymentV1.PaymentChannelType, paymentchannelconfig.Channel](paymentV1.PaymentChannelType_name, paymentV1.PaymentChannelType_value),
	}

	repo.init()

	return repo
}

func (r *PaymentChannelConfigRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.PaymentChannelConfigQuery, ent.PaymentChannelConfigSelect,
		ent.PaymentChannelConfigCreate, ent.PaymentChannelConfigCreateBulk,
		ent.PaymentChannelConfigUpdate, ent.PaymentChannelConfigUpdateOne,
		ent.PaymentChannelConfigDelete,
		predicate.PaymentChannelConfig,
		paymentV1.PaymentChannelConfig, ent.PaymentChannelConfig,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.channelConverter.NewConverterPair())
}

func (r *PaymentChannelConfigRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().PaymentChannelConfig.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query payment channel config count failed: %s", err.Error())
		return 0, paymentV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *PaymentChannelConfigRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentChannelConfigResponse, error) {
	if req == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentChannelConfig.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &paymentV1.ListPaymentChannelConfigResponse{Total: 0, Items: nil}, nil
	}

	return &paymentV1.ListPaymentChannelConfigResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListEnabled 列出启用中的渠道。C 端结算页渠道选择用；
// 返回脱敏的公开视图（无凭据/配置细节）。
func (r *PaymentChannelConfigRepo) ListEnabled(ctx context.Context) (*paymentV1.ListEnabledChannelsResponse, error) {
	builder := r.entClient.Client().PaymentChannelConfig.Query().
		Where(paymentchannelconfig.EnabledEQ(true))

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), &paginationV1.PagingRequest{
		NoPaging: trans.Ptr(true),
	})
	if err != nil {
		r.log.Errorf("query enabled payment channels failed: %s", err.Error())
		return nil, paymentV1.ErrorInternalServerError("query failed")
	}

	resp := &paymentV1.ListEnabledChannelsResponse{Items: []*paymentV1.PaymentChannel{}}
	if ret == nil {
		return resp, nil
	}
	for _, cfg := range ret.Items {
		resp.Items = append(resp.Items, &paymentV1.PaymentChannel{
			Id:      cfg.Id,
			Name:    cfg.Name,
			Channel: cfg.Channel,
			Sandbox: cfg.Sandbox,
		})
	}

	return resp, nil
}

func (r *PaymentChannelConfigRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().PaymentChannelConfig.Query().
		Where(paymentchannelconfig.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, paymentV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *PaymentChannelConfigRepo) Get(ctx context.Context, req *paymentV1.GetPaymentChannelConfigRequest) (*paymentV1.PaymentChannelConfig, error) {
	if req == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentChannelConfig.Query()

	var whereCond []func(s *sql.Selector)
	whereCond = append(whereCond, paymentchannelconfig.IDEQ(req.GetId()))

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *PaymentChannelConfigRepo) Create(ctx context.Context, req *paymentV1.CreatePaymentChannelConfigRequest) error {
	if req == nil || req.Data == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentChannelConfig.Create().
		SetName(req.Data.GetName()).
		SetNillableChannel(r.channelConverter.ToEntity(req.Data.Channel)).
		SetNillableAppID(req.Data.AppId).
		SetNillableMchID(req.Data.MchId).
		SetNillableSandbox(req.Data.Sandbox).
		SetNillableEnabled(req.Data.Enabled).
		SetNillableCredentials(req.Data.Credentials).
		SetNillableRemark(req.Data.Remark).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}


	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert payment channel config failed: %s", err.Error())
		return paymentV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *PaymentChannelConfigRepo) Update(ctx context.Context, req *paymentV1.UpdatePaymentChannelConfigRequest) error {
	if req == nil || req.Data == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return paymentV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &paymentV1.CreatePaymentChannelConfigRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().PaymentChannelConfig.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *paymentV1.PaymentChannelConfig) {
			builder.
				SetNillableName(req.Data.Name).
				SetNillableChannel(r.channelConverter.ToEntity(req.Data.Channel)).
				SetNillableAppID(req.Data.AppId).
				SetNillableMchID(req.Data.MchId).
				SetNillableSandbox(req.Data.Sandbox).
				SetNillableEnabled(req.Data.Enabled).
				SetNillableCredentials(req.Data.Credentials).
				SetNillableRemark(req.Data.Remark).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(paymentchannelconfig.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *PaymentChannelConfigRepo) Delete(ctx context.Context, req *paymentV1.DeletePaymentChannelConfigRequest) error {
	if req == nil {
		return paymentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().PaymentChannelConfig.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(paymentchannelconfig.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete payment channel config failed: %s", err.Error())
		return paymentV1.ErrorInternalServerError("delete data failed")
	}

	return nil
}
