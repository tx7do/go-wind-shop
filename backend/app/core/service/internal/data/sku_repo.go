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
	"go-wind-shop/app/core/service/internal/data/ent/sku"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.Sku, ent.Sku]

	repository *entCrud.Repository[
		ent.SkuQuery, ent.SkuSelect,
		ent.SkuCreate, ent.SkuCreateBulk,
		ent.SkuUpdate, ent.SkuUpdateOne,
		ent.SkuDelete,
		predicate.Sku,
		catalogV1.Sku, ent.Sku,
	]
}

func NewSkuRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *SkuRepo {
	repo := &SkuRepo{
		log:       ctx.NewLoggerHelper("sku/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.Sku, ent.Sku](),
	}

	repo.init()

	return repo
}

func (r *SkuRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.SkuQuery, ent.SkuSelect,
		ent.SkuCreate, ent.SkuCreateBulk,
		ent.SkuUpdate, ent.SkuUpdateOne,
		ent.SkuDelete,
		predicate.Sku,
		catalogV1.Sku, ent.Sku,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *SkuRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Sku.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *SkuRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Sku.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListSkuResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListSkuResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *SkuRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Sku.Query().
		Where(sku.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *SkuRepo) Get(ctx context.Context, req *catalogV1.GetSkuRequest) (*catalogV1.Sku, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Sku.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *catalogV1.GetSkuRequest_Id:
		whereCond = append(whereCond, sku.IDEQ(req.GetId()))
	case *catalogV1.GetSkuRequest_Code:
		whereCond = append(whereCond, sku.SkuCodeEQ(req.GetCode()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *SkuRepo) Create(ctx context.Context, req *catalogV1.CreateSkuRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Sku.Create().
		SetNillableProductID(req.Data.ProductId).
		SetNillableSkuCode(req.Data.SkuCode).
		SetNillableStockQty(req.Data.StockQty).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert sku failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *SkuRepo) Update(ctx context.Context, req *catalogV1.UpdateSkuRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &catalogV1.CreateSkuRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Sku.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.Sku) {
			builder.
				SetNillableProductID(req.Data.ProductId).
				SetNillableSkuCode(req.Data.SkuCode).
				SetNillableStockQty(req.Data.StockQty).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(sku.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *SkuRepo) Delete(ctx context.Context, req *catalogV1.DeleteSkuRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Sku.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(sku.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete sku failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete sku failed")
	}

	return nil
}

// AddStock 将指定 SKU 的库存加回 delta（用于订单超时取消/退款时释放占用库存）。
func (r *SkuRepo) AddStock(ctx context.Context, skuId uint32, delta int32) error {
	if skuId == 0 {
		return catalogV1.ErrorBadRequest("invalid sku id")
	}
	if err := r.entClient.Client().Sku.UpdateOneID(skuId).
		AddStockQty(delta).
		Exec(ctx); err != nil {
		r.log.Errorf("add stock for sku [%d] failed: %s", skuId, err.Error())
		return catalogV1.ErrorInternalServerError("add stock failed")
	}
	return nil
}
