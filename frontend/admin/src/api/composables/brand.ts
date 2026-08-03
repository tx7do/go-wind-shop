import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  catalogservicev1_BatchCreateBrandsRequest,
  catalogservicev1_Brand,
  catalogservicev1_BrandTranslation,
  catalogservicev1_CreateBrandRequest,
  catalogservicev1_CreateBrandTranslationRequest,
  catalogservicev1_DeleteBrandRequest,
  catalogservicev1_DeleteBrandTranslationRequest,
  catalogservicev1_GetBrandRequest,
  catalogservicev1_ListBrandResponse,
  catalogservicev1_UpdateBrandRequest,
  catalogservicev1_UpdateBrandTranslationRequest,
} from "@/api/generated/admin/service/v1";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// ==============================
// 品牌管理
// ==============================

export function useListBrands(
  query: PaginationQuery,
  options?: UseQueryOptions<catalogservicev1_ListBrandResponse, Error>
) {
  return useQuery({
    queryKey: ["listBrands", query],
    queryFn: () => apiClient.brandService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListBrands(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listBrands", params],
    queryFn: () => apiClient.brandService.List(params.toRawParams()),
    staleTime: 0,
    retry: 0,
  });
}

export function useGetBrand(
  req: catalogservicev1_GetBrandRequest,
  options?: UseQueryOptions<catalogservicev1_Brand, Error>
) {
  return useQuery({
    queryKey: ["getBrand", req],
    queryFn: () => apiClient.brandService.Get(req),
    ...options,
  });
}

export async function fetchGetBrand(req: catalogservicev1_GetBrandRequest) {
  return queryClient.fetchQuery({
    queryKey: ["getBrand", req],
    queryFn: () => apiClient.brandService.Get(req),
    staleTime: 0,
    retry: 0,
  });
}

export function useCreateBrand(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) =>
      apiClient.brandService.Create({ data: { ...values } as unknown as catalogservicev1_Brand }),
    ...options,
  });
}

export function useUpdateBrand(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.brandService.Update({
        id,
        data: { ...values } as unknown as catalogservicev1_Brand,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteBrand(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteBrandRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.brandService.Delete(data),
    ...options,
  });
}

export function useBatchCreateBrands(
  options?: UseMutationOptions<{}, Error, catalogservicev1_BatchCreateBrandsRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.brandService.BatchCreate(data),
    ...options,
  });
}

// ---- 品牌翻译 ----

export function useGetBrandTranslation(
  req: catalogservicev1_GetBrandRequest,
  options?: UseQueryOptions<catalogservicev1_BrandTranslation, Error>
) {
  return useQuery({
    queryKey: ["getBrandTranslation", req],
    queryFn: () => apiClient.brandService.GetTranslation(req),
    ...options,
  });
}

export function useCreateBrandTranslation(
  options?: UseMutationOptions<
    catalogservicev1_BrandTranslation,
    Error,
    catalogservicev1_CreateBrandTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.brandService.CreateTranslation(data),
    ...options,
  });
}

export function useUpdateBrandTranslation(
  options?: UseMutationOptions<
    catalogservicev1_BrandTranslation,
    Error,
    catalogservicev1_UpdateBrandTranslationRequest
  >
) {
  return useMutation({
    mutationFn: (data) => apiClient.brandService.UpdateTranslation(data),
    ...options,
  });
}

export function useDeleteBrandTranslation(
  options?: UseMutationOptions<{}, Error, catalogservicev1_DeleteBrandTranslationRequest>
) {
  return useMutation({
    mutationFn: (data) => apiClient.brandService.DeleteTranslation(data),
    ...options,
  });
}
