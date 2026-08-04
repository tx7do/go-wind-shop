import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateSkuAttributeCombinationsRequest,
  catalogservicev1_CreateSkuAttributeCombinationRequest,
  catalogservicev1_DeleteSkuAttributeCombinationRequest,
  catalogservicev1_GetSkuAttributeCombinationRequest,
  catalogservicev1_ListSkuAttributeCombinationResponse,
  catalogservicev1_SkuAttributeCombination,
  catalogservicev1_UpdateSkuAttributeCombinationRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// SKU 属性组合管理
// ==============================

export function useListSkuAttributeCombinations(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListSkuAttributeCombinationResponse, Error>
) {
  return useQuery({
    queryKey: ["listSkuAttributeCombinations", query],
    queryFn: () => apiClient.skuAttributeCombinationService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListSkuAttributeCombinations(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listSkuAttributeCombinations", params],
    queryFn: () => apiClient.skuAttributeCombinationService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetSkuAttributeCombination(
  req: catalogservicev1_GetSkuAttributeCombinationRequest,
  options?: UseQueryOptions<catalogservicev1_SkuAttributeCombination, Error>
) {
  return useQuery({
    queryKey: ["getSkuAttributeCombination", req],
    queryFn: () => apiClient.skuAttributeCombinationService.Get(req),
    ...options,
  });
}

export async function fetchGetSkuAttributeCombination(
  req: catalogservicev1_GetSkuAttributeCombinationRequest
) {
  return queryClient.fetchQuery({
    queryKey: ["getSkuAttributeCombination", req],
    queryFn: () => apiClient.skuAttributeCombinationService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateSkuAttributeCombination(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.skuAttributeCombinationService.Create({
        data: { ...values } as catalogservicev1_SkuAttributeCombination,
      }),
    ...options,
  });
}

export function useUpdateSkuAttributeCombination(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.skuAttributeCombinationService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteSkuAttributeCombination(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteSkuAttributeCombinationRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.skuAttributeCombinationService.Delete(data),
    ...options,
  });
}

export function useBatchCreateSkuAttributeCombinations(
  options?: UseMutationOptions<
    {},
    Error,
    catalogservicev1_BatchCreateSkuAttributeCombinationsRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.skuAttributeCombinationService.BatchCreate(data),
    ...options,
  });
}
