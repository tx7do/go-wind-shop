import type { RouteRecordRaw } from "vue-router";

import { acceptHMRUpdate, defineStore } from "pinia";
import type { MenuRecordRaw } from "@/core/router/types";

import { parseJwtPayload } from "@/utils/crypto";

/**
 * @zh_CN 访问令牌类型
 */
type AccessToken = null | string;

/**
 * @zh_CN 访问权限相关状态定义
 */
interface AccessState {
  /**
   * 权限码
   */
  accessCodes: string[];
  /**
   * 字段权限隐藏字段集（"资源.字段" 串，如 "User.email"，来自 GetMyPermissionCode）
   */
  hiddenFields: string[];
  /**
   * 可访问的菜单列表
   */
  accessMenus: MenuRecordRaw[];
  /**
   * 可访问的路由列表
   */
  accessRoutes: RouteRecordRaw[];
  /**
   * 登录 accessToken
   */
  accessToken: AccessToken;
  /**
   * accessToken 过期时间戳
   */
  accessTokenExpireTime?: number;
  /**
   * 是否已经检查过权限
   */
  isAccessChecked: boolean;
  /**
   * 登录是否过期
   */
  loginExpired: boolean;

  /**
   * 登录 refreshToken（随会话持久化，刷新页面后凭它静默续期）
   */
  refreshToken: AccessToken;

  /**
   * refreshToken 过期时间戳
   */
  refreshTokenExpireTime?: number;

  /**
   * 会话身份键：当前令牌对归属的 userId（刷新端点的查找键，非鉴权凭据）
   */
  sessionUserId?: number;

  /**
   * 会话身份键：当前令牌对的 jti。RT 按 (clientType, userId, jti) 定位且每次
   * 刷新轮换出新 jti，因此刷新成功后必须同步更新，否则下一次刷新查不到会话。
   */
  sessionJti?: string;

  /**
   * MFA 挑战操作标识：登录密码校验通过但需二次验证时，后端返回。
   * 非空表示当前处于 MFA 挑战待验证阶段，前端据此跳挑战页。
   */
  mfaOperationId: string | null;
}

/**
 * @zh_CN 访问权限相关状态管理
 */
export const useAccessStore = defineStore("core-access", {
  actions: {
    $reset() {
      this.accessToken = null;
      this.accessCodes = [];
      this.hiddenFields = [];
      this.accessMenus = [];
      this.accessRoutes = [];
      this.isAccessChecked = false;
      this.loginExpired = false;
      this.mfaOperationId = null;
      this.refreshToken = null;
      this.refreshTokenExpireTime = undefined;
      this.accessTokenExpireTime = undefined;
      this.sessionUserId = undefined;
      this.sessionJti = undefined;
    },
    /**
     * @zh_CN 检查 accessToken 是否过期
     */
    checkAccessTokenExpired(): boolean {
      if (!this.accessTokenExpireTime) {
        return true;
      }
      const now = Date.now();
      return now >= this.accessTokenExpireTime;
    },
    setAccessCodes(codes: string[]) {
      this.accessCodes = codes;
    },
    setHiddenFields(fields: string[]) {
      this.hiddenFields = fields;
    },
    setMfaOperationId(operationId: string | null) {
      this.mfaOperationId = operationId;
    },
    setRefreshToken(token: AccessToken) {
      this.refreshToken = token;
    },
    setRefreshTokenExpireTime(refreshTokenExpireTime: number) {
      this.refreshTokenExpireTime = refreshTokenExpireTime;
    },
    setAccessMenus(menus: MenuRecordRaw[]) {
      this.accessMenus = menus;
    },
    setAccessRoutes(routes: RouteRecordRaw[]) {
      this.accessRoutes = routes;
    },
    setAccessToken(token: AccessToken) {
      this.accessToken = token;
    },
    setAccessTokenExpireTime(accessTokenExpireTime: number) {
      this.accessTokenExpireTime = accessTokenExpireTime;
    },
    /**
     * 记录刷新端点所需的身份键（userId/jti）。jti 随每次刷新轮换，调用方须在
     * 登录成功与刷新成功后都调用本函数保持同步。
     */
    setSessionIdentity(userId: number, jti: string) {
      this.sessionUserId = userId > 0 ? userId : undefined;
      this.sessionJti = jti || undefined;
    },
    /**
     * @zh_CN 从 access token（JWT）载荷提取身份键的兜底读取
     */
    applySessionIdentityFromToken(token: string) {
      // 本后端签发的 AT 载荷键为 uid/jti（见 pkg/jwt claims 映射），user_id 仅作兼容兜底
      const payload = parseJwtPayload(token);
      this.setSessionIdentity(Number(payload?.uid ?? payload?.user_id ?? 0), String(payload?.jti ?? ""));
    },
    setIsAccessChecked(isAccessChecked: boolean) {
      this.isAccessChecked = isAccessChecked;
    },

    setLoginExpired(loginExpired: boolean) {
      this.loginExpired = loginExpired;
    },
  },
  persist: {
    // Token（access/refresh 及其过期时间）与会话身份键（sessionUserId/sessionJti）
    // 必须持久化到 localStorage：本项目后端无 HttpOnly cookie 续期链路（上游
    // go-wind-admin 的 Secure cookie 方案未移植，core 的 RT 是不透明随机串，
    // 依赖 Redis 会话而非自描述 JWT），刷新页面后唯一可用的会话凭证就是本地
    // 持久化的 token——不持久化则守卫看到 accessToken 为 null，一刷新即被踢
    // 回登录页。accessCodes/hiddenFields 非 secret，持久化避免刷新后重拉。
    // mfaOperationId 保持内存态：刷新即丢弃未完成的二次验证，要求重新登录。
    pick: [
      "accessToken",
      "accessTokenExpireTime",
      "refreshToken",
      "refreshTokenExpireTime",
      "sessionUserId",
      "sessionJti",
      "accessCodes",
      "hiddenFields",
    ],
  },
  state: (): AccessState => ({
    accessCodes: [],
    hiddenFields: [],
    accessMenus: [],
    accessRoutes: [],
    accessToken: null,
    accessTokenExpireTime: undefined,
    isAccessChecked: false,
    loginExpired: false,
    mfaOperationId: null,
    refreshToken: null,
    refreshTokenExpireTime: undefined,
    // 必须在初始化器中声明：options store 的 action 赋值只会同步已声明字段到
    // pinia.state（$state），未声明字段的赋值仅落在 store 实例上，持久化插件
    // 订阅的是 $state，声明缺失会导致这两个键永远写不进 localStorage。
    sessionUserId: undefined,
    sessionJti: undefined,
  }),
});

// 解决热更新问题
const hot = import.meta.hot;
if (hot) {
  hot.accept(acceptHMRUpdate(useAccessStore, hot));
}
