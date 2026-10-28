import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_GetPromotionRequest,
  catalogservicev1_ListPromotionResponse,
  catalogservicev1_Promotion,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 促销活动管理（全 CRUD）
// ==============================

export function useListPromotions(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListPromotionResponse, Error>
) {
  return useQuery({
    queryKey: ["listPromotions", query],
    queryFn: () => apiClient.promotionService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListPromotions(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listPromotions", params],
    queryFn: () => apiClient.promotionService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetPromotion(
  req: catalogservicev1_GetPromotionRequest,
  options?: UseQueryOptions<catalogservicev1_Promotion, Error>
) {
  return useQuery({
    queryKey: ["getPromotion", req],
    queryFn: () => apiClient.promotionService.Get(req),
    ...options,
  });
}

export async function fetchGetPromotion(req: catalogservicev1_GetPromotionRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getPromotion", req],
    queryFn: () => apiClient.promotionService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreatePromotion(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values: Record<string, any>) =>
      apiClient.promotionService.Create({ data: values as any }),
    ...options,
  });
}

export function useUpdatePromotion(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.promotionService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeletePromotion(
  options?: UseMutationOptions<{}, Error, { id: number }>
) {
  return useMutation({
    mutationFn: ({ id }: { id: number }) =>
      apiClient.promotionService.Delete({ id }),
    ...options,
  });
}
