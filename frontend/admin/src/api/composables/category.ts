import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateCategoriesRequest,
  catalogservicev1_Category,
  catalogservicev1_CategoryTranslation,
  catalogservicev1_CreateCategoryRequest,
  catalogservicev1_CreateCategoryTranslationRequest,
  catalogservicev1_DeleteCategoryRequest,
  catalogservicev1_DeleteCategoryTranslationRequest,
  catalogservicev1_GetCategoryRequest,
  catalogservicev1_ListCategoryResponse,
  catalogservicev1_UpdateCategoryRequest,
  catalogservicev1_UpdateCategoryTranslationRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 商品类目管理
// ==============================

export function useListCategories(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListCategoryResponse, Error>
) {
  return useQuery({
    queryKey: ["listCategories", query],
    queryFn: () => apiClient.categoryService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListCategories(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listCategories", params],
    queryFn: () => apiClient.categoryService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetCategory(
  req: catalogservicev1_GetCategoryRequest,
  options?: UseQueryOptions<catalogservicev1_Category, Error>
) {
  return useQuery({
    queryKey: ["getCategory", req],
    queryFn: () => apiClient.categoryService.Get(req),
    ...options,
  });
}

export async function fetchGetCategory(req: catalogservicev1_GetCategoryRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getCategory", req],
    queryFn: () => apiClient.categoryService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateCategory(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.categoryService.Create({ data: { ...values } as unknown as catalogservicev1_Category }),
    ...options,
  });
}

export function useUpdateCategory(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.categoryService.Update({
        id,
        data: { ...values } as unknown as catalogservicev1_Category,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteCategory(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteCategoryRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.Delete(data),
    ...options,
  });
}

export function useBatchCreateCategories(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateCategoriesRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.BatchCreate(data),
    ...options,
  });
}

// ---- 类目翻译 ----

export function useGetCategoryTranslation(
  req: catalogservicev1_GetCategoryRequest,
  options?: UseQueryOptions<catalogservicev1_CategoryTranslation, Error>
) {
  return useQuery({
    queryKey: ["getCategoryTranslation", req],
    queryFn: () => apiClient.categoryService.GetTranslation(req),
    ...options,
  });
}

export function useCreateCategoryTranslation(
  options?: UseMutationOptions<
    catalogservicev1_CategoryTranslation,
    Error,
    catalogservicev1_CreateCategoryTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.CreateTranslation(data),
    ...options,
  });
}

export function useUpdateCategoryTranslation(
  options?: UseMutationOptions<
    catalogservicev1_CategoryTranslation,
    Error,
    catalogservicev1_UpdateCategoryTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.UpdateTranslation(data),
    ...options,
  });
}

export function useDeleteCategoryTranslation(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteCategoryTranslationRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.DeleteTranslation(data),
    ...options,
  });
}
