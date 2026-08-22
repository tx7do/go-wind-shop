import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_GetStockAlertRequest,
  catalogservicev1_ListStockAlertResponse,
  catalogservicev1_StockAlert,
  catalogservicev1_UpdateStockAlertRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 库存预警记录管理（裁剪：仅 List/Get/Update，Update 唯一允许标记 RESOLVED）
// ==============================

export function useListStockAlerts(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListStockAlertResponse, Error>
) {
  return useQuery({
    queryKey: ["listStockAlerts", query],
    queryFn: () => apiClient.stockAlertService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListStockAlerts(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listStockAlerts", params],
    queryFn: () => apiClient.stockAlertService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetStockAlert(
  req: catalogservicev1_GetStockAlertRequest,
  options?: UseQueryOptions<catalogservicev1_StockAlert, Error>
) {
  return useQuery({
    queryKey: ["getStockAlert", req],
    queryFn: () => apiClient.stockAlertService.Get(req),
    ...options,
  });
}

export async function fetchGetStockAlert(req: catalogservicev1_GetStockAlertRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getStockAlert", req],
    queryFn: () => apiClient.stockAlertService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useUpdateStockAlert(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.stockAlertService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}
