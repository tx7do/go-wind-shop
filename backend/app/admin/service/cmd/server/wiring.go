package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"go-wind-shop/app/admin/service/internal/data"
	"go-wind-shop/app/admin/service/internal/server"
	"go-wind-shop/app/admin/service/internal/service"
	"go-wind-shop/pkg/middleware/auth"
)

// initApp 手写装配整个应用,是本服务的依赖注入点(替代原 google/wire 生成)。
// initApp assembles the whole application by hand — the dependency injection point.
//
// 装配严格单向分层,自上而下的阅读顺序即依赖方向:
// The wiring is strictly layered; reading top-down follows the dependency direction:
//
//	基础设施 → 服务客户端(data) → 服务层(service) → 传输层(server)
//
// 本服务是 BFF 网关:不直连数据库,仓储层由对 core-service 的 gRPC 服务客户端替代。
// 约定 / Conventions:
//   - 新增模块:先加服务客户端行,再加服务构造行,并传给 NewRestServer;漏接由编译器在调用处报错。
//     To add a module: append a service-client line, then a service constructor, then pass it to
//     NewRestServer; a missing connection is a compile error at the call site.
//   - 持有 cleanup 的资源创建成功后立即注册进 cleanups;任何一步失败,rollback 逆序执行已注册
//     的清理(shutdown 与中途失败共用同一条 LIFO 路径)。
//     Resources owning a cleanup register it immediately; rollback runs them LIFO both on
//     mid-way failure and on shutdown.
//   - 本文件只做构造与传参,不写业务逻辑。
//     Construction and parameter passing only; no business logic in this file.
func initApp(ctx *bootstrap.Context) (*kratos.App, func(), error) {
	// cleanup 注册表:rollback 时逆序执行。
	// Cleanup registry; rollback runs entries in reverse order.
	var cleanups []func()
	rollback := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	// ═══════════════════════ 一、基础设施 ═══════════════════════

	discovery := data.NewDiscovery(ctx)
	authenticationServiceClient := data.NewAuthenticationServiceClient(ctx, discovery)
	clientType := data.NewClientType()
	accessTokenChecker := auth.NewTokenChecker(ctx, authenticationServiceClient, clientType)
	engine := data.NewAuthorizer()
	redisClient, cleanupRedis, err := data.NewRedisClient(ctx)
	if err != nil {
		rollback()
		return nil, nil, err
	}
	cleanups = append(cleanups, cleanupRedis)

	captcha := data.NewCaptcha(redisClient)
	minIOClient := data.NewMinIoClient(ctx)

	// ═══════════════════════ 二、服务客户端(internal/data) ═══════════════════════

	apiAuditLogServiceClient := data.NewApiAuditLogServiceClient(ctx, discovery)
	loginAuditLogServiceClient := data.NewLoginAuditLogServiceClient(ctx, discovery)
	operationAuditLogServiceClient := data.NewOperationAuditLogServiceClient(ctx, discovery)
	userServiceClient := data.NewUserServiceClient(ctx, discovery)
	tenantServiceClient := data.NewTenantServiceClient(ctx, discovery)
	orgUnitServiceClient := data.NewOrgUnitServiceClient(ctx, discovery)
	positionServiceClient := data.NewPositionServiceClient(ctx, discovery)
	roleServiceClient := data.NewRoleServiceClient(ctx, discovery)
	userCredentialServiceClient := data.NewUserCredentialServiceClient(ctx, discovery)
	menuServiceClient := data.NewMenuServiceClient(ctx, discovery)
	apiServiceClient := data.NewApiServiceClient(ctx, discovery)
	permissionServiceClient := data.NewPermissionServiceClient(ctx, discovery)
	permissionGroupServiceClient := data.NewPermissionGroupServiceClient(ctx, discovery)
	taskServiceClient := data.NewTaskServiceClient(ctx, discovery)
	loginPolicyServiceClient := data.NewLoginPolicyServiceClient(ctx, discovery)
	languageServiceClient := data.NewLanguageServiceClient(ctx, discovery)
	categoryServiceClient := data.NewCategoryServiceClient(ctx, discovery)
	brandServiceClient := data.NewBrandServiceClient(ctx, discovery)
	productServiceClient := data.NewProductServiceClient(ctx, discovery)
	productAttributeServiceClient := data.NewProductAttributeServiceClient(ctx, discovery)
	productAttributeValueServiceClient := data.NewProductAttributeValueServiceClient(ctx, discovery)
	skuServiceClient := data.NewSkuServiceClient(ctx, discovery)
	skuPriceServiceClient := data.NewSkuPriceServiceClient(ctx, discovery)
	skuAttributeCombinationServiceClient := data.NewSkuAttributeCombinationServiceClient(ctx, discovery)
	cartServiceClient := data.NewCartServiceClient(ctx, discovery)
	cartItemServiceClient := data.NewCartItemServiceClient(ctx, discovery)
	orderServiceClient := data.NewOrderServiceClient(ctx, discovery)
	orderItemServiceClient := data.NewOrderItemServiceClient(ctx, discovery)
	paymentTransactionServiceClient := data.NewPaymentTransactionServiceClient(ctx, discovery)
	paymentRefundServiceClient := data.NewPaymentRefundServiceClient(ctx, discovery)
	shipmentServiceClient := data.NewShipmentServiceClient(ctx, discovery)
	couponTemplateServiceClient := data.NewCouponTemplateServiceClient(ctx, discovery)
	userCouponServiceClient := data.NewUserCouponServiceClient(ctx, discovery)
	shippingRateServiceClient := data.NewShippingRateServiceClient(ctx, discovery)
	stockAlertServiceClient := data.NewStockAlertServiceClient(ctx, discovery)
	taxRateServiceClient := data.NewTaxRateServiceClient(ctx, discovery)
	commentServiceClient := data.NewCommentServiceClient(ctx, discovery)
	interactionAdminServiceClient := data.NewInteractionAdminServiceClient(ctx, discovery)
	invoiceServiceClient := data.NewInvoiceServiceClient(ctx, discovery)
	fileServiceClient := data.NewFileServiceClient(ctx, discovery)
	internalMessageServiceClient := data.NewInternalMessageServiceClient(ctx, discovery)
	internalMessageCategoryServiceClient := data.NewInternalMessageCategoryServiceClient(ctx, discovery)
	internalMessageRecipientServiceClient := data.NewInternalMessageRecipientServiceClient(ctx, discovery)
	notificationServiceClient := data.NewNotificationServiceClient(ctx, discovery)
	notificationChannelServiceClient := data.NewNotificationChannelServiceClient(ctx, discovery)
	notificationRuleServiceClient := data.NewNotificationRuleServiceClient(ctx, discovery)
	notificationTemplateServiceClient := data.NewNotificationTemplateServiceClient(ctx, discovery)
	notificationPreferenceServiceClient := data.NewNotificationPreferenceServiceClient(ctx, discovery)
	dataAccessAuditLogServiceClient := data.NewDataAccessAuditLogServiceClient(ctx, discovery)
	policyEvaluationLogServiceClient := data.NewPolicyEvaluationLogServiceClient(ctx, discovery)
	permissionAuditLogServiceClient := data.NewPermissionAuditLogServiceClient(ctx, discovery)

	// ── register:client ── 新模块服务客户端在此行后注册(make register 工具锚点,勿删)

	// ═══════════════════════ 三、服务层(internal/service) ═══════════════════════

	userService := service.NewUserService(ctx, userServiceClient, tenantServiceClient, orgUnitServiceClient, positionServiceClient, roleServiceClient, userCredentialServiceClient)
	userProfileService := service.NewUserProfileService(ctx, userServiceClient, tenantServiceClient, orgUnitServiceClient, positionServiceClient, roleServiceClient, userCredentialServiceClient)
	roleService := service.NewRoleService(ctx, roleServiceClient, tenantServiceClient)
	tenantService := service.NewTenantService(ctx, userServiceClient, userCredentialServiceClient, tenantServiceClient, roleServiceClient)
	orgUnitService := service.NewOrgUnitService(ctx, orgUnitServiceClient, userServiceClient)
	positionService := service.NewPositionService(ctx, positionServiceClient, orgUnitServiceClient)
	menuService := service.NewMenuService(ctx, menuServiceClient)
	apiService := service.NewApiService(ctx, apiServiceClient)
	permissionGroupService := service.NewPermissionGroupService(ctx, permissionServiceClient, permissionGroupServiceClient)
	permissionService := service.NewPermissionService(ctx, permissionServiceClient, permissionGroupServiceClient, roleServiceClient, apiServiceClient, menuServiceClient)
	adminPortalService := service.NewRouterService(ctx, menuServiceClient, permissionServiceClient, roleServiceClient, userServiceClient)
	taskService := service.NewTaskService(ctx, taskServiceClient)
	authenticationService := service.NewAuthenticationService(ctx, authenticationServiceClient, captcha)
	loginPolicyService := service.NewLoginPolicyService(ctx, loginPolicyServiceClient)
	languageService := service.NewLanguageService(ctx, languageServiceClient)
	categoryService := service.NewCategoryService(ctx, categoryServiceClient)
	brandService := service.NewBrandService(ctx, brandServiceClient)
	productService := service.NewProductService(ctx, productServiceClient)
	productAttributeService := service.NewProductAttributeService(ctx, productAttributeServiceClient)
	productAttributeValueService := service.NewProductAttributeValueService(ctx, productAttributeValueServiceClient)
	skuService := service.NewSkuService(ctx, skuServiceClient)
	skuPriceService := service.NewSkuPriceService(ctx, skuPriceServiceClient)
	skuAttributeCombinationService := service.NewSkuAttributeCombinationService(ctx, skuAttributeCombinationServiceClient)
	cartService := service.NewCartService(ctx, cartServiceClient)
	cartItemService := service.NewCartItemService(ctx, cartItemServiceClient)
	orderService := service.NewOrderService(ctx, orderServiceClient)
	orderItemService := service.NewOrderItemService(ctx, orderItemServiceClient)
	paymentTransactionService := service.NewPaymentTransactionService(ctx, paymentTransactionServiceClient)
	paymentRefundService := service.NewPaymentRefundService(ctx, paymentRefundServiceClient)
	shipmentService := service.NewShipmentService(ctx, shipmentServiceClient)
	couponTemplateService := service.NewCouponTemplateService(ctx, couponTemplateServiceClient)
	userCouponService := service.NewUserCouponService(ctx, userCouponServiceClient)
	shippingRateService := service.NewShippingRateService(ctx, shippingRateServiceClient)
	stockAlertService := service.NewStockAlertService(ctx, stockAlertServiceClient)
	taxRateService := service.NewTaxRateService(ctx, taxRateServiceClient)
	commentService := service.NewCommentService(ctx, commentServiceClient)
	interactionAdminService := service.NewInteractionAdminService(ctx, interactionAdminServiceClient)
	invoiceService := service.NewInvoiceService(ctx, invoiceServiceClient)
	fileService := service.NewFileService(ctx, fileServiceClient)
	fileTransferService := service.NewFileTransferService(ctx, minIOClient, fileServiceClient)
	internalMessageService := service.NewInternalMessageService(ctx, internalMessageServiceClient, internalMessageCategoryServiceClient, internalMessageRecipientServiceClient, authenticationServiceClient, userServiceClient, clientType)
	internalMessageCategoryService := service.NewInternalMessageCategoryService(ctx, internalMessageCategoryServiceClient)
	internalMessageRecipientService := service.NewInternalMessageRecipientService(ctx, internalMessageServiceClient, internalMessageRecipientServiceClient)

	// ── Notification 通知域（移植自 go-wind-cms）──
	notificationService := service.NewNotificationService(ctx, notificationServiceClient)
	notificationChannelService := service.NewNotificationChannelService(ctx, notificationChannelServiceClient)
	notificationRuleService := service.NewNotificationRuleService(ctx, notificationRuleServiceClient)
	notificationTemplateService := service.NewNotificationTemplateService(ctx, notificationTemplateServiceClient)
	notificationPreferenceService := service.NewNotificationPreferenceService(ctx, notificationPreferenceServiceClient)
	apiAuditLogService := service.NewApiAuditLogService(ctx, apiAuditLogServiceClient, apiServiceClient)
	dataAccessAuditLogService := service.NewDataAccessAuditLogService(ctx, dataAccessAuditLogServiceClient)
	loginAuditLogService := service.NewLoginAuditLogService(ctx, loginAuditLogServiceClient)
	policyEvaluationLogService := service.NewPolicyEvaluationLogService(ctx, policyEvaluationLogServiceClient)
	operationAuditLogService := service.NewOperationAuditLogService(ctx, operationAuditLogServiceClient)
	permissionAuditLogService := service.NewPermissionAuditLogService(ctx, permissionAuditLogServiceClient)

	// ── register:service ── 新模块服务在此行后注册(make register 工具锚点,勿删)

	// ═══════════════════════ 四、传输层(internal/server) ═══════════════════════

	restMiddlewares := server.NewRestMiddleware(ctx, accessTokenChecker, engine, apiAuditLogServiceClient, loginAuditLogServiceClient, operationAuditLogServiceClient)

	httpServer := server.NewRestServer(ctx, restMiddlewares,
		userService,
		userProfileService,
		roleService,
		tenantService,
		orgUnitService,
		positionService,
		menuService,
		apiService,
		permissionGroupService,
		permissionService,
		adminPortalService,
		taskService,
		authenticationService,
		loginPolicyService,
		languageService,
		categoryService,
		brandService,
		productService,
		productAttributeService,
		productAttributeValueService,
		skuService,
		skuPriceService,
		skuAttributeCombinationService,
		cartService,
		cartItemService,
		orderService,
		orderItemService,
		paymentTransactionService,
		paymentRefundService,
		shipmentService,
		couponTemplateService,
		userCouponService,
		shippingRateService,
		stockAlertService,
		taxRateService,
		commentService,
		interactionAdminService,
		invoiceService,
		fileService,
		fileTransferService,
		internalMessageService,
		internalMessageCategoryService,
		internalMessageRecipientService,
		notificationService,
		notificationChannelService,
		notificationRuleService,
		notificationTemplateService,
		notificationPreferenceService,
		apiAuditLogService,
		dataAccessAuditLogService,
		loginAuditLogService,
		policyEvaluationLogService,
		operationAuditLogService,
		permissionAuditLogService,
		// register:rest-arg ── 新模块服务实参在此行后追加(make register 工具锚点,勿删)
	)

	grpcMiddlewares := server.NewGrpcMiddleware(ctx)

	grpcServer, err := server.NewGrpcServer(ctx, grpcMiddlewares)
	if err != nil {
		rollback()
		return nil, nil, err
	}

	sseServer := server.NewSseServer(ctx, internalMessageService)

	return newApp(ctx, httpServer, grpcServer, sseServer), rollback, nil
}
