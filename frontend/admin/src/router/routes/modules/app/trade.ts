import type { RouteRecordRaw } from "vue-router";
import { Layout } from "@/layouts";

const trade: RouteRecordRaw[] = [
  {
    path: "/trade",
    name: "TradeManagement",
    component: Layout,
    redirect: "/trade/orders",
    meta: {
      order: 2007,
      icon: "lucide:receipt",
      title: "routes.trade.moduleName",
      keepAlive: true,
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
        path: "finance-overview",
        name: "TradeFinanceOverview",
        meta: {
          order: 4,
          icon: "lucide:area-chart",
          title: "routes.trade.financeOverview",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/trade/finance-overview/index.vue"),
      },
      {
        path: "shipment",
        name: "TradeShipmentManagement",
        meta: {
          order: 5,
          icon: "lucide:truck",
          title: "routes.trade.shipment",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/trade/shipment/index.vue"),
      },
      {
        path: "shipping-rates",
        name: "TradeShippingRateManagement",
        meta: {
          order: 8,
          icon: "lucide:package",
          title: "routes.trade.shippingRate",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/trade/shipping-rate/index.vue"),
      },
      {
        path: "tax-rates",
        name: "TradeTaxRateManagement",
        meta: {
          order: 9,
          icon: "lucide:percent",
          title: "routes.trade.taxRate",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/trade/tax-rate/index.vue"),
      },
      {
        path: "coupon-templates",
        name: "TradeCouponTemplateManagement",
        meta: {
          order: 6,
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
          order: 7,
          icon: "lucide:ticket-percent",
          title: "routes.trade.userCoupon",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/trade/user-coupon/index.vue"),
      },
      {
        path: "comments",
        name: "TradeCommentManagement",
        meta: {
          order: 10,
          icon: "lucide:message-square",
          title: "routes.trade.comment",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/trade/comment/index.vue"),
      },
    ],
  },
];

export default trade;
