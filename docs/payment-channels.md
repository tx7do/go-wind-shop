# 支付渠道接入指南与合规边界

> 定位声明：本项目作为**代码交付物**提供支付能力，不涉及任何经营行为。
> 支付功能的实际使用方（部署方/运营方）承担其自身的经营合规责任。
> 本文档说明技术接入方式与设计时内置的合规护栏。

## 一、合规边界（为什么这样设计）

系统采用**「部署方自带资质」的直连商户模式**：

```
买家 → 支付渠道（微信/支付宝） → 部署方自己的商户号
```

- 交易资金由支付渠道**直接清算给部署方自己的商户号**，本系统不归集、
  不经手、不二次分配任何交易资金；
- 因此本模式**不构成**「二清」（无证经营支付业务），前提是部署方仅将
  本系统用于自营业务；
- 若部署方未来将其改造为**多商户平台**（商户入驻、平台代收再分账），
  则必须另行接入持牌机构分账产品（如微信电商收付通、支付宝分账），
  本项目代码**不提供**该形态的能力，枚举与文档均已明确此边界。

内置护栏清单：

| 护栏 | 位置 |
| --- | --- |
| 代码与仓库内零密钥，凭据全部由部署方在管理端录入 | `mall_payment_channel_configs.credentials`（AES-256-GCM 密文落库） |
| 凭据读取路径永不回传（含密文），只回 `credentials_configured` 布尔 | core `PaymentChannelConfigService` 读路径剥离 |
| 新渠道默认 `停用 + 沙箱`，运营显式开启 | ent schema 默认值 |
| MOCK 模拟渠道不出真实资金；生产应保持停用 | 渠道类型注释 + 管理页提示 |
| 模拟确认接口只放行 MOCK 流水（堵「免费购买」后门） | app BFF `MockConfirmPayment` |
| 回调凭验签保护，验签失败拒绝；金额不符拒绝确认 | `pkg/payment/gateway` + core `ConfirmPayment` |
| 支付金额取服务端订单 `total_amount`，不信任客户端 | core `CreatePaymentIntent` |
| **不提供、不预留虚拟货币支付**（境内属非法金融活动，银发〔2021〕237号） | `PaymentMethod` 枚举已删除全部 CRYPTO_* 值 |
| BALANCE/CREDIT 仅为线下/后台记账语义，**不构成在线储值钱包** | 枚举注释；未实现充值/提现/余额账户体系 |

> 重要：若要上线「用户储值余额（充值→消费）」，将触及预付卡/支付业务许可
> 监管，请先取得相应资质或咨询专业法律意见；本项目刻意未实现该体系。

## 二、部署方资质前提（真实收款）

| 项 | 说明 |
| --- | --- |
| 经营主体 | 个体工商户或公司营业执照（个人身份无法开通微信/支付宝商户收款） |
| 微信支付 | 商户号（mch_id）+ 商户平台「API安全」三件套：APIv3 密钥、商户证书序列号、商户私钥 `apiclient_key.pem` |
| 支付宝 | 开放平台应用 app_id + 应用私钥 + **支付宝公钥**（注意不是应用公钥） |
| ICP | 自营商城一般完成 ICP 备案即可；如改造为第三方入驻平台，涉及 EDI 等许可 |
| 开票义务 | 依法向消费者开具发票（系统已有 invoice 域） |

## 三、配置步骤

1. **设置凭据加密密钥**（BFF app 服务环境变量）：
   ```bash
   # 任意强随机串（SHA-256 派生 AES-256 密钥）。未设置时真实渠道凭据无法保存。
   PAYMENT_CREDENTIAL_KEY=<your-random-secret>
   ```
2. **设置回调公网地址**（BFF app 服务环境变量）：
   ```bash
   # 支付渠道服务器必须能访问到该地址；缺省为 http://localhost:10700（仅本地联调）
   PAYMENT_NOTIFY_BASE_URL=https://shop.example.com
   ```
   回调路由：`POST /app/v1/payments/notify/alipay`、`POST /app/v1/payments/notify/wechat`
   （免登录态，凭渠道验签保护；支付宝应答 `success` 明文，微信应答 JSON SUCCESS）。
3. **管理端录入渠道**：「商城 → 支付渠道配置」新增渠道，按渠道类型填入
   appid/商户号与凭据字段，验收无误后把沙箱开关切到生产并启用。
   - 渠道凭据为写入专用：保存后不再回显，编辑留空保持不变；
4. **开发/演示**：新增一个 MOCK 渠道并启用，C 端结算页会出现
   「模拟支付（演示）」，不出真实资金即可走通整条支付闭环。

