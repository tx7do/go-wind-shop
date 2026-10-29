import { useQuery, useMutation, type UseQueryOptions, type UseMutationOptions } from '@tanstack/vue-query';
import { toValue } from 'vue';
import { apiClient } from '@/api/client';
import { getCurrentLocale } from '@/utils/locale';
import type {
  paymentservicev1_ListEnabledChannelsRequest,
  paymentservicev1_ListEnabledChannelsResponse,
  paymentservicev1_CreatePaymentIntentRequest,
  paymentservicev1_CreatePaymentIntentAppResponse,
  paymentservicev1_MockConfirmPaymentRequest,
  paymentservicev1_MockConfirmPaymentResponse,
  paymentservicev1_PaymentTransaction,
} from '@/api/generated/app/service/v1';

// ==============================
// 支付渠道（C端）
// ListEnabledChannels：结算页渠道选择（登录态，仅公开视图字段）。
// ==============================
export async function fetchListEnabledPaymentChannels(
  params?: paymentservicev1_ListEnabledChannelsRequest,
) {
  return await apiClient.paymentChannelService.ListEnabledChannels(params ?? {});
}
export function useListPaymentChannels(
  params?: paymentservicev1_ListEnabledChannelsRequest,
  options?: UseQueryOptions<paymentservicev1_ListEnabledChannelsResponse, Error>,
) {
  return useQuery({
    queryKey: ['listPaymentChannels', params],
    queryFn: () => fetchListEnabledPaymentChannels(toValue(params)),
    ...options,
  });
}

// ==============================
// 创建支付意图（登录态）
// 服务端校验订单归属/待支付态并落 PENDING 流水；返回拉起支付所需信息
// （QR 码内容 / 跳转链接 / 模拟确认标记）。
// ==============================
export async function createPaymentIntent(
  request: paymentservicev1_CreatePaymentIntentRequest,
): Promise<paymentservicev1_CreatePaymentIntentAppResponse> {
  return await apiClient.paymentTransactionService.CreatePaymentIntent(request);
}
export function useCreatePaymentIntent(
  options?: UseMutationOptions<paymentservicev1_CreatePaymentIntentAppResponse, Error, paymentservicev1_CreatePaymentIntentRequest>,
) {
  return useMutation({
    mutationFn: (req: paymentservicev1_CreatePaymentIntentRequest) => createPaymentIntent(req),
    ...options,
  });
}

// ==============================
// 模拟渠道确认支付（仅 MOCK 渠道流水；开发/演示）
// ==============================
export async function mockConfirmPayment(
  request: paymentservicev1_MockConfirmPaymentRequest,
): Promise<paymentservicev1_MockConfirmPaymentResponse> {
  return await apiClient.paymentTransactionService.MockConfirmPayment(request);
}
export function useMockConfirmPayment(
  options?: UseMutationOptions<paymentservicev1_MockConfirmPaymentResponse, Error, paymentservicev1_MockConfirmPaymentRequest>,
) {
  return useMutation({
    mutationFn: (req: paymentservicev1_MockConfirmPaymentRequest) => mockConfirmPayment(req),
    ...options,
  });
}

// ==============================
// 支付流水查询（轮询支付结果用）
// ==============================
export async function fetchGetPaymentTransaction(id: number): Promise<paymentservicev1_PaymentTransaction> {
  return await apiClient.paymentTransactionService.Get({ id } as any);
}
