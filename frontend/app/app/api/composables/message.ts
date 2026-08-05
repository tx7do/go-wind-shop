import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/vue-query';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';

// ==============================
// 我的收件箱列表（Query）
// recipientUserId 由后端网关从登录态强制注入，前端 query 无需带 userId。
// 可选传 status 过滤（如只看未读：query 携带 status）。
// ==============================
export async function fetchListUserInbox(params: any) {
  return await apiClient.internalMessageRecipientService.ListUserInbox(params);
}
export function useListUserInbox(
  params: any,
  options?: UseQueryOptions<Awaited<ReturnType<typeof fetchListUserInbox>>, Error>,
) {
  return useQuery({
    queryKey: ['listUserInbox', params, getCurrentLocale()],
    queryFn: () => fetchListUserInbox(params),
    ...options,
  });
}

// ==============================
// 标记通知为已读（Mutation）
// recipientIds 由前端传入（要标记的收件箱记录ID），userId 由后端注入。
// ==============================
export async function markNotificationAsRead(recipientIds: number[]) {
  return await apiClient.internalMessageRecipientService.MarkNotificationAsRead({
    recipientIds,
    userId: 0,
  } as any);
}
export function useMarkNotificationAsRead(
  options?: UseMutationOptions<{}, Error, number[]>,
) {
  return useMutation({
    mutationFn: (recipientIds) => markNotificationAsRead(recipientIds),
    ...options,
  });
}

// ==============================
// 删除收件箱通知（Mutation）
// ==============================
export async function deleteNotificationFromInbox(recipientIds: number[]) {
  return await apiClient.internalMessageRecipientService.DeleteNotificationFromInbox({
    recipientIds,
    userId: 0,
  } as any);
}
export function useDeleteNotificationFromInbox(
  options?: UseMutationOptions<{}, Error, number[]>,
) {
  return useMutation({
    mutationFn: (recipientIds) => deleteNotificationFromInbox(recipientIds),
    ...options,
  });
}

// 缓存失效
export function invalidateUserInbox() {
  queryClient.invalidateQueries({ queryKey: ['listUserInbox'] });
}
