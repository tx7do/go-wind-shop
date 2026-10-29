package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/membertier"
	"go-wind-shop/app/core/service/internal/data/ent/order"

	"google.golang.org/protobuf/types/known/emptypb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type MemberTierService struct {
	catalogV1.UnimplementedMemberTierServiceServer

	log *logx.Helper

	repo      *data.MemberTierRepo
	entClient *entCrud.EntClient[*ent.Client]
}

func NewMemberTierService(ctx *bootstrap.Context, repo *data.MemberTierRepo, entClient *entCrud.EntClient[*ent.Client]) *MemberTierService {
	return &MemberTierService{
		log:       logx.New(ctx.NewLoggerHelper("member-tier/service/core-service")),
		repo:      repo,
		entClient: entClient,
	}
}

func (s *MemberTierService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListMemberTierResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *MemberTierService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountMemberTierResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}
	return &catalogV1.CountMemberTierResponse{Count: uint64(count)}, nil
}

func (s *MemberTierService) Get(ctx context.Context, req *catalogV1.GetMemberTierRequest) (*catalogV1.MemberTier, error) {
	return s.repo.Get(ctx, req)
}

// ListTiers 全部 ACTIVE 等级列表（level 升序；C端展示各等级权益）。
func (s *MemberTierService) ListTiers(ctx context.Context, req *catalogV1.ListActiveMemberTiersRequest) (*catalogV1.ListMemberTierResponse, error) {
	list, err := s.entClient.Client().MemberTier.Query().
		Where(membertier.StatusEQ(membertier.StatusMemberTierStatusActive)).
		Order(ent.Asc(membertier.FieldLevel)).
		All(ctx)
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("query failed")
	}
	items := make([]*catalogV1.MemberTier, 0, len(list))
	for _, it := range list {
		items = append(items, s.toDTO(it))
	}
	return &catalogV1.ListMemberTierResponse{Items: items, Total: uint64(len(items))}, nil
}

func (s *MemberTierService) Create(ctx context.Context, req *catalogV1.CreateMemberTierRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}
	if req.Data.GetName() == "" {
		return nil, catalogV1.ErrorBadRequest("name is required")
	}
	if req.Data.GetMinSpend() < 0 {
		return nil, catalogV1.ErrorBadRequest("min_spend must be non-negative")
	}
	if req.Data.GetPointsMultiplier() < 1 {
		data := req.Data
		data.PointsMultiplier = trans.Ptr(int32(1))
		_ = data
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *MemberTierService) Update(ctx context.Context, req *catalogV1.UpdateMemberTierRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *MemberTierService) ListActiveMemberTiers(ctx context.Context, req *catalogV1.ListActiveMemberTiersRequest) (*catalogV1.ListMemberTierResponse, error) {
	return s.ListTiers(ctx, req)
}

func (s *MemberTierService) Delete(ctx context.Context, req *catalogV1.DeleteMemberTierRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// cumulativeSpend 查用户累计已支付金额（PAID/FULFILLED/CLOSED）。
func (s *MemberTierService) cumulativeSpend(ctx context.Context, userId uint32) (int64, error) {
	var rows []struct{ Sum int64 }
	err := s.entClient.Client().Order.Query().
		Where(
			order.UserIDEQ(userId),
			order.StatusIn(
				order.StatusOrderStatusPaid,
				order.StatusOrderStatusFulfilled,
				order.StatusOrderStatusClosed,
			),
		).
		Aggregate(ent.As(ent.Sum(order.FieldTotalAmount), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Sum, nil
}

// GetUserTierInfo 用户等级信息：累计消费 + 匹配的最高等级（level 降序，首个 min_spend <= spend）。
func (s *MemberTierService) GetUserTierInfo(ctx context.Context, req *catalogV1.GetUserTierInfoRequest) (*catalogV1.GetUserTierInfoResponse, error) {
	userId := req.GetUserId()
	if userId == 0 {
		return nil, catalogV1.ErrorBadRequest("user_id is required")
	}

	spend, err := s.cumulativeSpend(ctx, userId)
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("cumulative spend query failed")
	}

	// ACTIVE 等级按 level 降序，找第一个 min_spend <= spend 的
	tiers, tErr := s.entClient.Client().MemberTier.Query().
		Where(membertier.StatusEQ(membertier.StatusMemberTierStatusActive)).
		Order(ent.Desc(membertier.FieldLevel)).
		All(ctx)
	if tErr != nil {
		return nil, catalogV1.ErrorInternalServerError("query tiers failed")
	}

	var matched *ent.MemberTier
	for _, t := range tiers {
		if t.MinSpend <= spend {
			matched = t
			break
		}
	}

	resp := &catalogV1.GetUserTierInfoResponse{CumulativeSpend: spend}
	if matched != nil {
		resp.Tier = &catalogV1.MemberTier{
			Id:               trans.Ptr(matched.ID),
			Name:             trans.Ptr(matched.Name),
			Level:            trans.Ptr(matched.Level),
			MinSpend:         trans.Ptr(matched.MinSpend),
			PointsMultiplier: trans.Ptr(matched.PointsMultiplier),
			BadgeIcon:        matched.BadgeIcon,
			Status:           catalogV1.MemberTier_ACTIVE.Enum(),
		}
	}
	return resp, nil
}

// GetPointsMultiplierForUser 查用户当前等级的积分倍率（积分入账用）。
// 无等级配置或无匹配时返回 1。仅供 AwardForOrder 调用。
func (s *MemberTierService) GetPointsMultiplierForUser(ctx context.Context, userId uint32) int32 {
	spend, err := s.cumulativeSpend(ctx, userId)
	if err != nil {
		return 1
	}
	tiers, tErr := s.entClient.Client().MemberTier.Query().
		Where(membertier.StatusEQ(membertier.StatusMemberTierStatusActive)).
		Order(ent.Desc(membertier.FieldLevel)).
		All(ctx)
	if tErr != nil {
		return 1
	}
	for _, t := range tiers {
		if t.MinSpend <= spend && t.PointsMultiplier > 0 {
			return t.PointsMultiplier
		}
	}
	return 1
}

func (s *MemberTierService) toDTO(it *ent.MemberTier) *catalogV1.MemberTier {
	dto := &catalogV1.MemberTier{
		Id:               trans.Ptr(it.ID),
		Name:             trans.Ptr(it.Name),
		Level:            trans.Ptr(it.Level),
		MinSpend:         trans.Ptr(it.MinSpend),
		PointsMultiplier: trans.Ptr(it.PointsMultiplier),
		BadgeIcon:        it.BadgeIcon,
		Currency:         trans.Ptr(it.Currency),
		TenantId:         it.TenantID,
	}
	if it.Status != nil {
		if v, ok := catalogV1.MemberTier_Status_value[string(*it.Status)]; ok {
			dto.Status = catalogV1.MemberTier_Status(v).Enum()
		}
	}
	if it.CreatedAt != nil {
		dto.CreatedAt = tspb.New(*it.CreatedAt)
	}
	return dto
}


