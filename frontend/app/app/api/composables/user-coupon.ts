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
  couponservicev1_UserCoupon,
  couponservicev1_GetUserCouponRequest,
  couponservicev1_QuoteRequest,
  couponservicev1_QuoteResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 用户优惠券列表（Query，仅本人）
// 用户隔离由 BFF fail-closed 注入 userId + core UserPrivacy 行级隔离双重保障。
// ==============================
export async function fetchListUserCoupons(params: any) {
  return await apiClient.userCouponService.List(params);
}
export function useListUserCoupons(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListUserCoupons>>, Error>,
) {
  return useQuery({
    queryKey: ['listUserCoupons', params, getCurrentLocale()],
    queryFn: () => fetchListUserCoupons(unref(params)),
    ...options,
  });
}
export async function fetchListUserCouponsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listUserCoupons', params, getCurrentLocale()],
    queryFn: () => fetchListUserCoupons(unref(params)),
    retry: 0,
  });
}

// ==============================
// 优惠券详情（Query，仅本人）
// ==============================
export async function fetchGetUserCoupon(req: couponservicev1_GetUserCouponRequest) {
  return await apiClient.userCouponService.Get(req);
}

// ==============================
// 试算报价（Mutation，预览抵扣金额）
// 不持锁不落库，最终抵扣以下单时事务内校验为准。
// ==============================
export async function quoteUserCoupon(req: couponservicev1_QuoteRequest): Promise<couponservicev1_QuoteResponse> {
  return await apiClient.userCouponService.Quote(req);
}
export function useQuoteUserCoupon(
  options?: UseMutationOptions<couponservicev1_QuoteResponse, Error, couponservicev1_QuoteRequest>,
) {
  return useMutation({
    mutationFn: (req) => quoteUserCoupon(req),
    ...options,
  });
}
