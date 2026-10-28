import {
  useMutation,
  useQuery,
  type UseMutationOptions,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import type {
  authenticationservicev1_GenerateCaptchaResponse,
  authenticationservicev1_LoginRequest,
  authenticationservicev1_LoginResponse,
} from "@/api/generated/admin/service/v1";
import { apiClient } from "@/api/client";
import { queryClient } from "@/plugins/vue-query";

// 直接导出函数，供非 Vue 上下文使用
export async function login(request: authenticationservicev1_LoginRequest) {
  return apiClient.authenticationService.Login(request);
}

export async function logout() {
  return apiClient.authenticationService.Logout({});
}

export async function generateCaptcha() {
  return apiClient.authenticationService.GenerateCaptcha({});
}

/**
 * 刷新令牌。端点在鉴权白名单内（AT 过期是刷新的前提），身份键（user_id/jti）
 * 由前端从持久化会话带回；RT 与 Redis 会话的值比对由后端完成真实鉴权。
 */
export interface RefreshTokenParams {
  refresh_token?: string;
  user_id?: number;
  jti?: string;
}

export async function refreshToken(request: RefreshTokenParams) {
  return apiClient.authenticationService.RefreshToken({
    grant_type: "refresh_token",
    refresh_token: request.refresh_token ?? "",
    user_id: request.user_id,
    jti: request.jti,
  });
}

// ------------------------------
// 登录（Mutation）
// ------------------------------
export function useLogin(
  options?: UseMutationOptions<
    authenticationservicev1_LoginResponse,
    Error,
    authenticationservicev1_LoginRequest
  >
) {
  return useMutation({
    mutationFn: (req) => login(req),
    ...options,
  });
}

// ------------------------------
// 登录（Mutation - GET）
// ------------------------------
export const loginMutation = queryClient.getMutationCache().build(queryClient, {
  mutationKey: ["login"],
  mutationFn: login,
  retry: 0,
});

// ------------------------------
// 登出（Mutation）
// ------------------------------
export function useLogout(options?: UseMutationOptions<{}, Error, {}>) {
  return useMutation({
    mutationFn: () => logout(),
    ...options,
  });
}

// ------------------------------
// 登出（Mutation - GET）
// ------------------------------
export const logoutMutation = queryClient.getMutationCache().build(queryClient, {
  mutationKey: ["logout"],
  mutationFn: logout,
  retry: 0,
});

// ------------------------------
// 刷新 Token（Mutation）
// ------------------------------
export function useRefreshToken(
  options?: UseMutationOptions<
    authenticationservicev1_LoginResponse,
    Error,
    RefreshTokenParams
  >
) {
  return useMutation({
    mutationFn: (req) => refreshToken(req),
    ...options,
  });
}

// ------------------------------
// 刷新 Token（Mutation - GET）
// ------------------------------
export const refreshTokenMutation = queryClient.getMutationCache().build(queryClient, {
  mutationKey: ["refreshToken"],
  mutationFn: refreshToken,
  retry: 0,
});

// ------------------------------
// 获取验证码（Query - GET）
// ------------------------------
export function useGenerateCaptcha(
  options?: UseQueryOptions<authenticationservicev1_GenerateCaptchaResponse, Error>
) {
  return useQuery({
    queryKey: ["captcha"],
    queryFn: () => generateCaptcha(),
    ...options,
  });
}

// ==============================================
// 获取验证码 【给 Store / 外部调用】不带 Hook 的方法
// ==============================================
export async function fetchGenerateCaptcha() {
  return queryClient.fetchQuery({
    queryKey: ["generateCaptcha"],
    queryFn: () => generateCaptcha(),
    staleTime: 0,
    retry: 0,
  });
}
