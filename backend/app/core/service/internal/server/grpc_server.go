package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-bootstrap/rpc"

	"go-wind-shop/app/core/service/internal/service"

	auditV1 "go-wind-shop/api/gen/go/audit/service/v1"
	addressV1 "go-wind-shop/api/gen/go/address/service/v1"
	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"
	cartV1 "go-wind-shop/api/gen/go/cart/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"
	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
	dictV1 "go-wind-shop/api/gen/go/dict/service/v1"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	interactionV1 "go-wind-shop/api/gen/go/interaction/service/v1"
	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
	invoiceV1 "go-wind-shop/api/gen/go/invoice/service/v1"
	notificationV1 "go-wind-shop/api/gen/go/notification/service/v1"
	aiV1 "go-wind-shop/api/gen/go/ai/service/v1"
	scriptV1 "go-wind-shop/api/gen/go/script/service/v1"
	configV1 "go-wind-shop/api/gen/go/config/service/v1"
	monitorAlertV1 "go-wind-shop/api/gen/go/monitor_alert/service/v1"
	onlineSessionV1 "go-wind-shop/api/gen/go/online_session/service/v1"
	accessKeyV1 "go-wind-shop/api/gen/go/access_key/service/v1"
	serverMonitorV1 "go-wind-shop/api/gen/go/server_monitor/service/v1"
	redisCacheV1 "go-wind-shop/api/gen/go/redis_cache/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
	permissionV1 "go-wind-shop/api/gen/go/permission/service/v1"
	storageV1 "go-wind-shop/api/gen/go/storage/service/v1"
	taskV1 "go-wind-shop/api/gen/go/task/service/v1"
	wishlistV1 "go-wind-shop/api/gen/go/wishlist/service/v1"

	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
	taxV1 "go-wind-shop/api/gen/go/tax/service/v1"

	"go-wind-shop/pkg/logx"
	"go-wind-shop/pkg/middleware/ent"
)

func NewGrpcMiddleware(ctx *bootstrap.Context) []middleware.Middleware {
	var ms []middleware.Middleware
	ms = append(ms, logging.Server(logx.AsKratos(ctx.GetLogger())))
	ms = append(ms, ent.Server())
	return ms
}

