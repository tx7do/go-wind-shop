import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  paymentservicev1_CreatePaymentRefundRequest,
  paymentservicev1_DeletePaymentRefundRequest,
  paymentservicev1_GetPaymentRefundRequest,
  paymentservicev1_ListPaymentRefundResponse,
  paymentservicev1_PaymentRefund,
  paymentservicev1_UpdatePaymentRefundRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 支付退款管理
// ==============================

export function useListPaymentRefunds(
  query: PaginationQuery,
  options?: UseQueryOptions<paymentservicev1_ListPaymentRefundResponse, Error>
) {
  return useQuery({
    queryKey: ["listPaymentRefunds", query],
    queryFn: () => apiClient.paymentRefundService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListPaymentRefunds(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listPaymentRefunds", params],
    queryFn: () => apiClient.paymentRefundService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetPaymentRefund(
  req: paymentservicev1_GetPaymentRefundRequest,
  options?: UseQueryOptions<paymentservicev1_PaymentRefund, Error>
) {
  return useQuery({
    queryKey: ["getPaymentRefund", req],
    queryFn: () => apiClient.paymentRefundService.Get(req),
    ...options,
  });
}

export async function fetchGetPaymentRefund(req: paymentservicev1_GetPaymentRefundRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getPaymentRefund", req],
    queryFn: () => apiClient.paymentRefundService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreatePaymentRefund(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.paymentRefundService.Create({
        data: { ...values },
      } as paymentservicev1_CreatePaymentRefundRequest),
    ...options,
  });
}

export function useUpdatePaymentRefund(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.paymentRefundService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as paymentservicev1_UpdatePaymentRefundRequest),
    ...options,
  });
}

export function useDeletePaymentRefund(
  options?: UseMutationOptions<{}, Error, paymentservicev1_DeletePaymentRefundRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.paymentRefundService.Delete(data),
    ...options,
  });
}
