# 生产部署：同源 nginx 路径分流

API/SSE 不开独立域名，全部经各自主域名的 nginx 反代转发：

| 域名 | 前端产物 | API 路径 | SSE 路径 |
|---|---|---|---|
| `shop.gowind.cloud` | Nuxt SSG `.output/public` | `/app/v1/` → 127.0.0.1:10700 | `/events` → 127.0.0.1:10701 |
| `admin.shop.gowind.cloud` | Vite `dist` | `/admin/v1/` → 127.0.0.1:10600 | `/events` → 127.0.0.1:10601 |

两个 conf 均为 **HTTPS 终态**（DNS 已通，certbot 已签发）：443 段与 80→443
重定向段带 `# managed by Certbot` 标记原样保留，不影响续期。

## 1. 构建前端产物（Node ≥ 20.19，全局 node 20.10 会炸 styleText）

```bash
export PATH="/c/Users/yangl/scoop/apps/nvm/current/nodejs/v20.20.0:$PATH"

# admin
cd frontend/admin && pnpm build          # 产物 dist/

# app（SSG，两个语言壳必须真实产出——nginx 深链接回退依赖它们）
cd frontend/app && node node_modules/nuxt/bin/nuxt.mjs generate --dotenv .env.production
# 产物 .output/public/   （构建约 2 分钟，必须前台跑完）
```

`.env.production` 已同源化：admin `VITE_APP_API_URL=''` + `VITE_APP_SSE_URL='/events'`；
app `NUXT_PUBLIC_API_BASE_URL=/`（不能空，nuxt.config 的 `||` 回退会兜到 localhost:8000）+
`NUXT_PUBLIC_SSE_URL=/events` + AES 密钥烧入（改密钥后必须重新 generate）。

## 2. 上传与启用

root 按服务器实际路径（conf 默认 `/var/www/go-wind-shop/{app,admin}`）：

```bash
rsync -a frontend/app/.output/public/  <server>:/var/www/go-wind-shop/app/
rsync -a frontend/admin/dist/          <server>:/var/www/go-wind-shop/admin/

# 覆盖服务器上的同名旧 conf。若旧 conf 不在 conf.d 而在 sites-enabled 等处，
# 就地替换那一份，避免同一 server_name 被加载两份。
scp deploy/nginx/shop.gowind.cloud.conf       <server>:/etc/nginx/conf.d/
scp deploy/nginx/admin.shop.gowind.cloud.conf <server>:/etc/nginx/conf.d/
ssh <server> 'nginx -t && nginx -s reload'
```

**证书路径核对**：`ls /etc/letsencrypt/live/`。admin 已实证是独立 lineage
（`live/admin.shop.gowind.cloud/`）；若 shop 是两域合并签发的单一 lineage，
把 shop conf 里的 `ssl_certificate*` 改成实际目录再 reload。

## 3. 要点

- **SSE 两个 `/events` 块已内置 `proxy_buffering off` + `proxy_read_timeout 3600s`**：
  没有这两项，AI 导购/token 续期的帧会被 nginx 攒住不吐、长连接被 60s 默认超时掐断。
- 代理 location 用 `^~` 前缀：抑制 `\.html$` / 资源扩展名正则抢匹配。
- HTML/入口页一律 no-cache，指纹化静态资源长缓存 immutable——发布即生效。
- `client_max_body_size 50m`：商品图等上传接口需要，服务器旧 conf 没配（默认 1m 会 413）。
- 后端四个端口建议只监听 `127.0.0.1`（或内网网卡），不暴露公网，入口只有 nginx。

## 4. 验证

```bash
curl --noproxy '*' -s https://shop.gowind.cloud/app/v1/login -o /dev/null -w '%{http_code}\n'   # 405/400 即通（非 502/404）
curl --noproxy '*' -sN https://shop.gowind.cloud/events?stream=x --max-time 3                    # 应有响应帧而非 502
curl --noproxy '*' -sI https://admin.shop.gowind.cloud/ | head -1                                # 200
curl --noproxy '*' -sI http://admin.shop.gowind.cloud/ | head -1                                 # 301 → https
```
