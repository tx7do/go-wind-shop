package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"go-wind-shop/app/app/service/internal/data"
	"go-wind-shop/app/app/service/internal/server"
	"go-wind-shop/app/app/service/internal/service"
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

	minIOClient := data.NewMinIoClient(ctx)

	// ═══════════════════════ 二、服务客户端(internal/data) ═══════════════════════

	userServiceClient := data.NewUserServiceClient(ctx, discovery)
	userCredentialServiceClient := data.NewUserCredentialServiceClient(ctx, discovery)
	fileServiceClient := data.NewFileServiceClient(ctx, discovery)
	tenantServiceClient := data.NewTenantServiceClient(ctx, discovery)
	orgUnitServiceClient := data.NewOrgUnitServiceClient(ctx, discovery)
	positionServiceClient := data.NewPositionServiceClient(ctx, discovery)
	roleServiceClient := data.NewRoleServiceClient(ctx, discovery)
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
	wishlistServiceClient := data.NewWishlistServiceClient(ctx, discovery)
	orderServiceClient := data.NewOrderServiceClient(ctx, discovery)
	orderItemServiceClient := data.NewOrderItemServiceClient(ctx, discovery)
	paymentTransactionServiceClient := data.NewPaymentTransactionServiceClient(ctx, discovery)
	paymentChannelConfigServiceClient := data.NewPaymentChannelConfigServiceClient(ctx, discovery)
	paymentRefundServiceClient := data.NewPaymentRefundServiceClient(ctx, discovery)
	shippingAddressServiceClient := data.NewShippingAddressServiceClient(ctx, discovery)
	shipmentServiceClient := data.NewShipmentServiceClient(ctx, discovery)
	internalMessageRecipientServiceClient := data.NewInternalMessageRecipientServiceClient(ctx, discovery)
	userCouponServiceClient := data.NewUserCouponServiceClient(ctx, discovery)
	couponTemplateServiceClient := data.NewCouponTemplateServiceClient(ctx, discovery)
	promotionServiceClient := data.NewPromotionServiceClient(ctx, discovery)
	commentServiceClient := data.NewCommentServiceClient(ctx, discovery)
	interactionServiceClient := data.NewInteractionServiceClient(ctx, discovery)
	invoiceServiceClient := data.NewInvoiceServiceClient(ctx, discovery)
	aiChatServiceClient := data.NewAiChatServiceClient(ctx, discovery)

	// ── register:client ── 新模块服务客户端在此行后注册(make register 工具锚点,勿删)
	memberTierServiceClient := data.NewMemberTierServiceClient(ctx, discovery)
	grouponServiceClient := data.NewGrouponServiceClient(ctx, discovery)
	pointsServiceClient := data.NewPointsServiceClient(ctx, discovery)
	seckillServiceClient := data.NewSeckillServiceClient(ctx, discovery)

	// ═══════════════════════ 三、服务层(internal/service) ═══════════════════════

	authenticationService := service.NewAuthenticationService(ctx, authenticationServiceClient, userServiceClient, userCredentialServiceClient, redisClient)
	fileTransferService := service.NewFileTransferService(ctx, minIOClient, fileServiceClient)
	userProfileService := service.NewUserProfileService(ctx, userServiceClient, tenantServiceClient, orgUnitServiceClient, positionServiceClient, roleServiceClient, userCredentialServiceClient)
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
	wishlistService := service.NewWishlistService(ctx, wishlistServiceClient)
	orderService := service.NewOrderService(ctx, orderServiceClient)
	orderItemService := service.NewOrderItemService(ctx, orderItemServiceClient)
	paymentTransactionService := service.NewPaymentTransactionService(ctx, paymentTransactionServiceClient, paymentChannelConfigServiceClient)
	paymentChannelService := service.NewPaymentChannelService(ctx, paymentChannelConfigServiceClient)
	paymentRefundService := service.NewPaymentRefundService(ctx, paymentRefundServiceClient)
	shippingAddressService := service.NewShippingAddressService(ctx, shippingAddressServiceClient)
	shipmentService := service.NewShipmentService(ctx, shipmentServiceClient)
	internalMessageRecipientService := service.NewInternalMessageRecipientService(ctx, internalMessageRecipientServiceClient)
	userCouponService := service.NewUserCouponService(ctx, userCouponServiceClient)
	couponTemplateService := service.NewCouponTemplateService(ctx, couponTemplateServiceClient)
	promotionService := service.NewPromotionService(ctx, promotionServiceClient)
	commentService := service.NewCommentService(ctx, commentServiceClient)
	interactionService := service.NewInteractionService(ctx, interactionServiceClient)
	invoiceService := service.NewInvoiceService(ctx, invoiceServiceClient)
	aiChatService := service.NewAiChatService(ctx, aiChatServiceClient, authenticationServiceClient)

	// ── register:service ── 新模块服务在此行后注册(make register 工具锚点,勿删)
	memberTierService := service.NewMemberTierService(ctx, memberTierServiceClient)
	grouponService := service.NewGrouponService(ctx, grouponServiceClient)
	pointsService := service.NewPointsService(ctx, pointsServiceClient)
	seckillService := service.NewSeckillService(ctx, seckillServiceClient)

	// ═══════════════════════ 四、传输层(internal/server) ═══════════════════════

	restMiddlewares := server.NewRestMiddleware(ctx, accessTokenChecker, engine)

	httpServer := server.NewRestServer(ctx, restMiddlewares,
		authenticationService,
		fileTransferService,
		userProfileService,
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
		wishlistService,
		orderService,
		orderItemService,
		paymentTransactionService,
		paymentChannelService,
		paymentRefundService,
		shippingAddressService,
		shipmentService,
		internalMessageRecipientService,
		memberTierService,
		grouponService,
		pointsService,
		seckillService,
		userCouponService,
		couponTemplateService,
		promotionService,		commentService,
		interactionService,
		invoiceService,
		aiChatService,
		// register:rest-arg ── 新模块服务实参在此行后追加(make register 工具锚点,勿删)
	)

	grpcMiddlewares := server.NewGrpcMiddleware(ctx)

	grpcServer, err := server.NewGrpcServer(ctx, grpcMiddlewares)
	if err != nil {
		rollback()
		return nil, nil, err
	}

	sseServer := server.NewSseServer(ctx, aiChatService)

	return newApp(ctx, httpServer, grpcServer, sseServer), rollback, nil
}
