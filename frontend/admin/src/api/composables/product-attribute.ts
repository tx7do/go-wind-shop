import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateProductAttributesRequest,
  catalogservicev1_CreateProductAttributeRequest,
  catalogservicev1_CreateProductAttributeTranslationRequest,
  catalogservicev1_DeleteProductAttributeRequest,
  catalogservicev1_DeleteProductAttributeTranslationRequest,
  catalogservicev1_GetProductAttributeRequest,
  catalogservicev1_ListProductAttributeResponse,
  catalogservicev1_ProductAttribute,
  catalogservicev1_ProductAttributeTranslation,
  catalogservicev1_UpdateProductAttributeRequest,
  catalogservicev1_UpdateProductAttributeTranslationRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 商品属性管理
// ==============================

export function useListProductAttributes(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListProductAttributeResponse, Error>
) {
  return useQuery({
    queryKey: ["listProductAttributes", query],
    queryFn: () => apiClient.productAttributeService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListProductAttributes(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listProductAttributes", params],
    queryFn: () => apiClient.productAttributeService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetProductAttribute(
  req: catalogservicev1_GetProductAttributeRequest,
  options?: UseQueryOptions<catalogservicev1_ProductAttribute, Error>
) {
  return useQuery({
    queryKey: ["getProductAttribute", req],
    queryFn: () => apiClient.productAttributeService.Get(req),
    ...options,
  });
}

export async function fetchGetProductAttribute(
  req: catalogservicev1_GetProductAttributeRequest
) {
  return queryClient.fetchQuery({
    queryKey: ["getProductAttribute", req],
    queryFn: () => apiClient.productAttributeService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateProductAttribute(
  options?: UseMutationOptions<{}, Error, Record<string, any>>
) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.productAttributeService.Create({
        data: { ...values } as unknown as catalogservicev1_ProductAttribute,
      }),
    ...options,
  });
}

export function useUpdateProductAttribute(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.productAttributeService.Update({
        id,
        data: { ...values } as unknown as catalogservicev1_ProductAttribute,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteProductAttribute(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteProductAttributeRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeService.Delete(data),
    ...options,
  });
}

export function useBatchCreateProductAttributes(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateProductAttributesRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeService.BatchCreate(data),
    ...options,
  });
}

// ---- 商品属性翻译 ----

export function useGetProductAttributeTranslation(
  req: catalogservicev1_GetProductAttributeRequest,
  options?: UseQueryOptions<catalogservicev1_ProductAttributeTranslation, Error>
) {
  return useQuery({
    queryKey: ["getProductAttributeTranslation", req],
    queryFn: () => apiClient.productAttributeService.GetTranslation(req),
    ...options,
  });
}

export function useCreateProductAttributeTranslation(
  options?: UseMutationOptions<
    catalogservicev1_ProductAttributeTranslation,
    Error,
    catalogservicev1_CreateProductAttributeTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeService.CreateTranslation(data),
    ...options,
  });
}

export function useUpdateProductAttributeTranslation(
  options?: UseMutationOptions<
    catalogservicev1_ProductAttributeTranslation,
    Error,
    catalogservicev1_UpdateProductAttributeTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeService.UpdateTranslation(data),
    ...options,
  });
}

export function useDeleteProductAttributeTranslation(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteProductAttributeTranslationRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productAttributeService.DeleteTranslation(data),
    ...options,
  });
}
