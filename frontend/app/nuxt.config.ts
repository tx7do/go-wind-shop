// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from '@tailwindcss/vite'
import { writeFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineNuxtConfig({
    compatibilityDate: '2025-07-15',
    devtools: {enabled: true},

    // 静态站点生成配置
    ssr: true,
    nitro: {
        prerender: {
            // 预渲染静态页面，动态路由页面由客户端渲染。
            // 两个语言壳必须真实产出（.output/public/<locale>/index.html），
            // nginx 的 SPA fallback（try_files … /<locale>/index.html）依赖它们存在。
            // 不能同时预渲染 '/'：i18n strategy=prefix 下 '/' 302 到 '/zh-CN/'，
            // Nitro 会把跳转目标重新入队，与显式的 '/zh-CN/' 任务并发渲染同一文件，
            // 两个写入方交错曾产出尾缀 '/html>' 的损坏 HTML（页面底部渲染出该文本）。
            // 根路径重定向页由下方 prerender:done 钩子生成，无需预渲染 '/'。
            routes: ['/zh-CN/', '/en-US/'],
            crawlLinks: true,
        },
        // SSG 构建后自动生成根路径 index.html 重定向页
        hooks: {
            'prerender:done'() {
                const outputDir = resolve(__dirname, '.output', 'public')
                writeFileSync(resolve(outputDir, 'index.html'), [
                    '<!DOCTYPE html>',
                    '<html><head>',
                    '<meta charset="utf-8">',
                    '<meta http-equiv="refresh" content="0;url=/zh-CN/">',
                    '<title>Redirecting...</title>',
                    '</head><body>',
                    '<script>location.replace("/zh-CN/"+location.search+location.hash)</script>',
                    '<noscript><meta http-equiv="refresh" content="0;url=/zh-CN/"></noscript>',
                    '<p><a href="/zh-CN/">Click here</a></p>',
                    '</body></html>',
                ].join('\n'))
            },
        },
    },

    // SPA fallback：未预渲染的路由返回 200 + index.html（由 nginx 配合）
    app: {
        baseURL: '/',
        buildAssetsDir: '_nuxt',
    },
    components: [
        {
            path: '~/components',
            extensions: ['.vue'],
            ignore: ['**/index.ts'],
        },
    ],
    css: [
        '~/assets/css/main.css',
        'katex/dist/katex.min.css',
    ],
    modules: [
        '@nuxtjs/i18n',
        '@pinia/nuxt',
        'pinia-plugin-persistedstate/nuxt',
        'shadcn-nuxt'
    ],
    vite: {
        plugins: [
            tailwindcss(),
        ],
    },
    i18n: {
        langDir: '../locales',
        locales: [
            {code: 'zh-CN', iso: 'zh-CN', name: '中文', file: 'zh-CN/index.ts'},
            {code: 'en-US', iso: 'en-US', name: 'English', file: 'en-US/index.ts'}
        ],
        defaultLocale: 'zh-CN',
        strategy: 'prefix',
        detectBrowserLanguage: false,
    },
    typescript: {
        tsConfig: {
            compilerOptions: {
                resolveJsonModule: true,
            },
        },
    },
    runtimeConfig: {
        public: {
            apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:8000',
            // SSE 网关地址。必须在此声明键位，NUXT_PUBLIC_SSE_URL 才会被映射进来并
            // 编进产物（AssistantChat 读 runtimeConfig.public.sseUrl）。
            sseUrl: process.env.NUXT_PUBLIC_SSE_URL || '',
            enableMock: process.env.NUXT_PUBLIC_ENABLE_MOCK === 'true',
            appTitle: process.env.NUXT_PUBLIC_APP_TITLE || 'GoWind Shop',
            appDescription: process.env.NUXT_PUBLIC_APP_DESCRIPTION || 'A modern e-commerce storefront built with Nuxt',
            defaultLocale: process.env.NUXT_PUBLIC_DEFAULT_LOCALE || 'zh-CN',
            tokenKey: process.env.NUXT_PUBLIC_TOKEN_KEY || 'access_token',
            refreshTokenKey: process.env.NUXT_PUBLIC_REFRESH_TOKEN_KEY || 'refresh_token',
            aesKey: process.env.NUXT_PUBLIC_AES_KEY || '',
            // pinia-plugin-persistedstate v4 默认用 cookie 存储（SSR 友好）。
            // 显式约束 cookie 属性：sameSite=lax 防 CSRF 跨站携带，secure 仅生产
            // 启用（开发 HTTP 下 secure 会导致 cookie 不写入，丢失登录态）。
            piniaPluginPersistedstate: {
                cookieOptions: {
                    sameSite: 'lax',
                    secure: process.env.NODE_ENV === 'production',
                },
            },
        },
    },
    pinia: {
        storesDirs: ['./app/stores/**'],
    },
})
