import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateSkuPricesRequest,
  catalogservicev1_CreateSkuPriceRequest,
  catalogservicev1_DeleteSkuPriceRequest,
  catalogservicev1_GetSkuPriceRequest,
  catalogservicev1_ListSkuPriceResponse,
  catalogservicev1_SkuPrice,
  catalogservicev1_UpdateSkuPriceRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// SKU 价格管理
// ==============================

export function useListSkuPrices(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListSkuPriceResponse, Error>
) {
  return useQuery({
    queryKey: ["listSkuPrices", query],
    queryFn: () => apiClient.skuPriceService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListSkuPrices(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listSkuPrices", params],
    queryFn: () => apiClient.skuPriceService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetSkuPrice(
  req: catalogservicev1_GetSkuPriceRequest,
  options?: UseQueryOptions<catalogservicev1_SkuPrice, Error>
) {
  return useQuery({
    queryKey: ["getSkuPrice", req],
    queryFn: () => apiClient.skuPriceService.Get(req),
    ...options,
  });
}

export async function fetchGetSkuPrice(req: catalogservicev1_GetSkuPriceRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getSkuPrice", req],
    queryFn: () => apiClient.skuPriceService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateSkuPrice(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.skuPriceService.Create({ data: { ...values } as catalogservicev1_SkuPrice }),
    ...options,
  });
}

export function useUpdateSkuPrice(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.skuPriceService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteSkuPrice(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteSkuPriceRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.skuPriceService.Delete(data),
    ...options,
  });
}

export function useBatchCreateSkuPrices(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateSkuPricesRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.skuPriceService.BatchCreate(data),
    ...options,
  });
}
