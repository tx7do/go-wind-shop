import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 商品属性列表（Query）
// ==============================
export async function fetchListProductAttributes(params: any) {
  return await apiClient.productAttributeService.List(params);
}
export function useListProductAttributes(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListProductAttributes>>, Error>,
) {
  return useQuery({
    queryKey: ['listProductAttributes', params, getCurrentLocale()],
    queryFn: () => fetchListProductAttributes(params),
    ...options,
  });
}
export async function fetchListProductAttributesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listProductAttributes', params, getCurrentLocale()],
    queryFn: () => fetchListProductAttributes(params),
    retry: 0,
  });
}

// ==============================
// 商品属性详情（Query，注入 locale）
// ==============================
export async function fetchGetProductAttribute(id: number) {
  return await apiClient.productAttributeService.Get({
    id,
    locale: getCurrentLocale(),
  } as any);
}
export function useGetProductAttribute(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetProductAttribute>>, Error>,
) {
  return useQuery({
    queryKey: ['getProductAttribute', id, getCurrentLocale()],
    queryFn: () => fetchGetProductAttribute(id),
    ...options,
  });
}
export async function fetchGetProductAttributeStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getProductAttribute', id, getCurrentLocale()],
    queryFn: () => fetchGetProductAttribute(id),
    retry: 0,
  });
}
