-- =====================================================================
-- 商城功能演示/测试数据（PostgreSQL）
-- =====================================================================
-- 覆盖域：商品/SKU/价格、促销（专价+满减）、订单+支付流水、库存操作流水、
--         商品评论（含商家回复）、站内信（C端消息中心）。
--
-- 设计约定：
--   * 幂等：所有 INSERT 带业务键 WHERE NOT EXISTS 守卫，可重复执行；
--   * 显式 id 后一律 setval 重置序列（防 MERGE/自增撞主键，见菜单序列坑）；
--   * 金额单位：分（最小货币单位）；
--   * 依赖：假定 sys_users 至少存在 id=1（admin）。demo-data.sql 之外可独立执行。
--
-- 执行：docker exec citus-server-standalone psql -U postgres -d go_wind_shop -f /path/to/test-data.sql
-- 或：psql -h 127.0.0.1 -U postgres -d go_wind_shop -f backend/sql/test-data.sql
-- =====================================================================

-- ---------------------------------------------------------------
-- 1. 商品（ACTIVE）+ 翻译
-- ---------------------------------------------------------------
-- 商品 3：无线蓝牙耳机 Pro（若不存在则建）
INSERT INTO public.mall_products (id, status, sort_order, created_by, created_at, updated_at)
SELECT 3, 'PRODUCT_STATUS_ACTIVE', 1, 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_products WHERE id = 3);

INSERT INTO public.mall_product_translations (product_id, language_code, name, slug, short_description, long_description, created_by, created_at, updated_at)
SELECT 3, 'zh-CN', '无线蓝牙耳机 Pro', 'wireless-earphone-pro', '主动降噪无线蓝牙耳机', '40mm 大动圈单元，ANC 主动降噪，蓝牙 5.3，单次续航 8 小时，配合充电盒总续航 32 小时。', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_product_translations WHERE product_id = 3 AND language_code = 'zh-CN');

-- 商品 4：智能运动手表 X2
INSERT INTO public.mall_products (id, status, sort_order, created_by, created_at, updated_at)
SELECT 4, 'PRODUCT_STATUS_ACTIVE', 2, 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_products WHERE id = 4);

INSERT INTO public.mall_product_translations (product_id, language_code, name, slug, short_description, long_description, created_by, created_at, updated_at)
SELECT 4, 'zh-CN', '智能运动手表 X2', 'smart-watch-x2', 'AMOLED 高刷屏，双频 GPS，14 天长续航', '1.43 英寸 AMOLED 常亮屏，双频 GNSS 精准轨迹，心率/血氧/睡眠监测，5ATM 防水，典型使用 14 天续航。', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_product_translations WHERE product_id = 4 AND language_code = 'zh-CN');

SELECT setval('mall_products_id_seq', (SELECT MAX(id) FROM public.mall_products));

-- ---------------------------------------------------------------
-- 2. SKU + 多币种价格
-- ---------------------------------------------------------------
-- SKU 1：耳机（黑色）
INSERT INTO public.mall_skus (id, product_id, sku_code, stock_qty, created_by, created_at, updated_at)
SELECT 1, 3, 'EARPHONE-PRO-BLK', 50, 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_skus WHERE id = 1);

-- SKU 2：手表（曜石黑）
INSERT INTO public.mall_skus (id, product_id, sku_code, stock_qty, created_by, created_at, updated_at)
SELECT 2, 4, 'WATCH-X2-BLK', 30, 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_skus WHERE id = 2);

SELECT setval('mall_skus_id_seq', (SELECT MAX(id) FROM public.mall_skus));

INSERT INTO public.mall_sku_prices (id, sku_id, currency, amount, created_by, created_at, updated_at)
SELECT 1, 1, 'CNY', '12800', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_sku_prices WHERE sku_id = 1 AND currency = 'CNY');

INSERT INTO public.mall_sku_prices (id, sku_id, currency, amount, created_by, created_at, updated_at)
SELECT 2, 2, 'CNY', '39900', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_sku_prices WHERE sku_id = 2 AND currency = 'CNY');

