import type { RouteRecordRaw } from "vue-router";
import { Layout } from "@/layouts";

const trade: RouteRecordRaw[] = [
  {
    path: "/trade",
    name: "TradeManagement",
    component: Layout,
    redirect: "/trade/fulfillment/orders",
    meta: {
      order: 3,
      icon: "lucide:receipt",
      title: "routes.trade.moduleName",
      keepAlive: true,
      authority: ["sys:platform_admin"],
    },
    children: [
      {
        path: "fulfillment",
        name: "TradeOrderFulfillmentGroup",
        meta: {
          order: 1,
          icon: "lucide:package-check",
          title: "routes.trade.orderFulfillment",
          authority: ["sys:platform_admin"],
        },
        children: [
          {
            path: "orders",
            name: "TradeOrderManagement",
            meta: {
              order: 1,
              icon: "lucide:clipboard-list",
              title: "routes.trade.order",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/order/index.vue"),
          },
          {
            path: "shipment",
            name: "TradeShipmentManagement",
            meta: {
              order: 2,
              icon: "lucide:truck",
              title: "routes.trade.shipment",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/shipment/index.vue"),
          },
          {
            path: "refunds",
            name: "TradeRefundManagement",
            meta: {
              order: 3,
              icon: "lucide:rotate-ccw",
              title: "routes.trade.refund",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/refund/index.vue"),
          },
          {
            path: "invoices",
            name: "TradeInvoiceManagement",
            meta: {
              order: 4,
              icon: "lucide:file-text",
              title: "routes.trade.invoice",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/invoice/index.vue"),
          },
        ],
      },
      {
        path: "finance",
        name: "TradePaymentFinanceGroup",
        meta: {
          order: 2,
          icon: "lucide:wallet",
          title: "routes.trade.paymentFinance",
          authority: ["sys:platform_admin"],
        },
        children: [
          {
            path: "payment-channels",
            name: "TradePaymentChannelManagement",
            meta: {
              order: 1,
              icon: "lucide:credit-card",
              title: "routes.trade.paymentChannel",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/payment-channel/index.vue"),
          },
          {
            path: "payment-transactions",
            name: "TradePaymentTransactionManagement",
            meta: {
              order: 2,
              icon: "lucide:credit-card",
              title: "routes.trade.paymentTransaction",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/payment-transaction/index.vue"),
          },
          {
            path: "finance-overview",
            name: "TradeFinanceOverview",
            meta: {
              order: 3,
              icon: "lucide:area-chart",
              title: "routes.trade.financeOverview",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/finance-overview/index.vue"),
          },
          {
            path: "tax-rates",
            name: "TradeTaxRateManagement",
            meta: {
              order: 4,
              icon: "lucide:percent",
              title: "routes.trade.taxRate",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/tax-rate/index.vue"),
          },
        ],
      },
      {
        path: "marketing",
        name: "TradeMarketingMemberGroup",
        meta: {
          order: 3,
          icon: "lucide:megaphone",
          title: "routes.trade.marketingMember",
          authority: ["sys:platform_admin"],
        },
        children: [
          {
            path: "promotions",
            name: "TradePromotionManagement",
            meta: {
              order: 1,
              icon: "lucide:badge-percent",
              title: "routes.trade.promotion",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/promotion/index.vue"),
          },
          {
            path: "seckills",
            name: "TradeSeckillManagement",
            meta: {
              order: 2,
              icon: "lucide:zap",
              title: "routes.trade.seckill",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/seckill/index.vue"),
          },
          {
            path: "groupons",
            name: "TradeGrouponManagement",
            meta: {
              order: 3,
              icon: "lucide:users",
              title: "routes.trade.groupon",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/groupon/index.vue"),
          },
          {
            path: "coupon-templates",
            name: "TradeCouponTemplateManagement",
            meta: {
              order: 4,
              icon: "lucide:ticket",
              title: "routes.trade.couponTemplate",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/coupon-template/index.vue"),
          },
          {
            path: "user-coupons",
            name: "TradeUserCouponManagement",
            meta: {
              order: 5,
              icon: "lucide:ticket-percent",
              title: "routes.trade.userCoupon",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/user-coupon/index.vue"),
          },
          {
            path: "points-ledgers",
            name: "TradePointsLedgerManagement",
            meta: {
              order: 6,
              icon: "lucide:star",
              title: "routes.trade.pointsLedger",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/points-ledger/index.vue"),
          },
          {
            path: "member-tiers",
            name: "TradeMemberTierManagement",
            meta: {
              order: 7,
              icon: "lucide:crown",
              title: "routes.trade.memberTier",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/member-tiers/index.vue"),
          },
        ],
      },
      {
        path: "inventory",
        name: "TradeInventoryLogisticsGroup",
        meta: {
          order: 4,
          icon: "lucide:warehouse",
          title: "routes.trade.inventoryLogistics",
          authority: ["sys:platform_admin"],
        },
        children: [
          {
            path: "low-stock",
            name: "TradeLowStockManagement",
            meta: {
              order: 1,
              icon: "lucide:alert-triangle",
              title: "routes.trade.lowStock",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/low-stock/index.vue"),
          },
          {
            path: "stock-alerts",
            name: "TradeStockAlertManagement",
            meta: {
              order: 2,
              icon: "lucide:shield-alert",
              title: "routes.trade.stockAlert",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/stock-alert/index.vue"),
          },
          {
            path: "stock-op-logs",
            name: "TradeStockOpLogManagement",
            meta: {
              order: 3,
              icon: "lucide:scroll-text",
              title: "routes.trade.stockOpLog",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/stock-op-log/index.vue"),
          },
          {
            path: "shipping-rates",
            name: "TradeShippingRateManagement",
            meta: {
              order: 4,
              icon: "lucide:package",
              title: "routes.trade.shippingRate",
              authority: ["sys:platform_admin"],
            },
            component: () => import("@/pages/app/trade/shipping-rate/index.vue"),
          },
        ],
      },
    ],
  },
];

export default trade;
