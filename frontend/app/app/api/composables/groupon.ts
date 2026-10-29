import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue } from 'vue';
import { apiClient } from '@/api/client';
import type {
  catalogservicev1_ListActiveGrouponsRequest,
  catalogservicev1_ListGrouponActivityResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 拼团活动（C端只读）
// ListActiveGroupons 匿名可读：当前生效中的拼团活动。
// ==============================
export async function fetchListActiveGroupons(
  params?: catalogservicev1_ListActiveGrouponsRequest,
) {
  return await apiClient.grouponService.ListActiveGroupons(params ?? {});
}
export function useListActiveGroupons(
  params?: catalogservicev1_ListActiveGrouponsRequest,
  options?: UseQueryOptions<catalogservicev1_ListGrouponActivityResponse, Error>,
) {
  return useQuery({
    queryKey: ['listActiveGroupons', params],
    queryFn: () => fetchListActiveGroupons(toValue(params)),
    ...options,
  });
}
