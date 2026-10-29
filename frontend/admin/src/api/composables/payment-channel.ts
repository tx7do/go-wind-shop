import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  paymentservicev1_CreatePaymentChannelConfigRequest,
  paymentservicev1_DeletePaymentChannelConfigRequest,
  paymentservicev1_GetPaymentChannelConfigRequest,
  paymentservicev1_ListPaymentChannelConfigResponse,
  paymentservicev1_PaymentChannelConfig,
  paymentservicev1_UpdatePaymentChannelConfigRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 支付渠道配置管理（全 CRUD）
// credentials 为写入专用字段：Get/List 恒不回填（后端剥离），仅提交新值时改写密钥。
// ==============================

export function useListPaymentChannelConfigs(
  query: PaginationQuery,
  options?: UseQueryOptions<paymentservicev1_ListPaymentChannelConfigResponse, Error>
) {
  return useQuery({
    queryKey: ["listPaymentChannelConfigs", query],
    queryFn: () => apiClient.paymentChannelConfigService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListPaymentChannelConfigs(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listPaymentChannelConfigs", params],
    queryFn: () => apiClient.paymentChannelConfigService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetPaymentChannelConfig(
  req: paymentservicev1_GetPaymentChannelConfigRequest,
  options?: UseQueryOptions<paymentservicev1_PaymentChannelConfig, Error>
) {
  return useQuery({
    queryKey: ["getPaymentChannelConfig", req],
    queryFn: () => apiClient.paymentChannelConfigService.Get(req),
    ...options,
  });
}

export async function fetchGetPaymentChannelConfig(req: paymentservicev1_GetPaymentChannelConfigRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getPaymentChannelConfig", req],
    queryFn: () => apiClient.paymentChannelConfigService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreatePaymentChannelConfig(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.paymentChannelConfigService.Create({
        data: { ...values } as paymentservicev1_PaymentChannelConfig,
      } as paymentservicev1_CreatePaymentChannelConfigRequest),
    ...options,
  });
}

export function useUpdatePaymentChannelConfig(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.paymentChannelConfigService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as paymentservicev1_UpdatePaymentChannelConfigRequest),
    ...options,
  });
}

export function useDeletePaymentChannelConfig(
  options?: UseMutationOptions<{}, Error, paymentservicev1_DeletePaymentChannelConfigRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.paymentChannelConfigService.Delete(data),
    ...options,
  });
}
