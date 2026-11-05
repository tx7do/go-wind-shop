package service

import (
	"context"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"

	"go-wind-shop/pkg/logx"
	"go-wind-shop/pkg/middleware/auth"
)

// requirePlatformAdmin 把平台专用配置与台账挡在平台管理员之内。
// 租户/套餐/菜单/API/权限等模块常态在租户套餐白名单内，叠加 authz.type
// 默认 noop 全放行，租户操作者即可打到平台专用变更端点。判定显式放这里
// （移植自上游 da2b3a5f）。
func requirePlatformAdmin(ctx context.Context, log *logx.Helper, resource string) error {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}

	if operator.GetIsPlatformAdmin() {
		return nil
	}

	log.Errorf("operator [%d] (tenant %d) is not a platform administrator, %s access denied",
		operator.GetUserId(), operator.GetTenantId(), resource)

	return adminV1.ErrorForbidden("%s is platform-level configuration, only platform administrators may access it", resource)
}
