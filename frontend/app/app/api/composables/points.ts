import { useMutation,
  useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue } from 'vue';
import { apiClient } from '@/api/client';
import type {
  catalogservicev1_ListPointsLedgerResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 会员积分（C端：余额 + 本人台账；登录态，隐私层钉定本人）
// ==============================
export async function fetchGetMyPointsBalance(): Promise<{ balance?: number }> {
  return await apiClient.pointsService.GetMyPointsBalance({} as any);
}
export function useGetMyPointsBalance(
  params?: Record<string, any>,
  options?: UseQueryOptions<{ balance?: number }, Error>,
) {
  return useQuery({
    queryKey: ['myPointsBalance', params],
    queryFn: () => fetchGetMyPointsBalance(),
    ...options,
  });
}

export function useListMyPointsLedgers(
  query: Record<string, any>,
  options?: UseQueryOptions<catalogservicev1_ListPointsLedgerResponse, Error>,
) {
  return useQuery({
    queryKey: ['myPointsLedgers', query],
    queryFn: () => apiClient.pointsService.ListMyPointsLedgers(query as any),
    ...options,
  });
}

// ==============================
// 每日签到（登录态；幂等）
// ==============================
export async function checkIn(): Promise<{ awarded?: number; balance?: number; alreadyCheckedIn?: boolean }> {
  return await apiClient.pointsService.CheckIn({});
}
export function useCheckIn(
  options?: UseMutationOptions<{ awarded?: number; balance?: number; alreadyCheckedIn?: boolean }, Error, void>,
) {
  return useMutation({
    mutationFn: () => checkIn(),
    ...options,
  });
}

export async function fetchGetCheckInStatus(): Promise<{ checkedInToday?: boolean; consecutiveDays?: number }> {
  return await apiClient.pointsService.GetCheckInStatus({});
}
export function useGetCheckInStatus(
  options?: UseQueryOptions<{ checkedInToday?: boolean; consecutiveDays?: number }, Error>,
) {
  return useQuery({
    queryKey: ['checkInStatus'],
    queryFn: () => fetchGetCheckInStatus(),
    ...options,
  });
}

// ==============================
// 积分兑换优惠券（登录态；事务内扣积分+发券）
// ==============================
export async function exchangePointsForCoupon(
  couponTemplateId: number,
): Promise<{ couponId?: number; pointsDeducted?: number; balance?: number }> {
  return await apiClient.pointsService.ExchangePointsForCoupon({ couponTemplateId } as any);
}
export function useExchangePointsForCoupon(
  options?: any,
) {
  return useMutation({
    mutationFn: (couponTemplateId: number) => exchangePointsForCoupon(couponTemplateId),
    ...options,
  });
}
