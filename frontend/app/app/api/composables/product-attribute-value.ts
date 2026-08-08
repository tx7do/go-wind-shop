import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import { unref } from 'vue';

// ==============================
// 商品属性值列表（Query）
// ==============================
export async function fetchListProductAttributeValues(params: any) {
  return await apiClient.productAttributeValueService.List(params);
}
export function useListProductAttributeValues(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListProductAttributeValues>>, Error>,
) {
  return useQuery({
    queryKey: ['listProductAttributeValues', params, getCurrentLocale()],
    queryFn: () => fetchListProductAttributeValues(unref(params)),
    ...options,
  });
}
export async function fetchListProductAttributeValuesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listProductAttributeValues', params, getCurrentLocale()],
    queryFn: () => fetchListProductAttributeValues(unref(params)),
    retry: 0,
  });
}

// ==============================
// 商品属性值详情（Query，注入 locale）
// ==============================
export async function fetchGetProductAttributeValue(id: number) {
  return await apiClient.productAttributeValueService.Get({
    id,
    locale: getCurrentLocale(),
  } as any);
}
export function useGetProductAttributeValue(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetProductAttributeValue>>, Error>,
) {
  return useQuery({
    queryKey: ['getProductAttributeValue', id, getCurrentLocale()],
    queryFn: () => fetchGetProductAttributeValue(id),
    ...options,
  });
}
export async function fetchGetProductAttributeValueStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getProductAttributeValue', id, getCurrentLocale()],
    queryFn: () => fetchGetProductAttributeValue(id),
    retry: 0,
  });
}
