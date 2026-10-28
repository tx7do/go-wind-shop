import {
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_GetStockOpLogRequest,
  catalogservicev1_ListStockOpLogResponse,
  catalogservicev1_StockOpLog,
} from "@/api/generated/admin/service/v1";
import { PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 库存操作流水（append-only 审计，只读）
// ==============================

export function useListStockOpLogs(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListStockOpLogResponse, Error>
) {
  return useQuery({
    queryKey: ["listStockOpLogs", query],
    queryFn: () => apiClient.stockOpLogService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListStockOpLogs(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listStockOpLogs", params],
    queryFn: () => apiClient.stockOpLogService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetStockOpLog(
  req: catalogservicev1_GetStockOpLogRequest,
  options?: UseQueryOptions<catalogservicev1_StockOpLog, Error>
) {
  return useQuery({
    queryKey: ["getStockOpLog", req],
    queryFn: () => apiClient.stockOpLogService.Get(req),
    ...options,
  });
}
