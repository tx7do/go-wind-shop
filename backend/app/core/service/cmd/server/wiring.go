package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/channel"
	"go-wind-shop/app/core/service/internal/server"
	"go-wind-shop/app/core/service/internal/service"
)

// initApp 手写装配整个应用,是本服务的依赖注入点(替代原 google/wire 生成)。
// initApp assembles the whole application by hand — the dependency injection point.
//
// 装配严格单向分层,自上而下的阅读顺序即依赖方向:
// The wiring is strictly layered; reading top-down follows the dependency direction:
//
//	基础设施 → 仓储层(data) → 服务层(service) → 传输层(server)
//
// 约定 / Conventions:
//   - 新增仓储/服务:在对应分层小节追加一行构造,并传给下游消费者;漏接由编译器在调用处报错。
//     To add a repo/service: append one constructor line in its layer section, then pass it to
//     downstream consumers; a missing connection is a compile error at the call site.
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

	authenticatorOption := data.NewAuthenticatorConfig(ctx)
	redisClient, cleanupRedis, err := data.NewRedisClient(ctx)
	if err != nil {
		rollback()
		return nil, nil, err
	}
	cleanups = append(cleanups, cleanupRedis)

	userTokenCache := data.NewUserTokenCache(ctx, redisClient)
	authenticator := data.NewAuthenticator(ctx, authenticatorOption, userTokenCache)

	entClient, cleanupEnt, err := data.NewEntClient(ctx)
	if err != nil {
		rollback()
		return nil, nil, err
	}
	cleanups = append(cleanups, cleanupEnt)

	// ── Plan 租户套餐域（移植自 go-wind-admin）──
	planRepo := data.NewPlanRepo(ctx, entClient)
	planModuleRepo := data.NewPlanModuleRepo(ctx, entClient)
	planQuotaRepo := data.NewPlanQuotaRepo(ctx, entClient)
	tenantUsageRepo := data.NewTenantUsageRepo(ctx, entClient, authenticator)
	mfaFactorRepo := data.NewUserMfaFactorRepo(ctx, entClient)
	mfaChallengeCache := data.NewMfaChallengeCache(ctx, redisClient)
	loginRateLimiter := data.NewLoginRateLimiter(ctx, redisClient)

	crypto := data.NewPasswordCrypto()
	minIOClient := data.NewMinIoClient(ctx)
	elasticsearchClient, cleanupSearch, err := data.NewElasticSearchClient(ctx)
	if err != nil {
		rollback()
		return nil, nil, err
	}
	cleanups = append(cleanups, cleanupSearch)

	// ═══════════════════════ 二、仓储层(internal/data) ═══════════════════════

	userCredentialRepo := data.NewUserCredentialRepo(ctx, entClient, crypto)
	userRoleRepo := data.NewUserRoleRepo(ctx, entClient)
	userOrgUnitRepo := data.NewUserOrgUnitRepo(ctx, entClient)
	userPositionRepo := data.NewUserPositionRepo(ctx, entClient)
	membershipRoleRepo := data.NewMembershipRoleRepo(ctx, entClient)
	membershipPositionRepo := data.NewMembershipPositionRepo(ctx, entClient)
	membershipOrgUnitRepo := data.NewMembershipOrgUnitRepo(ctx, entClient)
	membershipRepo := data.NewMembershipRepo(ctx, entClient, membershipRoleRepo, membershipPositionRepo, membershipOrgUnitRepo)
	userRepo := data.NewUserRepo(ctx, entClient, userRoleRepo, userOrgUnitRepo, userPositionRepo, membershipRepo)
	rolePermissionRepo := data.NewRolePermissionRepo(ctx, entClient)
	permissionApiRepo := data.NewPermissionApiRepo(ctx, entClient)
	permissionMenuRepo := data.NewPermissionMenuRepo(ctx, entClient)
	permissionRepo := data.NewPermissionRepo(ctx, entClient, permissionApiRepo, permissionMenuRepo)
	roleMetadataRepo := data.NewRoleMetadataRepo(ctx, entClient)
	roleRepo := data.NewRoleRepo(ctx, entClient, rolePermissionRepo, permissionRepo, roleMetadataRepo, userRoleRepo)
	tenantRepo := data.NewTenantRepo(ctx, entClient)
	loginPolicyRepo := data.NewLoginPolicyRepo(ctx, entClient)
	taskRepo := data.NewTaskRepo(ctx, entClient)
	fileRepo := data.NewFileRepo(ctx, entClient)
	languageRepo := data.NewLanguageRepo(ctx, entClient)
	positionRepo := data.NewPositionRepo(ctx, entClient)
	orgUnitRepo := data.NewOrgUnitRepo(ctx, entClient)
	menuRepo := data.NewMenuRepo(ctx, entClient)
	apiRepo := data.NewApiRepo(ctx, entClient)
	permissionGroupRepo := data.NewPermissionGroupRepo(ctx, entClient)
	permissionAuditLogRepo := data.NewPermissionAuditLogRepo(ctx, entClient)
	policyEvaluationLogRepo := data.NewPolicyEvaluationLogRepo(ctx, entClient)
	loginAuditLogRepo := data.NewLoginAuditLogRepo(ctx, entClient)
	apiAuditLogRepo := data.NewApiAuditLogRepo(ctx, entClient)
	operationAuditLogRepo := data.NewOperationAuditLogRepo(ctx, entClient)
	dataAccessAuditLogRepo := data.NewDataAccessAuditLogRepo(ctx, entClient)
	internalMessageRepo := data.NewInternalMessageRepo(ctx, entClient)
	internalMessageCategoryRepo := data.NewInternalMessageCategoryRepo(ctx, entClient)
	internalMessageRecipientRepo := data.NewInternalMessageRecipientRepo(ctx, entClient)
	categoryTranslationRepo := data.NewCategoryTranslationRepo(ctx, entClient)
	categoryRepo := data.NewCategoryRepo(ctx, entClient, categoryTranslationRepo)
	brandTranslationRepo := data.NewBrandTranslationRepo(ctx, entClient)
	brandRepo := data.NewBrandRepo(ctx, entClient, brandTranslationRepo)
	productTranslationRepo := data.NewProductTranslationRepo(ctx, entClient)
	productRepo := data.NewProductRepo(ctx, entClient, productTranslationRepo)
	productSearchRepo := data.NewProductSearchRepo(ctx, elasticsearchClient)
	productAttributeTranslationRepo := data.NewProductAttributeTranslationRepo(ctx, entClient)
	productAttributeRepo := data.NewProductAttributeRepo(ctx, entClient, productAttributeTranslationRepo)
	productAttributeValueTranslationRepo := data.NewProductAttributeValueTranslationRepo(ctx, entClient)
	productAttributeValueRepo := data.NewProductAttributeValueRepo(ctx, entClient, productAttributeValueTranslationRepo)
	skuRepo := data.NewSkuRepo(ctx, entClient)
	skuPriceRepo := data.NewSkuPriceRepo(ctx, entClient)
	skuAttributeCombinationRepo := data.NewSkuAttributeCombinationRepo(ctx, entClient)
	stockAlertRepo := data.NewStockAlertRepo(ctx, entClient)
	cartRepo := data.NewCartRepo(ctx, entClient)
	cartItemRepo := data.NewCartItemRepo(ctx, entClient)
	wishlistRepo := data.NewWishlistRepo(ctx, entClient)
	orderRepo := data.NewOrderRepo(ctx, entClient)
	orderItemRepo := data.NewOrderItemRepo(ctx, entClient)
	paymentTransactionRepo := data.NewPaymentTransactionRepo(ctx, entClient)
	paymentRefundRepo := data.NewPaymentRefundRepo(ctx, entClient)
	paymentChannelConfigRepo := data.NewPaymentChannelConfigRepo(ctx, entClient)
	shippingAddressRepo := data.NewShippingAddressRepo(ctx, entClient)
	shipmentRepo := data.NewShipmentRepo(ctx, entClient)
	couponTemplateRepo := data.NewCouponTemplateRepo(ctx, entClient)
	promotionRepo := data.NewPromotionRepo(ctx, entClient)
	stockOpLogRepo := data.NewStockOpLogRepo(ctx, entClient)
	userCouponRepo := data.NewUserCouponRepo(ctx, entClient)
	shippingRateRepo := data.NewShippingRateRepo(ctx, entClient)
	taxRateRepo := data.NewTaxRateRepo(ctx, entClient)
	commentRepo := data.NewCommentRepo(ctx, entClient)
	interactionRepo := data.NewInteractionRepo(ctx, entClient)
	invoiceRepo := data.NewInvoiceRepo(ctx, entClient)

	// ── register:repo ── 新模块仓储在此行后注册(make register 工具锚点,勿删)
	seckillActivityRepo := data.NewSeckillActivityRepo(ctx, entClient)

	// ═══════════════════════ 三、服务层(internal/service) ═══════════════════════

	authenticationService := service.NewAuthenticationService(ctx, authenticator, mfaFactorRepo, mfaChallengeCache, loginRateLimiter, userCredentialRepo, userRepo, roleRepo, tenantRepo, permissionRepo)
	loginPolicyService := service.NewLoginPolicyService(ctx, loginPolicyRepo)
	userCredentialService := service.NewUserCredentialService(ctx, userCredentialRepo)
	taskService := service.NewTaskService(ctx, taskRepo, userRepo)
	fileService := service.NewFileService(ctx, fileRepo, minIOClient)
	languageService := service.NewLanguageService(ctx, languageRepo)
	tenantService := service.NewTenantService(ctx, tenantRepo, tenantUsageRepo, userRepo, userCredentialRepo, roleRepo)
	mfaService := service.NewMfaService(ctx, mfaFactorRepo, mfaChallengeCache, authenticator, loginRateLimiter, userRepo)
	userService := service.NewUserService(ctx, userRepo, roleRepo, userCredentialRepo, positionRepo, orgUnitRepo, tenantRepo, membershipRepo)
	roleService := service.NewRoleService(ctx, roleRepo, tenantRepo, userRoleRepo)
	positionService := service.NewPositionService(ctx, positionRepo, orgUnitRepo)
	orgUnitService := service.NewOrgUnitService(ctx, orgUnitRepo, userRepo)
	menuService := service.NewMenuService(ctx, menuRepo)
	apiService := service.NewApiService(ctx, apiRepo)
	permissionService := service.NewPermissionService(ctx, permissionRepo, permissionGroupRepo, menuRepo, apiRepo, roleRepo)
	permissionGroupService := service.NewPermissionGroupService(ctx, permissionGroupRepo, permissionRepo)
	permissionAuditLogService := service.NewPermissionAuditLogService(ctx, permissionAuditLogRepo)
	policyEvaluationLogService := service.NewPolicyEvaluationLogService(ctx, policyEvaluationLogRepo)
	loginAuditLogService := service.NewLoginAuditLogService(ctx, loginAuditLogRepo)
	apiAuditLogService := service.NewApiAuditLogService(ctx, apiAuditLogRepo, apiRepo)
	operationAuditLogService := service.NewOperationAuditLogService(ctx, operationAuditLogRepo)
	dataAccessAuditLogService := service.NewDataAccessAuditLogService(ctx, dataAccessAuditLogRepo)
	internalMessageService := service.NewInternalMessageService(ctx, internalMessageRepo, internalMessageCategoryRepo, internalMessageRecipientRepo, userRepo)
	internalMessageCategoryService := service.NewInternalMessageCategoryService(ctx, internalMessageCategoryRepo)
	internalMessageRecipientService := service.NewInternalMessageRecipientService(ctx, internalMessageRepo, internalMessageRecipientRepo, userRepo)
	categoryService := service.NewCategoryService(ctx, categoryRepo)
	brandService := service.NewBrandService(ctx, brandRepo)
	productSearchService := service.NewProductSearchService(ctx, productSearchRepo, productRepo)
	productService := service.NewProductService(ctx, productRepo, productSearchService, taskService)
	productAttributeService := service.NewProductAttributeService(ctx, productAttributeRepo)
	productAttributeValueService := service.NewProductAttributeValueService(ctx, productAttributeValueRepo)
	skuService := service.NewSkuService(ctx, skuRepo, entClient)
	stockOpLogService := service.NewStockOpLogService(ctx, stockOpLogRepo)
	skuPriceService := service.NewSkuPriceService(ctx, skuPriceRepo)
	skuAttributeCombinationService := service.NewSkuAttributeCombinationService(ctx, skuAttributeCombinationRepo)
	stockAlertService := service.NewStockAlertService(ctx, stockAlertRepo)
	cartService := service.NewCartService(ctx, cartRepo)
	cartItemService := service.NewCartItemService(ctx, cartItemRepo)
	wishlistService := service.NewWishlistService(ctx, wishlistRepo)
	paymentChannelConfigService := service.NewPaymentChannelConfigService(ctx, paymentChannelConfigRepo)
	paymentGatewayResolver := service.NewPaymentGatewayResolver(ctx, paymentChannelConfigService)
	orderService := service.NewOrderService(ctx, orderRepo, orderItemRepo, skuRepo, cartItemRepo, entClient, internalMessageService, paymentGatewayResolver)
	orderItemService := service.NewOrderItemService(ctx, orderItemRepo)
	paymentTransactionService := service.NewPaymentTransactionService(ctx, paymentTransactionRepo, orderService, paymentChannelConfigService, paymentGatewayResolver, entClient)
	paymentRefundService := service.NewPaymentRefundService(ctx, paymentRefundRepo, paymentTransactionRepo, paymentTransactionService, paymentGatewayResolver)
	shippingAddressService := service.NewShippingAddressService(ctx, shippingAddressRepo)
	shipmentService := service.NewShipmentService(ctx, shipmentRepo, entClient)
	couponTemplateService := service.NewCouponTemplateService(ctx, couponTemplateRepo)
	promotionService := service.NewPromotionService(ctx, promotionRepo, entClient)
	userCouponService := service.NewUserCouponService(ctx, userCouponRepo, couponTemplateRepo, entClient)
	shippingRateService := service.NewShippingRateService(ctx, shippingRateRepo)
	taxRateService := service.NewTaxRateService(ctx, taxRateRepo)
	interactionService := service.NewInteractionService(ctx, interactionRepo)
	interactionAdminService := service.NewInteractionAdminService(ctx, interactionRepo, operationAuditLogRepo)
	invoiceService := service.NewInvoiceService(ctx, invoiceRepo)
	stockAlertScannerService := service.NewStockAlertScannerService(ctx, skuRepo, internalMessageService, stockAlertRepo)

	configRepo := data.NewConfigRepo(ctx, entClient, redisClient)
	dictEntryI18nRepo := data.NewDictEntryI18nRepo(ctx, entClient)
	dictEntryRepo := data.NewDictEntryRepo(ctx, entClient, dictEntryI18nRepo)
	dictTypeRepo := data.NewDictTypeRepo(ctx, entClient, dictEntryRepo)
	monitorAlertRuleRepo := data.NewMonitorAlertRuleRepo(ctx, entClient)
	accessKeyRepo := data.NewAccessKeyRepo(ctx, entClient)
	accessKeyRateLimiter := data.NewAccessKeyRateLimiter(ctx, redisClient)

	// ── Script 脚本域（移植自 go-wind-admin，ent 同构）──
	scriptRepo := data.NewScriptRepo(ctx, entClient)
	scriptLogRepo := data.NewScriptLogRepo(ctx, entClient)

	// ── AI 域（移植自 go-wind-cms）──
	aiProviderRepo := data.NewAiProviderRepo(ctx, entClient)
	aiConversationRepo := data.NewAiConversationRepo(ctx, entClient)
	aiMessageRepo := data.NewAiMessageRepo(ctx, entClient)
	aiUsageLogRepo := data.NewAiUsageLogRepo(ctx, entClient)
	aiKnowledgeRepo := data.NewAiKnowledgeRepo(ctx, entClient)
	aiAgentAppRepo := data.NewAiAgentAppRepo(ctx, entClient)
	aiAgentRunRepo := data.NewAiAgentRunRepo(ctx, entClient)

	// 评论服务依赖 configRepo（AI 反垃圾开关）与 AI 域 repos（审核外呼），故置于其后构造
	commentService := service.NewCommentService(ctx, commentRepo, configRepo, &service.CommentAIModerationDeps{
		ProviderRepo: aiProviderRepo,
		UsageLogRepo: aiUsageLogRepo,
	})

	// ── Notification 通知域（移植自 go-wind-cms）──
	notificationChannelRepo := data.NewNotificationChannelRepo(ctx, entClient)
	notificationDeliveryRepo := data.NewNotificationDeliveryRepo(ctx, entClient)
	notificationRuleRepo := data.NewNotificationRuleRepo(ctx, entClient)
	notificationTemplateRepo := data.NewNotificationTemplateRepo(ctx, entClient)
	notificationPreferenceRepo := data.NewNotificationPreferenceRepo(ctx, entClient)

	channelRegistry := channel.NewRegistry()
	channelRegistry.Register(channel.NewEmailSender(notificationChannelRepo))
	channelRegistry.Register(channel.NewWebhookSender(notificationChannelRepo))
	channelRegistry.Register(channel.NewSettingsSender(notificationChannelRepo))
	channelRegistry.Register(channel.NewSmsSender(notificationChannelRepo))
	// INTERNAL 渠道：shop 编排已落收件行，sender 仅确认送达（见 internal_channel_sender.go）
	channelRegistry.Register(channel.NewInternalChannelSender())
	notificationService := service.NewNotificationService(ctx, notificationDeliveryRepo, notificationRuleRepo, notificationTemplateRepo, channelRegistry)
	notificationChannelService := service.NewNotificationChannelService(ctx, notificationChannelRepo, notificationService)
	notificationRuleService := service.NewNotificationRuleService(ctx, notificationRuleRepo, notificationChannelRepo, notificationService)
	notificationTemplateService := service.NewNotificationTemplateService(ctx, notificationTemplateRepo)
	notificationPreferenceService := service.NewNotificationPreferenceService(ctx, notificationPreferenceRepo, internalMessageCategoryRepo)
	configService := service.NewConfigService(ctx, configRepo)
	dictTypeService := service.NewDictTypeService(ctx, dictTypeRepo)
	dictEntryService := service.NewDictEntryService(ctx, dictEntryRepo)
	monitorAlertService := service.NewMonitorAlertService(ctx, monitorAlertRuleRepo, entClient, redisClient)
	onlineSessionService := service.NewOnlineSessionService(ctx, authenticator)
	accessKeyService := service.NewAccessKeyService(ctx, accessKeyRepo, authenticator, accessKeyRateLimiter)
	serverMonitorService := service.NewServerMonitorService(ctx, entClient)
	redisCacheMonitorService := service.NewRedisCacheMonitorService(ctx, redisClient)
	monitorAlertService.RegisterNotifier(notificationService)
	internalMessageService.RegisterNotifier(notificationService)
	internalMessageRecipientService.RegisterNotifier(notificationService)
	aiAgentService := service.NewAiAgentService(ctx, aiAgentAppRepo, aiAgentRunRepo, aiProviderRepo, aiKnowledgeRepo, aiUsageLogRepo, redisClient)
	aiProviderService := service.NewAiProviderService(ctx, aiProviderRepo)
	aiChatService := service.NewAiChatService(ctx, aiConversationRepo, aiMessageRepo, aiProviderRepo, aiUsageLogRepo, aiKnowledgeRepo, configRepo, entClient)
	aiConversationService := service.NewAiConversationService(ctx, aiConversationRepo)
	aiMessageService := service.NewAiMessageService(ctx, aiMessageRepo)
	aiKnowledgeService := service.NewAiKnowledgeService(ctx, aiKnowledgeRepo, aiProviderRepo, aiUsageLogRepo, taskService)
	aiUsageLogService := service.NewAiUsageLogService(ctx, aiUsageLogRepo, aiProviderRepo)
	aiQueryService := service.NewAiQueryService(ctx, aiProviderRepo, aiUsageLogRepo, entClient, internalMessageRepo, internalMessageService)
	scriptRuntime := service.NewScriptRuntime(ctx, scriptRepo, redisClient, minIOClient, scriptLogRepo, aiProviderRepo, aiUsageLogRepo)
	scriptService := service.NewScriptService(ctx, scriptRepo, scriptRuntime)
	planService := service.NewPlanService(ctx, planRepo)
	planModuleService := service.NewPlanModuleService(ctx, planModuleRepo)
	planQuotaService := service.NewPlanQuotaService(ctx, planQuotaRepo)
	scriptLogService := service.NewScriptLogService(ctx, scriptLogRepo)

	// ── register:service ── 新模块服务在此行后注册(make register 工具锚点,勿删)
	seckillActivityService := service.NewSeckillActivityService(ctx, seckillActivityRepo, entClient)

	// ═══════════════════════ 四、传输层(internal/server) ═══════════════════════

	grpcMiddlewares := server.NewGrpcMiddleware(ctx)

	grpcServer, err := server.NewGrpcServer(ctx, grpcMiddlewares,
		authenticationService,
		loginPolicyService,
		notificationService,
		notificationChannelService,
		notificationRuleService,
		notificationTemplateService,
		notificationPreferenceService,
		configService,
		dictTypeService, dictEntryService,
		monitorAlertService,
		onlineSessionService,
		accessKeyService,
		serverMonitorService,
		redisCacheMonitorService,
		aiAgentService,
		aiProviderService, aiChatService, aiConversationService, aiMessageService,
		aiKnowledgeService, aiUsageLogService, aiQueryService,
		scriptService, scriptLogService,
		planService, planModuleService, planQuotaService,
		mfaService,
		userCredentialService,
		taskService,
		fileService,
		languageService,
		tenantService,
		userService,
		roleService,
		positionService,
		orgUnitService,
		menuService,
		apiService,
		permissionService,
		permissionGroupService,
		permissionAuditLogService,
		policyEvaluationLogService,
		loginAuditLogService,
		apiAuditLogService,
		operationAuditLogService,
		dataAccessAuditLogService,
		internalMessageService,
		internalMessageCategoryService,
		internalMessageRecipientService,
		categoryService,
		brandService,
		productService,
		productAttributeService,
		productAttributeValueService,
		skuService,
		skuPriceService,
		skuAttributeCombinationService,
		stockAlertService,
		cartService,
		cartItemService,
		wishlistService,
		orderService,
		orderItemService,
		seckillActivityService,
		paymentChannelConfigService,
		paymentTransactionService,
		paymentRefundService,
		shippingAddressService,
		shipmentService,
		couponTemplateService,
		promotionService,
		stockOpLogService,
		userCouponService,
		shippingRateService,
		taxRateService,
		commentService,
		interactionService,
		interactionAdminService,
		invoiceService,
		// register:grpc-arg ── 新模块服务实参在此行后追加(make register 工具锚点,勿删)
	)
	if err != nil {
		rollback()
		return nil, nil, err
	}

	asynqServer := server.NewAsynqServer(ctx, taskService, orderService, userCouponService, stockAlertScannerService, productSearchService, paymentTransactionService)

	return newApp(ctx, grpcServer, asynqServer), rollback, nil
}