## 四、支付链路（技术）

```
C端 checkout
  → POST /app/v1/mall/payments/intent {orderId, channelId, idempotencyKey}
      BFF: core CreatePaymentIntent（订单归属/待支付态/幂等校验 → PENDING 流水）
      BFF: 渠道网关下单（微信 Native prepay → code_url；支付宝 page pay → 跳转链接）
  → C端展示二维码 / 新窗口跳转支付宝；每 2.5s 轮询流水状态（上限 3 分钟）
  → 渠道网关异步回调 POST /app/v1/payments/notify/{channel}
      BFF: ResolveEnabledChannel（按类型取启用渠道 + 解密凭据）→ 验签解析
      BFF: core ConfirmPayment（幂等：PENDING→SUCCEEDED + 金额核验 → 订单 PAID）
  → C端轮询命中 SUCCEEDED → 跳转订单页
```

- 幂等键：支付流水 `idempotency_key`（租户内唯一索引）；网关 `out_trade_no`
  固定为支付流水 ID（数字串，兼容微信 6–32 位与支付宝 64 位限制）；
- 对账：流水 `gateway_trade_no` 回填渠道侧交易号，`business_ref_id` 与订单对齐；
- 超时：订单超时取消（order_sweep）与回调确认靠 `expected_status=PENDING_PAYMENT`
  乐观并发互斥——订单已取消时回调只置流水成功并告警日志，资金走退款流程；
- 退款：管理端对 SUCCEEDED 流水创建退款（默认 PENDING）→ 确认时**先调渠道
  退款 API**（微信 refunddomestic / 支付宝 trade.refund；同一退款单号渠道幂等）
  ——渠道受理才推进 SUCCEEDED，联动流水 REFUNDED + 库存回补 + 站内信；渠道
  拒绝则退款停留 PENDING 可重试。记账型渠道（BALANCE/OFFLINE）与 MOCK 无真实
  资金，直接放行。
- 关单联动：订单取消/超时取消（order_sweep）后 best-effort 调渠道关单
  （微信 closeorder / 支付宝 trade.close），让未支付流水的旧二维码/收银台
  链接失效；渠道侧已支付的关单失败由回调路径兜底（expected_status 拦截 +
  退款流程）。
- 空购物车/全无效项的下单请求直接拒绝（防 0 元空单走完支付退款链路）。

## 五、验收口径

| 渠道 | 验收方式 |
| --- | --- |
| MOCK | 本地全链路：下单 → 模拟确认 → 订单 PAID（无网络依赖） |
| 支付宝 | 沙箱（渠道配置勾选沙箱，用沙箱买家账号走 page pay）；生产前切生产网关小额验证 |
| 微信 Native | V3 无沙箱：真实商户号 0.01 元小额验收（配置 sandbox=true 时网关拒绝创建，防误刷生产） |
| 微信 H5 | 移动端浏览器跳转支付（h5_url 整页跳转；需买家客户端 IP，BFF 自动透传）；同 Native 走真实商户号验收 |

## 六、对账与自愈（payment_reconcile 周期任务）

- 每 10 分钟（sys_tasks 种子）扫描创建超过 10 分钟仍 PENDING 的 WECHAT/ALIPAY
  流水，主动调用渠道查单（微信 QueryOrderByOutTradeNo / 支付宝 trade.query）：
  - 渠道已支付 → 补确认（幂等；金额核验不过保留 PENDING 告警待人工）
  - 渠道已关闭/支付失败 → 流水收敛 FAILED（状态机白名单内）
  - 未支付/查不到 → 保持 PENDING（订单超时取消路径联动关单）
- 覆盖场景：网关异步回调丢失/延迟、网络分区；与 order_sweep（订单侧兜底）
  互补，构成支付链路的最终一致性双保险。

## 七、售后申请（C 端）

- 已支付订单详情页「申请退款」→ 弹窗填写原因（可选）→ 创建 PENDING 退款单
  （金额必须为正且 ≤ 实付；同流水存在 PENDING/SUCCEEDED 退款时拒绝重复申请）。
- 管理端确认后走第三节退款自动化（渠道受理 → 流水 REFUNDED + 库存回补 + 站内信）。
- 隔离：退款的交易读取走 UserPrivacy 行级隔离，买家只能对本人流水发起申请。

单元测试覆盖：`pkg/payment/gateway`（支付宝验签往返 + 篡改金额拒绝、
金额分↔元互转口径、MOCK 行为边界）。
