package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/go-crud/pagination"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/comment"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"
)

type CommentRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[commentV1.Comment, ent.Comment]

	statusConverter      *mapper.EnumTypeConverter[commentV1.Comment_Status, comment.Status]
	contentTypeConverter *mapper.EnumTypeConverter[commentV1.Comment_ContentType, comment.ContentType]

	repository *entCrud.Repository[
		ent.CommentQuery, ent.CommentSelect,
		ent.CommentCreate, ent.CommentCreateBulk,
		ent.CommentUpdate, ent.CommentUpdateOne,
		ent.CommentDelete,
		predicate.Comment,
		commentV1.Comment, ent.Comment,
	]
}

func NewCommentRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *CommentRepo {
	repo := &CommentRepo{
		log:                   ctx.NewLoggerHelper("comment/repo/core-service"),
		entClient:             entClient,
		mapper:                mapper.NewCopierMapper[commentV1.Comment, ent.Comment](),
		statusConverter:       mapper.NewEnumTypeConverter[commentV1.Comment_Status, comment.Status](commentV1.Comment_Status_name, commentV1.Comment_Status_value),
		contentTypeConverter:  mapper.NewEnumTypeConverter[commentV1.Comment_ContentType, comment.ContentType](commentV1.Comment_ContentType_name, commentV1.Comment_ContentType_value),
	}

	repo.init()

	return repo
}

func (r *CommentRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CommentQuery, ent.CommentSelect,
		ent.CommentCreate, ent.CommentCreateBulk,
		ent.CommentUpdate, ent.CommentUpdateOne,
		ent.CommentDelete,
		predicate.Comment,
		commentV1.Comment, ent.Comment,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
	r.mapper.AppendConverters(r.contentTypeConverter.NewConverterPair())
}

func (r *CommentRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().Comment.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query comment count failed: %s", err.Error())
		return 0, commentV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *CommentRepo) List(ctx context.Context, req *paginationV1.PagingRequest, treeTravel bool) (*commentV1.ListCommentResponse, error) {
	if req == nil {
		return nil, commentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Comment.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &commentV1.ListCommentResponse{Total: 0, Items: nil}, nil
	}

	if treeTravel {
		ret.Items = pagination.BuildTree(
			ret.Items,
			func(node *commentV1.Comment) *uint32 { return node.Id },
			func(node *commentV1.Comment) *uint32 { return node.ParentId },
			func(node *commentV1.Comment) *[]*commentV1.Comment { return &node.Children },
		)
	}

	return &commentV1.ListCommentResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *CommentRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Comment.Query().
		Where(comment.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, commentV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *CommentRepo) Get(ctx context.Context, req *commentV1.GetCommentRequest) (*commentV1.Comment, error) {
	if req == nil {
		return nil, commentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Comment.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *commentV1.GetCommentRequest_Id:
		whereCond = append(whereCond, comment.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *CommentRepo) Create(ctx context.Context, req *commentV1.CreateCommentRequest) error {
	if req == nil || req.Data == nil {
		return commentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Comment.Create().
		SetNillableContentType(r.contentTypeConverter.ToEntity(req.Data.ContentType)).
		SetNillableObjectID(req.Data.ObjectId).
		SetNillableContent(req.Data.Content).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableRating(ratingPtrToUint8(req.Data.Rating)).
		SetNillableParentID(req.Data.ParentId).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert comment failed: %s", err.Error())
		return commentV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *CommentRepo) Update(ctx context.Context, req *commentV1.UpdateCommentRequest) error {
	if req == nil || req.Data == nil {
		return commentV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return commentV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &commentV1.CreateCommentRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Comment.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *commentV1.Comment) {
			builder.
				SetNillableContentType(r.contentTypeConverter.ToEntity(req.Data.ContentType)).
				SetNillableObjectID(req.Data.ObjectId).
				SetNillableContent(req.Data.Content).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableParentID(req.Data.ParentId).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(comment.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *CommentRepo) Delete(ctx context.Context, req *commentV1.DeleteCommentRequest) error {
	if req == nil {
		return commentV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Comment.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(comment.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete comment failed: %s", err.Error())
		return commentV1.ErrorInternalServerError("delete comment failed")
	}

	return nil
}

// GetProductRating 计算指定商品的评分聚合。
// 仅统计 content_type=PRODUCT AND object_id=? AND status=APPROVED 的评论。
// - Mean(rating)：SQL AVG，自动忽略 NULL（无评分的评论不计入平均，但仍计入 count）。
// - Count()：SQL COUNT(*)，统计所有匹配行（含无评分）。
// 返回 ProductRatingSummary{Average, Count}；无数据时 Average=0, Count=0。
func (r *CommentRepo) GetProductRating(ctx context.Context, productID uint32) (*commentV1.ProductRatingSummary, error) {
	var rows []struct {
		Avg   *float64
		Count *int64
	}
	err := r.entClient.Client().Comment.Query().
		Where(
			comment.ContentTypeEQ(comment.ContentTypeCommentContentTypeProduct),
			comment.ObjectIDEQ(productID),
			comment.StatusEQ(comment.StatusCommentStatusApproved),
		).
		Aggregate(
			ent.As(ent.Mean(comment.FieldRating), "avg"),
			ent.As(ent.Count(), "count"),
		).
		Scan(ctx, &rows)
	if err != nil {
		r.log.Errorf("query product rating failed: %s", err.Error())
		return nil, commentV1.ErrorInternalServerError("query product rating failed")
	}

	summary := &commentV1.ProductRatingSummary{
		Average: 0,
		Count:   0,
	}
	if len(rows) > 0 {
		if rows[0].Avg != nil {
			summary.Average = *rows[0].Avg
		}
		if rows[0].Count != nil {
			summary.Count = uint64(*rows[0].Count)
		}
	}
	// 评分数为 0 时强制平均分 0（避免 Mean 返回 NULL 被解为 0 的歧义）。
	if summary.Count == 0 {
		summary.Average = 0
	}
	return summary, nil
}

// ratingPtrToUint8 将 *uint32（proto optional uint32）转为 *uint8（ent rating 列类型）。
// nil 透传。值范围校验由 BFF 层负责（1-5），此处仅做类型转换。
func ratingPtrToUint8(p *uint32) *uint8 {
	if p == nil {
		return nil
	}
	v := uint8(*p)
	return &v
}
