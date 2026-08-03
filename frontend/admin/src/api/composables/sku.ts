import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateSkusRequest,
  catalogservicev1_CreateSkuRequest,
  catalogservicev1_DeleteSkuRequest,
  catalogservicev1_GetSkuRequest,
  catalogservicev1_ListSkuResponse,
  catalogservicev1_Sku,
  catalogservicev1_UpdateSkuRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// SKU 管理
// ==============================

export function useListSkus(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListSkuResponse, Error>
) {
  return useQuery({
    queryKey: ["listSkus", query],
    queryFn: () => apiClient.skuService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListSkus(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listSkus", params],
    queryFn: () => apiClient.skuService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetSku(
  req: catalogservicev1_GetSkuRequest,
  options?: UseQueryOptions<catalogservicev1_Sku, Error>
) {
  return useQuery({
    queryKey: ["getSku", req],
    queryFn: () => apiClient.skuService.Get(req),
    ...options,
  });
}

export async function fetchGetSku(req: catalogservicev1_GetSkuRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getSku", req],
    queryFn: () => apiClient.skuService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateSku(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.skuService.Create({ data: { ...values } as catalogservicev1_Sku }),
    ...options,
  });
}

export function useUpdateSku(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.skuService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteSku(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteSkuRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.skuService.Delete(data),
    ...options,
  });
}

export function useBatchCreateSkus(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateSkusRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.skuService.BatchCreate(data),
    ...options,
  });
}
