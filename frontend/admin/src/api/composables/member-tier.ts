import {
  useMutation, type UseMutationOptions, useQuery, type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_CreateMemberTierRequest,
  catalogservicev1_DeleteMemberTierRequest,
  catalogservicev1_ListMemberTierResponse,
  catalogservicev1_MemberTier,
  catalogservicev1_UpdateMemberTierRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";

export function useListMemberTiers(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListMemberTierResponse, Error>
) {
  return useQuery({
    queryKey: ["listMemberTiers", query],
    queryFn: () => apiClient.memberTierService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListMemberTiers(params: PaginationQuery) {
  return apiClient.memberTierService.List(params.toRawParams());
}

export function useCreateMemberTier(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.memberTierService.Create({ data: { ...values } as catalogservicev1_MemberTier }),
    ...options,
  });
}

export function useUpdateMemberTier(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.memberTierService.Update({
        id, data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as catalogservicev1_UpdateMemberTierRequest),
    ...options,
  });
}

export function useDeleteMemberTier(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteMemberTierRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.memberTierService.Delete(data),
    ...options,
  });
}
