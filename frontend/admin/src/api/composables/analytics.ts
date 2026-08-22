import { PaginationQuery } from "@/core/transport/rest";
import { fetchListBrands } from "./brand";
import { fetchListCategories } from "./category";
import { fetchListProducts } from "./product";
import { fetchListOrders } from "./order";
import { fetchListPaymentTransactions } from "./payment-transaction";
import { fetchListPaymentRefunds } from "./payment-refund";

// ==============================
// 商城分析 (Analytics)
// ==============================
//
// 说明：
// 后端目前没有专用的聚合/统计 RPC，但每个实体的 List RPC 在服务端
// (go-crud/entgo ListWithPaging) 会把 PagingRequest.query 中的过滤条件
// 构造成 SQL WHERE，并应用到 countBuilder 上。因此 List*Response.total
// 反映的是“符合过滤条件的行数”，与当前页 items 无关。
//
// 利用这一特性，对订单/支付流水按 status / payment_method 分别发起带
// 过滤的 List 请求（pageSize=1，仅取 total），即可在无聚合端点的情况下
// 拼出分布数据。每次调用都会真实命中数据库计数，不缓存（全局
// staleTime:0 / gcTime:0，每次渲染都重新请求）。

/**
 * 订单状态枚举值（与 order.service.v1.Order.Status 的 JSON 名对应）。
 * 用于按状态发起计数请求。
 */
export const ORDER_STATUS_VALUES = [
  "PENDING_PAYMENT",
  "PAID",
  "CANCELLED",
  "FULFILLED",
  "CLOSED",
] as const;

/**
 * 支付方式枚举值（与 payment.service.v1.PaymentMethod 的 JSON 名对应）。
 * 仅枚举真实可能出现的渠道；UNSPECIFIED 不会被统计。
 */
export const PAYMENT_METHOD_VALUES = ["ALIPAY", "WECHAT"] as const;

export type DistributionEntry = { key: string; count: number };

/**
 * 按订单状态并发拉取计数分布。
 * 对每个状态发起一次 pageSize=1 的 List 请求，仅读取 total。
 */
export async function fetchOrderStatusCounts(): Promise<DistributionEntry[]> {
  const results = await Promise.all(
    ORDER_STATUS_VALUES.map(async (status) => {
      const resp = await fetchListOrders(
        new PaginationQuery({
          paging: { page: 1, pageSize: 1 },
          formValues: { status },
        })
      );
      return {
        key: status,
        count: Number(resp?.total ?? 0),
      } satisfies DistributionEntry;
    })
  );
  return results;
}

/**
 * 按支付方式并发拉取计数分布。
 */
export async function fetchPaymentMethodCounts(): Promise<DistributionEntry[]> {
  const results = await Promise.all(
    PAYMENT_METHOD_VALUES.map(async (method) => {
      const resp = await fetchListPaymentTransactions(
        new PaginationQuery({
          paging: { page: 1, pageSize: 1 },
          formValues: { payment_method: method },
        })
      );
      return {
        key: method,
        count: Number(resp?.total ?? 0),
      } satisfies DistributionEntry;
    })
  );
  return results;
}

export type CatalogTotals = {
  products: number;
  brands: number;
  categories: number;
};

/**
 * 拉取商品目录三类实体的总计数（无过滤，total 即全量）。
 */
export async function fetchCatalogTotals(): Promise<CatalogTotals> {
  const [products, brands, categories] = await Promise.all([
    fetchListProducts(new PaginationQuery({ paging: { page: 1, pageSize: 1 } })),
    fetchListBrands(new PaginationQuery({ paging: { page: 1, pageSize: 1 } })),
    fetchListCategories(new PaginationQuery({ paging: { page: 1, pageSize: 1 } })),
  ]);
  return {
    products: Number(products?.total ?? 0),
    brands: Number(brands?.total ?? 0),
    categories: Number(categories?.total ?? 0),
  };
}

export type RecentOrdersResult = {
  items: unknown[];
  total: number;
};

/**
 * 拉取最近订单（按创建时间倒序），用于首页最近订单表格。
 * 这是只读、非分页表格，仅取首页 limit 条。
 */
export async function fetchRecentOrders(limit: number): Promise<RecentOrdersResult> {
  const resp = await fetchListOrders(
    new PaginationQuery({
      paging: { page: 1, pageSize: limit },
      orderBy: ["-created_at"],
    })
  );
  return {
    items: (resp?.items ?? []) as unknown[],
    total: Number(resp?.total ?? 0),
  };
}

/**
 * 订单总数（全量，无过滤）。
 */
export async function fetchOrderTotal(): Promise<number> {
  const resp = await fetchListOrders(new PaginationQuery({ paging: { page: 1, pageSize: 1 } }));
  return Number(resp?.total ?? 0);
}

/**
 * 支付流水总数（全量，无过滤）。
 */
export async function fetchPaymentTransactionTotal(): Promise<number> {
  const resp = await fetchListPaymentTransactions(
    new PaginationQuery({ paging: { page: 1, pageSize: 1 } })
  );
  return Number(resp?.total ?? 0);
}

/**
 * 退款状态枚举值（与 payment.service.v1.PaymentRefund.Status 的 JSON 名对应）。
 * 仅枚举业务态（PENDING/SUCCEEDED/FAILED），STATUS_UNSPECIFIED 不统计。
 */
export const REFUND_STATUS_VALUES = ["PENDING", "SUCCEEDED", "FAILED"] as const;

/**
 * 按退款状态并发拉取计数分布。
 * 对每个状态发起一次 pageSize=1 的 List 请求，仅读取 total。
 */
export async function fetchRefundStatusCounts(): Promise<DistributionEntry[]> {
  const results = await Promise.all(
    REFUND_STATUS_VALUES.map(async (status) => {
      const resp = await fetchListPaymentRefunds(
        new PaginationQuery({
          paging: { page: 1, pageSize: 1 },
          formValues: { status },
        })
      );
      return {
        key: status,
        count: Number(resp?.total ?? 0),
      } satisfies DistributionEntry;
    })
  );
  return results;
}

/**
 * 退款总数（全量，无过滤）。
 */
export async function fetchRefundTotal(): Promise<number> {
  const resp = await fetchListPaymentRefunds(
    new PaginationQuery({ paging: { page: 1, pageSize: 1 } })
  );
  return Number(resp?.total ?? 0);
}

// 上述 fetch* 函数均通过 queryClient.fetchQuery 间接调用（见各
// useList*/fetchList* 实现），此处直接复用，无需在此重复注册 queryKey。
// 调用方在 useQuery 中包裹这些函数以获得响应式（全局
// staleTime:0 / gcTime:0，无缓存）。