// NewGrpcServer new a gRPC server.
func NewGrpcServer(
	ctx *bootstrap.Context,
	middlewares []middleware.Middleware,

	authenticationService *service.AuthenticationService,
	loginPolicyService *service.LoginPolicyService,
	notificationService *service.NotificationService,
	notificationChannelService *service.NotificationChannelService,
	notificationRuleService *service.NotificationRuleService,
	notificationTemplateService *service.NotificationTemplateService,
	notificationPreferenceService *service.NotificationPreferenceService,
	configService *service.ConfigService,
	dictTypeService *service.DictTypeService,
	dictEntryService *service.DictEntryService,
	monitorAlertService *service.MonitorAlertService,
	onlineSessionService *service.OnlineSessionService,
	accessKeyService *service.AccessKeyService,
	serverMonitorService *service.ServerMonitorService,
	redisCacheMonitorService *service.RedisCacheMonitorService,
	aiAgentService *service.AiAgentService,
	aiProviderService *service.AiProviderService,
	aiChatService *service.AiChatService,
	aiConversationService *service.AiConversationService,
	aiMessageService *service.AiMessageService,
	aiKnowledgeService *service.AiKnowledgeService,
	aiUsageLogService *service.AiUsageLogService,
	aiQueryService *service.AiQueryService,
	scriptService *service.ScriptService,
	scriptLogService *service.ScriptLogService,
	planService *service.PlanService,
	planModuleService *service.PlanModuleService,
	planQuotaService *service.PlanQuotaService,
	mfaService *service.MfaService,
	userCredentialService *service.UserCredentialService,

	taskService *service.TaskService,

	fileService *service.FileService,

	languageService *service.LanguageService,

	tenantService *service.TenantService,
	userService *service.UserService,
	roleService *service.RoleService,
	positionService *service.PositionService,
	orgUnitService *service.OrgUnitService,

	menuService *service.MenuService,
	apiService *service.ApiService,
	permissionService *service.PermissionService,
	permissionGroupService *service.PermissionGroupService,
	permissionAuditLogService *service.PermissionAuditLogService,
	policyEvaluationLogService *service.PolicyEvaluationLogService,

	loginAuditLogService *service.LoginAuditLogService,
	apiAuditLogService *service.ApiAuditLogService,
	operationAuditLogService *service.OperationAuditLogService,
	dataAccessAuditLogService *service.DataAccessAuditLogService,

	internalMessageService *service.InternalMessageService,
	internalMessageCategoryService *service.InternalMessageCategoryService,
	internalMessageRecipientService *service.InternalMessageRecipientService,

	categoryService *service.CategoryService,
	brandService *service.BrandService,
	productService *service.ProductService,
	productAttributeService *service.ProductAttributeService,
	productAttributeValueService *service.ProductAttributeValueService,
	skuService *service.SkuService,
	skuPriceService *service.SkuPriceService,
	skuAttributeCombinationService *service.SkuAttributeCombinationService,
	stockAlertService *service.StockAlertService,

		cartService *service.CartService,
		cartItemService *service.CartItemService,
		wishlistService *service.WishlistService,
		orderService *service.OrderService,
	orderItemService *service.OrderItemService,

	// register:grpc-param ── 新模块服务形参在此行后追加(make register 工具锚点,勿删)
		paymentTransactionService *service.PaymentTransactionService,
		paymentRefundService *service.PaymentRefundService,
		shippingAddressService *service.ShippingAddressService,
		shipmentService *service.ShipmentService,

		couponTemplateService *service.CouponTemplateService,
		userCouponService *service.UserCouponService,

		shippingRateService *service.ShippingRateService,
		taxRateService *service.TaxRateService,
		commentService *service.CommentService,

		interactionService *service.InteractionService,
		interactionAdminService *service.InteractionAdminService,
		invoiceService *service.InvoiceService,
	) (*grpc.Server, error) {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Grpc == nil {
		return nil, nil
	}

	srv, err := rpc.CreateGrpcServer(cfg, middlewares...)
	if err != nil {
		return nil, err
	}

	taskV1.RegisterTaskServiceServer(srv, taskService)

	authenticationV1.RegisterLoginPolicyServiceServer(srv, loginPolicyService)
	authenticationV1.RegisterAuthenticationServiceServer(srv, authenticationService)
	authenticationV1.RegisterUserCredentialServiceServer(srv, userCredentialService)

	dictV1.RegisterLanguageServiceServer(srv, languageService)

	permissionV1.RegisterApiServiceServer(srv, apiService)
	permissionV1.RegisterMenuServiceServer(srv, menuService)

	permissionV1.RegisterPermissionServiceServer(srv, permissionService)
	permissionV1.RegisterPermissionGroupServiceServer(srv, permissionGroupService)
	permissionV1.RegisterPolicyEvaluationLogServiceServer(srv, policyEvaluationLogService)
	permissionV1.RegisterRoleServiceServer(srv, roleService)

	identityV1.RegisterUserServiceServer(srv, userService)
	identityV1.RegisterOrgUnitServiceServer(srv, orgUnitService)
	identityV1.RegisterPositionServiceServer(srv, positionService)
	identityV1.RegisterTenantServiceServer(srv, tenantService)

	auditV1.RegisterLoginAuditLogServiceServer(srv, loginAuditLogService)
	auditV1.RegisterApiAuditLogServiceServer(srv, apiAuditLogService)
	auditV1.RegisterOperationAuditLogServiceServer(srv, operationAuditLogService)
	auditV1.RegisterDataAccessAuditLogServiceServer(srv, dataAccessAuditLogService)
	auditV1.RegisterPermissionAuditLogServiceServer(srv, permissionAuditLogService)

	storageV1.RegisterFileServiceServer(srv, fileService)

	internalMessageV1.RegisterInternalMessageServiceServer(srv, internalMessageService)
	internalMessageV1.RegisterInternalMessageCategoryServiceServer(srv, internalMessageCategoryService)
	internalMessageV1.RegisterInternalMessageRecipientServiceServer(srv, internalMessageRecipientService)

	catalogV1.RegisterCategoryServiceServer(srv, categoryService)
	catalogV1.RegisterBrandServiceServer(srv, brandService)
	catalogV1.RegisterProductServiceServer(srv, productService)
	catalogV1.RegisterProductAttributeServiceServer(srv, productAttributeService)
	catalogV1.RegisterProductAttributeValueServiceServer(srv, productAttributeValueService)
	catalogV1.RegisterSkuServiceServer(srv, skuService)
	catalogV1.RegisterSkuPriceServiceServer(srv, skuPriceService)
	catalogV1.RegisterSkuAttributeCombinationServiceServer(srv, skuAttributeCombinationService)
	catalogV1.RegisterStockAlertServiceServer(srv, stockAlertService)

		cartV1.RegisterCartServiceServer(srv, cartService)
		cartV1.RegisterCartItemServiceServer(srv, cartItemService)
		wishlistV1.RegisterWishlistServiceServer(srv, wishlistService)
		orderV1.RegisterOrderServiceServer(srv, orderService)
	orderV1.RegisterOrderItemServiceServer(srv, orderItemService)
	paymentV1.RegisterPaymentTransactionServiceServer(srv, paymentTransactionService)
	paymentV1.RegisterPaymentRefundServiceServer(srv, paymentRefundService)

	addressV1.RegisterShippingAddressServiceServer(srv, shippingAddressService)

	shippingV1.RegisterShipmentServiceServer(srv, shipmentService)

	couponV1.RegisterCouponTemplateServiceServer(srv, couponTemplateService)
	couponV1.RegisterUserCouponServiceServer(srv, userCouponService)

	shippingV1.RegisterShippingRateServiceServer(srv, shippingRateService)
	taxV1.RegisterTaxRateServiceServer(srv, taxRateService)

	commentV1.RegisterCommentServiceServer(srv, commentService)

	interactionV1.RegisterInteractionServiceServer(srv, interactionService)
	interactionV1.RegisterInteractionAdminServiceServer(srv, interactionAdminService)

	invoiceV1.RegisterInvoiceServiceServer(srv, invoiceService)

	notificationV1.RegisterNotificationServiceServer(srv, notificationService)
	notificationV1.RegisterNotificationChannelServiceServer(srv, notificationChannelService)
	notificationV1.RegisterNotificationRuleServiceServer(srv, notificationRuleService)
	notificationV1.RegisterNotificationTemplateServiceServer(srv, notificationTemplateService)
	notificationV1.RegisterNotificationPreferenceServiceServer(srv, notificationPreferenceService)
	configV1.RegisterConfigServiceServer(srv, configService)
	dictV1.RegisterDictTypeServiceServer(srv, dictTypeService)
	dictV1.RegisterDictEntryServiceServer(srv, dictEntryService)
	monitorAlertV1.RegisterMonitorAlertServiceServer(srv, monitorAlertService)
	onlineSessionV1.RegisterOnlineSessionServiceServer(srv, onlineSessionService)
	accessKeyV1.RegisterAccessKeyServiceServer(srv, accessKeyService)
	serverMonitorV1.RegisterServerMonitorServiceServer(srv, serverMonitorService)
	redisCacheV1.RegisterRedisCacheMonitorServiceServer(srv, redisCacheMonitorService)
	aiV1.RegisterAiAgentServiceServer(srv, aiAgentService)
	aiV1.RegisterAiProviderServiceServer(srv, aiProviderService)
	aiV1.RegisterAiChatServiceServer(srv, aiChatService)
	aiV1.RegisterAiConversationServiceServer(srv, aiConversationService)
	aiV1.RegisterAiMessageServiceServer(srv, aiMessageService)
	aiV1.RegisterAiKnowledgeBaseServiceServer(srv, aiKnowledgeService)
	aiV1.RegisterAiUsageLogServiceServer(srv, aiUsageLogService)
	aiV1.RegisterAiQueryServiceServer(srv, aiQueryService)
	scriptV1.RegisterScriptServiceServer(srv, scriptService)
	scriptV1.RegisterScriptLogServiceServer(srv, scriptLogService)
	identityV1.RegisterPlanServiceServer(srv, planService)
	identityV1.RegisterPlanModuleServiceServer(srv, planModuleService)
	identityV1.RegisterPlanQuotaServiceServer(srv, planQuotaService)
	authenticationV1.RegisterMFAServiceServer(srv, mfaService)

	// register:grpc-route ── 新模块服务路由在此行后注册(make register 工具锚点,勿删)

	return srv, nil
}
