import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import { unref } from 'vue';
import type {
  paymentservicev1_PaymentTransaction,
  paymentservicev1_CreatePaymentTransactionRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 创建支付流水（Mutation，结算“支付”步骤）
// ==============================
export async function createPaymentTransaction(data: paymentservicev1_PaymentTransaction) {
  const request: paymentservicev1_CreatePaymentTransactionRequest = { data };
  return await apiClient.paymentTransactionService.Create(request);
}
export function useCreatePaymentTransaction(
  options?: UseMutationOptions<{}, Error, paymentservicev1_PaymentTransaction>,
) {
  return useMutation({
    mutationFn: (data) => createPaymentTransaction(data),
    ...options,
  });
}
export async function fetchCreatePaymentTransactionStore(data: paymentservicev1_PaymentTransaction) {
  return queryClient.fetchQuery({
    queryKey: ['createPaymentTransaction', data],
    queryFn: () => createPaymentTransaction(data),
    retry: 0,
  });
}

// ==============================
// 支付流水列表（Query）
// ==============================
export async function fetchListPaymentTransactions(params: any) {
  return await apiClient.paymentTransactionService.List(params);
}
export function useListPaymentTransactions(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListPaymentTransactions>>, Error>,
) {
  return useQuery({
    queryKey: ['listPaymentTransactions', params, getCurrentLocale()],
    queryFn: () => fetchListPaymentTransactions(unref(params)),
    ...options,
  });
}
export async function fetchListPaymentTransactionsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listPaymentTransactions', params, getCurrentLocale()],
    queryFn: () => fetchListPaymentTransactions(unref(params)),
    retry: 0,
  });
}
