import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { makeUpdateMask } from '@/core/transport/rest';
import { getCurrentLocale } from '@/utils/locale';
import type {
  addressservicev1_ShippingAddress,
  addressservicev1_CreateShippingAddressRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 收货地址列表（Query）
// ==============================
export async function fetchListShippingAddresses(params: any) {
  return await apiClient.shippingAddressService.List(params);
}
export function useListShippingAddresses(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListShippingAddresses>>, Error>,
) {
  return useQuery({
    queryKey: ['listShippingAddresses', params, getCurrentLocale()],
    queryFn: () => fetchListShippingAddresses(params),
    ...options,
  });
}
export async function fetchListShippingAddressesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listShippingAddresses', params, getCurrentLocale()],
    queryFn: () => fetchListShippingAddresses(params),
    retry: 0,
  });
}

// ==============================
// 收货地址详情（Query）
// ==============================
export async function fetchGetShippingAddress(id: number) {
  return await apiClient.shippingAddressService.Get({ id } as any);
}
export function useGetShippingAddress(
  id: number,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchGetShippingAddress>>, Error>,
) {
  return useQuery({
    queryKey: ['getShippingAddress', id, getCurrentLocale()],
    queryFn: () => fetchGetShippingAddress(id),
    ...options,
  });
}

// ==============================
// 新增收货地址（Mutation）
// 后端网关会强制注入当前登录用户的 userId/tenantId，前端无需（也不应）传这些字段。
// ==============================
export async function createShippingAddress(data: addressservicev1_ShippingAddress) {
  const request: addressservicev1_CreateShippingAddressRequest = { data };
  return await apiClient.shippingAddressService.Create(request);
}
export function useCreateShippingAddress(
  options?: UseMutationOptions<{}, Error, addressservicev1_ShippingAddress>,
) {
  return useMutation({
    mutationFn: (data) => createShippingAddress(data),
    ...options,
  });
}

// ==============================
// 更新收货地址（Mutation）
// ==============================
export async function updateShippingAddress(id: number, values: Record<string, any> = {}) {
  const updateMask = makeUpdateMask(Object.keys(values ?? {}));
  return await apiClient.shippingAddressService.Update({
    id,
    // @ts-ignore proto generated code is error.
    data: { ...values, id },
    // @ts-ignore proto generated code is error.
    updateMask,
  });
}
export function useUpdateShippingAddress(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) => updateShippingAddress(id, values),
    ...options,
  });
}

// ==============================
// 删除收货地址（Mutation）
// ==============================
export async function deleteShippingAddress(id: number) {
  return await apiClient.shippingAddressService.Delete({ id } as any);
}
export function useDeleteShippingAddress(
  options?: UseMutationOptions<{}, Error, number>,
) {
  return useMutation({
    mutationFn: (id) => deleteShippingAddress(id),
    ...options,
  });
}

// ==============================
// 缓存失效工具：增删改后刷新列表
// ==============================
export function invalidateShippingAddresses() {
  queryClient.invalidateQueries({ queryKey: ['listShippingAddresses'] });
}
