import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  paymentservicev1_CreatePaymentTransactionRequest,
  paymentservicev1_DeletePaymentTransactionRequest,
  paymentservicev1_GetPaymentTransactionRequest,
  paymentservicev1_ListPaymentTransactionResponse,
  paymentservicev1_PaymentTransaction,
  paymentservicev1_UpdatePaymentTransactionRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 支付流水管理
// ==============================

export function useListPaymentTransactions(
  query: PaginationQuery,
  options?: UseQueryOptions<paymentservicev1_ListPaymentTransactionResponse, Error>
) {
  return useQuery({
    queryKey: ["listPaymentTransactions", query],
    queryFn: () => apiClient.paymentTransactionService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListPaymentTransactions(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listPaymentTransactions", params],
    queryFn: () => apiClient.paymentTransactionService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetPaymentTransaction(
  req: paymentservicev1_GetPaymentTransactionRequest,
  options?: UseQueryOptions<paymentservicev1_PaymentTransaction, Error>
) {
  return useQuery({
    queryKey: ["getPaymentTransaction", req],
    queryFn: () => apiClient.paymentTransactionService.Get(req),
    ...options,
  });
}

export async function fetchGetPaymentTransaction(
  req: paymentservicev1_GetPaymentTransactionRequest
) {
  return queryClient.fetchQuery({
    queryKey: ["getPaymentTransaction", req],
    queryFn: () => apiClient.paymentTransactionService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreatePaymentTransaction(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.paymentTransactionService.Create({
        data: { ...values },
      } as paymentservicev1_CreatePaymentTransactionRequest),
    ...options,
  });
}

export function useUpdatePaymentTransaction(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.paymentTransactionService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as paymentservicev1_UpdatePaymentTransactionRequest),
    ...options,
  });
}

export function useDeletePaymentTransaction(
  options?: UseMutationOptions<{}, Error, paymentservicev1_DeletePaymentTransactionRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.paymentTransactionService.Delete(data),
    ...options,
  });
}
