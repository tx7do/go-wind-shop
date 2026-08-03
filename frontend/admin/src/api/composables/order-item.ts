import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  orderservicev1_BatchCreateOrderItemsRequest,
  orderservicev1_CreateOrderItemRequest,
  orderservicev1_DeleteOrderItemRequest,
  orderservicev1_GetOrderItemRequest,
  orderservicev1_ListOrderItemResponse,
  orderservicev1_OrderItem,
  orderservicev1_UpdateOrderItemRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 订单项管理
// ==============================

export function useListOrderItems(
  query: PaginationQuery,
  options?: UseQueryOptions<orderservicev1_ListOrderItemResponse, Error>
) {
  return useQuery({
    queryKey: ["listOrderItems", query],
    queryFn: () => apiClient.orderItemService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListOrderItems(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listOrderItems", params],
    queryFn: () => apiClient.orderItemService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetOrderItem(
  req: orderservicev1_GetOrderItemRequest,
  options?: UseQueryOptions<orderservicev1_OrderItem, Error>
) {
  return useQuery({
    queryKey: ["getOrderItem", req],
    queryFn: () => apiClient.orderItemService.Get(req),
    ...options,
  });
}

export async function fetchGetOrderItem(req: orderservicev1_GetOrderItemRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getOrderItem", req],
    queryFn: () => apiClient.orderItemService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateOrderItem(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.orderItemService.Create({ data: { ...values } as orderservicev1_OrderItem }),
    ...options,
  });
}

export function useUpdateOrderItem(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.orderItemService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as orderservicev1_UpdateOrderItemRequest),
    ...options,
  });
}

export function useDeleteOrderItem(
  options?: UseMutationOptions<{}, Error, orderservicev1_DeleteOrderItemRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.orderItemService.Delete(data),
    ...options,
  });
}

export function useBatchCreateOrderItems(
  options?: UseMutationOptions<{}, Error, orderservicev1_BatchCreateOrderItemsRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.orderItemService.BatchCreate(data),
    ...options,
  });
}
