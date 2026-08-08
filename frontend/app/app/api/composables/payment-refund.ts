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
  paymentservicev1_PaymentRefund,
  paymentservicev1_CreatePaymentRefundRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 创建退款单（Mutation，买家侧“申请退款”）
// 退款单关联到具体的支付流水 transactionId；幂等键防重放。
// ==============================
export async function createPaymentRefund(data: paymentservicev1_PaymentRefund) {
  const request: paymentservicev1_CreatePaymentRefundRequest = { data };
  return await apiClient.paymentRefundService.Create(request);
}
export function useCreatePaymentRefund(
  options?: UseMutationOptions<{}, Error, paymentservicev1_PaymentRefund>,
) {
  return useMutation({
    mutationFn: (data) => createPaymentRefund(data),
    ...options,
  });
}

// ==============================
// 退款单列表（Query）
// ==============================
export async function fetchListPaymentRefunds(params: any) {
  return await apiClient.paymentRefundService.List(params);
}
export function useListPaymentRefunds(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListPaymentRefunds>>, Error>,
) {
  return useQuery({
    queryKey: ['listPaymentRefunds', params, getCurrentLocale()],
    queryFn: () => fetchListPaymentRefunds(unref(params)),
    ...options,
  });
}
export async function fetchListPaymentRefundsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listPaymentRefunds', params, getCurrentLocale()],
    queryFn: () => fetchListPaymentRefunds(unref(params)),
    retry: 0,
  });
}