SELECT setval('mall_sku_prices_id_seq', (SELECT MAX(id) FROM public.mall_sku_prices));

-- ---------------------------------------------------------------
-- 3. 促销：商品专价 + 店铺满减
-- ---------------------------------------------------------------
-- 专价：耳机限时 ¥99（原 ¥128）
INSERT INTO public.mall_promotions (id, name, promo_type, product_id, special_price, threshold_amount, reduction_amount, priority, valid_from, valid_until, status, currency, created_by, created_at, updated_at)
SELECT 1, '耳机限时专价', 'SPECIAL_PRICE', 3, 9900, 0, 0, 10, '2026-10-01', '2027-01-01', 'ACTIVE', 'CNY', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_promotions WHERE id = 1);

-- 满减：店铺满 ¥300 减 ¥40
INSERT INTO public.mall_promotions (id, name, promo_type, product_id, special_price, threshold_amount, reduction_amount, priority, valid_from, valid_until, status, currency, created_by, created_at, updated_at)
SELECT 2, '店铺满300减40', 'FULL_REDUCTION', 0, 0, 30000, 4000, 5, '2026-10-01', '2027-01-01', 'ACTIVE', 'CNY', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_promotions WHERE id = 2);

SELECT setval('mall_promotions_id_seq', (SELECT MAX(id) FROM public.mall_promotions));

-- ---------------------------------------------------------------
-- 4. 订单 + 订单项 + 支付流水
--    订单 1：PAID     耳机×1（专价 ¥99，原价 ¥128，优惠 ¥29）
--    订单 2：FULFILLED 手表×1（¥399）
--    订单 3：CLOSED   耳机×2（专价 ¥198）
-- ---------------------------------------------------------------
INSERT INTO public.mall_orders (id, tenant_id, currency, idempotency_key, user_id, status, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, recipient_name, recipient_phone, shipping_address, created_by, created_at, updated_at)
SELECT 1, 1, 'CNY', 'TEST-ORDER-001', 1, 'PAID', 9900, 12800, 2900, 0, 0, 'CN', '张三', '13800000001', '广东省深圳市南山区科技园路 1 号', 1, now() - interval '2 day', now() - interval '2 day'
WHERE NOT EXISTS (SELECT 1 FROM public.mall_orders WHERE idempotency_key = 'TEST-ORDER-001');

INSERT INTO public.mall_orders (id, tenant_id, currency, idempotency_key, user_id, status, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, recipient_name, recipient_phone, shipping_address, created_by, created_at, updated_at)
SELECT 2, 1, 'CNY', 'TEST-ORDER-002', 1, 'FULFILLED', 39900, 39900, 0, 0, 0, 'CN', '张三', '13800000001', '广东省深圳市南山区科技园路 1 号', 1, now() - interval '1 day', now()
WHERE NOT EXISTS (SELECT 1 FROM public.mall_orders WHERE idempotency_key = 'TEST-ORDER-002');

INSERT INTO public.mall_orders (id, tenant_id, currency, idempotency_key, user_id, status, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, recipient_name, recipient_phone, shipping_address, created_by, created_at, updated_at)
SELECT 3, 1, 'CNY', 'TEST-ORDER-003', 1, 'CLOSED', 19800, 25600, 5800, 0, 0, 'CN', '张三', '13800000001', '广东省深圳市南山区科技园路 1 号', 1, now() - interval '7 day', now() - interval '3 day'
WHERE NOT EXISTS (SELECT 1 FROM public.mall_orders WHERE idempotency_key = 'TEST-ORDER-003');

SELECT setval('mall_orders_id_seq', (SELECT MAX(id) FROM public.mall_orders));

INSERT INTO public.mall_order_items (tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_by, created_at, updated_at)
SELECT 1, o.id, 1, 1, '{"sku_id":1,"quantity":1,"currency":"CNY","unit_price":9900}', 1, 9900, 9900, 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-001'
  AND NOT EXISTS (SELECT 1 FROM public.mall_order_items i WHERE i.order_id = o.id AND i.sku_id = 1);

