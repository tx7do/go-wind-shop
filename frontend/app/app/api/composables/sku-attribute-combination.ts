import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// SKU 属性组合列表（Query）
// ==============================
export async function fetchListSkuAttributeCombinations(params: any) {
  return await apiClient.skuAttributeCombinationService.List(params);
}
export function useListSkuAttributeCombinations(
  params: any,
  options?: UseQueryOptions<
    Awaited<ReturnType<typeof fetchListSkuAttributeCombinations>>,
    Error
  >,
) {
  return useQuery({
    queryKey: ['listSkuAttributeCombinations', params, getCurrentLocale()],
    queryFn: () => fetchListSkuAttributeCombinations(params),
    ...options,
  });
}
export async function fetchListSkuAttributeCombinationsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listSkuAttributeCombinations', params, getCurrentLocale()],
    queryFn: () => fetchListSkuAttributeCombinations(params),
    retry: 0,
  });
}

// ==============================
// SKU 属性组合详情（Query）
// ==============================
export async function fetchGetSkuAttributeCombination(id: number) {
  return await apiClient.skuAttributeCombinationService.Get({ id } as any);
}
export function useGetSkuAttributeCombination(
  id: number,
  options?: UseQueryOptions<
    Awaited<ReturnType<typeof fetchGetSkuAttributeCombination>>,
    Error
  >,
) {
  return useQuery({
    queryKey: ['getSkuAttributeCombination', id],
    queryFn: () => fetchGetSkuAttributeCombination(id),
    ...options,
  });
}
export async function fetchGetSkuAttributeCombinationStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['getSkuAttributeCombination', id],
    queryFn: () => fetchGetSkuAttributeCombination(id),
    retry: 0,
  });
}
