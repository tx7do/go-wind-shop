import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import {
  useMutation,
  type UseMutationOptions,
} from '@tanstack/vue-query';
import { toValue, type MaybeRefOrGetter } from 'vue';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';
import { getCurrentLocale } from '@/utils/locale';
import type {
  commentservicev1_CreateCommentRequest,
  commentservicev1_ListCommentResponse,
  commentservicev1_ProductRatingSummary,
} from '@/api/generated/app/service/v1';

// ==============================
// 商品评论列表（Query）
// ==============================
export async function fetchListComments(params: any) {
  return await apiClient.commentService.List(params);
}
export function useListComments(
  params: any,
  options?: UseQueryOptions<commentservicev1_ListCommentResponse, Error>,
) {
  return useQuery({
    queryKey: ['listComments', params, getCurrentLocale()],
    queryFn: () => fetchListComments(toValue(params)),
    ...options,
  });
}
export async function fetchListCommentsStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['listComments', params, getCurrentLocale()],
    queryFn: () => fetchListComments(toValue(params)),
    retry: 0,
  });
}

// ==============================
// 发表评论（Mutation）
// ==============================
export async function createComment(request: commentservicev1_CreateCommentRequest) {
  return await apiClient.commentService.Create(request);
}
export function useCreateComment(
  options?: UseMutationOptions<{}, Error, commentservicev1_CreateCommentRequest>,
) {
  return useMutation({
    mutationFn: (request) => createComment(request),
    ...options,
  });
}

// ==============================
// 商品评分聚合（Query，匿名可读）
// ==============================

// ==============================
// 评论 AI 摘要（匿名可读；LLM 生成，服务端缓存）
// ==============================
export async function fetchGetProductReviewSummary(params: { productId: number }) {
  return await apiClient.commentService.GetProductReviewSummary(params);
}

export function useGetProductReviewSummary(
  productId: MaybeRefOrGetter<number>,
  options?: UseQueryOptions<any, Error>,
) {
  return useQuery({
    queryKey: ['productReviewSummary', productId],
    queryFn: () => fetchGetProductReviewSummary({ productId: toValue(productId) }),
    ...options,
  });
}

export function useGetProductRating(
  params: any,
  options?: UseQueryOptions<commentservicev1_ProductRatingSummary, Error>,
) {
  return useQuery({
    queryKey: ['getProductRating', params, getCurrentLocale()],
    queryFn: () => fetchGetProductRating(toValue(params)),
    ...options,
  });
}
export async function fetchGetProductRatingStore(params: any) {
  return queryClient.fetchQuery({
    queryKey: ['getProductRating', params, getCurrentLocale()],
    queryFn: () => fetchGetProductRating(toValue(params)),
    retry: 0,
  });
}
