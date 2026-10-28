import { refreshToken as refreshTokenService } from "@/api/composables";
import type { authenticationservicev1_LoginResponse } from "@/api/generated/admin/service/v1";
import { LOGIN_PATH } from "@/constants";
import { preferences } from "@/core/preferences";
import { globalSSEClient } from "@/core/transport/sse";
import { queryClient } from "@/plugins/vue-query";
import { resetAllStores, useAccessStore } from "@/stores";
import { router } from "@/router";

// ==============================
// 常量
// ==============================

/** Access Token 刷新间隔（1.5 小时） */
const ACCESS_TOKEN_REFRESH_INTERVAL = 90 * 60 * 1000;
/** Refresh Token 刷新间隔（12 小时） */
const REFRESH_TOKEN_REFRESH_INTERVAL = 12 * 60 * 60 * 1000;

/** 在 access token 到期前多久开始刷新 */
const SAFETY_BUFFER_MS = 5 * 60 * 1000;
/** 最小刷新间隔（避免立即重试风暴） */
const MIN_INTERVAL_MS = 3 * 1000;

// ==============================
// 模块级状态（单例）
// ==============================

let refreshTimer: null | ReturnType<typeof setTimeout> = null;
let refreshCallback: null | RefreshTokenFunc = null;
let isReauthenticating = false;

type RefreshTokenFunc = () => Promise<string> | string;

// ==============================
// 核心：刷新 Access Token
// ==============================

/**
 * 调刷新端点。端点在 BFF 鉴权白名单内（AT 过期是刷新的前提），身份键
 * （user_id/jti/refresh_token）全部从持久化的 access store 带回。
 */
async function requestRefreshToken(): Promise<authenticationservicev1_LoginResponse> {
  const accessStore = useAccessStore();
  return await refreshTokenService({
    grant_type: "refresh_token",
    refresh_token: accessStore.refreshToken ?? "",
    user_id: accessStore.sessionUserId,
    jti: accessStore.sessionJti,
  } as any);
}

/**
 * 将刷新响应应用到 store：更新 access token 与过期时间。
 * 注意：每次刷新 RT 都会轮换出新 jti（旧令牌对被原子吊销），必须从新 AT 的
 * JWT 载荷同步 sessionUserId/sessionJti，否则下一次刷新按旧 jti 查不到会话。
 */
function applyRefreshResponse(resp: any): string {
  const accessStore = useAccessStore();

  const newAccessToken = (resp as any).access_token;
  const newRefreshToken = (resp as any).refresh_token;

  let expiresIn = (resp as any).expires_in;
  let refreshExpiresIn = (resp as any).refresh_expires_in;

  const expiresInSec = Number(expiresIn);
  expiresIn =
    !Number.isFinite(expiresInSec) || expiresInSec <= 0
      ? Date.now() + ACCESS_TOKEN_REFRESH_INTERVAL
      : Date.now() + Math.floor(expiresInSec * 1000);

  const refreshExpiresInSec = Number(refreshExpiresIn);
  refreshExpiresIn =
    !Number.isFinite(refreshExpiresInSec) || refreshExpiresInSec <= 0
      ? Date.now() + REFRESH_TOKEN_REFRESH_INTERVAL
      : Date.now() + Math.floor(refreshExpiresInSec * 1000);

  accessStore.setAccessTokenExpireTime(expiresIn);
  accessStore.setRefreshTokenExpireTime(refreshExpiresIn);

  accessStore.setAccessToken(newAccessToken ?? null);
  accessStore.setRefreshToken(newRefreshToken ?? accessStore.refreshToken);
  if (newAccessToken) {
    accessStore.applySessionIdentityFromToken(newAccessToken);
  }

  return newAccessToken ?? "";
}

/**
 * 刷新访问令牌（定时器 / 401 拦截器入口）
 * 失败触发重新认证（跳登录页）
 */
