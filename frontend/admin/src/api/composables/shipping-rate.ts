import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  shippingservicev1_CreateShippingRateRequest,
  shippingservicev1_DeleteShippingRateRequest,
  shippingservicev1_GetShippingRateRequest,
  shippingservicev1_ListShippingRateResponse,
  shippingservicev1_ShippingRate,
  shippingservicev1_UpdateShippingRateRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 运费模板管理
// ==============================

export function useListShippingRates(
  query: PaginationQuery,
  options?: UseQueryOptions<shippingservicev1_ListShippingRateResponse, Error>
) {
  return useQuery({
    queryKey: ["listShippingRates", query],
    queryFn: () => apiClient.shippingRateService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListShippingRates(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listShippingRates", params],
    queryFn: () => apiClient.shippingRateService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetShippingRate(
  req: shippingservicev1_GetShippingRateRequest,
  options?: UseQueryOptions<shippingservicev1_ShippingRate, Error>
) {
  return useQuery({
    queryKey: ["getShippingRate", req],
    queryFn: () => apiClient.shippingRateService.Get(req),
    ...options,
  });
}

export async function fetchGetShippingRate(req: shippingservicev1_GetShippingRateRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getShippingRate", req],
    queryFn: () => apiClient.shippingRateService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateShippingRate(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.shippingRateService.Create({ data: { ...values } as shippingservicev1_ShippingRate }),
    ...options,
  });
}

export function useUpdateShippingRate(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.shippingRateService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteShippingRate(
  options?: UseMutationOptions<{}, Error, shippingservicev1_DeleteShippingRateRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.shippingRateService.Delete(data),
    ...options,
  });
}
