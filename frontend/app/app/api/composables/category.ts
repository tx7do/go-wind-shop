import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue, type MaybeRefOrGetter } from 'vue';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 类目列表（Query）
// ==============================
export async function fetchListCategories(params: any) {
  return await apiClient.categoryService.List(params);
}
export function useListCategories(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListCategories>>, Error>,
) {
  return useQuery({
    queryKey: ['listCategories', params, getCurrentLocale()],
    queryFn: () => fetchListCategories(params),
    ...options,
  });
}
export async function fetchListCategoriesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listCategories', params, getCurrentLocale()],
    queryFn: () => fetchListCategories(params),
    retry: 0,
  });
}

// ==============================
// 类目详情（Query，注入 locale）
// ==============================
export async function fetchGetCategory(id: number) {
  return await apiClient.categoryService.Get({
    id,
    locale: getCurrentLocale(),
  } as any);
}
export function useGetCategory(
  id: MaybeRefOrGetter<number>,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetCategory>>, Error>,
) {
  return useQuery({
    queryKey: ['getCategory', id, getCurrentLocale()],
    queryFn: () => fetchGetCategory(toValue(id)),
    ...options,
  });
}
export async function fetchGetCategoryStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getCategory', id, getCurrentLocale()],
    queryFn: () => fetchGetCategory(id),
    retry: 0,
  });
}
