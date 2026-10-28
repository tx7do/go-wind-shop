#!/bin/bash
# AI 整合 E2E 冒烟（脚本化登录套路见 memory gowind-shop-feature-batch-2026-10）
set -e
BASE_ADMIN=http://localhost:10600
BASE_APP=http://localhost:10700
RID=$(docker ps --format "{{.Names}}" | grep -i redis | head -1)
AES_KEY_HEX=66353164363661373364386130393237

enc() { printf '%s' "$1" | openssl enc -aes-128-cbc -K $AES_KEY_HEX -iv $AES_KEY_HEX -base64; }

echo "== 1. admin 登录 =="
CAP=$(curl -s $BASE_ADMIN/admin/v1/captcha)
CID=$(echo "$CAP" | python -c "import sys,json;print(json.load(sys.stdin)['captchaId'])")
CODE=$(docker exec "$RID" redis-cli -a '*Abcd123456' --no-auth-warning GET "gowind-shop:captcha:$CID" | tr -d '\r\n"')
echo "captcha: $CID -> $CODE"
LOGIN=$(curl -s -X POST $BASE_ADMIN/admin/v1/login \
  -H "Content-Type: application/json" -H "X-Captcha-Id: $CID" -H "X-Captcha-Value: $CODE" \
  -d "{\"grant_type\":\"password\",\"username\":\"admin\",\"password\":\"$(enc admin)\"}")
TOKEN=$(echo "$LOGIN" | python -c "import sys,json;d=json.load(sys.stdin);print(d.get('accessToken') or d.get('access_token',''))")
if [ -z "$TOKEN" ]; then echo "LOGIN FAILED: $LOGIN"; exit 1; fi
echo "admin token: ${TOKEN:0:20}..."
AH="Authorization: Bearer $TOKEN"

echo
echo "== 2. TaskMonitor:inspect（新移植接口） =="
curl -s -X POST $BASE_ADMIN/admin/v1/system-tasks:inspect -H "$AH" -H "Content-Type: application/json" -d '{}' | head -c 500
echo

echo
echo "== 3. admin Generate 代理（无 provider 时应报错而非 404/500 崩） =="
curl -s -X POST $BASE_ADMIN/admin/v1/ai/generate -H "$AH" -H "Content-Type: application/json" \
  -d '{"prompt":"测试","system":"你是测试助手"}' | head -c 300
echo

echo
echo "== 4. app 评论摘要（匿名；无 provider/评论不足应 available=false 而非报错） =="
curl -s "$BASE_APP/app/v1/mall/products/3/review-summary" | head -c 300
echo

echo
echo "== 5. app SSE 鉴权（无效 token 应 401） =="
curl -s -o /dev/null -w "%{http_code}\n" "http://localhost:10701/events?stream=bogus&token=bogus" --max-time 3 || true

echo
echo "== 6. app chat 未带 contextProductId（无 provider 时应报 provider 错误而非 500 崩） =="
APP_LOGIN=$(curl -s -X POST $BASE_APP/app/v1/login -H "Content-Type: application/json" \
  -d '{"grant_type":"password","username":"admin","password":"'$(enc admin)'","clientType":"CLIENT_TYPE_APP"}')
APP_TOKEN=$(echo "$APP_LOGIN" | python -c "import sys,json;d=json.load(sys.stdin);print(d.get('accessToken') or d.get('access_token',''))")
echo "app token: ${APP_TOKEN:0:16}..."
curl -s -X POST $BASE_APP/app/v1/ai/chat/completions -H "Authorization: Bearer $APP_TOKEN" -H "Content-Type: application/json" \
  -d '{"content":"你好"}' | head -c 300
echo

echo
echo "== 7. app chat 带 contextProductId=3（商品上下文透传，无 provider 降级验证同上） =="
curl -s -X POST $BASE_APP/app/v1/ai/chat/completions -H "Authorization: Bearer $APP_TOKEN" -H "Content-Type: application/json" \
  -d '{"content":"这个商品有什么优惠","contextProductId":3}' | head -c 300
echo

echo
echo "== 8. C端每日限额闸门（设限额=1 后再聊应 403，且不触碰 LLM） =="
docker exec "$RID" redis-cli -a '*Abcd123456' --no-auth-warning DEL "go_wind_shop:sys_config:ai.chat.daily_token_limit" >/dev/null 2>&1 || true
docker exec citus-server-standalone psql -U postgres -d go_wind_shop -c "UPDATE sys_configs SET value='1' WHERE key='ai.chat.daily_token_limit';" >/dev/null
sleep 2
curl -s -X POST $BASE_APP/app/v1/ai/chat/completions -H "Authorization: Bearer $APP_TOKEN" -H "Content-Type: application/json" \
  -d '{"content":"限额测试"}' | head -c 300
echo
docker exec citus-server-standalone psql -U postgres -d go_wind_shop -c "UPDATE sys_configs SET value='200000' WHERE key='ai.chat.daily_token_limit';" >/dev/null
echo "limit restored"
