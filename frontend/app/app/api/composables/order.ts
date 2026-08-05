import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { makeUpdateMask } from '@/core/transport/rest';
import { getCurrentLocale } from '@/utils/locale';
import type {
  orderservicev1_Order,
  orderservicev1_Order_Status,
  orderservicev1_CreateOrderRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 订单列表（Query）
// ==============================
export async function fetchListOrders(params: any) {
  return await apiClient.orderService.List(params);
}
export function useListOrders(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListOrders>>, Error>,
) {
  return useQuery({
    queryKey: ['listOrders', params, getCurrentLocale()],
    queryFn: () => fetchListOrders(params),
    ...options,
  });
}
export async function fetchListOrdersStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listOrders', params, getCurrentLocale()],
    queryFn: () => fetchListOrders(params),
    retry: 0,
  });
}

// ==============================
// 订单详情（Query）
// ==============================
export async function fetchGetOrder(id: number) {
  return await apiClient.orderService.Get({ id } as any);
}
export function useGetOrder(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetOrder>>, Error>,
) {
  return useQuery({
    queryKey: ['getOrder', id, getCurrentLocale()],
    queryFn: () => fetchGetOrder(id),
    ...options,
  });
}
export async function fetchGetOrderStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getOrder', id, getCurrentLocale()],
    queryFn: () => fetchGetOrder(id),
    retry: 0,
  });
}

// ==============================
// 按 idempotency_key + tenant_id 查询订单（下单后回查 orderId 用）
// ==============================
export async function fetchGetOrderByIdempotencyKey(
  idempotencyKey: string,
  tenantId: number,
) {
  return await apiClient.orderService.Get({
    idempotencyKey,
    tenantId,
  } as any);
}

// ==============================
// 创建订单 / 结算（Mutation）
// ==============================
export async function createOrder(data: orderservicev1_Order) {
  const request: orderservicev1_CreateOrderRequest = { data };
  return await apiClient.orderService.Create(request);
}
export function useCreateOrder(
  options?: UseMutationOptions<{}, Error, orderservicev1_Order>,
) {
  return useMutation({
    mutationFn: (data) => createOrder(data),
    ...options,
  });
}
export async function fetchCreateOrderStore(data: orderservicev1_Order) {
  return queryClient.fetchQuery({
    queryKey: ['createOrder', data],
    queryFn: () => createOrder(data),
    retry: 0,
  });
}

// ==============================
// 更新订单（Mutation）
// 用于买家侧的状态推进：取消订单（PENDING_PAYMENT → CANCELLED）、
// 确认收货（FULFILLED → CLOSED）等。借助乐观状态机 expectedStatus 防止终态被并发覆盖。
// ==============================
export async function updateOrder(
  id: number,
  values: Record<string, any> = {},
  expectedStatus?: orderservicev1_Order_Status[],
) {
  const updateMask = makeUpdateMask(Object.keys(values ?? {}));
  return await apiClient.orderService.Update({
    id,
    // @ts-ignore proto generated code is error.
    data: { ...values, id },
    // @ts-ignore proto generated code is error.
    updateMask,
    expectedStatus,
  });
}
export function useUpdateOrder(
  options?: UseMutationOptions<
    {},
    Error,
    { id: number; values: Record<string, any>; expectedStatus?: orderservicev1_Order_Status[] }
  >,
) {
  return useMutation({
    mutationFn: ({ id, values, expectedStatus }) =>
      updateOrder(id, values, expectedStatus),
    ...options,
  });
}
export async function fetchUpdateOrderStore(
  id: number,
  values: Record<string, any> = {},
  expectedStatus?: orderservicev1_Order_Status[],
) {
  return queryClient.fetchQuery({
    queryKey: ['updateOrder', id, values, expectedStatus],
    queryFn: () => updateOrder(id, values, expectedStatus),
    retry: 0,
  });
}

// ==============================
// 订单项列表（Query，用于订单详情）
// ==============================
export async function fetchListOrderItems(params: any) {
  return await apiClient.orderItemService.List(params);
}
export function useListOrderItems(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListOrderItems>>, Error>,
) {
  return useQuery({
    queryKey: ['listOrderItems', params, getCurrentLocale()],
    queryFn: () => fetchListOrderItems(params),
    ...options,
  });
}
export async function fetchListOrderItemsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listOrderItems', params, getCurrentLocale()],
    queryFn: () => fetchListOrderItems(params),
    retry: 0,
  });
}
