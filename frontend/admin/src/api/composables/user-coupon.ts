import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  couponservicev1_CreateUserCouponRequest,
  couponservicev1_DeleteUserCouponRequest,
  couponservicev1_GetUserCouponRequest,
  couponservicev1_ListUserCouponResponse,
  couponservicev1_UserCoupon,
} from "@/api/generated/admin/service/v1";
import type { PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 用户优惠券发放管理
// ==============================

export function useListUserCoupons(
  query: PaginationQuery,
  options?: UseQueryOptions<couponservicev1_ListUserCouponResponse, Error>
) {
  return useQuery({
    queryKey: ["listUserCoupons", query],
    queryFn: () => apiClient.userCouponService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListUserCoupons(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listUserCoupons", params],
    queryFn: () => apiClient.userCouponService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetUserCoupon(
  req: couponservicev1_GetUserCouponRequest,
  options?: UseQueryOptions<couponservicev1_UserCoupon, Error>
) {
  return useQuery({
    queryKey: ["getUserCoupon", req],
    queryFn: () => apiClient.userCouponService.Get(req),
    ...options,
  });
}

export async function fetchGetUserCoupon(req: couponservicev1_GetUserCouponRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getUserCoupon", req],
    queryFn: () => apiClient.userCouponService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateUserCoupon(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.userCouponService.Create({ data: { ...values } as couponservicev1_UserCoupon }),
    ...options,
  });
}

export function useDeleteUserCoupon(
  options?: UseMutationOptions<{}, Error, couponservicev1_DeleteUserCouponRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.userCouponService.Delete(data),
    ...options,
  });
}
