import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue, type MaybeRefOrGetter } from 'vue';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 商品列表（Query）
// ==============================
export async function fetchListProducts(params: any) {
  return await apiClient.productService.List(params);
}
export function useListProducts(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListProducts>>, Error>,
) {
  return useQuery({
    queryKey: ['listProducts', params, getCurrentLocale()],
    queryFn: () => fetchListProducts(toValue(params)),
    ...options,
  });
}
export async function fetchListProductsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listProducts', params, getCurrentLocale()],
    queryFn: () => fetchListProducts(toValue(params)),
    retry: 0,
  });
}

// ==============================
// 商品搜索（Elasticsearch 全文检索，Query）
// language 由 getCurrentLocale() 强制注入，服务端硬编码 status=ACTIVE。
// ==============================
export async function fetchSearchProducts(params: any) {
  return await apiClient.productService.SearchProducts(params);
}
// 搜索联想（输入联想，免鉴权）：按商品名前缀短语匹配返回候选
export async function fetchSearchSuggestions(query: string, language?: string, size?: number) {
  return await apiClient.productService.SearchProductSuggestions({
    query,
    language: language || getCurrentLocale(),
    size,
  } as any);
}

export function useSearchProducts(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchSearchProducts>>, Error>,
) {
  return useQuery({
    queryKey: ['searchProducts', params, getCurrentLocale()],
    queryFn: () => fetchSearchProducts(toValue(params)),
    ...options,
  });
}

// ==============================
// 商品详情（Query，注入 locale）
// ==============================
export async function fetchGetProduct(id: number) {
  return await apiClient.productService.Get({
    id,
    locale: getCurrentLocale(),
  } as any);
}
export function useGetProduct(
  id: MaybeRefOrGetter<number>,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetProduct>>, Error>,
) {
  return useQuery({
    queryKey: ['getProduct', id, getCurrentLocale()],
    queryFn: () => fetchGetProduct(toValue(id)),
    ...options,
  });
}
export async function fetchGetProductStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getProduct', id, getCurrentLocale()],
    queryFn: () => fetchGetProduct(id),
    retry: 0,
  });
}
