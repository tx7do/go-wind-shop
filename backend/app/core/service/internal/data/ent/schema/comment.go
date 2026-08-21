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

// Comment 商品评论表。
//
// Policy 注入 UserPrivacy：普通用户只能查询/变更 created_by = 自身 userID 的评论；
// 系统/平台视图放行，以便 admin 审核全量评论。与 user_coupon/order 同模式。
// content_type 仅 PRODUCT（shop 无文章/页面），object_id 为商品 ID。
// 树形回复由 mixin.Tree 提供 parent_id/children，参照 org_unit。
type Comment struct {
	ent.Schema
}

func (Comment) Policy() ent.Policy {
	return appPrivacy.UserPrivacy{}
}

func (Comment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "mall_comments",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("商品评论表"),
	}
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("content_type").
			Comment("内容类型（仅产品）").
			NamedValues(
				"CommentContentTypeProduct", "CONTENT_TYPE_PRODUCT",
			).
			Optional().
			Nillable(),

		field.Uint32("object_id").
			Comment("对象ID（商品ID）").
			Optional().
			Nillable(),

		field.String("content").
			Comment("评论内容").
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("评论状态").
			NamedValues(
				"CommentStatusPending", "STATUS_PENDING",
				"CommentStatusApproved", "STATUS_APPROVED",
				"CommentStatusRejected", "STATUS_REJECTED",
				"CommentStatusSpam", "STATUS_SPAM",
			).
			Optional().
			Nillable(),
	}
}

func (Comment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		mixin.Tree[Comment]{},
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		// 复合索引，优化按内容类型和对象ID查询评论
		index.Fields("content_type", "object_id"),
		// 单字段索引，用于按评论状态查询
		index.Fields("status"),
		// 单字段索引，用于树形结构中按父节点查询子评论
		index.Fields("parent_id"),
	}
}
