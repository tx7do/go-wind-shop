package data

import (
	"go-wind-shop/pkg/logx"
	"time"

	"github.com/redis/go-redis/v9"

	authzEngine "github.com/tx7do/kratos-authz/engine"
	"github.com/tx7do/kratos-authz/engine/noop"

	"github.com/go-kratos/kratos/v2/registry"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	redisx "go-wind-shop/pkg/redisx"
	bRegistry "github.com/tx7do/kratos-bootstrap/registry"
	"github.com/tx7do/kratos-bootstrap/rpc"

	"github.com/tx7do/go-utils/captcha"

	auditV1 "go-wind-shop/api/gen/go/audit/service/v1"
	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
	cartV1 "go-wind-shop/api/gen/go/cart/service/v1"
	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"
	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
	dictV1 "go-wind-shop/api/gen/go/dict/service/v1"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	interactionV1 "go-wind-shop/api/gen/go/interaction/service/v1"
	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
	notificationV1 "go-wind-shop/api/gen/go/notification/service/v1"
	aiV1 "go-wind-shop/api/gen/go/ai/service/v1"
	scriptV1 "go-wind-shop/api/gen/go/script/service/v1"
	configV1 "go-wind-shop/api/gen/go/config/service/v1"
	accessKeyV1 "go-wind-shop/api/gen/go/access_key/service/v1"
	onlineSessionV1 "go-wind-shop/api/gen/go/online_session/service/v1"
	monitorAlertV1 "go-wind-shop/api/gen/go/monitor_alert/service/v1"
	serverMonitorV1 "go-wind-shop/api/gen/go/server_monitor/service/v1"
	redisCacheV1 "go-wind-shop/api/gen/go/redis_cache/service/v1"
	invoiceV1 "go-wind-shop/api/gen/go/invoice/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
	permissionV1 "go-wind-shop/api/gen/go/permission/service/v1"
	storageV1 "go-wind-shop/api/gen/go/storage/service/v1"
	taskV1 "go-wind-shop/api/gen/go/task/service/v1"
	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
	taxV1 "go-wind-shop/api/gen/go/tax/service/v1"

	"go-wind-shop/pkg/oss"
	"go-wind-shop/pkg/serviceid"
)

func NewClientType() authenticationV1.ClientType {
	return authenticationV1.ClientType_admin
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(ctx *bootstrap.Context) (*redis.Client, func(), error) {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil, func() {}, nil
	}

	l := logx.New(ctx.NewLoggerHelper("redis/data/admin-service"))

	cli := redisx.NewClient(cfg.Data, l)

	return cli, func() {
		if err := cli.Close(); err != nil {
			l.Error(err)
		}
	}, nil
}

func NewCaptcha(rdb *redis.Client) *captcha.Captcha {
	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverString),
		captcha.WithExpire(10*time.Minute),
		captcha.WithKeyPrefix(serviceid.ProjectName+":captcha"),
		captcha.WithStringCount(6),
		captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
	)
	return captchaInstance
}

// NewDiscovery 创建服务发现客户端
func NewDiscovery(ctx *bootstrap.Context) registry.Discovery {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil
	}

	discovery, err := bRegistry.NewDiscovery(cfg.Registry)
	if err != nil {
		return nil
	}

	NewDtmDriver(discovery)

	return discovery
}

func NewMinIoClient(ctx *bootstrap.Context) *oss.MinIOClient {
	return oss.NewMinIoClient(ctx.GetConfig(), ctx.GetLogger())
}

// NewAuthorizer 创建权鉴器
func NewAuthorizer() authzEngine.Engine {
	return noop.State{}
}

func NewAuthenticationServiceClient(ctx *bootstrap.Context, r registry.Discovery) authenticationV1.AuthenticationServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return authenticationV1.NewAuthenticationServiceClient(cli)
}

func NewUserCredentialServiceClient(ctx *bootstrap.Context, r registry.Discovery) authenticationV1.UserCredentialServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return authenticationV1.NewUserCredentialServiceClient(cli)
}

