/**
 * 路由级鉴权中间件
 *
 * 挂在账户中心布局下的所有页面（通过 definePageMeta.middleware: 'auth'）。
 * 未登录或 token 已过期时重定向到登录页，取代各页面内重复的 isLogin 自检。
 */
export default defineNuxtRouteMiddleware(() => {
  const accessStore = useAccessStore()
  const token = accessStore.accessToken
  const ok = !!token?.value && !accessStore.loginExpired

  if (!ok) {
    const localePath = useLocalePath()
    return navigateTo(localePath('/login'))
  }
})
