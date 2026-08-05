import {getActivePinia} from 'pinia'

/**
 * 重置当前 pinia 实例下所有 store 的状态。
 *
 * 注意：pinia 实例由 @pinia/nuxt 创建并注入，持久化由
 * pinia-plugin-persistedstate/nuxt module 接管（cookie storage），
 * 不再需要手动 setupStore/createPersistedState。
 */
export function resetAllStores() {
    const pinia = getActivePinia() as any
    if (!pinia) {
        console.error('Pinia is not installed')
        return
    }
    const allStores = pinia._s
    if (!allStores) return
    for (const [_key, store] of allStores) {
        store.$reset()
    }
}
