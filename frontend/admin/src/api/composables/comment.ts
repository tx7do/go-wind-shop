import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  commentservicev1_CreateCommentRequest,
  commentservicev1_DeleteCommentRequest,
  commentservicev1_GetCommentRequest,
  commentservicev1_ListCommentResponse,
  commentservicev1_Comment,
  commentservicev1_UpdateCommentRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 商品评论管理
// ==============================

export function useListComments(
  query: PaginationQuery,
  options?: UseQueryOptions<commentservicev1_ListCommentResponse, Error>
) {
  return useQuery({
    queryKey: ["listComments", query],
    queryFn: () => apiClient.commentService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListComments(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listComments", params],
    queryFn: () => apiClient.commentService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetComment(
  req: commentservicev1_GetCommentRequest,
  options?: UseQueryOptions<commentservicev1_Comment, Error>
) {
  return useQuery({
    queryKey: ["getComment", req],
    queryFn: () => apiClient.commentService.Get(req),
    ...options,
  });
}

export async function fetchGetComment(req: commentservicev1_GetCommentRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getComment", req],
    queryFn: () => apiClient.commentService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateComment(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.commentService.Create({ data: { ...values } as commentservicev1_Comment }),
    ...options,
  });
}

export function useUpdateComment(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.commentService.Update({
        id,
        data: { ...values } as commentservicev1_Comment,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteComment(
  options?: UseMutationOptions<{}, Error, commentservicev1_DeleteCommentRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.commentService.Delete(data),
    ...options,
  });
}
