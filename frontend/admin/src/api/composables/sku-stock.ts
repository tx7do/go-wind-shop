import {
  useMutation,
  type UseMutationOptions,
} from "@tanstack/vue-query";
import { apiClient } from "@/api/client";

// ==============================
// SKU 库存手动调整（服务端行锁改写 + 操作流水）
// ==============================
export function useAdjustSkuStock(
  options?: UseMutationOptions<{}, Error, { id: number; values: { delta: number; remark?: string } }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: { delta: number; remark?: string } }) =>
      apiClient.skuService.AdjustStock({
        id,
        delta: values.delta,
        remark: values.remark || undefined,
      }),
    ...options,
  });
}
