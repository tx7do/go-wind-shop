import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue, type MaybeRefOrGetter } from 'vue';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 类目列表（Query）
// ==============================
// 后端 List 没有 ORDER BY，同一份数据多次请求顺序会变（首页快捷分类、分类页砖块
// 每次刷新都换位置）。列表接口消费方全部依赖这里的顺序，统一按 sortOrder → id 稳定排序。
// protobuf JSON 的 int64 是字符串，比较前先 Number()。
function sortByDisplayOrder<T>(resp: T): T {
  const items = (resp as any)?.items;
  if (!Array.isArray(items)) return resp;
  const num = (v: any) => {
    const n = Number(v);
    return Number.isFinite(n) ? n : 0;
  };
  items.sort((a: any, b: any) => num(a?.sortOrder) - num(b?.sortOrder) || num(a?.id) - num(b?.id));
  return resp;
}

export async function fetchListCategories(params: any) {
  return sortByDisplayOrder(await apiClient.categoryService.List(params));
}
export function useListCategories(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListCategories>>, Error>,
) {
  return useQuery({
    queryKey: ['listCategories', params, getCurrentLocale()],
    queryFn: () => fetchListCategories(toValue(params)),
    ...options,
  });
}
export async function fetchListCategoriesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listCategories', params, getCurrentLocale()],
    queryFn: () => fetchListCategories(toValue(params)),
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
