import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import { toValue } from 'vue';
import { apiClient } from '@/api/client';
import type {
  catalogservicev1_ListActiveSeckillsRequest,
  catalogservicev1_ListSeckillActivityResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 秒杀活动（C端只读）
// ListActiveSeckills 匿名可读：当前生效中且未售罄的秒杀活动。
// ==============================
export async function fetchListActiveSeckills(
  params?: catalogservicev1_ListActiveSeckillsRequest,
) {
  return await apiClient.seckillService.ListActiveSeckills(params ?? {});
}
export function useListActiveSeckills(
  params?: catalogservicev1_ListActiveSeckillsRequest,
  options?: UseQueryOptions<catalogservicev1_ListSeckillActivityResponse, Error>,
) {
  return useQuery({
    queryKey: ['listActiveSeckills', params],
    queryFn: () => fetchListActiveSeckills(toValue(params)),
    ...options,
  });
}
