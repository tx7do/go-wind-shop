import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { toValue } from 'vue';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import type {
  couponservicev1_ClaimCouponRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 可领优惠券模板列表（Query，匿名可读）
// 领券中心浏览入口。core 侧 ListClaimable 在 repo 层用 claimable=true AND
// status=ACTIVE 谓词过滤，service 层后过滤有效窗口（valid_from/until 时间判断）。
// 匿名可读由 TenantPrivacy 对无 viewer 放行保证（与商品 List 同机制）。
// ==============================
export async function fetchListCouponTemplates(params: any) {
  return await apiClient.couponTemplateService.List(params);
}
export function useListCouponTemplates(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListCouponTemplates>>, Error>,
) {
  return useQuery({
    queryKey: ['listCouponTemplates', params, getCurrentLocale()],
    queryFn: () => fetchListCouponTemplates(toValue(params)),
    ...options,
  });
}
export async function fetchListCouponTemplatesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listCouponTemplates', params, getCurrentLocale()],
    queryFn: () => fetchListCouponTemplates(toValue(params)),
    retry: 0,
  });
}

// ==============================
// 领取优惠券（Mutation，强制 auth）
// BFF 透传 template_id，user_id 由 core 从 viewer 强制，事务内 ForUpdate 原子校验
// claimable/status/有效窗口/限领。超限或不可领由 core 原子拒绝。
// ==============================
export async function claimCouponTemplate(req: couponservicev1_ClaimCouponRequest) {
  return await apiClient.userCouponService.Claim(req);
}
export function useClaimCouponTemplate(
  options?: UseMutationOptions<{}, Error, couponservicev1_ClaimCouponRequest>,
) {
  return useMutation({
    mutationFn: (req) => claimCouponTemplate(req),
    ...options,
  });
}
