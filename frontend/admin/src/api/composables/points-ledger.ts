import {
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_ListPointsLedgerResponse,
} from "@/api/generated/admin/service/v1";
import { type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 积分台账管理（admin 只读）
// ==============================
export function useListPointsLedgers(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListPointsLedgerResponse, Error>
) {
  return useQuery({
    queryKey: ["listPointsLedgers", query],
    queryFn: () => apiClient.pointsLedgerService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListPointsLedgers(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listPointsLedgers", params],
    queryFn: () => apiClient.pointsLedgerService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