func NewLoginPolicyServiceClient(ctx *bootstrap.Context, r registry.Discovery) authenticationV1.LoginPolicyServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return authenticationV1.NewLoginPolicyServiceClient(cli)
}

func NewUserServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.UserServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewUserServiceClient(cli)
}

func NewTenantServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.TenantServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewTenantServiceClient(cli)
}

func NewRoleServiceClient(ctx *bootstrap.Context, r registry.Discovery) permissionV1.RoleServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return permissionV1.NewRoleServiceClient(cli)
}

func NewOrgUnitServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.OrgUnitServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewOrgUnitServiceClient(cli)
}

func NewPositionServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.PositionServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewPositionServiceClient(cli)
}

func NewInternalMessageCategoryServiceClient(ctx *bootstrap.Context, r registry.Discovery) internalMessageV1.InternalMessageCategoryServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return internalMessageV1.NewInternalMessageCategoryServiceClient(cli)
}

func NewInternalMessageServiceClient(ctx *bootstrap.Context, r registry.Discovery) internalMessageV1.InternalMessageServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return internalMessageV1.NewInternalMessageServiceClient(cli)
}

func NewInternalMessageRecipientServiceClient(ctx *bootstrap.Context, r registry.Discovery) internalMessageV1.InternalMessageRecipientServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return internalMessageV1.NewInternalMessageRecipientServiceClient(cli)
}

func NewOssServiceClient(ctx *bootstrap.Context, r registry.Discovery) storageV1.OssServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return storageV1.NewOssServiceClient(cli)
}

func NewFileServiceClient(ctx *bootstrap.Context, r registry.Discovery) storageV1.FileServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return storageV1.NewFileServiceClient(cli)
}

func NewPermissionGroupServiceClient(ctx *bootstrap.Context, r registry.Discovery) permissionV1.PermissionGroupServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return permissionV1.NewPermissionGroupServiceClient(cli)
}

func NewPermissionServiceClient(ctx *bootstrap.Context, r registry.Discovery) permissionV1.PermissionServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return permissionV1.NewPermissionServiceClient(cli)
}

func NewApiServiceClient(ctx *bootstrap.Context, r registry.Discovery) permissionV1.ApiServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return permissionV1.NewApiServiceClient(cli)
}

func NewMenuServiceClient(ctx *bootstrap.Context, r registry.Discovery) permissionV1.MenuServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return permissionV1.NewMenuServiceClient(cli)
}

func NewPermissionAuditLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) auditV1.PermissionAuditLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return auditV1.NewPermissionAuditLogServiceClient(cli)
}

func NewPolicyEvaluationLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) permissionV1.PolicyEvaluationLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return permissionV1.NewPolicyEvaluationLogServiceClient(cli)
}

func NewApiAuditLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) auditV1.ApiAuditLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return auditV1.NewApiAuditLogServiceClient(cli)
}

func NewDataAccessAuditLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) auditV1.DataAccessAuditLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return auditV1.NewDataAccessAuditLogServiceClient(cli)
}

func NewLoginAuditLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) auditV1.LoginAuditLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return auditV1.NewLoginAuditLogServiceClient(cli)
}

func NewOperationAuditLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) auditV1.OperationAuditLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return auditV1.NewOperationAuditLogServiceClient(cli)
}

func NewLanguageServiceClient(ctx *bootstrap.Context, r registry.Discovery) dictV1.LanguageServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return dictV1.NewLanguageServiceClient(cli)
}

func NewTaskServiceClient(ctx *bootstrap.Context, r registry.Discovery) taskV1.TaskServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return taskV1.NewTaskServiceClient(cli)
}

func NewCategoryServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.CategoryServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewCategoryServiceClient(cli)
}

func NewBrandServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.BrandServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewBrandServiceClient(cli)
}

func NewProductServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.ProductServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewProductServiceClient(cli)
}

func NewProductAttributeServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.ProductAttributeServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewProductAttributeServiceClient(cli)
}

func NewProductAttributeValueServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.ProductAttributeValueServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewProductAttributeValueServiceClient(cli)
}

func NewSkuServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.SkuServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewSkuServiceClient(cli)
}

func NewSkuPriceServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.SkuPriceServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewSkuPriceServiceClient(cli)
}

func NewSkuAttributeCombinationServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.SkuAttributeCombinationServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewSkuAttributeCombinationServiceClient(cli)
}

func NewCartServiceClient(ctx *bootstrap.Context, r registry.Discovery) cartV1.CartServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return cartV1.NewCartServiceClient(cli)
}

func NewCartItemServiceClient(ctx *bootstrap.Context, r registry.Discovery) cartV1.CartItemServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return cartV1.NewCartItemServiceClient(cli)
}

func NewOrderServiceClient(ctx *bootstrap.Context, r registry.Discovery) orderV1.OrderServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return orderV1.NewOrderServiceClient(cli)
}

func NewOrderItemServiceClient(ctx *bootstrap.Context, r registry.Discovery) orderV1.OrderItemServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return orderV1.NewOrderItemServiceClient(cli)
}

func NewPaymentTransactionServiceClient(ctx *bootstrap.Context, r registry.Discovery) paymentV1.PaymentTransactionServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return paymentV1.NewPaymentTransactionServiceClient(cli)
}

func NewPaymentRefundServiceClient(ctx *bootstrap.Context, r registry.Discovery) paymentV1.PaymentRefundServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return paymentV1.NewPaymentRefundServiceClient(cli)
}

func NewShipmentServiceClient(ctx *bootstrap.Context, r registry.Discovery) shippingV1.ShipmentServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return shippingV1.NewShipmentServiceClient(cli)
}

func NewCouponTemplateServiceClient(ctx *bootstrap.Context, r registry.Discovery) couponV1.CouponTemplateServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return couponV1.NewCouponTemplateServiceClient(cli)
}

func NewPromotionServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.PromotionServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewPromotionServiceClient(cli)
}

func NewStockOpLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.StockOpLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewStockOpLogServiceClient(cli)
}

func NewUserCouponServiceClient(ctx *bootstrap.Context, r registry.Discovery) couponV1.UserCouponServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return couponV1.NewUserCouponServiceClient(cli)
}

func NewShippingRateServiceClient(ctx *bootstrap.Context, r registry.Discovery) shippingV1.ShippingRateServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return shippingV1.NewShippingRateServiceClient(cli)
}

func NewStockAlertServiceClient(ctx *bootstrap.Context, r registry.Discovery) catalogV1.StockAlertServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return catalogV1.NewStockAlertServiceClient(cli)
}

func NewTaxRateServiceClient(ctx *bootstrap.Context, r registry.Discovery) taxV1.TaxRateServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return taxV1.NewTaxRateServiceClient(cli)
}

func NewCommentServiceClient(ctx *bootstrap.Context, r registry.Discovery) commentV1.CommentServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return commentV1.NewCommentServiceClient(cli)
}

func NewInteractionAdminServiceClient(ctx *bootstrap.Context, r registry.Discovery) interactionV1.InteractionAdminServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return interactionV1.NewInteractionAdminServiceClient(cli)
}

func NewInvoiceServiceClient(ctx *bootstrap.Context, r registry.Discovery) invoiceV1.InvoiceServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return invoiceV1.NewInvoiceServiceClient(cli)
}

func NewNotificationServiceClient(ctx *bootstrap.Context, r registry.Discovery) notificationV1.NotificationServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return notificationV1.NewNotificationServiceClient(cli)
}

func NewNotificationChannelServiceClient(ctx *bootstrap.Context, r registry.Discovery) notificationV1.NotificationChannelServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return notificationV1.NewNotificationChannelServiceClient(cli)
}

