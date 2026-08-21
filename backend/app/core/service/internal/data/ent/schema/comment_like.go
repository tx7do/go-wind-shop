package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"

	appPrivacy "go-wind-shop/pkg/entgo/privacy"
)

// CommentLike 评论点赞 ledger 表，记录 (tenant, user, comment) 三元组的点赞关系。
// 计数本身不存于此，仅作为 InteractionService 递增 interaction_counter 的唯一写入依据。
// 复合 unique 索引天然防重复点赞，并兼作 (user, comment) 查询索引。
//
// Policy 注入 UserPrivacy：普通用户只能查询/变更 user_id = 自身 userID 的点赞记录。
// 系统/平台视图放行，以便 admin 清数据/重算计数。防同租户内越权看/改他人点赞。
type CommentLike struct {
	ent.Schema
}

func (CommentLike) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (CommentLike) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_comment_likes",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("评论点赞 ledger 表"),
	}
}

func (CommentLike) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("点赞用户ID").
			Optional().
			Nillable(),
		field.Uint32("comment_id").
			Comment("被点赞评论ID").
			Optional().
			Nillable(),
	}
}

func (CommentLike) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (CommentLike) Indexes() []ent.Index {
	return []ent.Index{
		// 复合 unique 索引：天然防重复点赞，兼作 (user, comment) 查询索引
		index.Fields("tenant_id", "user_id", "comment_id").
			Unique(),
	}
}
