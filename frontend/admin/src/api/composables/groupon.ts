import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_CreateGrouponActivityRequest,
  catalogservicev1_DeleteGrouponActivityRequest,
  catalogservicev1_GetGrouponActivityRequest,
  catalogservicev1_ListGrouponActivityResponse,
  catalogservicev1_GrouponActivity,
  catalogservicev1_UpdateGrouponActivityRequest,
  catalogservicev1_ListGrouponGroupsRequest,
  catalogservicev1_ListGrouponGroupResponse,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";

// ==============================
// 拼团活动管理（活动全 CRUD + 团只读）
// ==============================

export function useListGrouponActivities(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListGrouponActivityResponse, Error>
) {
  return useQuery({
    queryKey: ["listGrouponActivities", query],
    queryFn: () => apiClient.grouponActivityService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListGrouponActivities(params: PaginationQuery) {
  return apiClient.grouponActivityService.List(params.toRawParams());
}

export function useCreateGrouponActivity(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.grouponActivityService.Create({
        data: { ...values } as catalogservicev1_GrouponActivity,
      } as catalogservicev1_CreateGrouponActivityRequest),
    ...options,
  });
}

export function useUpdateGrouponActivity(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.grouponActivityService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      } as catalogservicev1_UpdateGrouponActivityRequest),
    ...options,
  });
}

export function useDeleteGrouponActivity(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteGrouponActivityRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.grouponActivityService.Delete(data),
    ...options,
  });
}

export function useListGrouponGroups(
  req: catalogservicev1_ListGrouponGroupsRequest | (() => catalogservicev1_ListGrouponGroupsRequest),
  options?: UseQueryOptions<catalogservicev1_ListGrouponGroupResponse, Error>
) {
  return useQuery({
    queryKey: ["listGrouponGroups", req],
    queryFn: () => apiClient.grouponActivityService.ListGroups(typeof req === "function" ? req() : req),
    ...options,
  });
}
