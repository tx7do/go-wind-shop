<script setup lang="ts">
import { Toaster } from '@/components/ui/sonner'

const isDev = import.meta.dev
const { t } = useI18n()
const config = useRuntimeConfig()

/**
 * 主题防闪：SSR 出的 <html> 不带 class，:root 亮色 token 会先画一帧，暗色用户看到白闪。
 *
 * 为什么在 app.vue 的 useHead 而不是 app/app.html：app.html 用的是 Nuxt 2 的
 * {{ HTML_ATTRS }} 占位符，Nuxt 4 不加载它（服务端 HTML 里没有该文件独有的 favicon
 * link），那段脚本一直是死代码。以前暗色靠 main.css 里 @media (prefers-color-scheme:
 * dark) 的副本兜首帧，但那份副本与 .dark 的 --card/--border 已经漂移成两套暗色，
 * 现已删除 —— 这段脚本是暗色首帧的唯一防线。
 * 改这里必须同步改 core/preferences/update-css-variables.ts 的 auto 分支。
 */
const THEME_NO_FLASH = `(function(){try{var P='gowind-shop';var keys=[P+':preferences-theme',P+':preferences'];var mode=null;for(var i=0;i<keys.length;i++){var raw=localStorage.getItem(keys[i]);if(!raw)continue;try{var d=JSON.parse(raw);d=(d&&typeof d==='object'&&'data'in d)?d.data:d;if(i===0){if(d==='dark'||d==='light'||d==='auto'){mode=d;break}}else{var m=d&&d.theme&&d.theme.mode;if(m==='dark'||m==='light'||m==='auto'){mode=m;break}}}catch(e){}}if(mode!=='dark'&&mode!=='light'){mode=(window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches)?'dark':'light'}var r=document.documentElement;r.classList.remove('dark','light');r.classList.add(mode)}catch(e){}})();`

useHead({
  titleTemplate: (title) => title ? `${title} - ${config.public.appTitle}` : config.public.appTitle,
  script: [{ innerHTML: THEME_NO_FLASH }],
})
</script>

<template>
  <div>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <Toaster />
    <ClientOnly>
      <VueQueryDevtools v-if="isDev" />
    </ClientOnly>
  </div>
</template>
