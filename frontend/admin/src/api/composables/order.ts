import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  orderservicev1_DeleteOrderRequest,
  orderservicev1_GetOrderRequest,
  orderservicev1_ListOrderResponse,
  orderservicev1_Order,
  orderservicev1_UpdateOrderRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 订单管理
// ==============================

export function useListOrders(
  query: PaginationQuery,
  options?: UseQueryOptions<orderservicev1_ListOrderResponse, Error>
) {
  return useQuery({
    queryKey: ["listOrders", query],
    queryFn: () => apiClient.orderService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListOrders(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listOrders", params],
    queryFn: () => apiClient.orderService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetOrder(
  req: orderservicev1_GetOrderRequest,
  options?: UseQueryOptions<orderservicev1_Order, Error>
) {
  return useQuery({
    queryKey: ["getOrder", req],
    queryFn: () => apiClient.orderService.Get(req),
    ...options,
  });
}

export async function fetchGetOrder(req: orderservicev1_GetOrderRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getOrder", req],
    queryFn: () => apiClient.orderService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useUpdateOrder(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.orderService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as orderservicev1_UpdateOrderRequest),
    ...options,
  });
}

export function useDeleteOrder(
  options?: UseMutationOptions<{}, Error, orderservicev1_DeleteOrderRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.orderService.Delete(data),
    ...options,
  });
}
