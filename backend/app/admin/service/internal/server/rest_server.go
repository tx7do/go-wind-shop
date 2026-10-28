package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"

	authzEngine "github.com/tx7do/kratos-authz/engine"
	authz "github.com/tx7do/kratos-authz/middleware"

	swaggerUI "github.com/tx7do/kratos-swagger-ui"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-bootstrap/rpc"

	"go-wind-shop/app/admin/service/cmd/server/assets"
	"go-wind-shop/app/admin/service/internal/service"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	auditV1 "go-wind-shop/api/gen/go/audit/service/v1"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	logx "go-wind-shop/pkg/logx"

	"go-wind-shop/pkg/metadata"
	"go-wind-shop/pkg/middleware/auth"
	applogging "go-wind-shop/pkg/middleware/logging"
)

// NewRestMiddleware 创建中间件
func NewRestMiddleware(
	ctx *bootstrap.Context,
	accessTokenChecker auth.AccessTokenChecker,
	authorizer authzEngine.Engine,
	apiAuditLogServiceClient auditV1.ApiAuditLogServiceClient,
	loginAuditLogServiceClient auditV1.LoginAuditLogServiceClient,
	operationAuditLogServiceClient auditV1.OperationAuditLogServiceClient,
) []middleware.Middleware {
	var ms []middleware.Middleware
	ms = append(ms, logging.Server(logx.AsKratos(ctx.GetLogger())))

	// add white list for authentication.
	// RefreshToken 必须在白名单内：access token 过期恰恰是触发刷新的前提场景，
	// 不能再要求有效 AT，否则 AT 一过期续期链路自己先 401（会话永远无法续命）。
	// 该端点的真实鉴权由 core 的 VerifyRefreshToken 完成——RT 是 256 位随机串，
	// 需与 Redis 会话中的值精确比对，身份键（user_id/jti）仅为查找键。
	rpc.AddWhiteList(
		adminV1.OperationAuthenticationServiceLogin,
		adminV1.OperationAuthenticationServiceGenerateCaptcha,
		adminV1.OperationAuthenticationServiceVerifyCaptcha,
		adminV1.OperationAuthenticationServiceRefreshToken,
	)

	ms = append(ms, applogging.Server(
		applogging.WithWriteApiLogFunc(func(ctx context.Context, data *auditV1.ApiAuditLog) error {
			ctx, _ = metadata.NewContext(ctx, metadata.NewUserOperator(0, 0, 0, identityV1.DataScope_ALL))
			// TODO 如果系统的负载比较小，可以同步写入数据库，否则，建议使用异步方式，即投递进队列。
			_, err := apiAuditLogServiceClient.Create(ctx, &auditV1.CreateApiAuditLogRequest{Data: data})
			return err
		}),
		applogging.WithWriteLoginLogFunc(func(ctx context.Context, data *auditV1.LoginAuditLog) error {
			// TODO 如果系统的负载比较小，可以同步写入数据库，否则，建议使用异步方式，即投递进队列。
			ctx, _ = metadata.NewContext(ctx, metadata.NewUserOperator(0, 0, 0, identityV1.DataScope_ALL))
			_, err := loginAuditLogServiceClient.Create(ctx, &auditV1.CreateLoginAuditLogRequest{Data: data})
			return err
		}),
		applogging.WithWriteOperationLogFunc(func(ctx context.Context, data *auditV1.OperationAuditLog) error {
			ctx, _ = metadata.NewContext(ctx, metadata.NewUserOperator(0, 0, 0, identityV1.DataScope_ALL))
			_, err := operationAuditLogServiceClient.Create(ctx, &auditV1.CreateOperationAuditLogRequest{Data: data})
			return err
		}),
	))

	ms = append(ms, selector.Server(
		auth.Server(
			auth.WithAccessTokenChecker(accessTokenChecker),
			auth.WithInjectMetadata(true),
			auth.WithInjectEnt(true),
		),
		authz.Server(authorizer),
	).Match(rpc.NewRestWhiteListMatcher()).Build())

	return ms
}