export async function refreshToken(): Promise<string> {
  const accessStore = useAccessStore();

  if (!accessStore.refreshToken) {
    await reauthenticate();
    return "";
  }

  try {
    const resp = await requestRefreshToken();

    const newAccessToken = applyRefreshResponse(resp);

    // token 已更新，重连 SSE 以使用新凭证
    reconnectSSEServer();

    return newAccessToken;
  } catch (error) {
    console.error("刷新 access token 失败", error);
    await reauthenticate();
    return "";
  }
}

/**
 * 静默续期：用持久化的 RT 换新 AT，成功返回 true；失败仅返回 false，
 * 不触发 reauthenticate/跳转——供路由守卫在 AT 过期时先行尝试，
 * 由守卫自行决定降级动作（守卫内并发 router.replace 会互相打断导航）。
 */
export async function attemptSilentRefresh(): Promise<boolean> {
  const accessStore = useAccessStore();

  // 身份键不齐（旧版本登录留下的会话）无法定位 Redis 会话，直接放弃
  if (!accessStore.refreshToken || !accessStore.sessionJti || !accessStore.sessionUserId) {
    return false;
  }
  // RT 已过有效期（本地时钟近似），不再浪费一次必败请求
  if (accessStore.refreshTokenExpireTime && Date.now() >= accessStore.refreshTokenExpireTime) {
    return false;
  }

  try {
    const resp = await requestRefreshToken();
    const newAccessToken = applyRefreshResponse(resp);
    if (!newAccessToken) {
      return false;
    }
    reconnectSSEServer();
    return true;
  } catch (error) {
    console.warn("[TokenRefresh] 静默续期失败:", error);
    return false;
  }
}

// ==============================
// 核心：重新认证
// ==============================

/**
 * 重新认证
 * 当 refresh token 失效时触发，根据配置决定弹窗或直接跳转登录页
 */
export async function reauthenticate(): Promise<void> {
  if (isReauthenticating) {
    return;
  }
  isReauthenticating = true;

  try {
    console.warn("Access token or refresh token is invalid or expired.");

    stopRefreshTimer();

    const accessStore = useAccessStore();
    accessStore.setAccessToken(null);
    accessStore.setRefreshToken(null);
    // 注意：setIsAccessChecked(false) 之前必须先读出原值用于下面的 modal 判定，
    // 否则下方 accessStore.isAccessChecked 永远是 false，modal 模式恒不触发。
    const wasAccessChecked = accessStore.isAccessChecked;
    accessStore.setIsAccessChecked(false);
    accessStore.setAccessCodes([]);

    if (preferences.app.loginExpiredMode === "modal" && wasAccessChecked) {
      accessStore.setLoginExpired(true);
    } else {
      await logoutToLoginPage();
    }
  } finally {
    isReauthenticating = false;
  }
}

// ==============================
// 登出跳转
// ==============================

/**
 * 停止刷新定时器 → 重置 Store 清凭证 → 关闭 SSE → 跳转登录页 → 跳转完成后清理 queryClient 缓存
 *
 * 顺序很关键：
 *  - resetAllStores（清 accessToken/refreshToken）必须在跳转前执行，否则登录页守卫
 *    （auth.guard.ts：已登录访问 /login 会被弹回首页）会把用户挡回业务页。
 *  - queryClient.clear() 必须在跳转完成后执行。若在源组件（如 /analytics）仍挂载时
 *    clear，TanStack Query 会因缓存被清空而对仍订阅的 useQuery 立即重建 query 并
 *    重新 fetch（queryObserver.shouldFetchOptionally：query!==prevQuery && isStale
 *    均为 true）。这些重发请求带着已清空的空 token 打后端 → 又触发 401 → reauthenticate
 *    循环，并与本次 router.replace(/login) 产生并发导航互相取消，污染后续登录跳转。
 *    跳转完成后再 clear，源组件已卸载、observer 已 unsubscribe，不会触发重发。
 * @param redirect 是否携带回跳地址
 */
