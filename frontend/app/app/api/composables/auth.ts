import {
  useMutation,
  type UseMutationOptions,
} from '@tanstack/vue-query';
import type {
  authenticationservicev1_LoginRequest,
  authenticationservicev1_LoginResponse,
  authenticationservicev1_RegisterUserRequest,
  authenticationservicev1_RegisterUserResponse,
  SendResetCodeRequest,
  SendResetCodeResponse,
  ResetPasswordRequest,
} from '@/api/generated/app/service/v1';
import { apiClient } from '@/api/client';
import { queryClient } from '@/plugins/vue-query';

// ==============================
// Service 层方法（使用 apiClient）
// ==============================

/**
 * 登录
 */
export async function login(request: authenticationservicev1_LoginRequest): Promise<authenticationservicev1_LoginResponse> {
  return apiClient.authenticationService.Login(request);
}

/**
 * 注册
 */
export async function register(request: authenticationservicev1_RegisterUserRequest): Promise<authenticationservicev1_RegisterUserResponse> {
  return apiClient.authenticationService.Register(request);
}

/**
 * 登出
 */
export async function logout() {
  return apiClient.authenticationService.Logout({});
}

/**
 * 刷新 Token
 */
export async function refreshToken(refreshTokenValue: string) {
  return apiClient.authenticationService.RefreshToken({
    grant_type: 'password',
    refresh_token: refreshTokenValue ?? '',
  });
}

// ==============================
// 登录（Mutation）
// ==============================
export function useLogin(
  options?: UseMutationOptions<
    authenticationservicev1_LoginResponse,
    Error,
    authenticationservicev1_LoginRequest
  >,
) {
  return useMutation({
    mutationFn: (req) => login(req),
    ...options,
  });
}

/**
 * 注册（Mutation）
 */
export function useRegister(
  options?: UseMutationOptions<
    authenticationservicev1_RegisterUserResponse,
    Error,
    authenticationservicev1_RegisterUserRequest
  >,
) {
  return useMutation({
    mutationFn: (req) => register(req),
    ...options,
  });
}

/**
 * 登录【给 Store / 外部调用】不带 Hook 的方法
 */
export async function fetchLogin(request: authenticationservicev1_LoginRequest) {
  return queryClient.fetchQuery({
    queryKey: ['login', request],
    queryFn: () => login(request),
    retry: 0,
  });
}

// ==============================
// 登出（Mutation）
// ==============================
export function useLogout(options?: UseMutationOptions<{}, Error, void>) {
  return useMutation({
    mutationFn: () => logout(),
    ...options,
  });
}

/**
 * 登出【给 Store / 外部调用】不带 Hook 的方法
 */
export async function fetchLogout() {
  return queryClient.fetchQuery({
    queryKey: ['logout'],
    queryFn: () => logout(),
    retry: 0,
  });
}

// ==============================
// 刷新 Token（Mutation）
// ==============================
export function useRefreshToken(
  options?: UseMutationOptions<
    authenticationservicev1_LoginResponse,
    Error,
    string
  >,
) {
  return useMutation({
    mutationFn: (token) => refreshToken(token),
    ...options,
  });
}

/**
 * 刷新 Token【给 Store / 外部调用】不带 Hook 的方法
 */
export async function fetchRefreshToken(refreshTokenValue: string) {
  return queryClient.fetchQuery({
    queryKey: ['refreshToken', refreshTokenValue],
    queryFn: () => refreshToken(refreshTokenValue),
    retry: 0,
  });
}

// ==============================
// 找回密码：发送验证码 / 重置密码
// ==============================

/**
 * 发送密码重置验证码到邮箱
 * 返回脱敏邮箱预览与有效期，用于前端提示与倒计时。
 */
export async function sendResetCode(email: string): Promise<SendResetCodeResponse> {
  const request: SendResetCodeRequest = { email };
  return apiClient.authenticationService.SendResetCode(request);
}
export function useSendResetCode(
  options?: UseMutationOptions<SendResetCodeResponse, Error, string>,
) {
  return useMutation({
    mutationFn: (email) => sendResetCode(email),
    ...options,
  });
}

/**
 * 校验验证码并重置密码
 */
export async function resetPassword(email: string, code: string, newPassword: string): Promise<void> {
  const request: ResetPasswordRequest = { email, code, newPassword };
  await apiClient.authenticationService.ResetPassword(request);
}
export function useResetPassword(
  options?: UseMutationOptions<void, Error, { email: string; code: string; newPassword: string }>,
) {
  return useMutation({
    mutationFn: ({ email, code, newPassword }) => resetPassword(email, code, newPassword),
    ...options,
  });
}
