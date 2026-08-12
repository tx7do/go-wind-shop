import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  couponservicev1_CreateCouponTemplateRequest,
  couponservicev1_DeleteCouponTemplateRequest,
  couponservicev1_GetCouponTemplateRequest,
  couponservicev1_ListCouponTemplateResponse,
  couponservicev1_CouponTemplate,
  couponservicev1_UpdateCouponTemplateRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 优惠券模板管理
// ==============================

export function useListCouponTemplates(
  query: PaginationQuery,
  options?: UseQueryOptions<couponservicev1_ListCouponTemplateResponse, Error>
) {
  return useQuery({
    queryKey: ["listCouponTemplates", query],
    queryFn: () => apiClient.couponTemplateService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListCouponTemplates(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listCouponTemplates", params],
    queryFn: () => apiClient.couponTemplateService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetCouponTemplate(
  req: couponservicev1_GetCouponTemplateRequest,
  options?: UseQueryOptions<couponservicev1_CouponTemplate, Error>
) {
  return useQuery({
    queryKey: ["getCouponTemplate", req],
    queryFn: () => apiClient.couponTemplateService.Get(req),
    ...options,
  });
}

export async function fetchGetCouponTemplate(req: couponservicev1_GetCouponTemplateRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getCouponTemplate", req],
    queryFn: () => apiClient.couponTemplateService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateCouponTemplate(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.couponTemplateService.Create({ data: { ...values } as couponservicev1_CouponTemplate }),
    ...options,
  });
}

export function useUpdateCouponTemplate(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.couponTemplateService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteCouponTemplate(
  options?: UseMutationOptions<{}, Error, couponservicev1_DeleteCouponTemplateRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.couponTemplateService.Delete(data),
    ...options,
  });
}
