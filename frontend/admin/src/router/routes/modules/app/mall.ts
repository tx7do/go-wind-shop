import type { RouteRecordRaw } from "vue-router";
import { Layout } from "@/layouts";

const mall: RouteRecordRaw[] = [
  {
    path: "/mall",
    name: "MallManagement",
    component: Layout,
    redirect: "/mall/categories",
    meta: {
      order: 2006,
      icon: "lucide:shopping-cart",
      title: "routes.mall.moduleName",
      keepAlive: true,
      authority: ["sys:platform_admin"],
    },
    children: [
      {
        path: "categories",
        name: "MallCategoryManagement",
        meta: {
          order: 1,
          icon: "lucide:folder-tree",
          title: "routes.mall.category",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/mall/category/index.vue"),
      },
      {
        path: "brands",
        name: "MallBrandManagement",
        meta: {
          order: 2,
          icon: "lucide:tag",
          title: "routes.mall.brand",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/mall/brand/index.vue"),
      },
      {
        path: "products",
        name: "MallProductManagement",
        meta: {
          order: 3,
          icon: "lucide:package",
          title: "routes.mall.product",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/mall/product/index.vue"),
      },
      {
        path: "product-attributes",
        name: "MallProductAttributeManagement",
        meta: {
          order: 4,
          icon: "lucide:list",
          title: "routes.mall.productAttribute",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/mall/product-attribute/index.vue"),
      },
      {
        path: "product-attribute-values",
        name: "MallProductAttributeValueManagement",
        meta: {
          order: 5,
          icon: "lucide:list-checks",
          title: "routes.mall.productAttributeValue",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/mall/product-attribute-value/index.vue"),
      },
      {
        path: "skus",
        name: "MallSkuManagement",
        meta: {
          order: 6,
          icon: "lucide:barcode",
          title: "routes.mall.sku",
          authority: ["sys:platform_admin"],
        },
        component: () => import("@/pages/app/mall/sku/index.vue"),
      },
    ],
  },
];

export default mall;