// NewRestServer new an REST server.
func NewRestServer(
	ctx *bootstrap.Context,

	middlewares []middleware.Middleware,

	userService *service.UserService,
	userProfileService *service.UserProfileService,
	roleService *service.RoleService,
	tenantService *service.TenantService,
	orgUnitService *service.OrgUnitService,
	positionService *service.PositionService,

	menuSvc *service.MenuService,
	apiService *service.ApiService,
	permissionGroupService *service.PermissionGroupService,
	permissionService *service.PermissionService,

	adminPortalService *service.AdminPortalService,
	taskService *service.TaskService,

	authenticationService *service.AuthenticationService,
	loginPolicyService *service.LoginPolicyService,

	languageService *service.LanguageService,

	categoryService *service.CategoryService,
	brandService *service.BrandService,
	productService *service.ProductService,
	productAttributeService *service.ProductAttributeService,
	productAttributeValueService *service.ProductAttributeValueService,
	skuService *service.SkuService,
	skuPriceService *service.SkuPriceService,
	skuAttributeCombinationService *service.SkuAttributeCombinationService,

	cartService *service.CartService,
	cartItemService *service.CartItemService,
	orderService *service.OrderService,
	orderItemService *service.OrderItemService,
	paymentTransactionService *service.PaymentTransactionService,
	paymentRefundService *service.PaymentRefundService,
	shipmentService *service.ShipmentService,

	couponTemplateService *service.CouponTemplateService,
	promotionService *service.PromotionService,
	stockOpLogService *service.StockOpLogService,
	userCouponService *service.UserCouponService,

		shippingRateService *service.ShippingRateService,
		stockAlertService *service.StockAlertService,
		taxRateService *service.TaxRateService,
		commentService *service.CommentService,
		interactionAdminService *service.InteractionAdminService,
		invoiceService *service.InvoiceService,

		fileSvc *service.FileService,
	fileTransferService *service.FileTransferService,

	internalMessageService *service.InternalMessageService,
	internalMessageCategoryService *service.InternalMessageCategoryService,
	internalMessageRecipientService *service.InternalMessageRecipientService,

	notificationService *service.NotificationService,
	notificationChannelService *service.NotificationChannelService,
	notificationRuleService *service.NotificationRuleService,
	notificationTemplateService *service.NotificationTemplateService,
	notificationPreferenceService *service.NotificationPreferenceService,
	auditExportService *service.AuditExportService,
	dictTypeService *service.DictTypeService,
	dictEntryService *service.DictEntryService,
	configService *service.ConfigService,
	accessKeyService *service.AccessKeyService,
	onlineSessionService *service.OnlineSessionService,
	monitorAlertService *service.MonitorAlertService,
	serverMonitorService *service.ServerMonitorService,
	redisCacheMonitorService *service.RedisCacheMonitorService,
	scriptService *service.ScriptService,
	scriptLogService *service.ScriptLogService,
	planService *service.PlanService,
	planModuleService *service.PlanModuleService,
	planQuotaService *service.PlanQuotaService,
	mfaService *service.MfaService,
	aiAgentService *service.AiAgentService,
	aiProviderService *service.AiProviderService,
	aiChatService *service.AiChatService,
	aiConversationService *service.AiConversationService,
	aiMessageService *service.AiMessageService,
	aiKnowledgeService *service.AiKnowledgeService,
	aiQueryService *service.AiQueryService,
	aiUsageLogService *service.AiUsageLogService,

	apiAuditLogService *service.ApiAuditLogService,
	dataAccessAuditLogService *service.DataAccessAuditLogService,
	loginAuditLogService *service.LoginAuditLogService,
	policyEvaluationLogService *service.PolicyEvaluationLogService,
	operationAuditLogService *service.OperationAuditLogService,
	permissionAuditLogService *service.PermissionAuditLogService,

	// register:rest-param ── 新模块服务形参在此行后追加(make register 工具锚点,勿删)
) *http.Server {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Rest == nil {
		return nil
	}

	srv, err := rpc.CreateRestServer(cfg, middlewares...)
	if err != nil {
		panic(err)
	}

	adminV1.RegisterAuthenticationServiceHTTPServer(srv, authenticationService)
	adminV1.RegisterLoginPolicyServiceHTTPServer(srv, loginPolicyService)

	adminV1.RegisterUserProfileServiceHTTPServer(srv, userProfileService)
	adminV1.RegisterUserServiceHTTPServer(srv, userService)
	adminV1.RegisterRoleServiceHTTPServer(srv, roleService)
	adminV1.RegisterTenantServiceHTTPServer(srv, tenantService)
	adminV1.RegisterOrgUnitServiceHTTPServer(srv, orgUnitService)
	adminV1.RegisterPositionServiceHTTPServer(srv, positionService)

	adminV1.RegisterAdminPortalServiceHTTPServer(srv, adminPortalService)
	adminV1.RegisterTaskServiceHTTPServer(srv, taskService)

	adminV1.RegisterLanguageServiceHTTPServer(srv, languageService)

	adminV1.RegisterCategoryServiceHTTPServer(srv, categoryService)
	adminV1.RegisterBrandServiceHTTPServer(srv, brandService)
	adminV1.RegisterProductServiceHTTPServer(srv, productService)
	adminV1.RegisterProductAttributeServiceHTTPServer(srv, productAttributeService)
	adminV1.RegisterProductAttributeValueServiceHTTPServer(srv, productAttributeValueService)
	adminV1.RegisterSkuServiceHTTPServer(srv, skuService)
	adminV1.RegisterSkuPriceServiceHTTPServer(srv, skuPriceService)
	adminV1.RegisterSkuAttributeCombinationServiceHTTPServer(srv, skuAttributeCombinationService)

	adminV1.RegisterCartServiceHTTPServer(srv, cartService)
	adminV1.RegisterCartItemServiceHTTPServer(srv, cartItemService)
	adminV1.RegisterOrderServiceHTTPServer(srv, orderService)
	adminV1.RegisterOrderItemServiceHTTPServer(srv, orderItemService)
	adminV1.RegisterPaymentTransactionServiceHTTPServer(srv, paymentTransactionService)
	adminV1.RegisterPaymentRefundServiceHTTPServer(srv, paymentRefundService)
	adminV1.RegisterShipmentServiceHTTPServer(srv, shipmentService)

	adminV1.RegisterCouponTemplateServiceHTTPServer(srv, couponTemplateService)
	adminV1.RegisterPromotionServiceHTTPServer(srv, promotionService)
	adminV1.RegisterStockOpLogServiceHTTPServer(srv, stockOpLogService)
	adminV1.RegisterUserCouponServiceHTTPServer(srv, userCouponService)

	adminV1.RegisterShippingRateServiceHTTPServer(srv, shippingRateService)
	adminV1.RegisterStockAlertServiceHTTPServer(srv, stockAlertService)
	adminV1.RegisterTaxRateServiceHTTPServer(srv, taxRateService)
	adminV1.RegisterCommentServiceHTTPServer(srv, commentService)
	adminV1.RegisterInteractionAdminServiceHTTPServer(srv, interactionAdminService)
	adminV1.RegisterInvoiceServiceHTTPServer(srv, invoiceService)

	adminV1.RegisterApiServiceHTTPServer(srv, apiService)
	adminV1.RegisterMenuServiceHTTPServer(srv, menuSvc)
	adminV1.RegisterPermissionGroupServiceHTTPServer(srv, permissionGroupService)
	adminV1.RegisterPermissionServiceHTTPServer(srv, permissionService)

	adminV1.RegisterApiAuditLogServiceHTTPServer(srv, apiAuditLogService)
	adminV1.RegisterDataAccessAuditLogServiceHTTPServer(srv, dataAccessAuditLogService)
	adminV1.RegisterLoginAuditLogServiceHTTPServer(srv, loginAuditLogService)
	adminV1.RegisterOperationAuditLogServiceHTTPServer(srv, operationAuditLogService)
	adminV1.RegisterPermissionAuditLogServiceHTTPServer(srv, permissionAuditLogService)
	adminV1.RegisterPolicyEvaluationLogServiceHTTPServer(srv, policyEvaluationLogService)

	adminV1.RegisterInternalMessageServiceHTTPServer(srv, internalMessageService)
	adminV1.RegisterInternalMessageCategoryServiceHTTPServer(srv, internalMessageCategoryService)
	adminV1.RegisterInternalMessageRecipientServiceHTTPServer(srv, internalMessageRecipientService)

	adminV1.RegisterNotificationServiceHTTPServer(srv, notificationService)
	adminV1.RegisterNotificationChannelServiceHTTPServer(srv, notificationChannelService)
	adminV1.RegisterNotificationRuleServiceHTTPServer(srv, notificationRuleService)
	adminV1.RegisterNotificationTemplateServiceHTTPServer(srv, notificationTemplateService)
	adminV1.RegisterNotificationPreferenceServiceHTTPServer(srv, notificationPreferenceService)

	// 审计日志服务端导出：手动注册（二进制文件响应 + Content-Disposition，
	// proto 生成路由不支持）。鉴权在 handler 内显式做（验 Bearer + 重建操作者）。
	srv.Route("/").GET("admin/v1/audit-logs:export", func(ctx http.Context) error {
		return auditExportService.ServeExport(ctx.Response(), ctx.Request())
	})
	srv.Route("/").GET("admin/v1/ai/usage-logs:export", func(ctx http.Context) error {
		return auditExportService.ServeAiUsageExport(ctx.Response(), ctx.Request())
	})
	srv.Route("/").GET("admin/v1/orders:export", func(ctx http.Context) error {
		return auditExportService.ServeOrdersExport(ctx.Response(), ctx.Request())
	})
	adminV1.RegisterDictTypeServiceHTTPServer(srv, dictTypeService)
	adminV1.RegisterDictEntryServiceHTTPServer(srv, dictEntryService)
	adminV1.RegisterConfigServiceHTTPServer(srv, configService)
	adminV1.RegisterAccessKeyServiceHTTPServer(srv, accessKeyService)
	adminV1.RegisterOnlineSessionServiceHTTPServer(srv, onlineSessionService)
	adminV1.RegisterMonitorAlertServiceHTTPServer(srv, monitorAlertService)
	adminV1.RegisterServerMonitorServiceHTTPServer(srv, serverMonitorService)
	adminV1.RegisterRedisCacheMonitorServiceHTTPServer(srv, redisCacheMonitorService)
	adminV1.RegisterScriptServiceHTTPServer(srv, scriptService)
	adminV1.RegisterScriptLogServiceHTTPServer(srv, scriptLogService)
	adminV1.RegisterPlanServiceHTTPServer(srv, planService)
	adminV1.RegisterPlanModuleServiceHTTPServer(srv, planModuleService)
	adminV1.RegisterPlanQuotaServiceHTTPServer(srv, planQuotaService)
	adminV1.RegisterMfaServiceHTTPServer(srv, mfaService)
	adminV1.RegisterAiAgentServiceHTTPServer(srv, aiAgentService)
	adminV1.RegisterAiProviderServiceHTTPServer(srv, aiProviderService)
	adminV1.RegisterAiChatServiceHTTPServer(srv, aiChatService)
	adminV1.RegisterAiConversationServiceHTTPServer(srv, aiConversationService)
	adminV1.RegisterAiMessageServiceHTTPServer(srv, aiMessageService)
	adminV1.RegisterAiKnowledgeBaseServiceHTTPServer(srv, aiKnowledgeService)
	adminV1.RegisterAiQueryServiceHTTPServer(srv, aiQueryService)
	adminV1.RegisterAiUsageLogServiceHTTPServer(srv, aiUsageLogService)

	// 注册文件传输服务，用于处理文件上传下载等功能
	// TODO 它不能够使用代码生成器生成的Handler，需要手动注册。代码生成器生成的Handler无法处理文件上传下载的请求。
	// 但，代码生成器生成代码可以提供给OpenAPI使用。
	registerFileTransferServiceHandler(srv, fileTransferService)
	adminV1.RegisterFileServiceHTTPServer(srv, fileSvc)

	// register:rest-route ── 新模块服务路由在此行后注册(make register 工具锚点,勿删)

	if cfg.GetServer().GetRest().GetEnableSwagger() {
		swaggerUI.RegisterSwaggerUIServerWithOption(
			srv,
			swaggerUI.WithTitle("GoWind Content Hub Admin API"),
			swaggerUI.WithMemoryData(assets.OpenApiData, "yaml"),
		)
	}

	return srv
}
