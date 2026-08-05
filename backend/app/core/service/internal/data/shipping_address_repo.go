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
	"go-wind-shop/app/core/service/internal/data/ent/shippingaddress"

	addressV1 "go-wind-shop/api/gen/go/address/service/v1"
)

type ShippingAddressRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[addressV1.ShippingAddress, ent.ShippingAddress]

	repository *entCrud.Repository[
		ent.ShippingAddressQuery, ent.ShippingAddressSelect,
		ent.ShippingAddressCreate, ent.ShippingAddressCreateBulk,
		ent.ShippingAddressUpdate, ent.ShippingAddressUpdateOne,
		ent.ShippingAddressDelete,
		predicate.ShippingAddress,
		addressV1.ShippingAddress, ent.ShippingAddress,
	]
}

func NewShippingAddressRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *ShippingAddressRepo {
	repo := &ShippingAddressRepo{
		log:       ctx.NewLoggerHelper("shipping-address/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[addressV1.ShippingAddress, ent.ShippingAddress](),
	}

	repo.init()

	return repo
}

func (r *ShippingAddressRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ShippingAddressQuery, ent.ShippingAddressSelect,
		ent.ShippingAddressCreate, ent.ShippingAddressCreateBulk,
		ent.ShippingAddressUpdate, ent.ShippingAddressUpdateOne,
		ent.ShippingAddressDelete,
		predicate.ShippingAddress,
		addressV1.ShippingAddress, ent.ShippingAddress,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *ShippingAddressRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().ShippingAddress.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if err != nil {
		r.log.Errorf("build count selector failed: %s", err.Error())
		return 0, addressV1.ErrorInternalServerError("query count failed")
	}
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query shipping address count failed: %s", err.Error())
		return 0, addressV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *ShippingAddressRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*addressV1.ListShippingAddressResponse, error) {
	if req == nil {
		return nil, addressV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ShippingAddress.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &addressV1.ListShippingAddressResponse{Total: 0, Items: nil}, nil
	}

	return &addressV1.ListShippingAddressResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *ShippingAddressRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().ShippingAddress.Query().
		Where(shippingaddress.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, addressV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ShippingAddressRepo) Get(ctx context.Context, req *addressV1.GetShippingAddressRequest) (*addressV1.ShippingAddress, error) {
	if req == nil {
		return nil, addressV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ShippingAddress.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *addressV1.GetShippingAddressRequest_Id:
		whereCond = append(whereCond, shippingaddress.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

// clearOtherDefaults 把指定用户（tenantId+userId）下的其它地址的 is_default 置 false，
// 用于"默认地址唯一"互斥：当新增/更新某地址为默认时，先清掉旧默认。
func (r *ShippingAddressRepo) clearOtherDefaults(ctx context.Context, tenantId, userId, exceptId uint32) error {
	if userId == 0 {
		return nil
	}
	builder := r.entClient.Client().ShippingAddress.Update()
	err := r.repository.UpdateX(ctx, builder, &addressV1.ShippingAddress{}, nil,
		func(_ *addressV1.ShippingAddress) {
			builder.
				SetIsDefault(false).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(
				sql.And(
					sql.EQ(shippingaddress.FieldUserID, userId),
					sql.EQ(shippingaddress.FieldIsDefault, true),
				),
			)
			if tenantId != 0 {
				s.Where(sql.EQ(shippingaddress.FieldTenantID, tenantId))
			}
			if exceptId != 0 {
				s.Where(sql.NEQ(shippingaddress.FieldID, exceptId))
			}
		},
	)
	if err != nil {
		r.log.Errorf("clear other defaults failed: %s", err.Error())
		return addressV1.ErrorInternalServerError("clear other defaults failed")
	}
	return nil
}

func (r *ShippingAddressRepo) Create(ctx context.Context, req *addressV1.CreateShippingAddressRequest) error {
	if req == nil || req.Data == nil {
		return addressV1.ErrorBadRequest("invalid parameter")
	}

	// 默认地址互斥：若新建为默认，先把该用户其它默认地址置 false。
	if req.Data.GetIsDefault() {
		if err := r.clearOtherDefaults(ctx, req.Data.GetTenantId(), req.Data.GetUserId(), 0); err != nil {
			return err
		}
	}

	builder := r.entClient.Client().ShippingAddress.Create().
		SetNillableUserID(req.Data.UserId).
		SetNillableRecipientName(req.Data.RecipientName).
		SetNillableRecipientPhone(req.Data.RecipientPhone).
		SetNillableRegion(req.Data.Region).
		SetNillableDetailAddress(req.Data.DetailAddress).
		SetNillablePostalCode(req.Data.PostalCode).
		SetNillableTag(req.Data.Tag).
		SetNillableIsDefault(req.Data.IsDefault).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert shipping address failed: %s", err.Error())
		return addressV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *ShippingAddressRepo) Update(ctx context.Context, req *addressV1.UpdateShippingAddressRequest) error {
	if req == nil || req.Data == nil {
		return addressV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return addressV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &addressV1.CreateShippingAddressRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	// 默认地址互斥：若更新为默认，先把该用户其它默认地址置 false（排除自身）。
	if req.Data.GetIsDefault() {
		var tenantId, userId uint32
		if req.Data.TenantId != nil {
			tenantId = req.Data.GetTenantId()
		}
		if req.Data.UserId != nil {
			userId = req.Data.GetUserId()
		} else {
			// userId 未随更新携带时，回查该地址的归属用户，确保互斥范围正确。
			existing, err := r.Get(ctx, &addressV1.GetShippingAddressRequest{
				QueryBy: &addressV1.GetShippingAddressRequest_Id{Id: req.GetId()},
			})
			if err == nil && existing != nil {
				userId = existing.GetUserId()
				if tenantId == 0 {
					tenantId = existing.GetTenantId()
				}
			}
		}
		if err := r.clearOtherDefaults(ctx, tenantId, userId, req.GetId()); err != nil {
			return err
		}
	}

	builder := r.entClient.Client().ShippingAddress.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *addressV1.ShippingAddress) {
			builder.
				SetNillableUserID(req.Data.UserId).
				SetNillableRecipientName(req.Data.RecipientName).
				SetNillableRecipientPhone(req.Data.RecipientPhone).
				SetNillableRegion(req.Data.Region).
				SetNillableDetailAddress(req.Data.DetailAddress).
				SetNillablePostalCode(req.Data.PostalCode).
				SetNillableTag(req.Data.Tag).
				SetNillableIsDefault(req.Data.IsDefault).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(shippingaddress.FieldID, req.GetId()))
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ShippingAddressRepo) Delete(ctx context.Context, req *addressV1.DeleteShippingAddressRequest) error {
	if req == nil {
		return addressV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ShippingAddress.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(shippingaddress.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete shipping address failed: %s", err.Error())
		return addressV1.ErrorInternalServerError("delete shipping address failed")
	}

	return nil
}
