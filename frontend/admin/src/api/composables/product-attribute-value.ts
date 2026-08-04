import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateProductAttributeValuesRequest,
  catalogservicev1_CreateProductAttributeValueRequest,
  catalogservicev1_CreateProductAttributeValueTranslationRequest,
  catalogservicev1_DeleteProductAttributeValueRequest,
  catalogservicev1_DeleteProductAttributeValueTranslationRequest,
  catalogservicev1_GetProductAttributeValueRequest,
  catalogservicev1_ListProductAttributeValueResponse,
  catalogservicev1_ProductAttributeValue,
  catalogservicev1_ProductAttributeValueTranslation,
  catalogservicev1_UpdateProductAttributeValueRequest,
  catalogservicev1_UpdateProductAttributeValueTranslationRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 商品属性值管理
// ==============================

export function useListProductAttributeValues(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListProductAttributeValueResponse, Error>
) {
  return useQuery({
    queryKey: ["listProductAttributeValues", query],
    queryFn: () => apiClient.productAttributeValueService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListProductAttributeValues(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listProductAttributeValues", params],
    queryFn: () => apiClient.productAttributeValueService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetProductAttributeValue(
  req: catalogservicev1_GetProductAttributeValueRequest,
  options?: UseQueryOptions<catalogservicev1_ProductAttributeValue, Error>
) {
  return useQuery({
    queryKey: ["getProductAttributeValue", req],
    queryFn: () => apiClient.productAttributeValueService.Get(req),
    ...options,
  });
}

export async function fetchGetProductAttributeValue(
  req: catalogservicev1_GetProductAttributeValueRequest
) {
  return queryClient.fetchQuery({
    queryKey: ["getProductAttributeValue", req],
    queryFn: () => apiClient.productAttributeValueService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateProductAttributeValue(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.productAttributeValueService.Create({
        data: { ...values } as unknown as catalogservicev1_ProductAttributeValue,
      }),
    ...options,
  });
}

export function useUpdateProductAttributeValue(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.productAttributeValueService.Update({
        id,
        data: { ...values } as unknown as catalogservicev1_ProductAttributeValue,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteProductAttributeValue(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteProductAttributeValueRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeValueService.Delete(data),
    ...options,
  });
}

export function useBatchCreateProductAttributeValues(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateProductAttributeValuesRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeValueService.BatchCreate(data),
    ...options,
  });
}

// ---- 商品属性值翻译 ----

export function useGetProductAttributeValueTranslation(
  req: catalogservicev1_GetProductAttributeValueRequest,
  options?: UseQueryOptions<catalogservicev1_ProductAttributeValueTranslation, Error>
) {
  return useQuery({
    queryKey: ["getProductAttributeValueTranslation", req],
    queryFn: () => apiClient.productAttributeValueService.GetTranslation(req),
    ...options,
  });
}

export function useCreateProductAttributeValueTranslation(
  options?: UseMutationOptions<
    catalogservicev1_ProductAttributeValueTranslation,
    Error,
    catalogservicev1_CreateProductAttributeValueTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeValueService.CreateTranslation(data),
    ...options,
  });
}

export function useUpdateProductAttributeValueTranslation(
  options?: UseMutationOptions<
    catalogservicev1_ProductAttributeValueTranslation,
    Error,
    catalogservicev1_UpdateProductAttributeValueTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeValueService.UpdateTranslation(data),
    ...options,
  });
}

export function useDeleteProductAttributeValueTranslation(
  options?: UseMutationOptions<
    {},
    Error,
    catalogservicev1_DeleteProductAttributeValueTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeValueService.DeleteTranslation(data),
    ...options,
  });
}