INSERT INTO public.mall_order_items (tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_by, created_at, updated_at)
SELECT 1, o.id, 1, 2, '{"sku_id":2,"quantity":1,"currency":"CNY","unit_price":39900}', 1, 39900, 39900, 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-002'
  AND NOT EXISTS (SELECT 1 FROM public.mall_order_items i WHERE i.order_id = o.id AND i.sku_id = 2);

INSERT INTO public.mall_order_items (tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_by, created_at, updated_at)
SELECT 1, o.id, 1, 1, '{"sku_id":1,"quantity":2,"currency":"CNY","unit_price":9900}', 2, 9900, 19800, 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-003'
  AND NOT EXISTS (SELECT 1 FROM public.mall_order_items i WHERE i.order_id = o.id AND i.sku_id = 1);

INSERT INTO public.mall_payment_transactions (tenant_id, currency, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, created_by, created_at, updated_at)
SELECT 1, 'CNY', 'TEST-PAY-001', 'BALANCE', 'BUSINESS_TYPE_CONSUME', o.id, 1, o.total_amount, 'SUCCEEDED', 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-001'
  AND NOT EXISTS (SELECT 1 FROM public.mall_payment_transactions WHERE idempotency_key = 'TEST-PAY-001');

INSERT INTO public.mall_payment_transactions (tenant_id, currency, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, created_by, created_at, updated_at)
SELECT 1, 'CNY', 'TEST-PAY-002', 'BALANCE', 'BUSINESS_TYPE_CONSUME', o.id, 1, o.total_amount, 'SUCCEEDED', 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-002'
  AND NOT EXISTS (SELECT 1 FROM public.mall_payment_transactions WHERE idempotency_key = 'TEST-PAY-002');

INSERT INTO public.mall_payment_transactions (tenant_id, currency, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, created_by, created_at, updated_at)
SELECT 1, 'CNY', 'TEST-PAY-003', 'BALANCE', 'BUSINESS_TYPE_CONSUME', o.id, 1, o.total_amount, 'SUCCEEDED', 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-003'
  AND NOT EXISTS (SELECT 1 FROM public.mall_payment_transactions WHERE idempotency_key = 'TEST-PAY-003');

SELECT setval('mall_payment_transactions_id_seq', (SELECT MAX(id) FROM public.mall_payment_transactions));

-- ---------------------------------------------------------------
-- 5. 库存操作流水（与订单对应；手动调整流水由操作产生）
-- ---------------------------------------------------------------
INSERT INTO public.mall_stock_op_logs (sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark, created_by, created_at, updated_at)
SELECT 1, 3, -1, 51, 50, 'ORDER_DEDUCT', o.id, NULL, 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-001'
  AND NOT EXISTS (SELECT 1 FROM public.mall_stock_op_logs l WHERE l.reason = 'ORDER_DEDUCT' AND l.ref_id = o.id AND l.sku_id = 1);

INSERT INTO public.mall_stock_op_logs (sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark, created_by, created_at, updated_at)
SELECT 2, 4, -1, 31, 30, 'ORDER_DEDUCT', o.id, NULL, 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-002'
  AND NOT EXISTS (SELECT 1 FROM public.mall_stock_op_logs l WHERE l.reason = 'ORDER_DEDUCT' AND l.ref_id = o.id AND l.sku_id = 2);

INSERT INTO public.mall_stock_op_logs (sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark, created_by, created_at, updated_at)
SELECT 1, 3, -2, 52, 50, 'ORDER_DEDUCT', o.id, NULL, 1, o.created_at, o.created_at
FROM public.mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-003'
  AND NOT EXISTS (SELECT 1 FROM public.mall_stock_op_logs l WHERE l.reason = 'ORDER_DEDUCT' AND l.ref_id = o.id AND l.sku_id = 1);

SELECT setval('mall_stock_op_logs_id_seq', (SELECT MAX(id) FROM public.mall_stock_op_logs));

