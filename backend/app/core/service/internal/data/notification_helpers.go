package data

import (
	"context"
	"time"

	"github.com/tx7do/go-crud/viewer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// viewerUserIDFromContext 从请求上下文的查看者身份中取用户 ID。
// （随 Notification 域从 go-wind-cms 移植，cms 同名实现见 user_credential_repo.go）
func viewerUserIDFromContext(ctx context.Context) (uint32, bool) {
	vc, exist := viewer.FromContext(ctx)
	if !exist || vc == nil {
		return 0, false
	}
	uid := vc.UserID()
	if uid == 0 {
		return 0, false
	}
	return uint32(uid), true
}

// derefString 解引用字符串指针，nil 返回空串。
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// tsOf proto Timestamp → *time.Time（nil 安全）。
func tsOf(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	tm := t.AsTime()
	return &tm
}
