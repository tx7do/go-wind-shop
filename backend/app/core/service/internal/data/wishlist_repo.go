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
	"go-wind-shop/app/core/service/internal/data/ent/wishlist"

	wishlistV1 "go-wind-shop/api/gen/go/wishlist/service/v1"
)

type WishlistRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[wishlistV1.Wishlist, ent.Wishlist]

	repository *entCrud.Repository[
		ent.WishlistQuery, ent.WishlistSelect,
		ent.WishlistCreate, ent.WishlistCreateBulk,
		ent.WishlistUpdate, ent.WishlistUpdateOne,
		ent.WishlistDelete,
		predicate.Wishlist,
		wishlistV1.Wishlist, ent.Wishlist,
	]
}

func NewWishlistRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *WishlistRepo {
	repo := &WishlistRepo{
		log:       ctx.NewLoggerHelper("wishlist/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[wishlistV1.Wishlist, ent.Wishlist](),
	}

	repo.init()

	return repo
}

func (r *WishlistRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.WishlistQuery, ent.WishlistSelect,
		ent.WishlistCreate, ent.WishlistCreateBulk,
		ent.WishlistUpdate, ent.WishlistUpdateOne,
		ent.WishlistDelete,
		predicate.Wishlist,
		wishlistV1.Wishlist, ent.Wishlist,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *WishlistRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*wishlistV1.ListWishlistResponse, error) {
	if req == nil {
		return nil, wishlistV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Wishlist.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &wishlistV1.ListWishlistResponse{Total: 0, Items: nil}, nil
	}

	return &wishlistV1.ListWishlistResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *WishlistRepo) Create(ctx context.Context, req *wishlistV1.CreateWishlistRequest) error {
	if req == nil || req.Data == nil {
		return wishlistV1.ErrorBadRequest("invalid parameter")
	}

	// user_id 由 UserPrivacy 隐私层在 Create 时强制覆盖为当前 viewer，
	// 故此处不设 user_id——客户端传入的 userId 会被忽略。
	builder := r.entClient.Client().Wishlist.Create().
		SetNillableProductID(req.Data.ProductId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert wishlist failed: %s", err.Error())
		return wishlistV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *WishlistRepo) Delete(ctx context.Context, req *wishlistV1.DeleteWishlistRequest) error {
	if req == nil {
		return wishlistV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Wishlist.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(wishlist.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete wishlist failed: %s", err.Error())
		return wishlistV1.ErrorInternalServerError("delete wishlist failed")
	}

	return nil
}