-- ---------------------------------------------------------------
-- 6. 商品评论（含商家回复）
--    评论 1/2（耳机 + 商家回复）已由功能验证产生；此处补手表的评论
-- ---------------------------------------------------------------
INSERT INTO public.mall_comments (tenant_id, content_type, object_id, content, status, rating, parent_id, created_by, created_at, updated_at)
SELECT 0, 'CONTENT_TYPE_PRODUCT', 4, '续航给力，双频 GPS 定位很准，跑步骑行都够用。', 'STATUS_APPROVED', 5, NULL, 1, now() - interval '1 day', now() - interval '1 day'
WHERE NOT EXISTS (SELECT 1 FROM public.mall_comments WHERE object_id = 4 AND content_type = 'CONTENT_TYPE_PRODUCT' AND parent_id IS NULL);

-- 商家回复
INSERT INTO public.mall_comments (tenant_id, content_type, object_id, content, status, parent_id, created_by, created_at, updated_at)
SELECT 0, 'CONTENT_TYPE_PRODUCT', 4, '感谢您的认可，X2 的双频 GPS 是本代最大升级点，祝运动愉快！', 'STATUS_APPROVED', (SELECT id FROM public.mall_comments WHERE object_id = 4 AND parent_id IS NULL LIMIT 1), 1, now() - interval '20 hour', now() - interval '20 hour'
WHERE (SELECT id FROM public.mall_comments WHERE object_id = 4 AND parent_id IS NULL LIMIT 1) IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM public.mall_comments WHERE object_id = 4 AND parent_id IS NOT NULL);

SELECT setval('mall_comments_id_seq', (SELECT MAX(id) FROM public.mall_comments));

-- ---------------------------------------------------------------
-- 7. 站内信（C端消息中心：未读 badge 数据源）
-- ---------------------------------------------------------------
INSERT INTO public.internal_messages (id, title, content, status, type, sender_id, created_at, updated_at)
SELECT 10, '订单已发货', '您的订单 #2 已发货，请注意查收物流信息。', 'PUBLISHED', 'NOTIFICATION', 0, now() - interval '1 day', now() - interval '1 day'
WHERE NOT EXISTS (SELECT 1 FROM public.internal_messages WHERE id = 10);

INSERT INTO public.internal_messages (id, title, content, status, type, sender_id, created_at, updated_at)
SELECT 11, '店铺满减活动上线', '全店满 ¥300 减 ¥40，活动截至 2027-01-01，先到先得！', 'PUBLISHED', 'NOTIFICATION', 0, now() - interval '4 hour', now() - interval '4 hour'
WHERE NOT EXISTS (SELECT 1 FROM public.internal_messages WHERE id = 11);

SELECT setval('internal_messages_id_seq', (SELECT MAX(id) FROM public.internal_messages));

INSERT INTO public.internal_message_recipients (tenant_id, message_id, recipient_user_id, status, created_at, updated_at)
SELECT 1, m.id, 1, 'SENT', m.created_at, m.created_at
FROM public.internal_messages m
WHERE m.id = 10
  AND NOT EXISTS (SELECT 1 FROM public.internal_message_recipients r WHERE r.message_id = 10 AND r.recipient_user_id = 1);

INSERT INTO public.internal_message_recipients (tenant_id, message_id, recipient_user_id, status, created_at, updated_at)
SELECT 1, m.id, 1, 'SENT', m.created_at, m.created_at
FROM public.internal_messages m
WHERE m.id = 11
  AND NOT EXISTS (SELECT 1 FROM public.internal_message_recipients r WHERE r.message_id = 11 AND r.recipient_user_id = 1);

SELECT setval('internal_message_recipients_id_seq', (SELECT MAX(id) FROM public.internal_message_recipients));

-- ---------------------------------------------------------------
-- 完成
-- ---------------------------------------------------------------
DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['mall_products', 'mall_product_translations', 'mall_skus', 'mall_sku_prices',
                             'mall_promotions', 'mall_orders', 'mall_order_items', 'mall_payment_transactions',
                             'mall_stock_op_logs', 'mall_comments', 'internal_messages', 'internal_message_recipients']
    LOOP
        BEGIN
            EXECUTE format('SELECT setval(%L, (SELECT MAX(id) FROM public.%I))', t || '_id_seq', t);
        EXCEPTION WHEN undefined_table THEN
            NULL;
        END;
    END LOOP;
END $$;
