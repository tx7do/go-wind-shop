import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateProductsRequest,
  catalogservicev1_CreateProductRequest,
  catalogservicev1_CreateProductTranslationRequest,
  catalogservicev1_DeleteProductRequest,
  catalogservicev1_DeleteProductTranslationRequest,
  catalogservicev1_GetProductRequest,
  catalogservicev1_ListProductResponse,
  catalogservicev1_Product,
  catalogservicev1_ProductTranslation,
  catalogservicev1_UpdateProductRequest,
  catalogservicev1_UpdateProductTranslationRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 商品管理
// ==============================

export function useListProducts(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListProductResponse, Error>
) {
  return useQuery({
    queryKey: ["listProducts", query],
    queryFn: () => apiClient.productService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListProducts(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listProducts", params],
    queryFn: () => apiClient.productService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetProduct(
  req: catalogservicev1_GetProductRequest,
  options?: UseQueryOptions<catalogservicev1_Product, Error>
) {
  return useQuery({
    queryKey: ["getProduct", req],
    queryFn: () => apiClient.productService.Get(req),
    ...options,
  });
}

export async function fetchGetProduct(req: catalogservicev1_GetProductRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getProduct", req],
    queryFn: () => apiClient.productService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateProduct(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.productService.Create({
        data: { ...values } as unknown as catalogservicev1_Product,
      }),
    ...options,
  });
}

export function useUpdateProduct(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.productService.Update({
        id,
        data: { ...values } as unknown as catalogservicev1_Product,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteProduct(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteProductRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.Delete(data),
    ...options,
  });
}

export function useBatchCreateProducts(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateProductsRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.BatchCreate(data),
    ...options,
  });
}

// ---- 商品翻译 ----

export function useGetProductTranslation(
  req: catalogservicev1_GetProductRequest,
  options?: UseQueryOptions<catalogservicev1_ProductTranslation, Error>
) {
  return useQuery({
    queryKey: ["getProductTranslation", req],
    queryFn: () => apiClient.productService.GetTranslation(req),
    ...options,
  });
}

export function useCreateProductTranslation(
  options?: UseMutationOptions<
    catalogservicev1_ProductTranslation,
    Error,
    catalogservicev1_CreateProductTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.CreateTranslation(data),
    ...options,
  });
}

export function useUpdateProductTranslation(
  options?: UseMutationOptions<
    catalogservicev1_ProductTranslation,
    Error,
    catalogservicev1_UpdateProductTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.UpdateTranslation(data),
    ...options,
  });
}

export function useDeleteProductTranslation(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteProductTranslationRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.DeleteTranslation(data),
    ...options,
  });
}
