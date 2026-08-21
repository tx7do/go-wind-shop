import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  invoiceservicev1_CreateInvoiceRequest,
  invoiceservicev1_DeleteInvoiceRequest,
  invoiceservicev1_GetInvoiceRequest,
  invoiceservicev1_Invoice,
  invoiceservicev1_ListInvoiceResponse,
  invoiceservicev1_UpdateInvoiceRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 发票管理
// ==============================

export function useListInvoices(
  query: PaginationQuery,
  options?: UseQueryOptions<invoiceservicev1_ListInvoiceResponse, Error>
) {
  return useQuery({
    queryKey: ["listInvoices", query],
    queryFn: () => apiClient.invoiceService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListInvoices(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listInvoices", params],
    queryFn: () => apiClient.invoiceService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetInvoice(
  req: invoiceservicev1_GetInvoiceRequest,
  options?: UseQueryOptions<invoiceservicev1_Invoice, Error>
) {
  return useQuery({
    queryKey: ["getInvoice", req],
    queryFn: () => apiClient.invoiceService.Get(req),
    ...options,
  });
}

export async function fetchGetInvoice(req: invoiceservicev1_GetInvoiceRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getInvoice", req],
    queryFn: () => apiClient.invoiceService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateInvoice(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.invoiceService.Create({ data: { ...values } as invoiceservicev1_Invoice }),
    ...options,
  });
}

export function useUpdateInvoice(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.invoiceService.Update({
        id,
        data: { ...values } as invoiceservicev1_Invoice,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteInvoice(
  options?: UseMutationOptions<{}, Error, invoiceservicev1_DeleteInvoiceRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.invoiceService.Delete(data),
    ...options,
  });
}
