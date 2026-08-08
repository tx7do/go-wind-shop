import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  shippingservicev1_CreateShipmentRequest,
  shippingservicev1_DeleteShipmentRequest,
  shippingservicev1_GetShipmentRequest,
  shippingservicev1_ListShipmentResponse,
  shippingservicev1_Shipment,
  shippingservicev1_Shipment_Status,
  shippingservicev1_UpdateShipmentRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 物流单管理
// ==============================
//
// 物流单为独立实体，状态机：
//   PENDING（创建初始态） → SHIPPED（发货，后端同事务推进关联订单 PAID→FULFILLED） → DELIVERED（终态）
//
// Update 调用必须透传 expectedStatus（状态机前置条件），否则后端按 affected_rows=0 返回 Conflict。

export function useListShipments(
  query: PaginationQuery,
  options?: UseQueryOptions<shippingservicev1_ListShipmentResponse, Error>
) {
  return useQuery({
    queryKey: ["listShipments", query],
    queryFn: () => apiClient.shipmentService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListShipments(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listShipments", params],
    queryFn: () => apiClient.shipmentService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetShipment(
  req: shippingservicev1_GetShipmentRequest,
  options?: UseQueryOptions<shippingservicev1_Shipment, Error>
) {
  return useQuery({
    queryKey: ["getShipment", req],
    queryFn: () => apiClient.shipmentService.Get(req),
    ...options,
  });
}

export async function fetchGetShipment(req: shippingservicev1_GetShipmentRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getShipment", req],
    queryFn: () => apiClient.shipmentService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateShipment(
  options?: UseMutationOptions<{}, Error, { data: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ data }: { data: Record<string, any> }) =>
      apiClient.shipmentService.Create({
        data,
      } as shippingservicev1_CreateShipmentRequest),
    ...options,
  });
}

export function useUpdateShipment(
  options?: UseMutationOptions<
    {},
    Error,
    {
      id: number;
      values: Record<string, any>;
      expectedStatus: shippingservicev1_Shipment_Status[];
    }
  >
) {
  return useMutation({
    mutationFn: ({ id, values, expectedStatus }) =>
      apiClient.shipmentService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
        expectedStatus,
      } as shippingservicev1_UpdateShipmentRequest),
    ...options,
  });
}

export function useDeleteShipment(
  options?: UseMutationOptions<{}, Error, shippingservicev1_DeleteShipmentRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.shipmentService.Delete(data),
    ...options,
  });
}
