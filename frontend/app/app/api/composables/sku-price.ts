import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import { unref } from 'vue';

// ==============================
// SKU 价格列表（Query）
// ==============================
export async function fetchListSkuPrices(params: any) {
  return await apiClient.skuPriceService.List(params);
}
export function useListSkuPrices(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListSkuPrices>>, Error>,
) {
  return useQuery({
    queryKey: ['listSkuPrices', params, getCurrentLocale()],
    queryFn: () => fetchListSkuPrices(unref(params)),
    ...options,
  });
}
export async function fetchListSkuPricesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listSkuPrices', params, getCurrentLocale()],
    queryFn: () => fetchListSkuPrices(unref(params)),
    retry: 0,
  });
}

// ==============================
// SKU 价格详情（Query）
// ==============================
export async function fetchGetSkuPrice(id: number) {
  return await apiClient.skuPriceService.Get({ id } as any);
}
export function useGetSkuPrice(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetSkuPrice>>, Error>,
) {
  return useQuery({
    queryKey: ['getSkuPrice', id],
    queryFn: () => fetchGetSkuPrice(id),
    ...options,
  });
}
export async function fetchGetSkuPriceStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getSkuPrice', id],
    queryFn: () => fetchGetSkuPrice(id),
    retry: 0,
  });
}
