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
  cartservicev1_CartItem,
  cartservicev1_CreateCartItemRequest,
  cartservicev1_UpdateCartItemRequest,
  cartservicev1_DeleteCartItemRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 购物车项列表（Query）
// ==============================
export async function fetchListCartItems(params: any) {
  return await apiClient.cartItemService.List(params);
}
export function useListCartItems(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListCartItems>>, Error>,
) {
  return useQuery({
    queryKey: ['listCartItems', params, getCurrentLocale()],
    queryFn: () => fetchListCartItems(params),
    ...options,
  });
}
export async function fetchListCartItemsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listCartItems', params, getCurrentLocale()],
    queryFn: () => fetchListCartItems(params),
    retry: 0,
  });
}

// ==============================
// 添加购物车项（Mutation）
// ==============================
export async function createCartItem(data: cartservicev1_CartItem) {
  const request: cartservicev1_CreateCartItemRequest = { data };
  return await apiClient.cartItemService.Create(request);
}
export function useCreateCartItem(
  options?: UseMutationOptions<{}, Error, cartservicev1_CartItem>,
) {
  return useMutation({
    mutationFn: (data) => createCartItem(data),
    ...options,
  });
}
export async function fetchCreateCartItemStore(data: cartservicev1_CartItem) {
  return queryClient.fetchQuery({
    queryKey: ['createCartItem', data],
    queryFn: () => createCartItem(data),
    retry: 0,
  });
}

// ==============================
// 更新购物车项数量（Mutation）
// ==============================
export async function updateCartItem(id: number, values: Record<string, any> = {}) {
  const cleaned = omit(values);
  const updateMask = makeUpdateMask(Object.keys(cleaned ?? []));
  return await apiClient.cartItemService.Update({
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
export function useUpdateCartItem(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) => updateCartItem(id, values),
    ...options,
  });
}
export async function fetchUpdateCartItemStore(id: number, values: Record<string, any> = {}) {
  return queryClient.fetchQuery({
    queryKey: ['updateCartItem', id, values],
    queryFn: () => updateCartItem(id, values),
    retry: 0,
  });
}

// ==============================
// 移除购物车项（Mutation）
// ==============================
export async function deleteCartItem(id: number) {
  const request: cartservicev1_DeleteCartItemRequest = { id };
  return await apiClient.cartItemService.Delete(request);
}
export function useDeleteCartItem(options?: UseMutationOptions<{}, Error, number>) {
  return useMutation({
    mutationFn: (id) => deleteCartItem(id),
    ...options,
  });
}
export async function fetchDeleteCartItemStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['deleteCartItem', id],
    queryFn: () => deleteCartItem(id),
    retry: 0,
  });
}