func NewNotificationRuleServiceClient(ctx *bootstrap.Context, r registry.Discovery) notificationV1.NotificationRuleServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return notificationV1.NewNotificationRuleServiceClient(cli)
}

func NewNotificationTemplateServiceClient(ctx *bootstrap.Context, r registry.Discovery) notificationV1.NotificationTemplateServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return notificationV1.NewNotificationTemplateServiceClient(cli)
}

func NewNotificationPreferenceServiceClient(ctx *bootstrap.Context, r registry.Discovery) notificationV1.NotificationPreferenceServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return notificationV1.NewNotificationPreferenceServiceClient(cli)
}

func NewDictTypeServiceClient(ctx *bootstrap.Context, r registry.Discovery) dictV1.DictTypeServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return dictV1.NewDictTypeServiceClient(cli)
}

func NewDictEntryServiceClient(ctx *bootstrap.Context, r registry.Discovery) dictV1.DictEntryServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return dictV1.NewDictEntryServiceClient(cli)
}

func NewConfigServiceClient(ctx *bootstrap.Context, r registry.Discovery) configV1.ConfigServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return configV1.NewConfigServiceClient(cli)
}

func NewAccessKeyServiceClient(ctx *bootstrap.Context, r registry.Discovery) accessKeyV1.AccessKeyServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return accessKeyV1.NewAccessKeyServiceClient(cli)
}

func NewOnlineSessionServiceClient(ctx *bootstrap.Context, r registry.Discovery) onlineSessionV1.OnlineSessionServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return onlineSessionV1.NewOnlineSessionServiceClient(cli)
}

func NewMonitorAlertServiceClient(ctx *bootstrap.Context, r registry.Discovery) monitorAlertV1.MonitorAlertServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return monitorAlertV1.NewMonitorAlertServiceClient(cli)
}

func NewServerMonitorServiceClient(ctx *bootstrap.Context, r registry.Discovery) serverMonitorV1.ServerMonitorServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return serverMonitorV1.NewServerMonitorServiceClient(cli)
}

func NewRedisCacheMonitorServiceClient(ctx *bootstrap.Context, r registry.Discovery) redisCacheV1.RedisCacheMonitorServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return redisCacheV1.NewRedisCacheMonitorServiceClient(cli)
}

func NewAiAgentServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiAgentServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiAgentServiceClient(cli)
}

func NewAiChatServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiChatServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiChatServiceClient(cli)
}

func NewAiConversationServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiConversationServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiConversationServiceClient(cli)
}

func NewAiQueryServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiQueryServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiQueryServiceClient(cli)
}

func NewAiUsageLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiUsageLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiUsageLogServiceClient(cli)
}

func NewAiKnowledgeBaseServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiKnowledgeBaseServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiKnowledgeBaseServiceClient(cli)
}

func NewAiMessageServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiMessageServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiMessageServiceClient(cli)
}

func NewAiProviderServiceClient(ctx *bootstrap.Context, r registry.Discovery) aiV1.AiProviderServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return aiV1.NewAiProviderServiceClient(cli)
}

func NewScriptServiceClient(ctx *bootstrap.Context, r registry.Discovery) scriptV1.ScriptServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}
	return scriptV1.NewScriptServiceClient(cli)
}

func NewScriptLogServiceClient(ctx *bootstrap.Context, r registry.Discovery) scriptV1.ScriptLogServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}
	return scriptV1.NewScriptLogServiceClient(cli)
}

func NewPlanServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.PlanServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewPlanServiceClient(cli)
}

func NewPlanModuleServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.PlanModuleServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewPlanModuleServiceClient(cli)
}

func NewPlanQuotaServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.PlanQuotaServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return identityV1.NewPlanQuotaServiceClient(cli)
}

func NewMFAServiceClient(ctx *bootstrap.Context, r registry.Discovery) authenticationV1.MFAServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return authenticationV1.NewMFAServiceClient(cli)
}
