import { useQuery, type UseQueryOptions } from '@tanstack/vue-query';
import {
  useMutation,
  type UseMutationOptions,
} from '@tanstack/vue-query';
import { toValue, type MaybeRefOrGetter } from 'vue';
import { apiClient } from '@/api/client';
import type {
  interactionservicev1_GetCountsRequest,
  interactionservicev1_GetCountsResponse,
  interactionservicev1_GetInteractionStatusRequest,
  interactionservicev1_GetInteractionStatusResponse,
  interactionservicev1_LikeRequest,
  interactionservicev1_LikeResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// 点赞状态批量查询（Query，需登录）
// ==============================
export async function fetchInteractionStatus(params: interactionservicev1_GetInteractionStatusRequest) {
  return await apiClient.interactionService.GetInteractionStatus(params);
}
export function useGetInteractionStatus(
  params: any,
  options?: UseQueryOptions<interactionservicev1_GetInteractionStatusResponse, Error>,
) {
  return useQuery({
    queryKey: ['interactionStatus', params],
    queryFn: () => fetchInteractionStatus(toValue(params) as interactionservicev1_GetInteractionStatusRequest),
    ...options,
  });
}

// ==============================
// 点赞计数批量查询（Query，公开）
// ==============================
export async function fetchCounts(params: interactionservicev1_GetCountsRequest) {
  return await apiClient.interactionService.GetCounts(params);
}
export function useGetCounts(
  params: any,
  options?: UseQueryOptions<interactionservicev1_GetCountsResponse, Error>,
) {
  return useQuery({
    queryKey: ['interactionCounts', params],
    queryFn: () => fetchCounts(toValue(params) as interactionservicev1_GetCountsRequest),
    ...options,
  });
}

// ==============================
// 点赞 / 取消点赞（Mutation，需登录）
// ==============================
export async function likeComment(request: interactionservicev1_LikeRequest) {
  return await apiClient.interactionService.Like(request);
}
export function useLike(
  options?: UseMutationOptions<interactionservicev1_LikeResponse, Error, interactionservicev1_LikeRequest>,
) {
  return useMutation({
    mutationFn: (request) => likeComment(request),
    ...options,
  });
}

export async function unlikeComment(request: interactionservicev1_LikeRequest) {
  return await apiClient.interactionService.Unlike(request);
}
export function useUnlike(
  options?: UseMutationOptions<interactionservicev1_LikeResponse, Error, interactionservicev1_LikeRequest>,
) {
  return useMutation({
    mutationFn: (request) => unlikeComment(request),
    ...options,
  });
}
