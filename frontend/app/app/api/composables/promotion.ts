import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue } from 'vue';
import { apiClient } from '@/api/client';
import { getCurrentLocale } from '@/utils/locale';
import type {
  catalogservicev1_ListActivePromotionsRequest,
  catalogservicev1_ListPromotionResponse,
  catalogservicev1_QuoteOrderPromotionsRequest,
  catalogservicev1_QuoteOrderPromotionsResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 促销活动（C端只读）
// ListActivePromotions 匿名可读；QuoteOrderPromotions 需登录（BFF 强制）。
// ==============================
export async function fetchListActivePromotions(
  params?: catalogservicev1_ListActivePromotionsRequest,
) {
  return await apiClient.promotionService.ListActivePromotions(params ?? {});
}
export function useListActivePromotions(
  params?: catalogservicev1_ListActivePromotionsRequest,
  options?: UseQueryOptions<catalogservicev1_ListPromotionResponse, Error>,
) {
  return useQuery({
    queryKey: ['listActivePromotions', params, getCurrentLocale()],
    queryFn: () => fetchListActivePromotions(toValue(params)),
    ...options,
  });
}

export async function quoteOrderPromotions(
  items: { skuId: number; quantity: number }[],
  currency?: string,
): Promise<catalogservicev1_QuoteOrderPromotionsResponse> {
  const request: catalogservicev1_QuoteOrderPromotionsRequest = {
    items: items.map((it) => ({ skuId: it.skuId, quantity: it.quantity })),
  };
  if (currency) {
    request.currency = currency;
  }
  return await apiClient.promotionService.QuoteOrderPromotions(request);
}