export async function logoutToLoginPage(redirect: boolean = true): Promise<void> {
  console.log("logoutToLoginPage");

  stopRefreshTimer();

  // 清凭证必须在跳转前：登录页守卫依 accessToken 判断是否已登录
  resetAllStores();

  const accessStore = useAccessStore();
  accessStore.setLoginExpired(false);

  globalSSEClient.close();

  // 记录回跳地址：必须在跳转前读取，跳转后 currentRoute 即变为登录页
  const fromFullPath = router.currentRoute.value.fullPath;
  if (fromFullPath === LOGIN_PATH) return;

  await router.replace({
    path: LOGIN_PATH,
    query: redirect
      ? {
          redirect: encodeURIComponent(fromFullPath),
        }
      : {},
  });

  // 跳转完成、源组件卸载后再清 queryClient 缓存，避免触发 useQuery 重发风暴
  queryClient.clear();
}

// ==============================
// 定时刷新管理
// ==============================

function computeNextInterval(): number {
  const accessStore = useAccessStore();
  const now = Date.now();

  const accessExpire = accessStore.accessTokenExpireTime ?? 0;
  const refreshExpire = accessStore.refreshTokenExpireTime ?? 0;

  const refreshRemaining = refreshExpire - now;
  if (refreshExpire && refreshRemaining <= SAFETY_BUFFER_MS) {
    return MIN_INTERVAL_MS;
  }

  const accessRemaining = accessExpire - now;
  if (!accessExpire || accessRemaining <= 0) {
    return MIN_INTERVAL_MS;
  }

  if (accessRemaining <= SAFETY_BUFFER_MS) {
    return MIN_INTERVAL_MS;
  }

  return Math.floor(
    Math.max(
      MIN_INTERVAL_MS,
      Math.min(ACCESS_TOKEN_REFRESH_INTERVAL, (accessRemaining - SAFETY_BUFFER_MS) * 0.8)
    )
  );
}

function _startRefreshTimer(cb?: RefreshTokenFunc): void {
  stopRefreshTimer();

  if (cb) refreshCallback = cb;
  if (!refreshCallback) return;

  const schedule = async () => {
    try {
      const accessStore = useAccessStore();
      const now = Date.now();
      const refreshExpire = accessStore.refreshTokenExpireTime ?? 0;
      if (!accessStore.refreshToken) {
        await reauthenticate();
        return;
      }
      if (refreshExpire && refreshExpire - now <= SAFETY_BUFFER_MS) {
        await reauthenticate();
        return;
      }

      await refreshCallback?.();
    } catch (error) {
      console.error("refreshToken 定时刷新失败", error);
    } finally {
      if (refreshCallback) {
        const next = computeNextInterval();
        refreshTimer = globalThis.setTimeout(schedule, next);
      }
    }
  };

  refreshTimer = globalThis.setTimeout(schedule, computeNextInterval());
}

export function stopRefreshTimer(): void {
  if (refreshTimer !== null) {
    globalThis.clearTimeout(refreshTimer);
    refreshTimer = null;
    refreshCallback = null;
  }
}

export function startRefreshTimer(): void {
  _startRefreshTimer(refreshToken);
}

// ==============================
// SSE 连接
// ==============================

export function connectSSEServer(): void {
  const accessStore = useAccessStore();

  const token = accessStore.accessToken ?? "";
  const targetSseUrl = `${import.meta.env.VITE_APP_SSE_URL}?stream=${encodeURIComponent(token)}`;

  globalSSEClient.setHeaders({ Authorization: `Bearer ${token}` });
  globalSSEClient.connect(targetSseUrl);
}

/**
 * 使用新 token 重连 SSE（关闭旧连接 → 更新凭证 → 重新连接）
 * 适用于 token 刷新后 SSE 连接携带的凭证已过期的场景
 */
function reconnectSSEServer(): void {
  const accessStore = useAccessStore();

  const token = accessStore.accessToken ?? "";
  const targetSseUrl = `${import.meta.env.VITE_APP_SSE_URL}?stream=${encodeURIComponent(token)}`;

  globalSSEClient.setHeaders({ Authorization: `Bearer ${token}` });
  globalSSEClient.reconnect(targetSseUrl);
}
