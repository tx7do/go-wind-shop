import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { makeUpdateMask, omit } from '@/core/transport/rest';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import type {
  cartservicev1_Cart,
  cartservicev1_CreateCartRequest,
  cartservicev1_UpdateCartRequest,
  cartservicev1_DeleteCartRequest,
  cartservicev1_ListCartResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 购物车列表（Query）
// ==============================
export async function fetchListCarts(params: any) {
  return await apiClient.cartService.List(params);
}
export function useListCarts(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListCarts>>, Error>,
) {
  return useQuery({
    queryKey: ['listCarts', params, getCurrentLocale()],
    queryFn: () => fetchListCarts(params),
    ...options,
  });
}
export async function fetchListCartsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listCarts', params, getCurrentLocale()],
    queryFn: () => fetchListCarts(params),
    retry: 0,
  });
}

// ==============================
// 创建购物车（Mutation）
// ==============================
export async function createCart(data: cartservicev1_Cart) {
  const request: cartservicev1_CreateCartRequest = { data };
  return await apiClient.cartService.Create(request);
}
export function useCreateCart(
  options?: UseMutationOptions<{}, Error, cartservicev1_Cart>,
) {
  return useMutation({
    mutationFn: (data) => createCart(data),
    ...options,
  });
}
export async function fetchCreateCartStore(data: cartservicev1_Cart) {
  return queryClient.fetchQuery({
    queryKey: ['createCart', data],
    queryFn: () => createCart(data),
    retry: 0,
  });
}

// ==============================
// 更新购物车（Mutation）
// ==============================
export async function updateCart(id: number, values: Record<string, any> = {}) {
  const cleaned = omit(values);
  const updateMask = makeUpdateMask(Object.keys(cleaned ?? []));
  return await apiClient.cartService.Update({
    id,
    // @ts-ignore proto generated code is error.
    data: {
      ...cleaned,
      id,
    },
    // @ts-ignore proto generated code is error.
    updateMask,
  });
}
export function useUpdateCart(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) => updateCart(id, values),
    ...options,
  });
}
export async function fetchUpdateCartStore(id: number, values: Record<string, any> = {}) {
  return queryClient.fetchQuery({
    queryKey: ['updateCart', id, values],
    queryFn: () => updateCart(id, values),
    retry: 0,
  });
}

// ==============================
// 删除购物车（Mutation）
// ==============================
export async function deleteCart(id: number) {
  const request: cartservicev1_DeleteCartRequest = { id };
  return await apiClient.cartService.Delete(request);
}
export function useDeleteCart(options?: UseMutationOptions<{}, Error, number>) {
  return useMutation({
    mutationFn: (id) => deleteCart(id),
    ...options,
  });
}
export async function fetchDeleteCartStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['deleteCart', id],
    queryFn: () => deleteCart(id),
    retry: 0,
  });
}
