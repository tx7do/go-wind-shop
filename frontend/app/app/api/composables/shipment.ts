import {
  useQuery,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { unref, type ComputedRef, type Ref } from 'vue';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 物流单列表（Query，app 侧只读）
// ==============================
export async function fetchListShipments(params: any) {
  return await apiClient.shipmentService.List(params);
}
export function useListShipments(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListShipments>>, Error>,
) {
  return useQuery({
    queryKey: ['listShipments', params, getCurrentLocale()],
    queryFn: () => fetchListShipments(unref(params)),
    ...options,
  });
}

// ==============================
// 物流单详情（Query，app 侧只读）
// ==============================
export async function fetchGetShipment(id: number) {
  return await apiClient.shipmentService.Get({ id } as any);
}
type ShipmentIdSource = number | Ref<number> | ComputedRef<number>;
export function useGetShipment(
  id: ShipmentIdSource,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetShipment>>, Error>,
) {
  return useQuery({
    queryKey: ['getShipment', id, getCurrentLocale()],
    queryFn: () => fetchGetShipment(unref(id)),
    ...options,
  });
}

export function invalidateShipments() {
  queryClient.invalidateQueries({ queryKey: ['listShipments'] });
}
