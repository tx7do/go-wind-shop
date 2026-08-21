import {
  useQuery,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import { unref } from 'vue';
import type {
  invoiceservicev1_GetInvoiceRequest,
} from '@/api/generated/app/service/v1';

// ==============================
// 发票列表（Query，仅本人）
// 裁剪写 RPC：仅 List/Get，用户不能自己开发票。
// 用户隔离由 BFF fail-closed 注入 userId + core UserPrivacy 行级隔离双重保障。
// ==============================
export async function fetchListInvoices(params: any) {
  return await apiClient.invoiceService.List(params);
}
export function useListInvoices(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListInvoices>>, Error>,
) {
  return useQuery({
    queryKey: ['listInvoices', params, getCurrentLocale()],
    queryFn: () => fetchListInvoices(unref(params)),
    ...options,
  });
}
export async function fetchListInvoicesStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listInvoices', params, getCurrentLocale()],
    queryFn: () => fetchListInvoices(unref(params)),
    retry: 0,
  });
}

// ==============================
// 发票详情（Query，仅本人）
// ==============================
export async function fetchGetInvoice(req: invoiceservicev1_GetInvoiceRequest) {
  return await apiClient.invoiceService.Get(req);
}
