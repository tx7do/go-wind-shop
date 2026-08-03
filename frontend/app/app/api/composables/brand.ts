import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 品牌列表（Query）
// ==============================
export async function fetchListBrands(params: any) {
  return await apiClient.brandService.List(params);
}
export function useListBrands(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListBrands>>, Error>,
) {
  return useQuery({
    queryKey: ['listBrands', params, getCurrentLocale()],
    queryFn: () => fetchListBrands(params),
    ...options,
  });
}
export async function fetchListBrandsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listBrands', params, getCurrentLocale()],
    queryFn: () => fetchListBrands(params),
    retry: 0,
  });
}

// ==============================
// 品牌详情（Query，注入 locale）
// ==============================
export async function fetchGetBrand(id: number) {
  return await apiClient.brandService.Get({
    id,
    locale: getCurrentLocale(),
  } as any);
}
export function useGetBrand(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetBrand>>, Error>,
) {
  return useQuery({
    queryKey: ['getBrand', id, getCurrentLocale()],
    queryFn: () => fetchGetBrand(id),
    ...options,
  });
}
export async function fetchGetBrandStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getBrand', id, getCurrentLocale()],
    queryFn: () => fetchGetBrand(id),
    retry: 0,
  });
}
