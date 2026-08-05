import { refreshToken as refreshTokenService } from "@/api/composables";
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
 * 刷新访问令牌
 * 使用 refresh_token 换取新的 access_token 和 refresh_token
 */
export async function refreshToken(): Promise<string> {
  const accessStore = useAccessStore();

  if (!accessStore.refreshToken) {
    await reauthenticate();
    return "";
  }

  try {
    const resp = await refreshTokenService(accessStore.refreshToken ?? "");

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
    accessStore.setRefreshToken(newRefreshToken ?? null);

    // token 已更新，重连 SSE 以使用新凭证
    reconnectSSEServer();

    return newAccessToken ?? "";
  } catch (error) {
    console.error("刷新 access token 失败", error);
    await reauthenticate();
    return "";
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
    accessStore.setIsAccessChecked(false);
    accessStore.setAccessCodes([]);

    if (preferences.app.loginExpiredMode === "modal" && accessStore.isAccessChecked) {
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
