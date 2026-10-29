package server

import (
	"go-wind-shop/pkg/logx"
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"

	authzEngine "github.com/tx7do/kratos-authz/engine"
	authz "github.com/tx7do/kratos-authz/middleware"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-bootstrap/rpc"

	swaggerUI "github.com/tx7do/kratos-swagger-ui"

	"go-wind-shop/app/app/service/cmd/server/assets"
	"go-wind-shop/app/app/service/internal/service"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	auditV1 "go-wind-shop/api/gen/go/audit/service/v1"

	"go-wind-shop/pkg/middleware/auth"
	applogging "go-wind-shop/pkg/middleware/logging"
)

// NewRestMiddleware 创建中间件
func NewRestMiddleware(
	ctx *bootstrap.Context,
	accessTokenChecker auth.AccessTokenChecker,
	authorizer authzEngine.Engine,
) []middleware.Middleware {
	var ms []middleware.Middleware
	ms = append(ms, logging.Server(logx.AsKratos(ctx.GetLogger())))

	// add white list for authentication.
	rpc.AddWhiteList(
		appV1.OperationAuthenticationServiceLogin,
		appV1.OperationAuthenticationServiceRegister,
		appV1.OperationAuthenticationServiceSendResetCode,
		appV1.OperationAuthenticationServiceResetPassword,

		appV1.OperationCategoryServiceList,
		appV1.OperationCategoryServiceGet,
		appV1.OperationBrandServiceList,
		appV1.OperationBrandServiceGet,
		appV1.OperationProductServiceList,
		appV1.OperationProductServiceGet,
		appV1.OperationProductServiceSearchProducts,
		appV1.OperationProductServiceSearchProductSuggestions,
		appV1.OperationPromotionServiceListActivePromotions,
		appV1.OperationSeckillServiceListActiveSeckills,
		appV1.OperationProductAttributeServiceList,
		appV1.OperationProductAttributeServiceGet,
		appV1.OperationProductAttributeValueServiceList,
		appV1.OperationProductAttributeValueServiceGet,
		appV1.OperationSkuServiceList,
		appV1.OperationSkuServiceGet,
		appV1.OperationSkuPriceServiceList,
		appV1.OperationSkuPriceServiceGet,
		appV1.OperationSkuAttributeCombinationServiceList,
		appV1.OperationSkuAttributeCombinationServiceGet,

		appV1.OperationInteractionServiceGetCounts,

		appV1.OperationCommentServiceGetProductRating,
		appV1.OperationCommentServiceGetProductReviewSummary,
		appV1.OperationAiChatServiceRewriteSearchQuery,
		appV1.OperationCouponTemplateServiceList,

		// 支付网关异步回调（服务端到服务端，无登录态；凭渠道验签保护）。
		// kratos 裸路由的 operation = 路由模板（见 kratos http server）。
		"app/v1/payments/notify/{channel}",
	)

	ms = append(ms, applogging.Server(
		applogging.WithWriteApiLogFunc(func(ctx context.Context, data *auditV1.ApiAuditLog) error {
			return nil
		}),
		applogging.WithWriteLoginLogFunc(func(ctx context.Context, data *auditV1.LoginAuditLog) error {
			return nil
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

	authenticationService *service.AuthenticationService,
	fileTransferService *service.FileTransferService,
	userProfileService *service.UserProfileService,

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
	wishlistService *service.WishlistService,
	orderService *service.OrderService,
	orderItemService *service.OrderItemService,
	paymentTransactionService *service.PaymentTransactionService,
	paymentChannelService *service.PaymentChannelService,
	paymentRefundService *service.PaymentRefundService,
	shippingAddressService *service.ShippingAddressService,
	shipmentService *service.ShipmentService,
	internalMessageRecipientService *service.InternalMessageRecipientService,

	// register:rest-param ── 新模块服务形参在此行后追加(make register 工具锚点,勿删)
	seckillService *service.SeckillService,
		userCouponService *service.UserCouponService,
		couponTemplateService *service.CouponTemplateService,
		promotionService *service.PromotionService,
		commentService *service.CommentService,
		interactionService *service.InteractionService,
		invoiceService *service.InvoiceService,
		aiChatService *service.AiChatService,
	) *http.Server {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Rest == nil {
		return nil
	}

	srv, err := rpc.CreateRestServer(cfg, middlewares...)
	if err != nil {
		panic(err)
	}

	appV1.RegisterAuthenticationServiceHTTPServer(srv, authenticationService)
	appV1.RegisterFileTransferServiceHTTPServer(srv, fileTransferService)
	appV1.RegisterUserProfileServiceHTTPServer(srv, userProfileService)

	appV1.RegisterCategoryServiceHTTPServer(srv, categoryService)
	appV1.RegisterBrandServiceHTTPServer(srv, brandService)
	appV1.RegisterProductServiceHTTPServer(srv, productService)
	appV1.RegisterProductAttributeServiceHTTPServer(srv, productAttributeService)
	appV1.RegisterProductAttributeValueServiceHTTPServer(srv, productAttributeValueService)
	appV1.RegisterSkuServiceHTTPServer(srv, skuService)
	appV1.RegisterSkuPriceServiceHTTPServer(srv, skuPriceService)
	appV1.RegisterSkuAttributeCombinationServiceHTTPServer(srv, skuAttributeCombinationService)

	appV1.RegisterCartServiceHTTPServer(srv, cartService)
	appV1.RegisterCartItemServiceHTTPServer(srv, cartItemService)
	appV1.RegisterWishlistServiceHTTPServer(srv, wishlistService)
	appV1.RegisterOrderServiceHTTPServer(srv, orderService)
	appV1.RegisterOrderItemServiceHTTPServer(srv, orderItemService)
	appV1.RegisterPaymentTransactionServiceHTTPServer(srv, paymentTransactionService)
	appV1.RegisterPaymentChannelServiceHTTPServer(srv, paymentChannelService)

	// 支付网关异步回调：手动注册（渠道表单/JSON 报文 + 验签应答口径，
	// proto 生成路由不适用）。鉴权白名单放行，安全由渠道验签保证。
	srv.Route("/").POST("/app/v1/payments/notify/{channel}", func(ctx http.Context) error {
		status, body := paymentTransactionService.HandleNotify(
			ctx, ctx.Vars().Get("channel"), ctx.Request(),
		)
		return ctx.String(status, body)
	})
	appV1.RegisterPaymentRefundServiceHTTPServer(srv, paymentRefundService)
	appV1.RegisterShippingAddressServiceHTTPServer(srv, shippingAddressService)
	appV1.RegisterShipmentServiceHTTPServer(srv, shipmentService)
	appV1.RegisterInternalMessageRecipientServiceHTTPServer(srv, internalMessageRecipientService)
	appV1.RegisterUserCouponServiceHTTPServer(srv, userCouponService)
	appV1.RegisterCouponTemplateServiceHTTPServer(srv, couponTemplateService)
	appV1.RegisterPromotionServiceHTTPServer(srv, promotionService)
	appV1.RegisterCommentServiceHTTPServer(srv, commentService)
	appV1.RegisterInteractionServiceHTTPServer(srv, interactionService)
	appV1.RegisterInvoiceServiceHTTPServer(srv, invoiceService)
	appV1.RegisterAiChatServiceHTTPServer(srv, aiChatService)

	// register:rest-route ── 新模块服务路由在此行后注册(make register 工具锚点,勿删)
	appV1.RegisterSeckillServiceHTTPServer(srv, seckillService)
	if cfg.GetServer().GetRest().GetEnableSwagger() {
		swaggerUI.RegisterSwaggerUIServerWithOption(
			srv,
			swaggerUI.WithTitle("GoWind Content Hub App API"),
			swaggerUI.WithMemoryData(assets.OpenApiData, "yaml"),
		)
	}

	return srv
}
