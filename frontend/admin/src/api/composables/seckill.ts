import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_CreateSeckillActivityRequest,
  catalogservicev1_DeleteSeckillActivityRequest,
  catalogservicev1_GetSeckillActivityRequest,
  catalogservicev1_ListSeckillActivityResponse,
  catalogservicev1_SeckillActivity,
  catalogservicev1_UpdateSeckillActivityRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 秒杀活动管理（全 CRUD）
// ==============================

export function useListSeckillActivities(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListSeckillActivityResponse, Error>
) {
  return useQuery({
    queryKey: ["listSeckillActivities", query],
    queryFn: () => apiClient.seckillActivityService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListSeckillActivities(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listSeckillActivities", params],
    queryFn: () => apiClient.seckillActivityService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateSeckillActivity(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.seckillActivityService.Create({
        data: { ...values } as catalogservicev1_SeckillActivity,
      } as catalogservicev1_CreateSeckillActivityRequest),
    ...options,
  });
}

export function useUpdateSeckillActivity(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.seckillActivityService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as catalogservicev1_UpdateSeckillActivityRequest),
    ...options,
  });
}

export function useDeleteSeckillActivity(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteSeckillActivityRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.seckillActivityService.Delete(data),
    ...options,
  });
}
