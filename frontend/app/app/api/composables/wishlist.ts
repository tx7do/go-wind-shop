import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import { unref } from 'vue';
import type {
  wishlistservicev1_Wishlist,
  wishlistservicev1_CreateWishlistRequest,
  wishlistservicev1_DeleteWishlistRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 收藏夹列表（Query）
// ==============================
export async function fetchListWishlist(params: any) {
  return await apiClient.wishlistService.List(params);
}
export function useListWishlist(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListWishlist>>, Error>,
) {
  return useQuery({
    queryKey: ['listWishlist', params, getCurrentLocale()],
    queryFn: () => fetchListWishlist(unref(params)),
    ...options,
  });
}
export async function fetchListWishlistStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listWishlist', params, getCurrentLocale()],
    queryFn: () => fetchListWishlist(unref(params)),
    retry: 0,
  });
}

// ==============================
// 添加收藏（Mutation）
// ==============================
export async function createWishlist(data: wishlistservicev1_Wishlist) {
  const request: wishlistservicev1_CreateWishlistRequest = { data };
  return await apiClient.wishlistService.Create(request);
}
export function useCreateWishlist(
  options?: UseMutationOptions<{}, Error, wishlistservicev1_Wishlist>,
) {
  return useMutation({
    mutationFn: (data) => createWishlist(data),
    ...options,
  });
}
export async function fetchCreateWishlistStore(data: wishlistservicev1_Wishlist) {
  return queryClient.fetchQuery({
    queryKey: ['createWishlist', data],
    queryFn: () => createWishlist(data),
    retry: 0,
  });
}

// ==============================
// 取消收藏（Mutation）
// ==============================
export async function deleteWishlist(id: number) {
  const request: wishlistservicev1_DeleteWishlistRequest = { id };
  return await apiClient.wishlistService.Delete(request);
}
export function useDeleteWishlist(options?: UseMutationOptions<{}, Error, number>) {
  return useMutation({
    mutationFn: (id) => deleteWishlist(id),
    ...options,
  });
}
export async function fetchDeleteWishlistStore(id: number) {
  return queryClient.fetchQuery({
    queryKey: ['deleteWishlist', id],
    queryFn: () => deleteWishlist(id),
    retry: 0,
  });
}
