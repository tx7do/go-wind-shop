import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  taxservicev1_CreateTaxRateRequest,
  taxservicev1_DeleteTaxRateRequest,
  taxservicev1_GetTaxRateRequest,
  taxservicev1_ListTaxRateResponse,
  taxservicev1_TaxRate,
  taxservicev1_UpdateTaxRateRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 税率规则管理
// ==============================

export function useListTaxRates(
  query: PaginationQuery,
  options?: UseQueryOptions<taxservicev1_ListTaxRateResponse, Error>
) {
  return useQuery({
    queryKey: ["listTaxRates", query],
    queryFn: () => apiClient.taxRateService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListTaxRates(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listTaxRates", params],
    queryFn: () => apiClient.taxRateService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetTaxRate(
  req: taxservicev1_GetTaxRateRequest,
  options?: UseQueryOptions<taxservicev1_TaxRate, Error>
) {
  return useQuery({
    queryKey: ["getTaxRate", req],
    queryFn: () => apiClient.taxRateService.Get(req),
    ...options,
  });
}

export async function fetchGetTaxRate(req: taxservicev1_GetTaxRateRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getTaxRate", req],
    queryFn: () => apiClient.taxRateService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateTaxRate(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.taxRateService.Create({ data: { ...values } as taxservicev1_TaxRate }),
    ...options,
  });
}

export function useUpdateTaxRate(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.taxRateService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteTaxRate(
  options?: UseMutationOptions<{}, Error, taxservicev1_DeleteTaxRateRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.taxRateService.Delete(data),
    ...options,
  });
}
