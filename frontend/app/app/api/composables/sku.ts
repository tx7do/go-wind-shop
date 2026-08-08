import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import { unref } from 'vue';

// ==============================
// SKU 列表（Query）
// ==============================
export async function fetchListSkus(params: any) {
  return await apiClient.skuService.List(params);
}
export function useListSkus(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListSkus>>, Error>,
) {
  return useQuery({
    queryKey: ['listSkus', params, getCurrentLocale()],
    queryFn: () => fetchListSkus(unref(params)),
    ...options,
  });
}
export async function fetchListSkusStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listSkus', params, getCurrentLocale()],
    queryFn: () => fetchListSkus(unref(params)),
    retry: 0,
  });
}

// ==============================
// SKU 详情（Query）
// ==============================
export async function fetchGetSku(id: number) {
  return await apiClient.skuService.Get({ id } as any);
}
export function useGetSku(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetSku>>, Error>,
) {
  return useQuery({
    queryKey: ['getSku', id],
    queryFn: () => fetchGetSku(id),
    ...options,
  });
}
export async function fetchGetSkuStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getSku', id],
    queryFn: () => fetchGetSku(id),
    retry: 0,
  });
}
