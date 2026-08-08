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
    ],
  },
];

export default trade;
