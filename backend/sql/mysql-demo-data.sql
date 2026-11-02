-- =====================================================================
-- 商城演示库种子（MySQL）—— 以内网 go_wind_shop 实测数据为基准
-- =====================================================================
-- 与 postgresql-demo-data.sql 同源同内容（MySQL 方言版）：
--   A. 内网业务表数据快照：目录（类目/品牌/商品/属性/SKU/价格）、促销、购物车、
--      订单/支付/退款/发货、库存流水与预警、评论+点赞、优惠券/秒杀/拼团/发票/
--      税率/运费/会员积分/收藏、站内信，以及 admin 账号+凭证、语言注册表。
--      基准：商品 100-119（智能家居目录，图 /demo-images/*.webp）、SKU 1000+、
--      订单 100-111、租户 0 + admin(id=1)。
--   B. 原 test-data.sql 增补段（原始 ID 原样保留，带幂等守卫）：商品 3/4
--      （无线蓝牙耳机 Pro / 智能运动手表 X2）、SKU 1/2、促销 1、订单 1-3、
--      支付流水、库存流水、手表评论+商家回复、站内信 10/11（与快照去重）。
--   C. 演示增补层：顾客用户 4 名（密码 admin）、促销 3-7、订单 112-123、
--      支付/退款/发货、评论 164-189（含待审核与商家回复）、站内信 35-38、
--      购物车/库存预警/库存流水。
-- 约定：TRUNCATE 后全量重灌、可重复执行；金额单位：分；
--       MySQL 无序列，显式 id 插入后 AUTO_INCREMENT 自动抬升，无需校正；
--       internal_message_categories 由运行时管理，不在本文件范围。
--
-- 执行：mysql -h 127.0.0.1 -u root -p go_wind_shop < backend/sql/mysql-demo-data.sql
-- =====================================================================

-- 关闭外键检查（自引用树表 TRUNCATE 需要）
SET FOREIGN_KEY_CHECKS = 0;
SET NAMES utf8mb4;

-- ── 一、清空业务表 ─────────────────────────────────────────────────
TRUNCATE TABLE sys_languages;
TRUNCATE TABLE sys_users;
TRUNCATE TABLE sys_user_credentials;
TRUNCATE TABLE internal_message_recipients;
TRUNCATE TABLE internal_messages;
TRUNCATE TABLE mall_brand_translations;
TRUNCATE TABLE mall_brands;
TRUNCATE TABLE mall_cart_items;
TRUNCATE TABLE mall_carts;
TRUNCATE TABLE mall_categories;
TRUNCATE TABLE mall_category_translations;
TRUNCATE TABLE mall_comment_likes;
TRUNCATE TABLE mall_comments;
TRUNCATE TABLE mall_coupon_templates;
TRUNCATE TABLE mall_groupon_activities;
TRUNCATE TABLE mall_groupon_groups;
TRUNCATE TABLE mall_invoices;
TRUNCATE TABLE mall_member_tiers;
TRUNCATE TABLE mall_order_items;
TRUNCATE TABLE mall_orders;
TRUNCATE TABLE mall_payment_channel_configs;
TRUNCATE TABLE mall_payment_refunds;
TRUNCATE TABLE mall_payment_transactions;
TRUNCATE TABLE mall_points_ledgers;
TRUNCATE TABLE mall_product_attribute_translations;
TRUNCATE TABLE mall_product_attribute_value_translations;
TRUNCATE TABLE mall_product_attribute_values;
TRUNCATE TABLE mall_product_attributes;
TRUNCATE TABLE mall_product_translations;
TRUNCATE TABLE mall_products;
TRUNCATE TABLE mall_promotions;
TRUNCATE TABLE mall_seckill_activities;
TRUNCATE TABLE mall_shipments;
TRUNCATE TABLE mall_shipping_rates;
TRUNCATE TABLE mall_sku_attribute_combinations;
TRUNCATE TABLE mall_sku_prices;
TRUNCATE TABLE mall_skus;
TRUNCATE TABLE mall_stock_alerts;
TRUNCATE TABLE mall_stock_op_logs;
TRUNCATE TABLE mall_tax_rates;
TRUNCATE TABLE mall_user_coupons;
TRUNCATE TABLE mall_wishlist;

SET FOREIGN_KEY_CHECKS = 1;

START TRANSACTION;

-- ── 二、内网数据快照 ────────────────────────────────────────────────
--
-- PostgreSQL database dump
--


-- Dumped from database version 16.15 (Debian 16.15-1.pgdg12+2)
-- Dumped by pg_dump version 16.15 (Debian 16.15-1.pgdg12+2)


--
-- Data for Name: internal_message_recipients; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (4, '2026-10-27 04:56:26.551016', '2026-10-27 04:56:26.551016', NULL, 1, 10, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (5, '2026-10-28 00:56:26.5571', '2026-10-28 00:56:26.5571', NULL, 1, 11, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (8, '2026-10-29 01:35:18.449787', NULL, NULL, 0, 14, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (9, '2026-10-29 02:04:04.478118', NULL, NULL, 0, 15, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (10, '2026-10-29 02:14:56.322189', NULL, NULL, 0, 16, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (11, '2026-10-29 02:26:34.912287', NULL, NULL, 0, 17, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (12, '2026-10-29 02:26:35.916859', NULL, NULL, 0, 18, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (13, '2026-10-29 02:39:41.242088', NULL, NULL, 0, 19, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (14, '2026-10-29 02:39:45.82795', NULL, NULL, 0, 20, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (15, '2026-10-29 02:40:14.590483', NULL, NULL, 0, 21, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (16, '2026-10-29 03:03:26.519041', NULL, NULL, 0, 22, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (17, '2026-10-29 03:15:55.313084', NULL, NULL, 0, 23, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (18, '2026-10-29 03:17:36.444804', NULL, NULL, 0, 24, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (19, '2026-10-29 03:39:18.686408', NULL, NULL, 0, 25, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (20, '2026-10-29 06:22:41.562947', NULL, NULL, 0, 26, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (21, '2026-10-29 06:25:18.578498', NULL, NULL, 0, 27, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (22, '2026-10-29 06:28:34.020818', NULL, NULL, 0, 28, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (23, '2026-10-29 06:28:35.056086', NULL, NULL, 0, 29, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (24, '2026-10-29 06:29:44.26476', NULL, NULL, 0, 30, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (25, '2026-10-29 07:55:32.40661', NULL, NULL, 0, 31, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (27, '2026-10-30 23:15:01.294306', NULL, NULL, 0, 33, 1, 'SENT', NULL, NULL);
INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at) VALUES (28, '2026-10-30 23:15:01.32874', NULL, NULL, 0, 34, 1, 'SENT', NULL, NULL);


--
-- Data for Name: internal_messages; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (10, '2026-10-27 04:56:26.551016', '2026-10-27 04:56:26.551016', NULL, NULL, NULL, NULL, 0, '订单已发货', '您的订单 #2 已发货，请注意查收物流信息。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (11, '2026-10-28 00:56:26.5571', '2026-10-28 00:56:26.5571', NULL, NULL, NULL, NULL, 0, '店铺满减活动上线', '全店满 ¥300 减 ¥40，活动截至 2027-01-01，先到先得！', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (14, '2026-10-29 01:35:18.443719', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #4 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (15, '2026-10-29 02:04:04.471225', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #5 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (16, '2026-10-29 02:14:56.315714', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #6 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (17, '2026-10-29 02:26:34.907958', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #7 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (18, '2026-10-29 02:26:35.913268', NULL, NULL, 0, NULL, NULL, 0, '退款成功通知', '您的订单 #7 的退款已处理成功，退款金额 0.00 元，请注意查收。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (19, '2026-10-29 02:39:41.231799', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #15 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (20, '2026-10-29 02:39:45.823744', NULL, NULL, 0, NULL, NULL, 0, '退款成功通知', '您的订单 #15 的退款已处理成功，退款金额 198.00 元，请注意查收。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (21, '2026-10-29 02:40:14.587131', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #17 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (22, '2026-10-29 03:03:26.511613', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #11 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (23, '2026-10-29 03:15:55.307153', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #19 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (24, '2026-10-29 03:17:36.437994', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #8 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (25, '2026-10-29 03:39:18.679684', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #18 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (26, '2026-10-29 06:22:41.553749', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #24 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (27, '2026-10-29 06:25:18.571286', NULL, NULL, 0, NULL, NULL, 0, '退款成功通知', '您的订单 #24 的退款已处理成功，退款金额 297.00 元，请注意查收。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (28, '2026-10-29 06:28:34.015836', NULL, NULL, 0, NULL, NULL, 0, '订单支付成功', '您的订单 #25 已支付成功，我们将尽快为您安排发货。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (29, '2026-10-29 06:28:35.053045', NULL, NULL, 0, NULL, NULL, 0, '退款成功通知', '您的订单 #25 的退款已处理成功，退款金额 99.00 元，请注意查收。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (30, '2026-10-29 06:29:44.26061', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #20 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (31, '2026-10-29 07:55:32.396308', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #29 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (32, '2026-10-30 23:15:01.241948', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #100 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (33, '2026-10-30 23:15:01.290393', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #105 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, title, content, sender_id, category_id, status, type) VALUES (34, '2026-10-30 23:15:01.324956', NULL, NULL, 0, NULL, NULL, 0, '订单已取消', '您的订单 #111 已取消。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');


--
-- Data for Name: mall_brand_translations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (300, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'zh-CN', '风铃', 'windbell', '居家智能生态');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (301, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'en-US', 'Windbell', 'windbell', '居家智能生态');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (302, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'zh-CN', '曜石', 'obsidian', '高端手机与配件');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (303, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'en-US', 'Obsidian', 'obsidian', '高端手机与配件');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (304, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'zh-CN', '云雀', 'lark', '音频设备');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (305, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'en-US', 'Lark', 'lark', '音频设备');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (306, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'zh-CN', '极目', 'vista', '穿戴与影像');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (307, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'en-US', 'Vista', 'vista', '穿戴与影像');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (308, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'zh-CN', '拾光', 'lumen', '照明与传感');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (309, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'en-US', 'Lumen', 'lumen', '照明与传感');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (310, NULL, NULL, NULL, NULL, NULL, NULL, 105, 'zh-CN', '川行', 'rivero', '出行与运动装备');
INSERT INTO mall_brand_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, brand_id, language_code, name, slug, description) VALUES (311, NULL, NULL, NULL, NULL, NULL, NULL, 105, 'en-US', 'Rivero', 'rivero', '出行与运动装备');


--
-- Data for Name: mall_brands; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_brands (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, logo_url) VALUES (100, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 0, '/demo-images/brand-100.webp');
INSERT INTO mall_brands (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, logo_url) VALUES (101, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 1, '/demo-images/brand-101.webp');
INSERT INTO mall_brands (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, logo_url) VALUES (102, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 2, '/demo-images/brand-102.webp');
INSERT INTO mall_brands (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, logo_url) VALUES (103, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 3, '/demo-images/brand-103.webp');
INSERT INTO mall_brands (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, logo_url) VALUES (104, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 4, '/demo-images/brand-104.webp');
INSERT INTO mall_brands (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, logo_url) VALUES (105, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 5, '/demo-images/brand-105.webp');


--
-- Data for Name: mall_cart_items; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_cart_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, cart_id, user_id, sku_id, quantity) VALUES (100, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 1, 1012, 1);
INSERT INTO mall_cart_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, cart_id, user_id, sku_id, quantity) VALUES (101, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 1, 1026, 2);
INSERT INTO mall_cart_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, cart_id, user_id, sku_id, quantity) VALUES (102, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 1, 1030, 1);
INSERT INTO mall_cart_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, cart_id, user_id, sku_id, quantity) VALUES (103, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 1, 1046, 3);
INSERT INTO mall_cart_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, cart_id, user_id, sku_id, quantity) VALUES (104, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 1, 1090, 1);


--
-- Data for Name: mall_carts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_carts (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id) VALUES (1, '2026-10-28 04:16:30.783667', NULL, NULL, 1, NULL, NULL, 0, 1);


--
-- Data for Name: mall_categories; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (3, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 1, '/1', 0, '/demo-images/cat-home.webp', NULL);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (4, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 2, '/2', 0, '/demo-images/cat-accessory.webp', NULL);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (5, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 3, '/3', 0, '/demo-images/cat-phone.webp', NULL);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (6, '2026-10-08 00:00:00', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 4, '/4', 0, '/demo-images/cat-wearable.webp', NULL);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (100, '2026-10-10 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 1, '/3/100', 1, '/demo-images/cat-lighting.webp', 3);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (101, '2026-10-11 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 2, '/3/101', 1, '/demo-images/cat-security.webp', 3);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (102, '2026-10-12 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 3, '/4/102', 1, '/demo-images/cat-charging.webp', 4);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (103, '2026-10-13 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 4, '/4/103', 1, '/demo-images/cat-audio.webp', 4);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (104, '2026-10-14 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 5, '/5/104', 1, '/demo-images/cat-flagship.webp', 5);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (105, '2026-10-15 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 6, '/5/105', 1, '/demo-images/cat-foldable.webp', 5);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (106, '2026-10-16 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 7, '/6/106', 1, '/demo-images/cat-watch.webp', 6);
INSERT INTO mall_categories (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, path, depth, image_url, parent_id) VALUES (107, '2026-10-17 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 8, '/6/107', 1, '/demo-images/cat-band.webp', 6);


--
-- Data for Name: mall_category_translations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (16, NULL, NULL, NULL, NULL, NULL, NULL, 3, 'zh-CN', '智能家居', NULL, NULL, NULL);
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (17, NULL, NULL, NULL, NULL, NULL, NULL, 4, 'zh-CN', '配件', NULL, NULL, NULL);
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (18, NULL, NULL, NULL, NULL, NULL, NULL, 5, 'zh-CN', '智能手机', NULL, NULL, NULL);
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (19, NULL, NULL, NULL, NULL, NULL, NULL, 6, 'zh-CN', '智能穿戴', NULL, NULL, NULL);
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (200, NULL, NULL, NULL, NULL, NULL, NULL, 3, 'en-US', 'Smart Home', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (201, NULL, NULL, NULL, NULL, NULL, NULL, 4, 'en-US', 'Accessories', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (202, NULL, NULL, NULL, NULL, NULL, NULL, 5, 'en-US', 'Smartphones', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (203, NULL, NULL, NULL, NULL, NULL, NULL, 6, 'en-US', 'Wearables', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (210, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'zh-CN', '智能照明', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (211, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'en-US', 'Lighting', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (212, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'zh-CN', '安防监控', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (213, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'en-US', 'Security', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (214, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'zh-CN', '充电器', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (215, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'en-US', 'Chargers', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (216, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'zh-CN', '音频配件', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (217, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'en-US', 'Audio', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (218, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'zh-CN', '旗舰手机', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (219, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'en-US', 'Flagship', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (220, NULL, NULL, NULL, NULL, NULL, NULL, 105, 'zh-CN', '折叠屏', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (221, NULL, NULL, NULL, NULL, NULL, NULL, 105, 'en-US', 'Foldable', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (222, NULL, NULL, NULL, NULL, NULL, NULL, 106, 'zh-CN', '智能手表', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (223, NULL, NULL, NULL, NULL, NULL, NULL, 106, 'en-US', 'Watches', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (224, NULL, NULL, NULL, NULL, NULL, NULL, 107, 'zh-CN', '运动手环', '', '', '');
INSERT INTO mall_category_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, category_id, language_code, name, slug, description, full_path) VALUES (225, NULL, NULL, NULL, NULL, NULL, NULL, 107, 'en-US', 'Bands', '', '', '');


--
-- Data for Name: mall_comment_likes; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (100, NULL, NULL, NULL, 0, 1, 100);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (101, NULL, NULL, NULL, 0, 1, 101);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (102, NULL, NULL, NULL, 0, 1, 106);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (103, NULL, NULL, NULL, 0, 1, 108);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (104, NULL, NULL, NULL, 0, 1, 110);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (105, NULL, NULL, NULL, 0, 1, 114);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (106, NULL, NULL, NULL, 0, 1, 117);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (107, NULL, NULL, NULL, 0, 1, 118);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (108, NULL, NULL, NULL, 0, 1, 126);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (109, NULL, NULL, NULL, 0, 1, 130);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (110, NULL, NULL, NULL, 0, 1, 131);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (111, NULL, NULL, NULL, 0, 1, 132);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (112, NULL, NULL, NULL, 0, 1, 134);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (113, NULL, NULL, NULL, 0, 1, 135);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (114, NULL, NULL, NULL, 0, 1, 139);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (115, NULL, NULL, NULL, 0, 1, 141);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (116, NULL, NULL, NULL, 0, 1, 143);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (117, NULL, NULL, NULL, 0, 1, 145);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (118, NULL, NULL, NULL, 0, 1, 146);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (119, NULL, NULL, NULL, 0, 1, 150);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (120, NULL, NULL, NULL, 0, 1, 152);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (121, NULL, NULL, NULL, 0, 1, 153);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (122, NULL, NULL, NULL, 0, 1, 154);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (123, NULL, NULL, NULL, 0, 1, 156);
INSERT INTO mall_comment_likes (id, created_at, updated_at, deleted_at, tenant_id, user_id, comment_id) VALUES (124, NULL, NULL, NULL, 0, 1, 157);


--
-- Data for Name: mall_comments; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (100, '2026-10-26 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 100, '客厅换了四只，暖光和冷光切换很自然，App 调色也顺手。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (101, '2026-10-29 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 100, '定时开灯很实用，早上起床不再摸黑找开关。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (102, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 100, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 101);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (103, '2026-10-28 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 100, '联动场景偶尔慢半拍，但整体够用。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (104, '2026-10-28 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 100, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 103);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (105, '2026-10-18 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 101, '指纹识别挺快，老人小孩都能用，就是体积比想象中大。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (106, '2026-10-10 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 101, '上门安装很专业，门框适配没费事。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (107, '2026-10-29 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 101, '已反馈给产品团队，后续版本会优化。', 'STATUS_APPROVED', NULL, 106);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (108, '2026-10-20 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 102, '手机、耳机、手表一起充，桌面清爽多了。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (109, '2026-10-28 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 102, '睡前放上去就能充，指示灯亮度也克制。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (110, '2026-10-25 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 102, '家里第二块，床头和书房各放一个。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (111, '2026-10-12 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 103, '夜视画面清晰，手机回看很流畅，就是云台转动有点声音。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (112, '2026-10-18 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 104, '墨水屏不刺眼，但刷新频率偏低，看实时数据要等几秒。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (113, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 104, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 112);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (114, '2026-10-14 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 104, '和加湿器联动很省心，湿度低了会自动开启。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (115, '2026-10-11 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 105, '出差带一个就够，笔记本也能充到满速。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (116, '2026-10-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 105, '已反馈给产品团队，后续版本会优化。', 'STATUS_APPROVED', NULL, 115);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (117, '2026-10-21 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 105, '体积小是小，插脚偶尔会松，需要摆正。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (118, '2026-10-11 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 105, '颜值在线，送人很合适。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (119, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 105, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 118);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (120, '2026-10-22 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 106, '低音比想象中好，浴室用也不怕水。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (121, '2026-10-21 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 106, '家里第二台，露营带出去很合适。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (122, '2026-10-26 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 106, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 121);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (123, '2026-10-24 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 106, '做工细节到位，就是按键手感偏硬。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (124, '2026-10-20 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 107, '双 HDMI 同时接显示器没问题，网口也很稳定。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (125, '2026-10-20 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 107, '说明书有点简略，好在接口标识清楚。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (126, '2026-10-17 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 107, '包装严实，物流第二天就到。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (127, '2026-10-27 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 107, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 126);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (128, '2026-10-17 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 108, '按键确实安静，但握感偏小，大手需要适应。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (129, '2026-10-27 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 108, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 128);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (130, '2026-10-18 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 109, '吸力很稳，就是铝合金边缘有点硌手。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (131, '2026-10-29 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 110, '双向快充很实用，飞机上也能带。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (132, '2026-10-26 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 111, '屏幕通透，120Hz 滑动很跟手。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (133, '2026-10-26 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 111, '已反馈给产品团队，后续版本会优化。', 'STATUS_APPROVED', NULL, 132);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (134, '2026-10-11 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 112, '长焦拍演唱会效果不错，卫星消息在户外很安心。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (135, '2026-10-13 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 113, '折痕控制得不错，分屏办公效率高。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (136, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 113, '已反馈给产品团队，后续版本会优化。', 'STATUS_APPROVED', NULL, 135);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (137, '2026-10-21 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 114, '千元档里质感算好的，双卡切换方便。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (138, '2026-10-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 114, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 137);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (139, '2026-10-25 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 114, '做工细节到位，接缝均匀。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (140, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 114, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 139);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (141, '2026-10-13 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 114, '性价比很高，但夜景拍照一般。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (142, '2026-10-27 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 114, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 141);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (143, '2026-10-26 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 115, '一英寸大底夜景很能打，就是机身偏重。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (144, '2026-10-29 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 115, '已反馈给产品团队，后续版本会优化。', 'STATUS_APPROVED', NULL, 143);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (145, '2026-10-28 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 115, '退货流程顺畅，两天到账。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (146, '2026-10-15 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 115, '可变光圈拍人像虚化自然。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (147, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 115, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 146);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (148, '2026-10-28 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 116, '客服响应快，换新流程顺利，但配送慢了一天。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (149, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 116, '感谢您的支持，使用中有任何问题随时联系我们。', 'STATUS_APPROVED', NULL, 148);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (150, '2026-10-14 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 116, '包装严实，物流第二天就到。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (151, '2026-10-21 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 116, '用了一周没遇到卡顿，GPS 轨迹很准。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (152, '2026-10-12 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 117, '血氧和心率监测挺准，戴着睡觉也轻。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (153, '2026-10-17 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 117, '续航一周多，通勤不用天天充。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (154, '2026-10-20 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 117, '颜值在线，送人很合适。', 'STATUS_APPROVED', 4, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (155, '2026-10-27 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 117, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 154);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (156, '2026-10-26 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 118, '定位很准，视频通话孩子很喜欢。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (157, '2026-10-22 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 118, '上课禁用模式很实用，家长放心。', 'STATUS_APPROVED', 5, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (158, '2026-10-26 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 118, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 157);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (159, '2026-10-26 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 118, '功能比想象中丰富，但表带偏硬。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (160, '2026-10-26 23:03:27.011658', '2026-10-31 13:23:11.840173', NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 119, '屏幕通透，户外亮度够用，但镜腿夹得有点紧。', 'STATUS_APPROVED', 3, NULL);
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, content_type, object_id, content, status, rating, parent_id) VALUES (161, '2026-10-27 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CONTENT_TYPE_PRODUCT', 119, '感谢认可，欢迎参加会员日活动。', 'STATUS_APPROVED', NULL, 160);


--
-- Data for Name: mall_coupon_templates; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (2, '2026-10-29 11:02:45.134922', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 1000, 0, '2026-10-29 11:02:45', '2026-11-28 11:02:45', 50, 1, 2, 'ACTIVE', false, 50);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (100, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 3000, 0, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 189, 'ACTIVE', true, 0);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (101, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 5000, 0, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 172, 'ACTIVE', true, 0);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (102, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 10000, 0, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 19, 'ACTIVE', true, 0);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (103, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'PERCENTAGE', 0, 10, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 75, 'ACTIVE', true, 0);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (104, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'PERCENTAGE', 0, 15, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 221, 'ACTIVE', false, 0);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (105, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 2000, 0, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 53, 'ACTIVE', true, 500);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (106, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 8800, 0, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 98, 'ACTIVE', true, 1200);
INSERT INTO mall_coupon_templates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, discount_type, discount_value, discount_percentage, valid_from, valid_until, max_redemptions, max_redemptions_per_user, redeemed_count, status, claimable, points_price) VALUES (107, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'FIXED_AMOUNT', 6600, 0, '2026-10-25 23:03:27.011658', '2026-11-29 23:03:27.011658', 1000, 5, 64, 'EXPIRED', true, 0);


--
-- Data for Name: mall_groupon_activities; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_groupon_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, group_price, group_size, group_valid_hours, per_user_limit, valid_from, valid_until, currency, status) VALUES (100, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '温湿度传感器三人团', 1026, 14900, 3, 48, 1, '2026-10-27 23:03:27.011658', '2026-11-13 23:03:27.011658', 'CNY', 'ACTIVE');
INSERT INTO mall_groupon_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, group_price, group_size, group_valid_hours, per_user_limit, valid_from, valid_until, currency, status) VALUES (101, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'Nova 12 Ultra 两人团', 1090, 499900, 2, 48, 1, '2026-10-27 23:03:27.011658', '2026-11-13 23:03:27.011658', 'CNY', 'ACTIVE');
INSERT INTO mall_groupon_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, group_price, group_size, group_valid_hours, per_user_limit, valid_from, valid_until, currency, status) VALUES (102, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '充电器五人团', 1030, 9900, 5, 48, 1, '2026-10-27 23:03:27.011658', '2026-11-13 23:03:27.011658', 'CNY', 'ACTIVE');
INSERT INTO mall_groupon_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, group_price, group_size, group_valid_hours, per_user_limit, valid_from, valid_until, currency, status) VALUES (103, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '智能门锁拼团', 1006, 109900, 3, 48, 1, '2026-10-27 23:03:27.011658', '2026-11-13 23:03:27.011658', 'CNY', 'ACTIVE');
INSERT INTO mall_groupon_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, group_price, group_size, group_valid_hours, per_user_limit, valid_from, valid_until, currency, status) VALUES (104, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '音箱新人团', 1039, 32900, 2, 48, 1, '2026-10-27 23:03:27.011658', '2026-11-13 23:03:27.011658', 'CNY', 'ACTIVE');


--
-- Data for Name: mall_groupon_groups; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_groupon_groups (id, created_at, updated_at, deleted_at, tenant_id, activity_id, leader_order_id, member_count, paid_count, target_size, expire_at, status) VALUES (100, '2026-10-28 23:03:27.011658', NULL, NULL, 0, 100, 101, 2, 2, 3, '2026-11-05 23:59:59', 'OPENING');
INSERT INTO mall_groupon_groups (id, created_at, updated_at, deleted_at, tenant_id, activity_id, leader_order_id, member_count, paid_count, target_size, expire_at, status) VALUES (101, '2026-10-28 23:03:27.011658', NULL, NULL, 0, 101, 102, 2, 2, 2, '2026-10-31 23:03:27.011658', 'SUCCEEDED');
INSERT INTO mall_groupon_groups (id, created_at, updated_at, deleted_at, tenant_id, activity_id, leader_order_id, member_count, paid_count, target_size, expire_at, status) VALUES (102, '2026-10-28 23:03:27.011658', NULL, NULL, 0, 102, 103, 1, 1, 5, '2026-11-05 23:59:59', 'OPENING');
INSERT INTO mall_groupon_groups (id, created_at, updated_at, deleted_at, tenant_id, activity_id, leader_order_id, member_count, paid_count, target_size, expire_at, status) VALUES (103, '2026-10-28 23:03:27.011658', NULL, NULL, 0, 103, 104, 3, 3, 3, '2026-10-31 23:03:27.011658', 'SUCCEEDED');
INSERT INTO mall_groupon_groups (id, created_at, updated_at, deleted_at, tenant_id, activity_id, leader_order_id, member_count, paid_count, target_size, expire_at, status) VALUES (104, '2026-10-28 23:03:27.011658', NULL, NULL, 0, 104, 105, 1, 0, 2, '2026-10-31 23:03:27.011658', 'EXPIRED');


--
-- Data for Name: mall_invoices; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_invoices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, order_id, user_id, invoice_number, invoice_type, status, buyer_name, buyer_tax_id, buyer_address, buyer_phone, amount) VALUES (100, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 101, 1, 'INV-2026-10000', 'PERSONAL', 'ISSUED', '张三', '', '浙江省 杭州市 西湖区 文三路 100 号 3 幢 501 室', '13800001234', 407500);
INSERT INTO mall_invoices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, order_id, user_id, invoice_number, invoice_type, status, buyer_name, buyer_tax_id, buyer_address, buyer_phone, amount) VALUES (101, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 102, 1, 'INV-2026-10001', 'COMPANY', 'PENDING', '李四', '91330100MA27XXXX01', '广东省 深圳市 南山区 科技园南区 A8 栋 12 层', '13900002345', 427500);
INSERT INTO mall_invoices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, order_id, user_id, invoice_number, invoice_type, status, buyer_name, buyer_tax_id, buyer_address, buyer_phone, amount) VALUES (102, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 106, 1, 'INV-2026-10002', 'PERSONAL', 'ISSUED', '王五', '', '北京市 朝阳区 望京街道阜通东大街 6 号院 2 号楼', '13700003456', 25800);
INSERT INTO mall_invoices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, order_id, user_id, invoice_number, invoice_type, status, buyer_name, buyer_tax_id, buyer_address, buyer_phone, amount) VALUES (103, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 107, 1, 'INV-2026-10003', 'COMPANY', 'CANCELLED', '张三', '91330100MA27XXXX03', '浙江省 杭州市 西湖区 文三路 100 号 3 幢 501 室', '13800001234', 721800);
INSERT INTO mall_invoices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, order_id, user_id, invoice_number, invoice_type, status, buyer_name, buyer_tax_id, buyer_address, buyer_phone, amount) VALUES (104, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 110, 1, 'INV-2026-10004', 'PERSONAL', 'ISSUED', '李四', '', '广东省 深圳市 南山区 科技园南区 A8 栋 12 层', '13900002345', 283600);
INSERT INTO mall_invoices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, order_id, user_id, invoice_number, invoice_type, status, buyer_name, buyer_tax_id, buyer_address, buyer_phone, amount) VALUES (105, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 104, 1, 'INV-2026-10005', 'COMPANY', 'PENDING', '王五', '91330100MA27XXXX05', '北京市 朝阳区 望京街道阜通东大街 6 号院 2 号楼', '13700003456', 809700);


--
-- Data for Name: mall_member_tiers; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_member_tiers (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, level, min_spend, points_multiplier, badge_icon, currency, status) VALUES (100, '2026-09-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '普通会员', 1, 0, 1, 'carbon:user', 'CNY', 'ACTIVE');
INSERT INTO mall_member_tiers (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, level, min_spend, points_multiplier, badge_icon, currency, status) VALUES (101, '2026-09-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '青铜会员', 2, 50000, 1, 'carbon:star', 'CNY', 'ACTIVE');
INSERT INTO mall_member_tiers (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, level, min_spend, points_multiplier, badge_icon, currency, status) VALUES (102, '2026-09-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '白银会员', 3, 200000, 2, 'carbon:star', 'CNY', 'ACTIVE');
INSERT INTO mall_member_tiers (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, level, min_spend, points_multiplier, badge_icon, currency, status) VALUES (103, '2026-09-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '黄金会员', 4, 500000, 3, 'carbon:trophy', 'CNY', 'ACTIVE');
INSERT INTO mall_member_tiers (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, level, min_spend, points_multiplier, badge_icon, currency, status) VALUES (104, '2026-09-30 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '铂金会员', 5, 1200000, 5, 'carbon:award', 'CNY', 'ACTIVE');


--
-- Data for Name: mall_order_items; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1000, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 100, 1, 1100, '{"sku_id" : 1100, "quantity" : 2, "currency" : "CNY", "unit_price" : 171900}', 2, 171900, 343800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1001, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 100, 1, 1090, '{"sku_id" : 1090, "quantity" : 2, "currency" : "CNY", "unit_price" : 549900}', 2, 549900, 1099800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1002, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 100, 1, 1060, '{"sku_id" : 1060, "quantity" : 1, "currency" : "CNY", "unit_price" : 15900}', 1, 15900, 15900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1003, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 101, 1, 1002, '{"sku_id" : 1002, "quantity" : 2, "currency" : "CNY", "unit_price" : 11900}', 2, 11900, 23800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1004, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 101, 1, 1100, '{"sku_id" : 1100, "quantity" : 2, "currency" : "CNY", "unit_price" : 171900}', 2, 171900, 343800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1005, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 101, 1, 1012, '{"sku_id" : 1012, "quantity" : 1, "currency" : "CNY", "unit_price" : 39900}', 1, 39900, 39900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1006, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 102, 1, 1046, '{"sku_id" : 1046, "quantity" : 2, "currency" : "CNY", "unit_price" : 33900}', 2, 33900, 67800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1007, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 102, 1, 1100, '{"sku_id" : 1100, "quantity" : 2, "currency" : "CNY", "unit_price" : 171900}', 2, 171900, 343800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1008, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 102, 1, 1060, '{"sku_id" : 1060, "quantity" : 1, "currency" : "CNY", "unit_price" : 15900}', 1, 15900, 15900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1009, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 103, 1, 1080, '{"sku_id" : 1080, "quantity" : 2, "currency" : "CNY", "unit_price" : 905900}', 2, 905900, 1811800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1010, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 103, 1, 1060, '{"sku_id" : 1060, "quantity" : 2, "currency" : "CNY", "unit_price" : 15900}', 2, 15900, 31800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1011, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 103, 1, 1046, '{"sku_id" : 1046, "quantity" : 2, "currency" : "CNY", "unit_price" : 33900}', 2, 33900, 67800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1012, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 104, 1, 1006, '{"sku_id" : 1006, "quantity" : 2, "currency" : "CNY", "unit_price" : 129900}', 2, 129900, 259800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1013, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 104, 1, 1090, '{"sku_id" : 1090, "quantity" : 1, "currency" : "CNY", "unit_price" : 549900}', 1, 549900, 549900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1014, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 105, 1, 1100, '{"sku_id" : 1100, "quantity" : 1, "currency" : "CNY", "unit_price" : 171900}', 1, 171900, 171900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1015, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 106, 1, 1024, '{"sku_id" : 1024, "quantity" : 2, "currency" : "CNY", "unit_price" : 12900}', 2, 12900, 25800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1016, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 107, 1, 1100, '{"sku_id" : 1100, "quantity" : 1, "currency" : "CNY", "unit_price" : 171900}', 1, 171900, 171900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1017, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 107, 1, 1090, '{"sku_id" : 1090, "quantity" : 1, "currency" : "CNY", "unit_price" : 549900}', 1, 549900, 549900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1018, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 108, 1, 1100, '{"sku_id" : 1100, "quantity" : 1, "currency" : "CNY", "unit_price" : 171900}', 1, 171900, 171900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1019, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 109, 1, 1006, '{"sku_id" : 1006, "quantity" : 1, "currency" : "CNY", "unit_price" : 129900}', 1, 129900, 129900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1020, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 109, 1, 1100, '{"sku_id" : 1100, "quantity" : 1, "currency" : "CNY", "unit_price" : 171900}', 1, 171900, 171900);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1021, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 109, 1, 1012, '{"sku_id" : 1012, "quantity" : 2, "currency" : "CNY", "unit_price" : 39900}', 2, 39900, 79800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1022, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 110, 1, 1002, '{"sku_id" : 1002, "quantity" : 2, "currency" : "CNY", "unit_price" : 11900}', 2, 11900, 23800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1023, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 110, 1, 1006, '{"sku_id" : 1006, "quantity" : 2, "currency" : "CNY", "unit_price" : 129900}', 2, 129900, 259800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1024, '2026-10-26 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 111, 1, 1006, '{"sku_id" : 1006, "quantity" : 2, "currency" : "CNY", "unit_price" : 129900}', 2, 129900, 259800);
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal) VALUES (1025, '2026-10-26 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 111, 1, 1100, '{"sku_id" : 1100, "quantity" : 2, "currency" : "CNY", "unit_price" : 171900}', 2, 171900, 343800);


--
-- Data for Name: mall_orders; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (100, '2026-10-15 23:03:27.011658', '2026-10-30 23:15:01.22969', NULL, 1, NULL, NULL, 0, 'CNY', 'seed-100', 'idem-seed-100', 1, 1454500, 1459500, 5000, 0, 0, '浙江省 杭州市 西湖区', 'CANCELLED', '张三', '13800001234', '浙江省 杭州市 西湖区 文三路 100 号 3 幢 501 室', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (101, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-101', 'idem-seed-101', 1, 407500, 407500, 0, 0, 0, '广东省 深圳市 南山区', 'PAID', '李四', '13900002345', '广东省 深圳市 南山区 科技园南区 A8 栋 12 层', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (102, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-102', 'idem-seed-102', 1, 427500, 427500, 0, 0, 0, '北京市 朝阳区', 'FULFILLED', '王五', '13700003456', '北京市 朝阳区 望京街道阜通东大街 6 号院 2 号楼', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (103, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-103', 'idem-seed-103', 1, 1911400, 1911400, 0, 0, 0, '浙江省 杭州市 西湖区', 'CANCELLED', '张三', '13800001234', '浙江省 杭州市 西湖区 文三路 100 号 3 幢 501 室', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (104, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-104', 'idem-seed-104', 1, 809700, 809700, 0, 0, 0, '广东省 深圳市 南山区', 'CLOSED', '李四', '13900002345', '广东省 深圳市 南山区 科技园南区 A8 栋 12 层', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (105, '2026-10-20 23:03:27.011658', '2026-10-30 23:15:01.283739', NULL, 1, NULL, NULL, 0, 'CNY', 'seed-105', 'idem-seed-105', 1, 168900, 171900, 3000, 0, 0, '北京市 朝阳区', 'CANCELLED', '王五', '13700003456', '北京市 朝阳区 望京街道阜通东大街 6 号院 2 号楼', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (106, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-106', 'idem-seed-106', 1, 25800, 25800, 0, 0, 0, '浙江省 杭州市 西湖区', 'FULFILLED', '张三', '13800001234', '浙江省 杭州市 西湖区 文三路 100 号 3 幢 501 室', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (107, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-107', 'idem-seed-107', 1, 721800, 721800, 0, 0, 0, '广东省 深圳市 南山区', 'FULFILLED', '李四', '13900002345', '广东省 深圳市 南山区 科技园南区 A8 栋 12 层', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (108, '2026-10-23 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-108', 'idem-seed-108', 1, 171900, 171900, 0, 0, 0, '北京市 朝阳区', 'CANCELLED', '王五', '13700003456', '北京市 朝阳区 望京街道阜通东大街 6 号院 2 号楼', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (109, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-109', 'idem-seed-109', 1, 378600, 381600, 3000, 0, 0, '浙江省 杭州市 西湖区', 'CLOSED', '张三', '13800001234', '浙江省 杭州市 西湖区 文三路 100 号 3 幢 501 室', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (110, '2026-10-25 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'CNY', 'seed-110', 'idem-seed-110', 1, 283600, 283600, 0, 0, 0, '广东省 深圳市 南山区', 'PAID', '李四', '13900002345', '广东省 深圳市 南山区 科技园南区 A8 栋 12 层', NULL, NULL);
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address, groupon_activity_id, groupon_group_id) VALUES (111, '2026-10-26 23:03:27.011658', '2026-10-30 23:15:01.317919', NULL, 1, NULL, NULL, 0, 'CNY', 'seed-111', 'idem-seed-111', 1, 603600, 603600, 0, 0, 0, '北京市 朝阳区', 'CANCELLED', '王五', '13700003456', '北京市 朝阳区 望京街道阜通东大街 6 号院 2 号楼', NULL, NULL);


--
-- Data for Name: mall_payment_channel_configs; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_payment_channel_configs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, channel, app_id, mch_id, sandbox, enabled, credentials, remark) VALUES (1, '2026-10-29 01:22:34.963477', '2026-10-29 01:33:53.992148', NULL, NULL, NULL, NULL, 0, '模拟支付（演示）', 'PAYMENT_CHANNEL_MOCK', NULL, NULL, true, true, NULL, 'smoke');


--
-- Data for Name: mall_payment_refunds; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_payment_refunds (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, transaction_id, user_id, amount, status, reason) VALUES (8, '2026-10-25 23:03:27.011658', '2026-10-25 23:03:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-refund-ref-109', 'demo-refund-109', 16, 1, 378600, 'SUCCEEDED', '商品与描述不符，协商退款');
INSERT INTO mall_payment_refunds (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, transaction_id, user_id, amount, status, reason) VALUES (9, '2026-10-30 23:03:27.011658', '2026-10-30 23:03:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-refund-ref-106', 'demo-refund-106', 14, 1, 25800, 'PENDING', '外包装破损，申请仅退款');


--
-- Data for Name: mall_payment_transactions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (14, '2026-10-21 23:06:27.011658', '2026-10-21 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-106', 'demo-pay-106', 'WECHAT', 'BUSINESS_TYPE_CONSUME', 106, 1, 25800, 'SUCCEEDED', 'DEMO-000106');
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (15, '2026-10-16 23:06:27.011658', '2026-10-16 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-101', 'demo-pay-101', 'BALANCE', 'BUSINESS_TYPE_CONSUME', 101, 1, 407500, 'SUCCEEDED', 'DEMO-000101');
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (16, '2026-10-24 23:06:27.011658', '2026-10-24 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-109', 'demo-pay-109', 'WECHAT', 'BUSINESS_TYPE_CONSUME', 109, 1, 378600, 'REFUNDED', 'DEMO-000109');
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (17, '2026-10-19 23:06:27.011658', '2026-10-19 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-104', 'demo-pay-104', 'BALANCE', 'BUSINESS_TYPE_CONSUME', 104, 1, 809700, 'SUCCEEDED', 'DEMO-000104');
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (18, '2026-10-17 23:06:27.011658', '2026-10-17 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-102', 'demo-pay-102', 'ALIPAY', 'BUSINESS_TYPE_CONSUME', 102, 1, 427500, 'SUCCEEDED', 'DEMO-000102');
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (19, '2026-10-22 23:06:27.011658', '2026-10-22 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-107', 'demo-pay-107', 'BALANCE', 'BUSINESS_TYPE_CONSUME', 107, 1, 721800, 'SUCCEEDED', 'DEMO-000107');
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no) VALUES (20, '2026-10-25 23:06:27.011658', '2026-10-25 23:06:27.011658', NULL, 1, NULL, NULL, 0, 'CNY', 'demo-ref-110', 'demo-pay-110', 'BALANCE', 'BUSINESS_TYPE_CONSUME', 110, 1, 283600, 'SUCCEEDED', 'DEMO-000110');


--
-- Data for Name: mall_points_ledgers; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (4, '2026-10-29 07:41:17.556284', NULL, NULL, 0, 1, 5, 302, 'POINTS_TYPE_CHECK_IN', '每日签到', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (5, '2026-10-29 11:05:47.023256', NULL, NULL, 0, 1, -50, 252, 'POINTS_TYPE_DEDUCT', '积分兑换优惠券', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (6, '2026-10-30 16:37:35.523714', NULL, NULL, 0, 1, -50, 202, 'POINTS_TYPE_DEDUCT', '积分兑换优惠券', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (100, '2026-10-10 23:03:27.011658', NULL, NULL, 0, 1, 100, 100, 'POINTS_TYPE_EARN', '下单返积分', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (101, '2026-10-12 23:03:27.011658', NULL, NULL, 0, 1, 200, 300, 'POINTS_TYPE_CHECK_IN', '签到奖励', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (102, '2026-10-14 23:03:27.011658', NULL, NULL, 0, 1, -500, 0, 'POINTS_TYPE_DEDUCT', '兑换优惠券', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (103, '2026-10-16 23:03:27.011658', NULL, NULL, 0, 1, 300, 300, 'POINTS_TYPE_EARN', '评价晒单', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (104, '2026-10-18 23:03:27.011658', NULL, NULL, 0, 1, 150, 450, 'POINTS_TYPE_EARN', '会员日翻倍', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (105, '2026-10-20 23:03:27.011658', NULL, NULL, 0, 1, -1200, 0, 'POINTS_TYPE_DEDUCT', '兑换满减券', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (106, '2026-10-22 23:03:27.011658', NULL, NULL, 0, 1, 800, 800, 'POINTS_TYPE_EARN', '订单完成返积分', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (107, '2026-10-24 23:03:27.011658', NULL, NULL, 0, 1, 60, 860, 'POINTS_TYPE_EARN', '生日礼包', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (108, '2026-10-26 23:03:27.011658', NULL, NULL, 0, 1, -30, 830, 'POINTS_TYPE_DEDUCT', '活动扣减', NULL);
INSERT INTO mall_points_ledgers (id, created_at, updated_at, deleted_at, tenant_id, user_id, delta, balance_after, type, reason, ref_order_id) VALUES (109, '2026-10-28 23:03:27.011658', NULL, NULL, 0, 1, 500, 1330, 'POINTS_TYPE_CHECK_IN', '连续签到第 7 天', NULL);


--
-- Data for Name: mall_product_attribute_translations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (1, '2026-10-28 11:10:11.510135', NULL, NULL, 1, NULL, NULL, 1, 'zh-CN', '颜色');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (2, '2026-10-28 11:10:11.510135', NULL, NULL, 1, NULL, NULL, 1, 'en-US', 'Color');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (400, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'zh-CN', '配色');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (401, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'en-US', 'Finish');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (402, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'zh-CN', '版本');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (403, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'en-US', 'Edition');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (404, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'zh-CN', '存储');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (405, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'en-US', 'Storage');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (406, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'zh-CN', '尺码');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (407, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'en-US', 'Size');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (408, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'zh-CN', '套餐');
INSERT INTO mall_product_attribute_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, attribute_id, language_code, name) VALUES (409, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'en-US', 'Bundle');


--
-- Data for Name: mall_product_attribute_value_translations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (9, NULL, NULL, NULL, NULL, NULL, NULL, 2, 'zh-CN', '曜石黑');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (10, NULL, NULL, NULL, NULL, NULL, NULL, 3, 'zh-CN', '皓月白');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5010, NULL, NULL, NULL, NULL, NULL, NULL, 501, 'zh-CN', '星辰银');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5011, NULL, NULL, NULL, NULL, NULL, NULL, 501, 'en-US', 'Silver');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5020, NULL, NULL, NULL, NULL, NULL, NULL, 502, 'zh-CN', '暮山紫');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5021, NULL, NULL, NULL, NULL, NULL, NULL, 502, 'en-US', 'Violet');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5030, NULL, NULL, NULL, NULL, NULL, NULL, 503, 'zh-CN', '松柏绿');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5031, NULL, NULL, NULL, NULL, NULL, NULL, 503, 'en-US', 'Pine');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5040, NULL, NULL, NULL, NULL, NULL, NULL, 504, 'zh-CN', '标准版');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5041, NULL, NULL, NULL, NULL, NULL, NULL, 504, 'en-US', 'Standard');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5050, NULL, NULL, NULL, NULL, NULL, NULL, 505, 'zh-CN', 'Pro 版');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5051, NULL, NULL, NULL, NULL, NULL, NULL, 505, 'en-US', 'Pro');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5060, NULL, NULL, NULL, NULL, NULL, NULL, 506, 'zh-CN', '128GB');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5061, NULL, NULL, NULL, NULL, NULL, NULL, 506, 'en-US', '128GB');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5070, NULL, NULL, NULL, NULL, NULL, NULL, 507, 'zh-CN', '256GB');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5071, NULL, NULL, NULL, NULL, NULL, NULL, 507, 'en-US', '256GB');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5080, NULL, NULL, NULL, NULL, NULL, NULL, 508, 'zh-CN', '512GB');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5081, NULL, NULL, NULL, NULL, NULL, NULL, 508, 'en-US', '512GB');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5090, NULL, NULL, NULL, NULL, NULL, NULL, 509, 'zh-CN', 'S');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5091, NULL, NULL, NULL, NULL, NULL, NULL, 509, 'en-US', 'S');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5100, NULL, NULL, NULL, NULL, NULL, NULL, 510, 'zh-CN', 'M');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5101, NULL, NULL, NULL, NULL, NULL, NULL, 510, 'en-US', 'M');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5110, NULL, NULL, NULL, NULL, NULL, NULL, 511, 'zh-CN', 'L');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5111, NULL, NULL, NULL, NULL, NULL, NULL, 511, 'en-US', 'L');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5120, NULL, NULL, NULL, NULL, NULL, NULL, 512, 'zh-CN', '单机装');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5121, NULL, NULL, NULL, NULL, NULL, NULL, 512, 'en-US', 'Solo');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5130, NULL, NULL, NULL, NULL, NULL, NULL, 513, 'zh-CN', '尊享套装');
INSERT INTO mall_product_attribute_value_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, value_id, language_code, display_name) VALUES (5131, NULL, NULL, NULL, NULL, NULL, NULL, 513, 'en-US', 'Deluxe Kit');


--
-- Data for Name: mall_product_attribute_values; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (2, NULL, NULL, NULL, NULL, NULL, NULL, 1, 1);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (3, NULL, NULL, NULL, NULL, NULL, NULL, 2, 1);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (501, NULL, NULL, NULL, NULL, NULL, NULL, 0, 100);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (502, NULL, NULL, NULL, NULL, NULL, NULL, 1, 100);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (503, NULL, NULL, NULL, NULL, NULL, NULL, 2, 100);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (504, NULL, NULL, NULL, NULL, NULL, NULL, 0, 101);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (505, NULL, NULL, NULL, NULL, NULL, NULL, 1, 101);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (506, NULL, NULL, NULL, NULL, NULL, NULL, 0, 102);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (507, NULL, NULL, NULL, NULL, NULL, NULL, 1, 102);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (508, NULL, NULL, NULL, NULL, NULL, NULL, 2, 102);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (509, NULL, NULL, NULL, NULL, NULL, NULL, 0, 103);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (510, NULL, NULL, NULL, NULL, NULL, NULL, 1, 103);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (511, NULL, NULL, NULL, NULL, NULL, NULL, 2, 103);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (512, NULL, NULL, NULL, NULL, NULL, NULL, 0, 104);
INSERT INTO mall_product_attribute_values (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, attribute_id) VALUES (513, NULL, NULL, NULL, NULL, NULL, NULL, 1, 104);


--
-- Data for Name: mall_product_attributes; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_product_attributes (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order) VALUES (1, '2026-10-28 11:10:11.499362', NULL, NULL, 1, NULL, NULL, 0);
INSERT INTO mall_product_attributes (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order) VALUES (100, NULL, NULL, NULL, NULL, NULL, NULL, 0);
INSERT INTO mall_product_attributes (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order) VALUES (101, NULL, NULL, NULL, NULL, NULL, NULL, 1);
INSERT INTO mall_product_attributes (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order) VALUES (102, NULL, NULL, NULL, NULL, NULL, NULL, 2);
INSERT INTO mall_product_attributes (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order) VALUES (103, NULL, NULL, NULL, NULL, NULL, NULL, 3);
INSERT INTO mall_product_attributes (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order) VALUES (104, NULL, NULL, NULL, NULL, NULL, NULL, 4);


--
-- Data for Name: mall_product_translations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1000, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'zh-CN', '智能灯泡 Lumen 二代', 'p-100', '1600 万色可调，支持分组与定时', '智能灯泡 Lumen 二代 是拾光面向日常使用场景打造的智能照明产品。1600 万色可调，支持分组与定时。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1001, NULL, NULL, NULL, NULL, NULL, NULL, 100, 'en-US', 'Lumen Bulb Gen 2', 'p-100-en', 'Lumen Bulb Gen 2', '智能灯泡 Lumen 二代 是拾光面向日常使用场景打造的智能照明产品。1600 万色可调，支持分组与定时。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1002, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'zh-CN', '智能门锁 Vault Pro', 'p-101', '指纹+人脸双模，C 级锁芯', '智能门锁 Vault Pro 是风铃面向日常使用场景打造的安防监控产品。指纹+人脸双模，C 级锁芯。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1003, NULL, NULL, NULL, NULL, NULL, NULL, 101, 'en-US', 'Vault Pro Lock', 'p-101-en', 'Vault Pro Lock', '智能门锁 Vault Pro 是风铃面向日常使用场景打造的安防监控产品。指纹+人脸双模，C 级锁芯。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1004, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'zh-CN', '三合一无线充电板', 'p-102', '手机 / 耳机 / 手表同时回血', '三合一无线充电板 是曜石面向日常使用场景打造的充电器产品。手机 / 耳机 / 手表同时回血。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1005, NULL, NULL, NULL, NULL, NULL, NULL, 102, 'en-US', 'Trio Charging Pad', 'p-102-en', 'Trio Charging Pad', '三合一无线充电板 是曜石面向日常使用场景打造的充电器产品。手机 / 耳机 / 手表同时回血。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1006, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'zh-CN', '智能摄像头 Sentinel 2K', 'p-103', '2K 全彩夜视，本地存储不订阅', '智能摄像头 Sentinel 2K 是风铃面向日常使用场景打造的安防监控产品。2K 全彩夜视，本地存储不订阅。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1007, NULL, NULL, NULL, NULL, NULL, NULL, 103, 'en-US', 'Sentinel 2K Camera', 'p-103-en', 'Sentinel 2K Camera', '智能摄像头 Sentinel 2K 是风铃面向日常使用场景打造的安防监控产品。2K 全彩夜视，本地存储不订阅。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1008, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'zh-CN', '温湿度传感器 Breeze', 'p-104', '墨水屏显示，联动空调与加湿器', '温湿度传感器 Breeze 是拾光面向日常使用场景打造的智能照明产品。墨水屏显示，联动空调与加湿器。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1009, NULL, NULL, NULL, NULL, NULL, NULL, 104, 'en-US', 'Breeze Sensor', 'p-104-en', 'Breeze Sensor', '温湿度传感器 Breeze 是拾光面向日常使用场景打造的智能照明产品。墨水屏显示，联动空调与加湿器。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1010, NULL, NULL, NULL, NULL, NULL, NULL, 105, 'zh-CN', '氮化镓充电器 65W', 'p-105', '口红大小，笔记本也能喂饱', '氮化镓充电器 65W 是曜石面向日常使用场景打造的充电器产品。口红大小，笔记本也能喂饱。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1011, NULL, NULL, NULL, NULL, NULL, NULL, 105, 'en-US', 'GaN Charger 65W', 'p-105-en', 'GaN Charger 65W', '氮化镓充电器 65W 是曜石面向日常使用场景打造的充电器产品。口红大小，笔记本也能喂饱。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1012, NULL, NULL, NULL, NULL, NULL, NULL, 106, 'zh-CN', '便携蓝牙音箱 Echo Mini', 'p-106', 'IP67 防水，24 小时续航', '便携蓝牙音箱 Echo Mini 是云雀面向日常使用场景打造的音频配件产品。IP67 防水，24 小时续航。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1013, NULL, NULL, NULL, NULL, NULL, NULL, 106, 'en-US', 'Echo Mini Speaker', 'p-106-en', 'Echo Mini Speaker', '便携蓝牙音箱 Echo Mini 是云雀面向日常使用场景打造的音频配件产品。IP67 防水，24 小时续航。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1014, NULL, NULL, NULL, NULL, NULL, NULL, 107, 'zh-CN', 'Type-C 扩展坞 8 合 1', 'p-107', '双 HDMI + 千兆网口 + SD 读卡', 'Type-C 扩展坞 8 合 1 是曜石面向日常使用场景打造的充电器产品。双 HDMI + 千兆网口 + SD 读卡。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1015, NULL, NULL, NULL, NULL, NULL, NULL, 107, 'en-US', 'USB-C Hub 8-in-1', 'p-107-en', 'USB-C Hub 8-in-1', 'Type-C 扩展坞 8 合 1 是曜石面向日常使用场景打造的充电器产品。双 HDMI + 千兆网口 + SD 读卡。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1016, NULL, NULL, NULL, NULL, NULL, NULL, 108, 'zh-CN', '无线静音鼠标 Glide', 'p-108', '4200 DPI，两档静音微动', '无线静音鼠标 Glide 是曜石面向日常使用场景打造的充电器产品。4200 DPI，两档静音微动。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1017, NULL, NULL, NULL, NULL, NULL, NULL, 108, 'en-US', 'Glide Wireless Mouse', 'p-108-en', 'Glide Wireless Mouse', '无线静音鼠标 Glide 是曜石面向日常使用场景打造的充电器产品。4200 DPI，两档静音微动。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1018, NULL, NULL, NULL, NULL, NULL, NULL, 109, 'zh-CN', '磁吸手机支架 Aero', 'p-109', '航空铝，360° 无级悬停', '磁吸手机支架 Aero 是曜石面向日常使用场景打造的音频配件产品。航空铝，360° 无级悬停。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1019, NULL, NULL, NULL, NULL, NULL, NULL, 109, 'en-US', 'Aero MagSafe Stand', 'p-109-en', 'Aero MagSafe Stand', '磁吸手机支架 Aero 是曜石面向日常使用场景打造的音频配件产品。航空铝，360° 无级悬停。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1020, NULL, NULL, NULL, NULL, NULL, NULL, 110, 'zh-CN', '移动电源 Pebble 10000', 'p-110', '10000mAh 双向 30W，可上飞机', '移动电源 Pebble 10000 是曜石面向日常使用场景打造的充电器产品。10000mAh 双向 30W，可上飞机。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1021, NULL, NULL, NULL, NULL, NULL, NULL, 110, 'en-US', 'Pebble Power Bank', 'p-110-en', 'Pebble Power Bank', '移动电源 Pebble 10000 是曜石面向日常使用场景打造的充电器产品。10000mAh 双向 30W，可上飞机。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1022, NULL, NULL, NULL, NULL, NULL, NULL, 111, 'zh-CN', '智能手机 Nova 12', 'p-111', '6.7 英寸 120Hz 直屏，5000mAh', '智能手机 Nova 12 是曜石面向日常使用场景打造的旗舰手机产品。6.7 英寸 120Hz 直屏，5000mAh。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1023, NULL, NULL, NULL, NULL, NULL, NULL, 111, 'en-US', 'Nova 12 Smartphone', 'p-111-en', 'Nova 12 Smartphone', '智能手机 Nova 12 是曜石面向日常使用场景打造的旗舰手机产品。6.7 英寸 120Hz 直屏，5000mAh。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1024, NULL, NULL, NULL, NULL, NULL, NULL, 112, 'zh-CN', '智能手机 Nova 12 Pro', 'p-112', '潜望长焦 + 卫星消息', '智能手机 Nova 12 Pro 是曜石面向日常使用场景打造的旗舰手机产品。潜望长焦 + 卫星消息。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1025, NULL, NULL, NULL, NULL, NULL, NULL, 112, 'en-US', 'Nova 12 Pro', 'p-112-en', 'Nova 12 Pro', '智能手机 Nova 12 Pro 是曜石面向日常使用场景打造的旗舰手机产品。潜望长焦 + 卫星消息。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1026, NULL, NULL, NULL, NULL, NULL, NULL, 113, 'zh-CN', '折叠手机 Fold X', 'p-113', '8 英寸折叠屏，悬停分屏', '折叠手机 Fold X 是曜石面向日常使用场景打造的折叠屏产品。8 英寸折叠屏，悬停分屏。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1027, NULL, NULL, NULL, NULL, NULL, NULL, 113, 'en-US', 'Fold X', 'p-113-en', 'Fold X', '折叠手机 Fold X 是曜石面向日常使用场景打造的折叠屏产品。8 英寸折叠屏，悬停分屏。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1028, NULL, NULL, NULL, NULL, NULL, NULL, 114, 'zh-CN', '智能手机 Lite 5G', 'p-114', '千元档长续航，双卡双待', '智能手机 Lite 5G 是曜石面向日常使用场景打造的旗舰手机产品。千元档长续航，双卡双待。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1029, NULL, NULL, NULL, NULL, NULL, NULL, 114, 'en-US', 'Nova Lite 5G', 'p-114-en', 'Nova Lite 5G', '智能手机 Lite 5G 是曜石面向日常使用场景打造的旗舰手机产品。千元档长续航，双卡双待。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1030, NULL, NULL, NULL, NULL, NULL, NULL, 115, 'zh-CN', '智能手机 Nova 12 Ultra', 'p-115', '一英寸大底主摄，可变光圈', '智能手机 Nova 12 Ultra 是曜石面向日常使用场景打造的旗舰手机产品。一英寸大底主摄，可变光圈。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1031, NULL, NULL, NULL, NULL, NULL, NULL, 115, 'en-US', 'Nova 12 Ultra', 'p-115-en', 'Nova 12 Ultra', '智能手机 Nova 12 Ultra 是曜石面向日常使用场景打造的旗舰手机产品。一英寸大底主摄，可变光圈。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1032, NULL, NULL, NULL, NULL, NULL, NULL, 116, 'zh-CN', '智能手表 Watch S3', 'p-116', '双频 GPS，14 天续航，蓝牙通话', '智能手表 Watch S3 是极目面向日常使用场景打造的智能手表产品。双频 GPS，14 天续航，蓝牙通话。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1033, NULL, NULL, NULL, NULL, NULL, NULL, 116, 'en-US', 'Watch S3', 'p-116-en', 'Watch S3', '智能手表 Watch S3 是极目面向日常使用场景打造的智能手表产品。双频 GPS，14 天续航，蓝牙通话。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1034, NULL, NULL, NULL, NULL, NULL, NULL, 117, 'zh-CN', '运动手环 Band 7', 'p-117', '1.62 英寸 AMOLED，血氧心率', '运动手环 Band 7 是川行面向日常使用场景打造的运动手环产品。1.62 英寸 AMOLED，血氧心率。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1035, NULL, NULL, NULL, NULL, NULL, NULL, 117, 'en-US', 'Band 7', 'p-117-en', 'Band 7', '运动手环 Band 7 是川行面向日常使用场景打造的运动手环产品。1.62 英寸 AMOLED，血氧心率。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1036, NULL, NULL, NULL, NULL, NULL, NULL, 118, 'zh-CN', '儿童电话手表 KidSafe', 'p-118', '4G 视频通话，楼层级定位', '儿童电话手表 KidSafe 是极目面向日常使用场景打造的智能手表产品。4G 视频通话，楼层级定位。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1037, NULL, NULL, NULL, NULL, NULL, NULL, 118, 'en-US', 'KidSafe Watch', 'p-118-en', 'KidSafe Watch', '儿童电话手表 KidSafe 是极目面向日常使用场景打造的智能手表产品。4G 视频通话，楼层级定位。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1038, NULL, NULL, NULL, NULL, NULL, NULL, 119, 'zh-CN', '智能眼镜 Vista AR', 'p-119', '43g 轻量，实时翻译字幕', '智能眼镜 Vista AR 是极目面向日常使用场景打造的智能手表产品。43g 轻量，实时翻译字幕。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');
INSERT INTO mall_product_translations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, language_code, name, slug, short_description, long_description) VALUES (1039, NULL, NULL, NULL, NULL, NULL, NULL, 119, 'en-US', 'Vista AR Glasses', 'p-119-en', 'Vista AR Glasses', '智能眼镜 Vista AR 是极目面向日常使用场景打造的智能手表产品。43g 轻量，实时翻译字幕。整机采用低功耗方案与可回收包装材料，随机附赠说明书与一年质保；支持 7 天无理由退货、全国联保与上门取件。');


--
-- Data for Name: mall_products; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (100, '2026-09-30 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 1, 'PRODUCT_STATUS_ACTIVE', 100, 104, '/demo-images/prod-100.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (101, '2026-10-01 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 2, 'PRODUCT_STATUS_ACTIVE', 101, 100, '/demo-images/prod-101.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (102, '2026-10-02 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 3, 'PRODUCT_STATUS_ACTIVE', 102, 101, '/demo-images/prod-102.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (103, '2026-10-03 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 4, 'PRODUCT_STATUS_ACTIVE', 101, 100, '/demo-images/prod-103.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (104, '2026-10-04 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 5, 'PRODUCT_STATUS_ACTIVE', 100, 104, '/demo-images/prod-104.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (105, '2026-10-05 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 6, 'PRODUCT_STATUS_ACTIVE', 102, 101, '/demo-images/prod-105.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (106, '2026-10-06 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 7, 'PRODUCT_STATUS_ACTIVE', 103, 102, '/demo-images/prod-106.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (107, '2026-10-07 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 8, 'PRODUCT_STATUS_ACTIVE', 102, 101, '/demo-images/prod-107.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (108, '2026-10-08 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 9, 'PRODUCT_STATUS_ACTIVE', 102, 101, '/demo-images/prod-108.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (109, '2026-10-09 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 10, 'PRODUCT_STATUS_ACTIVE', 103, 101, '/demo-images/prod-109.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (110, '2026-10-10 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 11, 'PRODUCT_STATUS_ACTIVE', 102, 101, '/demo-images/prod-110.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (111, '2026-10-11 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 12, 'PRODUCT_STATUS_ACTIVE', 104, 101, '/demo-images/prod-111.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (112, '2026-10-12 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 13, 'PRODUCT_STATUS_ACTIVE', 104, 101, '/demo-images/prod-112.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (113, '2026-10-13 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 14, 'PRODUCT_STATUS_ACTIVE', 105, 101, '/demo-images/prod-113.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (114, '2026-10-14 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 15, 'PRODUCT_STATUS_ACTIVE', 104, 101, '/demo-images/prod-114.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (115, '2026-10-15 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 16, 'PRODUCT_STATUS_ACTIVE', 104, 101, '/demo-images/prod-115.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (116, '2026-10-16 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 17, 'PRODUCT_STATUS_ACTIVE', 106, 103, '/demo-images/prod-116.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (117, '2026-10-17 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 18, 'PRODUCT_STATUS_ACTIVE', 107, 105, '/demo-images/prod-117.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (118, '2026-10-18 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 19, 'PRODUCT_STATUS_ACTIVE', 106, 103, '/demo-images/prod-118.webp');
INSERT INTO mall_products (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, status, category_id, brand_id, image_url) VALUES (119, '2026-10-19 23:03:27.011658', '2026-10-31 13:58:20.702562', NULL, 1, NULL, NULL, 20, 'PRODUCT_STATUS_ACTIVE', 106, 103, '/demo-images/prod-119.webp');


--
-- Data for Name: mall_promotions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_promotions (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, name, promo_type, product_id, special_price, threshold_amount, reduction_amount, priority, valid_from, valid_until, status) VALUES (2, '2026-10-28 04:55:59.773229', '2026-10-28 04:55:59.773229', NULL, 1, NULL, NULL, 0, 'CNY', '店铺满300减40', 'FULL_REDUCTION', 0, 0, 30000, 4000, 5, '2026-10-01 00:00:00', '2027-01-01 00:00:00', 'ACTIVE');


--
-- Data for Name: mall_seckill_activities; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_seckill_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, seckill_price, total_stock, sold_count, per_user_limit, priority, valid_from, valid_until, currency, status) VALUES (100, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '三合一充电板限时抢购', 1012, 29900, 50, 18, 1, 100, '2026-10-29 23:03:27.011658', '2026-11-30 23:59:59', 'CNY', 'ACTIVE');
INSERT INTO mall_seckill_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, seckill_price, total_stock, sold_count, per_user_limit, priority, valid_from, valid_until, currency, status) VALUES (101, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '温湿度传感器秒杀', 1026, 14900, 60, 12, 1, 80, '2026-10-29 23:03:27.011658', '2026-11-30 23:59:59', 'CNY', 'ACTIVE');
INSERT INTO mall_seckill_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, seckill_price, total_stock, sold_count, per_user_limit, priority, valid_from, valid_until, currency, status) VALUES (102, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 'Nova 12 Ultra 新品首发', 1090, 499900, 30, 12, 1, 60, '2026-10-29 23:03:27.011658', '2026-11-30 23:59:59', 'CNY', 'ACTIVE');
INSERT INTO mall_seckill_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, seckill_price, total_stock, sold_count, per_user_limit, priority, valid_from, valid_until, currency, status) VALUES (103, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '移动电源秒杀夜场', 1061, 14900, 10, 0, 1, 40, '2026-10-31 12:00:00', '2026-11-02 23:59:59', 'CNY', 'ACTIVE');
INSERT INTO mall_seckill_activities (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, name, sku_id, seckill_price, total_stock, sold_count, per_user_limit, priority, valid_from, valid_until, currency, status) VALUES (104, '2026-10-24 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, '智能门锁返场（已结束）', 1006, 99900, 100, 100, 1, 20, '2026-10-22 23:03:27.011658', '2026-10-24 23:03:27.011658', 'CNY', 'INACTIVE');


--
-- Data for Name: mall_shipments; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_shipments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, carrier, tracking_number, status, tracking_events) VALUES (1, '2026-10-25 07:33:18.231339', '2026-10-26 07:33:18.231339', NULL, 1, NULL, NULL, 0, 106, 1, '顺丰速运', 'SF2026102800001', 'DELIVERED', '[{"time":"2026-10-28 10:00","event":"包裹已签收"},{"time":"2026-10-27 21:00","event":"运输中"},{"time":"2026-10-27 09:00","event":"已发货"}]');
INSERT INTO mall_shipments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, carrier, tracking_number, status, tracking_events) VALUES (2, '2026-10-27 07:33:18.231339', '2026-10-28 07:33:18.231339', NULL, 1, NULL, NULL, 0, 107, 1, '京东物流', 'JD2026102900002', 'DELIVERED', '[{"time":"2026-10-29 14:00","event":"包裹已签收"},{"time":"2026-10-28 20:00","event":"运输中"}]');
INSERT INTO mall_shipments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, carrier, tracking_number, status, tracking_events) VALUES (4, '2026-10-30 07:33:18.231339', '2026-10-30 07:33:18.231339', NULL, 1, NULL, NULL, 0, 109, 1, '顺丰速运', 'SF2026103100004', 'SHIPPED', '[{"time":"2026-10-31 09:00","event":"已发货"}]');
INSERT INTO mall_shipments (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, order_id, user_id, carrier, tracking_number, status, tracking_events) VALUES (5, '2026-10-31 07:33:18.231339', '2026-10-31 07:33:18.231339', NULL, 1, NULL, NULL, 0, 110, 1, '圆通速递', 'YT2026103100005', 'PENDING', '[]');


--
-- Data for Name: mall_shipping_rates; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_shipping_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, base_fee, per_unit_fee, status) VALUES (2, '2026-10-23 07:32:55.389343', '2026-10-23 07:32:55.389343', NULL, 1, NULL, NULL, 0, 'CNY', '北京', 800, 200, 'ACTIVE');
INSERT INTO mall_shipping_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, base_fee, per_unit_fee, status) VALUES (3, '2026-10-23 07:32:55.389343', '2026-10-23 07:32:55.389343', NULL, 1, NULL, NULL, 0, 'CNY', '上海', 800, 200, 'ACTIVE');
INSERT INTO mall_shipping_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, base_fee, per_unit_fee, status) VALUES (5, '2026-10-23 07:34:03.477116', '2026-10-23 07:34:03.477116', NULL, 1, NULL, NULL, 0, 'CNY', '广东', 1000, 300, 'ACTIVE');
INSERT INTO mall_shipping_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, base_fee, per_unit_fee, status) VALUES (6, '2026-10-23 07:34:03.477116', '2026-10-23 07:34:03.477116', NULL, 1, NULL, NULL, 0, 'CNY', '浙江', 1000, 250, 'ACTIVE');
INSERT INTO mall_shipping_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, base_fee, per_unit_fee, status) VALUES (7, '2026-10-23 07:34:03.477116', '2026-10-23 07:34:03.477116', NULL, 1, NULL, NULL, 0, 'CNY', '四川', 1500, 400, 'INACTIVE');


--
-- Data for Name: mall_sku_attribute_combinations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1000, NULL, NULL, NULL, NULL, NULL, NULL, 1000, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1001, NULL, NULL, NULL, NULL, NULL, NULL, 1000, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1002, NULL, NULL, NULL, NULL, NULL, NULL, 1001, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1003, NULL, NULL, NULL, NULL, NULL, NULL, 1001, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1004, NULL, NULL, NULL, NULL, NULL, NULL, 1002, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1005, NULL, NULL, NULL, NULL, NULL, NULL, 1002, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1006, NULL, NULL, NULL, NULL, NULL, NULL, 1003, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1007, NULL, NULL, NULL, NULL, NULL, NULL, 1003, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1008, NULL, NULL, NULL, NULL, NULL, NULL, 1004, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1009, NULL, NULL, NULL, NULL, NULL, NULL, 1004, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1010, NULL, NULL, NULL, NULL, NULL, NULL, 1005, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1011, NULL, NULL, NULL, NULL, NULL, NULL, 1005, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1012, NULL, NULL, NULL, NULL, NULL, NULL, 1006, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1013, NULL, NULL, NULL, NULL, NULL, NULL, 1006, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1014, NULL, NULL, NULL, NULL, NULL, NULL, 1007, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1015, NULL, NULL, NULL, NULL, NULL, NULL, 1007, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1016, NULL, NULL, NULL, NULL, NULL, NULL, 1008, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1017, NULL, NULL, NULL, NULL, NULL, NULL, 1008, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1018, NULL, NULL, NULL, NULL, NULL, NULL, 1009, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1019, NULL, NULL, NULL, NULL, NULL, NULL, 1009, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1020, NULL, NULL, NULL, NULL, NULL, NULL, 1010, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1021, NULL, NULL, NULL, NULL, NULL, NULL, 1010, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1022, NULL, NULL, NULL, NULL, NULL, NULL, 1011, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1023, NULL, NULL, NULL, NULL, NULL, NULL, 1011, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1024, NULL, NULL, NULL, NULL, NULL, NULL, 1012, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1025, NULL, NULL, NULL, NULL, NULL, NULL, 1012, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1026, NULL, NULL, NULL, NULL, NULL, NULL, 1013, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1027, NULL, NULL, NULL, NULL, NULL, NULL, 1013, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1028, NULL, NULL, NULL, NULL, NULL, NULL, 1014, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1029, NULL, NULL, NULL, NULL, NULL, NULL, 1014, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1030, NULL, NULL, NULL, NULL, NULL, NULL, 1015, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1031, NULL, NULL, NULL, NULL, NULL, NULL, 1015, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1032, NULL, NULL, NULL, NULL, NULL, NULL, 1016, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1033, NULL, NULL, NULL, NULL, NULL, NULL, 1016, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1034, NULL, NULL, NULL, NULL, NULL, NULL, 1017, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1035, NULL, NULL, NULL, NULL, NULL, NULL, 1017, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1036, NULL, NULL, NULL, NULL, NULL, NULL, 1018, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1037, NULL, NULL, NULL, NULL, NULL, NULL, 1018, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1038, NULL, NULL, NULL, NULL, NULL, NULL, 1019, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1039, NULL, NULL, NULL, NULL, NULL, NULL, 1019, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1040, NULL, NULL, NULL, NULL, NULL, NULL, 1020, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1041, NULL, NULL, NULL, NULL, NULL, NULL, 1020, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1042, NULL, NULL, NULL, NULL, NULL, NULL, 1021, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1043, NULL, NULL, NULL, NULL, NULL, NULL, 1021, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1044, NULL, NULL, NULL, NULL, NULL, NULL, 1022, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1045, NULL, NULL, NULL, NULL, NULL, NULL, 1022, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1046, NULL, NULL, NULL, NULL, NULL, NULL, 1023, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1047, NULL, NULL, NULL, NULL, NULL, NULL, 1023, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1048, NULL, NULL, NULL, NULL, NULL, NULL, 1024, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1049, NULL, NULL, NULL, NULL, NULL, NULL, 1024, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1050, NULL, NULL, NULL, NULL, NULL, NULL, 1025, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1051, NULL, NULL, NULL, NULL, NULL, NULL, 1025, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1052, NULL, NULL, NULL, NULL, NULL, NULL, 1026, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1053, NULL, NULL, NULL, NULL, NULL, NULL, 1026, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1054, NULL, NULL, NULL, NULL, NULL, NULL, 1027, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1055, NULL, NULL, NULL, NULL, NULL, NULL, 1027, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1056, NULL, NULL, NULL, NULL, NULL, NULL, 1028, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1057, NULL, NULL, NULL, NULL, NULL, NULL, 1028, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1058, NULL, NULL, NULL, NULL, NULL, NULL, 1029, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1059, NULL, NULL, NULL, NULL, NULL, NULL, 1029, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1060, NULL, NULL, NULL, NULL, NULL, NULL, 1030, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1061, NULL, NULL, NULL, NULL, NULL, NULL, 1030, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1062, NULL, NULL, NULL, NULL, NULL, NULL, 1031, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1063, NULL, NULL, NULL, NULL, NULL, NULL, 1031, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1064, NULL, NULL, NULL, NULL, NULL, NULL, 1032, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1065, NULL, NULL, NULL, NULL, NULL, NULL, 1032, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1066, NULL, NULL, NULL, NULL, NULL, NULL, 1033, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1067, NULL, NULL, NULL, NULL, NULL, NULL, 1033, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1068, NULL, NULL, NULL, NULL, NULL, NULL, 1034, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1069, NULL, NULL, NULL, NULL, NULL, NULL, 1034, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1070, NULL, NULL, NULL, NULL, NULL, NULL, 1035, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1071, NULL, NULL, NULL, NULL, NULL, NULL, 1035, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1072, NULL, NULL, NULL, NULL, NULL, NULL, 1036, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1073, NULL, NULL, NULL, NULL, NULL, NULL, 1036, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1074, NULL, NULL, NULL, NULL, NULL, NULL, 1037, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1075, NULL, NULL, NULL, NULL, NULL, NULL, 1037, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1076, NULL, NULL, NULL, NULL, NULL, NULL, 1038, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1077, NULL, NULL, NULL, NULL, NULL, NULL, 1038, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1078, NULL, NULL, NULL, NULL, NULL, NULL, 1039, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1079, NULL, NULL, NULL, NULL, NULL, NULL, 1039, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1080, NULL, NULL, NULL, NULL, NULL, NULL, 1040, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1081, NULL, NULL, NULL, NULL, NULL, NULL, 1040, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1082, NULL, NULL, NULL, NULL, NULL, NULL, 1041, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1083, NULL, NULL, NULL, NULL, NULL, NULL, 1041, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1084, NULL, NULL, NULL, NULL, NULL, NULL, 1042, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1085, NULL, NULL, NULL, NULL, NULL, NULL, 1042, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1086, NULL, NULL, NULL, NULL, NULL, NULL, 1043, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1087, NULL, NULL, NULL, NULL, NULL, NULL, 1043, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1088, NULL, NULL, NULL, NULL, NULL, NULL, 1044, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1089, NULL, NULL, NULL, NULL, NULL, NULL, 1044, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1090, NULL, NULL, NULL, NULL, NULL, NULL, 1045, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1091, NULL, NULL, NULL, NULL, NULL, NULL, 1045, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1092, NULL, NULL, NULL, NULL, NULL, NULL, 1046, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1093, NULL, NULL, NULL, NULL, NULL, NULL, 1046, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1094, NULL, NULL, NULL, NULL, NULL, NULL, 1047, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1095, NULL, NULL, NULL, NULL, NULL, NULL, 1047, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1096, NULL, NULL, NULL, NULL, NULL, NULL, 1048, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1097, NULL, NULL, NULL, NULL, NULL, NULL, 1048, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1098, NULL, NULL, NULL, NULL, NULL, NULL, 1049, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1099, NULL, NULL, NULL, NULL, NULL, NULL, 1049, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1100, NULL, NULL, NULL, NULL, NULL, NULL, 1050, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1101, NULL, NULL, NULL, NULL, NULL, NULL, 1050, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1102, NULL, NULL, NULL, NULL, NULL, NULL, 1051, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1103, NULL, NULL, NULL, NULL, NULL, NULL, 1051, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1104, NULL, NULL, NULL, NULL, NULL, NULL, 1052, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1105, NULL, NULL, NULL, NULL, NULL, NULL, 1052, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1106, NULL, NULL, NULL, NULL, NULL, NULL, 1053, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1107, NULL, NULL, NULL, NULL, NULL, NULL, 1053, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1108, NULL, NULL, NULL, NULL, NULL, NULL, 1054, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1109, NULL, NULL, NULL, NULL, NULL, NULL, 1054, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1110, NULL, NULL, NULL, NULL, NULL, NULL, 1055, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1111, NULL, NULL, NULL, NULL, NULL, NULL, 1055, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1112, NULL, NULL, NULL, NULL, NULL, NULL, 1056, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1113, NULL, NULL, NULL, NULL, NULL, NULL, 1056, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1114, NULL, NULL, NULL, NULL, NULL, NULL, 1057, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1115, NULL, NULL, NULL, NULL, NULL, NULL, 1057, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1116, NULL, NULL, NULL, NULL, NULL, NULL, 1058, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1117, NULL, NULL, NULL, NULL, NULL, NULL, 1058, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1118, NULL, NULL, NULL, NULL, NULL, NULL, 1059, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1119, NULL, NULL, NULL, NULL, NULL, NULL, 1059, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1120, NULL, NULL, NULL, NULL, NULL, NULL, 1060, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1121, NULL, NULL, NULL, NULL, NULL, NULL, 1060, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1122, NULL, NULL, NULL, NULL, NULL, NULL, 1061, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1123, NULL, NULL, NULL, NULL, NULL, NULL, 1061, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1124, NULL, NULL, NULL, NULL, NULL, NULL, 1062, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1125, NULL, NULL, NULL, NULL, NULL, NULL, 1062, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1126, NULL, NULL, NULL, NULL, NULL, NULL, 1063, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1127, NULL, NULL, NULL, NULL, NULL, NULL, 1063, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1128, NULL, NULL, NULL, NULL, NULL, NULL, 1064, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1129, NULL, NULL, NULL, NULL, NULL, NULL, 1064, 104, 512);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1130, NULL, NULL, NULL, NULL, NULL, NULL, 1065, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1131, NULL, NULL, NULL, NULL, NULL, NULL, 1065, 104, 513);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1132, NULL, NULL, NULL, NULL, NULL, NULL, 1066, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1133, NULL, NULL, NULL, NULL, NULL, NULL, 1066, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1134, NULL, NULL, NULL, NULL, NULL, NULL, 1067, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1135, NULL, NULL, NULL, NULL, NULL, NULL, 1067, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1136, NULL, NULL, NULL, NULL, NULL, NULL, 1068, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1137, NULL, NULL, NULL, NULL, NULL, NULL, 1068, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1138, NULL, NULL, NULL, NULL, NULL, NULL, 1069, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1139, NULL, NULL, NULL, NULL, NULL, NULL, 1069, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1140, NULL, NULL, NULL, NULL, NULL, NULL, 1070, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1141, NULL, NULL, NULL, NULL, NULL, NULL, 1070, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1142, NULL, NULL, NULL, NULL, NULL, NULL, 1071, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1143, NULL, NULL, NULL, NULL, NULL, NULL, 1071, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1144, NULL, NULL, NULL, NULL, NULL, NULL, 1072, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1145, NULL, NULL, NULL, NULL, NULL, NULL, 1072, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1146, NULL, NULL, NULL, NULL, NULL, NULL, 1073, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1147, NULL, NULL, NULL, NULL, NULL, NULL, 1073, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1148, NULL, NULL, NULL, NULL, NULL, NULL, 1074, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1149, NULL, NULL, NULL, NULL, NULL, NULL, 1074, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1150, NULL, NULL, NULL, NULL, NULL, NULL, 1075, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1151, NULL, NULL, NULL, NULL, NULL, NULL, 1075, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1152, NULL, NULL, NULL, NULL, NULL, NULL, 1076, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1153, NULL, NULL, NULL, NULL, NULL, NULL, 1076, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1154, NULL, NULL, NULL, NULL, NULL, NULL, 1077, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1155, NULL, NULL, NULL, NULL, NULL, NULL, 1077, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1156, NULL, NULL, NULL, NULL, NULL, NULL, 1078, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1157, NULL, NULL, NULL, NULL, NULL, NULL, 1078, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1158, NULL, NULL, NULL, NULL, NULL, NULL, 1079, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1159, NULL, NULL, NULL, NULL, NULL, NULL, 1079, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1160, NULL, NULL, NULL, NULL, NULL, NULL, 1080, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1161, NULL, NULL, NULL, NULL, NULL, NULL, 1080, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1162, NULL, NULL, NULL, NULL, NULL, NULL, 1081, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1163, NULL, NULL, NULL, NULL, NULL, NULL, 1081, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1164, NULL, NULL, NULL, NULL, NULL, NULL, 1082, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1165, NULL, NULL, NULL, NULL, NULL, NULL, 1082, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1166, NULL, NULL, NULL, NULL, NULL, NULL, 1083, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1167, NULL, NULL, NULL, NULL, NULL, NULL, 1083, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1168, NULL, NULL, NULL, NULL, NULL, NULL, 1084, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1169, NULL, NULL, NULL, NULL, NULL, NULL, 1084, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1170, NULL, NULL, NULL, NULL, NULL, NULL, 1085, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1171, NULL, NULL, NULL, NULL, NULL, NULL, 1085, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1172, NULL, NULL, NULL, NULL, NULL, NULL, 1086, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1173, NULL, NULL, NULL, NULL, NULL, NULL, 1086, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1174, NULL, NULL, NULL, NULL, NULL, NULL, 1087, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1175, NULL, NULL, NULL, NULL, NULL, NULL, 1087, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1176, NULL, NULL, NULL, NULL, NULL, NULL, 1088, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1177, NULL, NULL, NULL, NULL, NULL, NULL, 1088, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1178, NULL, NULL, NULL, NULL, NULL, NULL, 1089, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1179, NULL, NULL, NULL, NULL, NULL, NULL, 1089, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1180, NULL, NULL, NULL, NULL, NULL, NULL, 1090, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1181, NULL, NULL, NULL, NULL, NULL, NULL, 1090, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1182, NULL, NULL, NULL, NULL, NULL, NULL, 1091, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1183, NULL, NULL, NULL, NULL, NULL, NULL, 1091, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1184, NULL, NULL, NULL, NULL, NULL, NULL, 1092, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1185, NULL, NULL, NULL, NULL, NULL, NULL, 1092, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1186, NULL, NULL, NULL, NULL, NULL, NULL, 1093, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1187, NULL, NULL, NULL, NULL, NULL, NULL, 1093, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1188, NULL, NULL, NULL, NULL, NULL, NULL, 1094, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1189, NULL, NULL, NULL, NULL, NULL, NULL, 1094, 102, 506);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1190, NULL, NULL, NULL, NULL, NULL, NULL, 1095, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1191, NULL, NULL, NULL, NULL, NULL, NULL, 1095, 102, 507);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1192, NULL, NULL, NULL, NULL, NULL, NULL, 1096, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1193, NULL, NULL, NULL, NULL, NULL, NULL, 1096, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1194, NULL, NULL, NULL, NULL, NULL, NULL, 1097, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1195, NULL, NULL, NULL, NULL, NULL, NULL, 1097, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1196, NULL, NULL, NULL, NULL, NULL, NULL, 1098, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1197, NULL, NULL, NULL, NULL, NULL, NULL, 1098, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1198, NULL, NULL, NULL, NULL, NULL, NULL, 1099, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1199, NULL, NULL, NULL, NULL, NULL, NULL, 1099, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1200, NULL, NULL, NULL, NULL, NULL, NULL, 1100, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1201, NULL, NULL, NULL, NULL, NULL, NULL, 1100, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1202, NULL, NULL, NULL, NULL, NULL, NULL, 1101, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1203, NULL, NULL, NULL, NULL, NULL, NULL, 1101, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1204, NULL, NULL, NULL, NULL, NULL, NULL, 1102, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1205, NULL, NULL, NULL, NULL, NULL, NULL, 1102, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1206, NULL, NULL, NULL, NULL, NULL, NULL, 1103, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1207, NULL, NULL, NULL, NULL, NULL, NULL, 1103, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1208, NULL, NULL, NULL, NULL, NULL, NULL, 1104, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1209, NULL, NULL, NULL, NULL, NULL, NULL, 1104, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1210, NULL, NULL, NULL, NULL, NULL, NULL, 1105, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1211, NULL, NULL, NULL, NULL, NULL, NULL, 1105, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1212, NULL, NULL, NULL, NULL, NULL, NULL, 1106, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1213, NULL, NULL, NULL, NULL, NULL, NULL, 1106, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1214, NULL, NULL, NULL, NULL, NULL, NULL, 1107, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1215, NULL, NULL, NULL, NULL, NULL, NULL, 1107, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1216, NULL, NULL, NULL, NULL, NULL, NULL, 1108, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1217, NULL, NULL, NULL, NULL, NULL, NULL, 1108, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1218, NULL, NULL, NULL, NULL, NULL, NULL, 1109, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1219, NULL, NULL, NULL, NULL, NULL, NULL, 1109, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1220, NULL, NULL, NULL, NULL, NULL, NULL, 1110, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1221, NULL, NULL, NULL, NULL, NULL, NULL, 1110, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1222, NULL, NULL, NULL, NULL, NULL, NULL, 1111, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1223, NULL, NULL, NULL, NULL, NULL, NULL, 1111, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1224, NULL, NULL, NULL, NULL, NULL, NULL, 1112, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1225, NULL, NULL, NULL, NULL, NULL, NULL, 1112, 103, 509);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1226, NULL, NULL, NULL, NULL, NULL, NULL, 1113, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1227, NULL, NULL, NULL, NULL, NULL, NULL, 1113, 103, 510);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1228, NULL, NULL, NULL, NULL, NULL, NULL, 1114, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1229, NULL, NULL, NULL, NULL, NULL, NULL, 1114, 101, 504);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1230, NULL, NULL, NULL, NULL, NULL, NULL, 1115, 100, 501);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1231, NULL, NULL, NULL, NULL, NULL, NULL, 1115, 101, 505);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1232, NULL, NULL, NULL, NULL, NULL, NULL, 1116, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1233, NULL, NULL, NULL, NULL, NULL, NULL, 1116, 101, 504);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1234, NULL, NULL, NULL, NULL, NULL, NULL, 1117, 100, 502);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1235, NULL, NULL, NULL, NULL, NULL, NULL, 1117, 101, 505);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1236, NULL, NULL, NULL, NULL, NULL, NULL, 1118, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1237, NULL, NULL, NULL, NULL, NULL, NULL, 1118, 101, 504);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1238, NULL, NULL, NULL, NULL, NULL, NULL, 1119, 100, 503);
INSERT INTO mall_sku_attribute_combinations (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, attribute_id, attribute_value_id) VALUES (1239, NULL, NULL, NULL, NULL, NULL, NULL, 1119, 101, 505);


--
-- Data for Name: mall_sku_prices; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1000, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1000, '5900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1001, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1001, '8900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1002, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1002, '11900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1003, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1003, '14900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1004, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1004, '17900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1005, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1005, '20900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1006, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1006, '129900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1007, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1007, '132900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1008, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1008, '135900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1009, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1009, '138900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1010, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1010, '141900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1011, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1011, '144900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1012, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1012, '39900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1013, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1013, '42900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1014, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1014, '45900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1015, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1015, '48900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1016, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1016, '51900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1017, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1017, '54900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1018, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1018, '45900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1019, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1019, '48900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1020, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1020, '51900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1021, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1021, '54900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1022, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1022, '57900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1023, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1023, '60900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1024, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1024, '12900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1025, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1025, '15900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1026, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1026, '18900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1027, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1027, '21900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1028, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1028, '24900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1029, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1029, '27900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1030, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1030, '12900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1031, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1031, '15900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1032, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1032, '18900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1033, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1033, '21900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1034, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1034, '24900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1035, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1035, '27900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1036, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1036, '29900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1037, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1037, '32900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1038, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1038, '35900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1039, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1039, '38900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1040, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1040, '41900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1041, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1041, '44900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1042, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1042, '21900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1043, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1043, '24900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1044, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1044, '27900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1045, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1045, '30900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1046, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1046, '33900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1047, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1047, '36900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1048, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1048, '9900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1049, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1049, '12900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1050, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1050, '15900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1051, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1051, '18900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1052, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1052, '21900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1053, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1053, '24900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1054, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1054, '6900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1055, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1055, '9900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1056, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1056, '12900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1057, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1057, '15900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1058, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1058, '18900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1059, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1059, '21900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1060, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1060, '15900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1061, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1061, '18900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1062, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1062, '21900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1063, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1063, '24900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1064, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1064, '27900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1065, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1065, '30900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1066, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1066, '249900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1067, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1067, '272900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1068, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1068, '255900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1069, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1069, '278900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1070, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1070, '261900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1071, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1071, '284900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1072, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1072, '369900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1073, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1073, '392900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1074, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1074, '375900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1075, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1075, '398900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1076, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1076, '381900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1077, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1077, '404900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1078, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1078, '899900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1079, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1079, '922900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1080, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1080, '905900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1081, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1081, '928900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1082, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1082, '911900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1083, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1083, '934900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1084, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1084, '129900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1085, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1085, '152900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1086, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1086, '135900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1087, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1087, '158900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1088, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1088, '141900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1089, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1089, '164900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1090, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1090, '549900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1091, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1091, '572900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1092, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1092, '555900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1093, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1093, '578900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1094, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1094, '561900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1095, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1095, '584900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1096, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1096, '159900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1097, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1097, '162900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1098, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1098, '165900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1099, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1099, '168900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1100, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1100, '171900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1101, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1101, '174900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1102, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1102, '24900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1103, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1103, '27900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1104, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1104, '30900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1105, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1105, '33900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1106, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1106, '36900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1107, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1107, '39900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1108, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1108, '89900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1109, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1109, '92900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1110, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1110, '95900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1111, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1111, '98900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1112, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1112, '101900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1113, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1113, '104900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1114, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1114, '229900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1115, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1115, '242900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1116, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1116, '235900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1117, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1117, '248900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1118, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1118, '241900');
INSERT INTO mall_sku_prices (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, currency, sku_id, amount) VALUES (1119, NULL, NULL, NULL, NULL, NULL, NULL, 'CNY', 1119, '254900');


--
-- Data for Name: mall_skus; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1000, '2026-10-02 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 100, 'LUMEN-1000', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1001, '2026-10-02 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 100, 'LUMEN-1001', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1002, '2026-10-02 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 100, 'LUMEN-1002', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1003, '2026-10-02 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 100, 'LUMEN-1003', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1004, '2026-10-02 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 100, 'LUMEN-1004', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1005, '2026-10-02 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 100, 'LUMEN-1005', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1006, '2026-10-03 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 101, 'VAULT-1006', 48);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1007, '2026-10-03 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 101, 'VAULT-1007', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1008, '2026-10-03 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 101, 'VAULT-1008', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1009, '2026-10-03 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 101, 'VAULT-1009', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1010, '2026-10-03 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 101, 'VAULT-1010', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1011, '2026-10-03 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 101, 'VAULT-1011', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1012, '2026-10-04 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 102, 'TRIO-1012', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1013, '2026-10-04 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 102, 'TRIO-1013', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1014, '2026-10-04 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 102, 'TRIO-1014', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1015, '2026-10-04 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 102, 'TRIO-1015', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1016, '2026-10-04 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 102, 'TRIO-1016', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1017, '2026-10-04 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 102, 'TRIO-1017', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1018, '2026-10-05 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 103, 'SENTINEL-1018', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1019, '2026-10-05 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 103, 'SENTINEL-1019', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1020, '2026-10-05 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 103, 'SENTINEL-1020', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1021, '2026-10-05 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 103, 'SENTINEL-1021', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1022, '2026-10-05 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 103, 'SENTINEL-1022', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1023, '2026-10-05 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 103, 'SENTINEL-1023', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1024, '2026-10-06 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 104, 'BREEZE-1024', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1025, '2026-10-06 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 104, 'BREEZE-1025', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1026, '2026-10-06 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 104, 'BREEZE-1026', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1027, '2026-10-06 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 104, 'BREEZE-1027', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1028, '2026-10-06 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 104, 'BREEZE-1028', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1029, '2026-10-06 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 104, 'BREEZE-1029', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1030, '2026-10-07 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 105, 'GAN-1030', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1031, '2026-10-07 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 105, 'GAN-1031', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1032, '2026-10-07 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 105, 'GAN-1032', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1033, '2026-10-07 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 105, 'GAN-1033', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1034, '2026-10-07 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 105, 'GAN-1034', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1035, '2026-10-07 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 105, 'GAN-1035', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1036, '2026-10-08 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 106, 'ECHO-1036', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1037, '2026-10-08 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 106, 'ECHO-1037', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1038, '2026-10-08 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 106, 'ECHO-1038', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1039, '2026-10-08 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 106, 'ECHO-1039', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1040, '2026-10-08 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 106, 'ECHO-1040', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1041, '2026-10-08 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 106, 'ECHO-1041', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1042, '2026-10-09 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 107, 'USB-C-1042', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1043, '2026-10-09 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 107, 'USB-C-1043', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1044, '2026-10-09 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 107, 'USB-C-1044', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1045, '2026-10-09 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 107, 'USB-C-1045', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1046, '2026-10-09 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 107, 'USB-C-1046', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1047, '2026-10-09 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 107, 'USB-C-1047', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1048, '2026-10-10 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 108, 'GLIDE-1048', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1049, '2026-10-10 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 108, 'GLIDE-1049', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1050, '2026-10-10 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 108, 'GLIDE-1050', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1051, '2026-10-10 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 108, 'GLIDE-1051', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1052, '2026-10-10 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 108, 'GLIDE-1052', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1053, '2026-10-10 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 108, 'GLIDE-1053', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1054, '2026-10-11 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 109, 'AERO-1054', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1055, '2026-10-11 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 109, 'AERO-1055', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1056, '2026-10-11 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 109, 'AERO-1056', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1057, '2026-10-11 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 109, 'AERO-1057', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1058, '2026-10-11 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 109, 'AERO-1058', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1059, '2026-10-11 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 109, 'AERO-1059', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1060, '2026-10-12 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 110, 'PEBBLE-1060', 1);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1061, '2026-10-12 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 110, 'PEBBLE-1061', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1062, '2026-10-12 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 110, 'PEBBLE-1062', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1063, '2026-10-12 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 110, 'PEBBLE-1063', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1064, '2026-10-12 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 110, 'PEBBLE-1064', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1065, '2026-10-12 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 110, 'PEBBLE-1065', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1066, '2026-10-13 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 111, 'NOVA-1066', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1067, '2026-10-13 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 111, 'NOVA-1067', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1068, '2026-10-13 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 111, 'NOVA-1068', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1069, '2026-10-13 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 111, 'NOVA-1069', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1070, '2026-10-13 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 111, 'NOVA-1070', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1071, '2026-10-13 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 111, 'NOVA-1071', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1072, '2026-10-14 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 112, 'NOVA-1072', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1073, '2026-10-14 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 112, 'NOVA-1073', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1074, '2026-10-14 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 112, 'NOVA-1074', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1075, '2026-10-14 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 112, 'NOVA-1075', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1076, '2026-10-14 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 112, 'NOVA-1076', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1077, '2026-10-14 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 112, 'NOVA-1077', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1078, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 113, 'FOLD-1078', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1079, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 113, 'FOLD-1079', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1080, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 113, 'FOLD-1080', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1081, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 113, 'FOLD-1081', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1082, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 113, 'FOLD-1082', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1083, '2026-10-15 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 113, 'FOLD-1083', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1084, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 114, 'NOVA-1084', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1085, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 114, 'NOVA-1085', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1086, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 114, 'NOVA-1086', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1087, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 114, 'NOVA-1087', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1088, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 114, 'NOVA-1088', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1089, '2026-10-16 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 114, 'NOVA-1089', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1090, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 115, 'NOVA-1090', 5);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1091, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 115, 'NOVA-1091', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1092, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 115, 'NOVA-1092', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1093, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 115, 'NOVA-1093', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1094, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 115, 'NOVA-1094', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1095, '2026-10-17 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 115, 'NOVA-1095', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1096, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 116, 'WATCH-1096', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1097, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 116, 'WATCH-1097', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1098, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 116, 'WATCH-1098', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1099, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 116, 'WATCH-1099', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1100, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 116, 'WATCH-1100', 80);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1101, '2026-10-18 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 116, 'WATCH-1101', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1102, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 117, 'BAND-1102', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1103, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 117, 'BAND-1103', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1104, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 117, 'BAND-1104', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1105, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 117, 'BAND-1105', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1106, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 117, 'BAND-1106', 46);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1107, '2026-10-19 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 117, 'BAND-1107', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1108, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 118, 'KIDSAFE-1108', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1109, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 118, 'KIDSAFE-1109', 28);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1110, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 118, 'KIDSAFE-1110', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1111, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 118, 'KIDSAFE-1111', 3);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1112, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 118, 'KIDSAFE-1112', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1113, '2026-10-20 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 118, 'KIDSAFE-1113', 12);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1114, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 119, 'VISTA-1114', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1115, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 119, 'VISTA-1115', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1116, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 119, 'VISTA-1116', 75);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1117, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 119, 'VISTA-1117', 130);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1118, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 119, 'VISTA-1118', 0);
INSERT INTO mall_skus (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, product_id, sku_code, stock_qty) VALUES (1119, '2026-10-21 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 119, 'VISTA-1119', 0);


--
-- Data for Name: mall_stock_alerts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_stock_alerts (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, stock_qty_at_trigger, threshold, status) VALUES (3, '2026-10-26 07:33:18.246778', '2026-10-27 07:33:18.246778', NULL, 1, NULL, NULL, 1000, 2, 8, 'RESOLVED');
INSERT INTO mall_stock_alerts (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, stock_qty_at_trigger, threshold, status) VALUES (4, '2026-10-26 07:33:18.246778', '2026-10-27 07:33:18.246778', NULL, 1, NULL, NULL, 1001, 6, 8, 'RESOLVED');


--
-- Data for Name: mall_stock_op_logs; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark) VALUES (43, '2026-10-30 23:15:01.221742', NULL, NULL, 0, NULL, NULL, 1100, 116, 2, 75, 77, 'CANCEL_RESTORE', 100, NULL);
INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark) VALUES (44, '2026-10-30 23:15:01.226856', NULL, NULL, 0, NULL, NULL, 1090, 115, 2, 3, 5, 'CANCEL_RESTORE', 100, NULL);
INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark) VALUES (45, '2026-10-30 23:15:01.229065', NULL, NULL, 0, NULL, NULL, 1060, 110, 1, 0, 1, 'CANCEL_RESTORE', 100, NULL);
INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark) VALUES (46, '2026-10-30 23:15:01.283085', NULL, NULL, 0, NULL, NULL, 1100, 116, 1, 77, 78, 'CANCEL_RESTORE', 105, NULL);
INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark) VALUES (47, '2026-10-30 23:15:01.315203', NULL, NULL, 0, NULL, NULL, 1006, 101, 2, 46, 48, 'CANCEL_RESTORE', 111, NULL);
INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark) VALUES (48, '2026-10-30 23:15:01.317407', NULL, NULL, 0, NULL, NULL, 1100, 116, 2, 78, 80, 'CANCEL_RESTORE', 111, NULL);


--
-- Data for Name: mall_tax_rates; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_tax_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, tax_rate, status) VALUES (3, '2026-10-23 07:33:18.240397', '2026-10-23 07:33:18.240397', NULL, 1, NULL, NULL, 0, 'CNY', '北京', 13, 'ACTIVE');
INSERT INTO mall_tax_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, tax_rate, status) VALUES (4, '2026-10-23 07:33:18.240397', '2026-10-23 07:33:18.240397', NULL, 1, NULL, NULL, 0, 'CNY', '上海', 13, 'ACTIVE');
INSERT INTO mall_tax_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, tax_rate, status) VALUES (5, '2026-10-23 07:33:18.240397', '2026-10-23 07:33:18.240397', NULL, 1, NULL, NULL, 0, 'CNY', '广东', 13, 'ACTIVE');
INSERT INTO mall_tax_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, tax_rate, status) VALUES (6, '2026-10-23 07:33:18.240397', '2026-10-23 07:33:18.240397', NULL, 1, NULL, NULL, 0, 'CNY', '浙江', 9, 'ACTIVE');
INSERT INTO mall_tax_rates (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, currency, region, tax_rate, status) VALUES (7, '2026-10-23 07:33:18.240397', '2026-10-23 07:33:18.240397', NULL, 1, NULL, NULL, 0, 'CNY', '四川', 9, 'ACTIVE');


--
-- Data for Name: mall_user_coupons; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (1, '2026-10-29 11:05:47.026507', NULL, NULL, NULL, NULL, NULL, 0, 1, 2, 'UNUSED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (2, '2026-10-30 16:37:35.529021', NULL, NULL, NULL, NULL, NULL, 0, 1, 2, 'UNUSED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (3, '2026-10-31 06:44:27.703489', NULL, NULL, 1, NULL, NULL, 0, 3, 101, 'USED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (100, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 100, 'UNUSED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (101, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 101, 'UNUSED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (102, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 102, 'USED', '2026-10-30 23:03:27.011658', 101, 3000);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (103, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 103, 'UNUSED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (104, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 105, 'UNUSED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (105, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 107, 'EXPIRED', NULL, NULL, 0);
INSERT INTO mall_user_coupons (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, coupon_template_id, status, redeemed_at, redeemed_order_id, applied_discount_amount) VALUES (106, '2026-10-22 23:03:27.011658', NULL, NULL, 1, NULL, NULL, 0, 1, 104, 'UNUSED', NULL, NULL, 0);


--
-- Data for Name: mall_wishlist; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (101, '2026-10-29 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 103);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (102, '2026-10-28 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 106);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (103, '2026-10-27 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 109);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (104, '2026-10-26 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 112);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (105, '2026-10-25 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 115);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (106, '2026-10-24 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 118);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (107, '2026-10-23 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 101);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (108, '2026-10-22 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 104);
INSERT INTO mall_wishlist (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, tenant_id, user_id, product_id) VALUES (109, '2026-10-21 23:03:27.011658', NULL, NULL, NULL, NULL, NULL, 0, 1, 107);


--
-- Data for Name: sys_languages; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (1, '2026-10-27 12:23:24.730777', NULL, NULL, NULL, NULL, NULL, 0, true, 'zh-CN', '中文（简体）', '简体中文', true);
INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (2, '2026-10-27 12:23:24.739468', NULL, NULL, NULL, NULL, NULL, 100, true, 'zh-TW', '中文（繁体）', '繁體中文', false);
INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (3, '2026-10-27 12:23:24.749213', NULL, NULL, NULL, NULL, NULL, 1, true, 'en-US', '英语', 'English', false);
INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (4, '2026-10-27 12:23:24.758692', NULL, NULL, NULL, NULL, NULL, 100, true, 'ja-JP', '日语', '日本語', false);
INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (5, '2026-10-27 12:23:24.764927', NULL, NULL, NULL, NULL, NULL, 100, true, 'ko-KR', '韩语', '한국어', false);
INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (6, '2026-10-27 12:23:24.770697', NULL, NULL, NULL, NULL, NULL, 100, true, 'es-ES', '西班牙语', 'Español', false);
INSERT INTO sys_languages (id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, sort_order, is_enabled, language_code, language_name, native_name, is_default) VALUES (7, '2026-10-27 12:23:24.777327', NULL, NULL, NULL, NULL, NULL, 100, true, 'fr-FR', '法语', 'Français', false);


--
-- Data for Name: sys_user_credentials; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO sys_user_credentials (id, created_at, updated_at, deleted_at, tenant_id, user_id, identity_type, identifier, credential_type, credential, is_primary, status, extra_info, provider, provider_account_id, activate_token_hash, activate_token_expires_at, activate_token_used_at, reset_token_hash, reset_token_expires_at, reset_token_used_at) VALUES (1, '2026-10-27 12:23:24.852763', NULL, NULL, 0, 1, 'USERNAME', 'admin', 'PASSWORD_HASH', '$2a$10$nEu9uNeIkDpF//oeexvfBOWv9dDIjQV/6ImiIK/JaeC61NAiQc0VG', true, 'ENABLED', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);


--
-- Data for Name: sys_users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO sys_users (id, created_by, updated_by, deleted_by, created_at, updated_at, deleted_at, remark, tenant_id, username, nickname, realname, email, mobile, telephone, avatar, address, region, description, gender, last_login_at, last_login_ip, locked_until, status) VALUES (1, NULL, NULL, NULL, '2026-10-27 12:23:24.788677', NULL, NULL, NULL, 0, 'admin', '鹳狸猿', '喵个咪', 'admin@gmail.com', '', '', NULL, '', '中国', NULL, 'SECRET', NULL, NULL, NULL, 'NORMAL');


--
-- Name: internal_message_recipients_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: internal_messages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_brand_translations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_brands_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_cart_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_carts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_category_translations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_comment_likes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_comments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_coupon_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_groupon_activities_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_groupon_groups_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_invoices_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_member_tiers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_order_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_payment_channel_configs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_payment_refunds_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_payment_transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_points_ledgers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_product_attribute_translations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_product_attribute_value_translations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_product_attribute_values_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_product_attributes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_product_translations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_products_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_promotions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_seckill_activities_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_shipments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_shipping_rates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_sku_attribute_combinations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_sku_prices_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_skus_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_stock_alerts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_stock_op_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_tax_rates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_user_coupons_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: mall_wishlist_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: sys_languages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: sys_user_credentials_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- Name: sys_users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--



--
-- PostgreSQL database dump complete
--

-- ── 三、test-data.sql 增补段（原始 ID，幂等守卫与快照自动去重）───────

-- ---------------------------------------------------------------
-- 1. 商品（ACTIVE）+ 翻译
-- ---------------------------------------------------------------
-- 商品 3：无线蓝牙耳机 Pro（若不存在则建）
INSERT INTO mall_products (id, status, sort_order, image_url, created_by, created_at, updated_at)
SELECT 3, 'PRODUCT_STATUS_ACTIVE', 1, '/demo-images/cat-wearable.webp', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_products WHERE id = 3);

INSERT INTO mall_product_translations (product_id, language_code, name, slug, short_description, long_description, created_by, created_at, updated_at)
SELECT 3, 'zh-CN', '无线蓝牙耳机 Pro', 'wireless-earphone-pro', '主动降噪无线蓝牙耳机', '40mm 大动圈单元，ANC 主动降噪，蓝牙 5.3，单次续航 8 小时，配合充电盒总续航 32 小时。', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_product_translations WHERE product_id = 3 AND language_code = 'zh-CN');

-- 商品 4：智能运动手表 X2
INSERT INTO mall_products (id, status, sort_order, image_url, created_by, created_at, updated_at)
SELECT 4, 'PRODUCT_STATUS_ACTIVE', 2, '/demo-images/cat-watch.webp', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_products WHERE id = 4);

INSERT INTO mall_product_translations (product_id, language_code, name, slug, short_description, long_description, created_by, created_at, updated_at)
SELECT 4, 'zh-CN', '智能运动手表 X2', 'smart-watch-x2', 'AMOLED 高刷屏，双频 GPS，14 天长续航', '1.43 英寸 AMOLED 常亮屏，双频 GNSS 精准轨迹，心率/血氧/睡眠监测，5ATM 防水，典型使用 14 天续航。', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_product_translations WHERE product_id = 4 AND language_code = 'zh-CN');


-- ---------------------------------------------------------------
-- 2. SKU + 多币种价格
-- ---------------------------------------------------------------
-- SKU 1：耳机（黑色）
INSERT INTO mall_skus (id, product_id, sku_code, stock_qty, created_by, created_at, updated_at)
SELECT 1, 3, 'EARPHONE-PRO-BLK', 50, 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_skus WHERE id = 1);

-- SKU 2：手表（曜石黑）
INSERT INTO mall_skus (id, product_id, sku_code, stock_qty, created_by, created_at, updated_at)
SELECT 2, 4, 'WATCH-X2-BLK', 30, 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_skus WHERE id = 2);


INSERT INTO mall_sku_prices (id, sku_id, currency, amount, created_by, created_at, updated_at)
SELECT 1, 1, 'CNY', '12800', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_sku_prices WHERE sku_id = 1 AND currency = 'CNY');

INSERT INTO mall_sku_prices (id, sku_id, currency, amount, created_by, created_at, updated_at)
SELECT 2, 2, 'CNY', '39900', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_sku_prices WHERE sku_id = 2 AND currency = 'CNY');


-- ---------------------------------------------------------------
-- 3. 促销：商品专价 + 店铺满减
-- ---------------------------------------------------------------
-- 专价：耳机限时 ¥99（原 ¥128）
INSERT INTO mall_promotions (id, name, promo_type, product_id, special_price, threshold_amount, reduction_amount, priority, valid_from, valid_until, status, currency, created_by, created_at, updated_at)
SELECT 1, '耳机限时专价', 'SPECIAL_PRICE', 3, 9900, 0, 0, 10, '2026-10-01', '2027-01-01', 'ACTIVE', 'CNY', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_promotions WHERE id = 1);

-- 满减：店铺满 ¥300 减 ¥40
INSERT INTO mall_promotions (id, name, promo_type, product_id, special_price, threshold_amount, reduction_amount, priority, valid_from, valid_until, status, currency, created_by, created_at, updated_at)
SELECT 2, '店铺满300减40', 'FULL_REDUCTION', 0, 0, 30000, 4000, 5, '2026-10-01', '2027-01-01', 'ACTIVE', 'CNY', 1, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM mall_promotions WHERE id = 2);


-- ---------------------------------------------------------------
-- 4. 订单 + 订单项 + 支付流水
--    订单 1：PAID     耳机×1（专价 ¥99，原价 ¥128，优惠 ¥29）
--    订单 2：FULFILLED 手表×1（¥399）
--    订单 3：CLOSED   耳机×2（专价 ¥198）
-- ---------------------------------------------------------------
INSERT INTO mall_orders (id, tenant_id, currency, idempotency_key, user_id, status, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, recipient_name, recipient_phone, shipping_address, created_by, created_at, updated_at)
SELECT 1, 1, 'CNY', 'TEST-ORDER-001', 1, 'PAID', 9900, 12800, 2900, 0, 0, 'CN', '张三', '13800000001', '广东省深圳市南山区科技园路 1 号', 1, NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY
WHERE NOT EXISTS (SELECT 1 FROM mall_orders WHERE idempotency_key = 'TEST-ORDER-001');

INSERT INTO mall_orders (id, tenant_id, currency, idempotency_key, user_id, status, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, recipient_name, recipient_phone, shipping_address, created_by, created_at, updated_at)
SELECT 2, 1, 'CNY', 'TEST-ORDER-002', 1, 'FULFILLED', 39900, 39900, 0, 0, 0, 'CN', '张三', '13800000001', '广东省深圳市南山区科技园路 1 号', 1, NOW() - INTERVAL 1 DAY, now()
WHERE NOT EXISTS (SELECT 1 FROM mall_orders WHERE idempotency_key = 'TEST-ORDER-002');

INSERT INTO mall_orders (id, tenant_id, currency, idempotency_key, user_id, status, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, recipient_name, recipient_phone, shipping_address, created_by, created_at, updated_at)
SELECT 3, 1, 'CNY', 'TEST-ORDER-003', 1, 'CLOSED', 19800, 25600, 5800, 0, 0, 'CN', '张三', '13800000001', '广东省深圳市南山区科技园路 1 号', 1, NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 3 DAY
WHERE NOT EXISTS (SELECT 1 FROM mall_orders WHERE idempotency_key = 'TEST-ORDER-003');


INSERT INTO mall_order_items (tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_by, created_at, updated_at)
SELECT 1, o.id, 1, 1, '{"sku_id":1,"quantity":1,"currency":"CNY","unit_price":9900}', 1, 9900, 9900, 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-001'
  AND NOT EXISTS (SELECT 1 FROM mall_order_items i WHERE i.order_id = o.id AND i.sku_id = 1);

INSERT INTO mall_order_items (tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_by, created_at, updated_at)
SELECT 1, o.id, 1, 2, '{"sku_id":2,"quantity":1,"currency":"CNY","unit_price":39900}', 1, 39900, 39900, 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-002'
  AND NOT EXISTS (SELECT 1 FROM mall_order_items i WHERE i.order_id = o.id AND i.sku_id = 2);

INSERT INTO mall_order_items (tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_by, created_at, updated_at)
SELECT 1, o.id, 1, 1, '{"sku_id":1,"quantity":2,"currency":"CNY","unit_price":9900}', 2, 9900, 19800, 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-003'
  AND NOT EXISTS (SELECT 1 FROM mall_order_items i WHERE i.order_id = o.id AND i.sku_id = 1);

INSERT INTO mall_payment_transactions (tenant_id, currency, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, created_by, created_at, updated_at)
SELECT 1, 'CNY', 'TEST-PAY-001', 'BALANCE', 'BUSINESS_TYPE_CONSUME', o.id, 1, o.total_amount, 'SUCCEEDED', 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-001'
  AND NOT EXISTS (SELECT 1 FROM mall_payment_transactions WHERE idempotency_key = 'TEST-PAY-001');

INSERT INTO mall_payment_transactions (tenant_id, currency, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, created_by, created_at, updated_at)
SELECT 1, 'CNY', 'TEST-PAY-002', 'BALANCE', 'BUSINESS_TYPE_CONSUME', o.id, 1, o.total_amount, 'SUCCEEDED', 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-002'
  AND NOT EXISTS (SELECT 1 FROM mall_payment_transactions WHERE idempotency_key = 'TEST-PAY-002');

INSERT INTO mall_payment_transactions (tenant_id, currency, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, created_by, created_at, updated_at)
SELECT 1, 'CNY', 'TEST-PAY-003', 'BALANCE', 'BUSINESS_TYPE_CONSUME', o.id, 1, o.total_amount, 'SUCCEEDED', 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-003'
  AND NOT EXISTS (SELECT 1 FROM mall_payment_transactions WHERE idempotency_key = 'TEST-PAY-003');


-- ---------------------------------------------------------------
-- 5. 库存操作流水（与订单对应；手动调整流水由操作产生）
-- ---------------------------------------------------------------
INSERT INTO mall_stock_op_logs (sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark, created_by, created_at, updated_at)
SELECT 1, 3, -1, 51, 50, 'ORDER_DEDUCT', o.id, NULL, 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-001'
  AND NOT EXISTS (SELECT 1 FROM mall_stock_op_logs l WHERE l.reason = 'ORDER_DEDUCT' AND l.ref_id = o.id AND l.sku_id = 1);

INSERT INTO mall_stock_op_logs (sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark, created_by, created_at, updated_at)
SELECT 2, 4, -1, 31, 30, 'ORDER_DEDUCT', o.id, NULL, 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-002'
  AND NOT EXISTS (SELECT 1 FROM mall_stock_op_logs l WHERE l.reason = 'ORDER_DEDUCT' AND l.ref_id = o.id AND l.sku_id = 2);

INSERT INTO mall_stock_op_logs (sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark, created_by, created_at, updated_at)
SELECT 1, 3, -2, 52, 50, 'ORDER_DEDUCT', o.id, NULL, 1, o.created_at, o.created_at
FROM mall_orders o
WHERE o.idempotency_key = 'TEST-ORDER-003'
  AND NOT EXISTS (SELECT 1 FROM mall_stock_op_logs l WHERE l.reason = 'ORDER_DEDUCT' AND l.ref_id = o.id AND l.sku_id = 1);


-- ---------------------------------------------------------------
-- 6. 商品评论（含商家回复）
--    评论 1/2（耳机 + 商家回复）已由功能验证产生；此处补手表的评论
-- ---------------------------------------------------------------
INSERT INTO mall_comments (tenant_id, content_type, object_id, content, status, rating, parent_id, created_by, created_at, updated_at)
SELECT 0, 'CONTENT_TYPE_PRODUCT', 4, '续航给力，双频 GPS 定位很准，跑步骑行都够用。', 'STATUS_APPROVED', 5, NULL, 1, NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY
WHERE NOT EXISTS (SELECT 1 FROM mall_comments WHERE object_id = 4 AND content_type = 'CONTENT_TYPE_PRODUCT' AND parent_id IS NULL);

-- 商家回复
INSERT INTO mall_comments (tenant_id, content_type, object_id, content, status, parent_id, created_by, created_at, updated_at)
SELECT 0, 'CONTENT_TYPE_PRODUCT', 4, '感谢您的认可，X2 的双频 GPS 是本代最大升级点，祝运动愉快！', 'STATUS_APPROVED', (SELECT id FROM mall_comments WHERE object_id = 4 AND parent_id IS NULL LIMIT 1), 1, NOW() - INTERVAL 20 HOUR, NOW() - INTERVAL 20 HOUR
WHERE (SELECT id FROM mall_comments WHERE object_id = 4 AND parent_id IS NULL LIMIT 1) IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM mall_comments WHERE object_id = 4 AND parent_id IS NOT NULL);


-- ---------------------------------------------------------------
-- 7. 站内信（C端消息中心：未读 badge 数据源）
-- ---------------------------------------------------------------
INSERT INTO internal_messages (id, title, content, status, type, sender_id, created_at, updated_at)
SELECT 10, '订单已发货', '您的订单 #2 已发货，请注意查收物流信息。', 'PUBLISHED', 'NOTIFICATION', 0, NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY
WHERE NOT EXISTS (SELECT 1 FROM internal_messages WHERE id = 10);

INSERT INTO internal_messages (id, title, content, status, type, sender_id, created_at, updated_at)
SELECT 11, '店铺满减活动上线', '全店满 ¥300 减 ¥40，活动截至 2027-01-01，先到先得！', 'PUBLISHED', 'NOTIFICATION', 0, NOW() - INTERVAL 4 HOUR, NOW() - INTERVAL 4 HOUR
WHERE NOT EXISTS (SELECT 1 FROM internal_messages WHERE id = 11);


INSERT INTO internal_message_recipients (tenant_id, message_id, recipient_user_id, status, created_at, updated_at)
SELECT 1, m.id, 1, 'SENT', m.created_at, m.created_at
FROM internal_messages m
WHERE m.id = 10
  AND NOT EXISTS (SELECT 1 FROM internal_message_recipients r WHERE r.message_id = 10 AND r.recipient_user_id = 1);

INSERT INTO internal_message_recipients (tenant_id, message_id, recipient_user_id, status, created_at, updated_at)
SELECT 1, m.id, 1, 'SENT', m.created_at, m.created_at
FROM internal_messages m
WHERE m.id = 11
  AND NOT EXISTS (SELECT 1 FROM internal_message_recipients r WHERE r.message_id = 11 AND r.recipient_user_id = 1);

-- ── 三B、演示增补层（顾客用户/促销/订单/支付/退款/发货/评论/站内信/预警）
--      约定沿用内网基准：tenant_id=0、user_id=1；金额单位：分；
--      订单/库存流水金额与 SKU 现价自洽，库存流水期末值对齐 SKU 现存库存。
--      id 段避让：test-data 段经序列接续占用 支付21-23/订单项1026-1028/
--      评论162-163/库存流水49-51/收件人23-24，本段从其后起步。

-- 1. 顾客用户（凭证同 admin，密码 admin）
INSERT INTO sys_users (id, created_by, created_at, updated_at, tenant_id, username, nickname, realname, email, mobile, gender, status, last_login_at)
VALUES
    (2, 1, NOW() - INTERVAL 30 DAY, NOW() - INTERVAL 1 DAY,  0, 'suxiaotong', '苏晓桐', '苏晓桐', 'suxiaotong@demo.com', '13911112222', 'FEMALE', 'NORMAL', NOW() - INTERVAL 1 DAY),
    (3, 1, NOW() - INTERVAL 29 DAY, NOW() - INTERVAL 2 DAY,  0, 'chenyi',     '陈屹',   '陈屹',   'chenyi@demo.com',     '13733334444', 'MALE',   'NORMAL', NOW() - INTERVAL 2 DAY),
    (4, 1, NOW() - INTERVAL 26 DAY, NOW() - INTERVAL 3 HOUR, 0, 'heyutong',   '何雨桐', '何雨桐', 'heyutong@demo.com',   '13655556666', 'FEMALE', 'NORMAL', NOW() - INTERVAL 3 HOUR),
    (5, 1, NOW() - INTERVAL 20 DAY, NOW() - INTERVAL 6 HOUR,  0, 'linzhiyuan', '林知远', '林知远', 'linzhiyuan@demo.com', '13577778888', 'MALE',   'NORMAL', NOW() - INTERVAL 6 HOUR);

INSERT INTO sys_user_credentials (id, created_at, updated_at, deleted_at, tenant_id, user_id, identity_type, identifier, credential_type, credential, is_primary, status)
VALUES
    (2, NOW() - INTERVAL 30 DAY, NOW() - INTERVAL 30 DAY, NULL, 0, 2, 'USERNAME', 'suxiaotong', 'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', true,  'ENABLED'),
    (3, NOW() - INTERVAL 29 DAY, NOW() - INTERVAL 29 DAY, NULL, 0, 3, 'USERNAME', 'chenyi',     'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', true,  'ENABLED'),
    (4, NOW() - INTERVAL 26 DAY, NOW() - INTERVAL 26 DAY, NULL, 0, 4, 'USERNAME', 'heyutong',   'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', true,  'ENABLED'),
    (5, NOW() - INTERVAL 20 DAY, NOW() - INTERVAL 20 DAY, NULL, 0, 5, 'USERNAME', 'linzhiyuan', 'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', true,  'ENABLED');

-- 2. 促销：三档专价 + 高门槛满减 + 一条已结束活动（状态标签齐全）
INSERT INTO mall_promotions (id, created_at, updated_at, deleted_at, created_by, tenant_id, currency, name, promo_type, product_id, special_price, threshold_amount, reduction_amount, priority, valid_from, valid_until, status)
VALUES
    (3, NOW() - INTERVAL 14 DAY, NOW() - INTERVAL 14 DAY, NULL, 1, 0, 'CNY', 'Watch S3 限时直降',   'SPECIAL_PRICE',  116, 129900, 0,     0,    10, '2026-10-20 00:00:00', '2026-12-31 00:00:00', 'ACTIVE'),
    (4, NOW() - INTERVAL 10 DAY, NOW() - INTERVAL 10 DAY, NULL, 1, 0, 'CNY', 'Nova 12 新品尝鲜价',  'SPECIAL_PRICE',  111, 219900, 0,     0,    10, '2026-10-25 00:00:00', '2026-11-11 00:00:00', 'ACTIVE'),
    (5, NOW() - INTERVAL 8 DAY,  NOW() - INTERVAL 8 DAY,  NULL, 1, 0, 'CNY', 'Lumen 二代清仓特惠',  'SPECIAL_PRICE',  100, 4900,   0,     0,     5, '2026-10-28 00:00:00', '2026-11-30 00:00:00', 'ACTIVE'),
    (6, NOW() - INTERVAL 6 DAY,  NOW() - INTERVAL 6 DAY,  NULL, 1, 0, 'CNY', '店铺满500减80',       'FULL_REDUCTION', 0,   0,      50000, 8000, 8,  '2026-10-25 00:00:00', '2026-11-11 00:00:00', 'ACTIVE'),
    (7, NOW() - INTERVAL 40 DAY, NOW() - INTERVAL 40 DAY, NULL, 1, 0, 'CNY', 'Echo Mini 首发尝鲜',  'SPECIAL_PRICE',  106, 26900,  0,     0,     5, '2026-09-01 00:00:00', '2026-10-20 00:00:00', 'INACTIVE');

-- 3. 订单 112-123：近 30 天分布，状态覆盖待支付/已支付/已履约/已关闭/已取消
INSERT INTO mall_orders (id, created_at, updated_at, deleted_at, created_by, tenant_id, currency, business_ref_id, idempotency_key, user_id, total_amount, original_amount, discount_amount, shipping_fee, tax_amount, shipping_region, status, recipient_name, recipient_phone, shipping_address)
VALUES
    (112, NOW() - INTERVAL 20 DAY, NOW() - INTERVAL 20 DAY, NULL, 1, 0, 'CNY', 'ORD-DEMO-1112', 'idem-enrich-1112', 1, 23700,  23700,  0,    0, 0, 'CN', 'PENDING_PAYMENT', '苏晓桐', '13911112222', '浙江省杭州市西湖区文三路 88 号'),
    (113, NOW() - INTERVAL 18 DAY, NOW() - INTERVAL 17 DAY, NULL, 1, 0, 'CNY', 'ORD-DEMO-1113', 'idem-enrich-1113', 1, 129900, 129900, 0,    0, 0, 'CN', 'PAID',            '陈屹',   '13733334444', '四川省成都市高新区天府大道北段 1700 号'),
    (114, NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY, NULL, 1, 0, 'CNY', 'ORD-DEMO-1114', 'idem-enrich-1114', 1, 65700,  65700,  0,    0, 0, 'CN', 'PAID',            '何雨桐', '13655556666', '湖北省武汉市洪山区光谷大道 77 号'),
    (115, NOW() - INTERVAL 13 DAY, NOW() - INTERVAL 10 DAY, NULL, 1, 0, 'CNY', 'ORD-DEMO-1115', 'idem-enrich-1115', 1, 129900, 159900, 30000, 0, 0, 'CN', 'FULFILLED',       '林知远', '13577778888', '江苏省南京市建邺区江东中路 369 号'),
    (116, NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 8 DAY,  NULL, 1, 0, 'CNY', 'ORD-DEMO-1116', 'idem-enrich-1116', 1, 274800, 274800, 0,    0, 0, 'CN', 'FULFILLED',       '苏晓桐', '13911112222', '浙江省杭州市滨江区江南大道 588 号'),
    (117, NOW() - INTERVAL 9 DAY,  NOW() - INTERVAL 6 DAY,  NULL, 1, 0, 'CNY', 'ORD-DEMO-1117', 'idem-enrich-1117', 1, 17700,  17700,  0,    0, 0, 'CN', 'FULFILLED',       '陈屹',   '13733334444', '四川省成都市锦江区春熙路 99 号'),
    (118, NOW() - INTERVAL 7 DAY,  NOW() - INTERVAL 6 DAY,  NULL, 1, 0, 'CNY', 'ORD-DEMO-1118', 'idem-enrich-1118', 1, 43800,  51800,  8000, 0, 0, 'CN', 'PAID',            '何雨桐', '13655556666', '湖北省长沙市岳麓区麓谷大道 658 号'),
    (119, NOW() - INTERVAL 5 DAY,  NOW() - INTERVAL 2 DAY,  NULL, 1, 0, 'CNY', 'ORD-DEMO-1119', 'idem-enrich-1119', 1, 89900,  89900,  0,    0, 0, 'CN', 'CLOSED',          '林知远', '13577778888', '江苏省苏州市工业园区星湖街 328 号'),
    (120, NOW() - INTERVAL 4 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 1, 0, 'CNY', 'ORD-DEMO-1120', 'idem-enrich-1120', 1, 15900,  15900,  0,    0, 0, 'CN', 'CANCELLED',       '苏晓桐', '13911112222', '浙江省杭州市余杭区文一西路 969 号'),
    (121, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 1, 0, 'CNY', 'ORD-DEMO-1121', 'idem-enrich-1121', 1, 48700,  56700,  8000, 0, 0, 'CN', 'PAID',            '陈屹',   '13733334444', '四川省成都市武侯区人民南路四段 3 号'),
    (122, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 20 HOUR, NULL, 1, 0, 'CNY', 'ORD-DEMO-1122', 'idem-enrich-1122', 1, 549900, 549900, 0, 0, 0, 'CN', 'PAID',            '何雨桐', '13655556666', '湖南省长沙市开福区芙蓉中路 469 号'),
    (123, NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 2 HOUR, NULL, 1, 0, 'CNY', 'ORD-DEMO-1123', 'idem-enrich-1123', 1, 45800,  45800,  0,    0, 0, 'CN', 'PENDING_PAYMENT', '林知远', '13577778888', '江苏省南京市鼓楼区中山北路 8 号');

-- 订单项（单价×数量=小计；115 挂 Watch S3 专价、118/121 命中满500减80）
INSERT INTO mall_order_items (id, created_at, updated_at, deleted_at, created_by, tenant_id, order_id, user_id, sku_id, sku_snapshot, quantity, unit_price, subtotal)
VALUES
    (1029, NOW() - INTERVAL 20 DAY, NOW() - INTERVAL 20 DAY, NULL, 1, 0, 112, 1, 1054, '{"sku":"AERO-1054","name":"磁吸手机支架 Aero","sku_id":1054,"quantity":2,"currency":"CNY","unit_price":6900}', 2, 6900, 13800),
    (1030, NOW() - INTERVAL 20 DAY, NOW() - INTERVAL 20 DAY, NULL, 1, 0, 112, 1, 1048, '{"sku":"GLIDE-1048","name":"无线静音鼠标 Glide","sku_id":1048,"quantity":1,"currency":"CNY","unit_price":9900}', 1, 9900, 9900),
    (1031, NOW() - INTERVAL 18 DAY, NOW() - INTERVAL 18 DAY, NULL, 1, 0, 113, 1, 1006, '{"sku":"VAULT-1006","name":"智能门锁 Vault Pro","sku_id":1006,"quantity":1,"currency":"CNY","unit_price":129900}', 1, 129900, 129900),
    (1032, NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY, NULL, 1, 0, 114, 1, 1012, '{"sku":"TRIO-1012","name":"三合一无线充电板","sku_id":1012,"quantity":1,"currency":"CNY","unit_price":39900}', 1, 39900, 39900),
    (1033, NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY, NULL, 1, 0, 114, 1, 1030, '{"sku":"GAN-1030","name":"氮化镓充电器 65W","sku_id":1030,"quantity":2,"currency":"CNY","unit_price":12900}', 2, 12900, 25800),
    (1034, NOW() - INTERVAL 13 DAY, NOW() - INTERVAL 13 DAY, NULL, 1, 0, 115, 1, 1096, '{"sku":"WATCH-1096","name":"智能手表 Watch S3","sku_id":1096,"quantity":1,"currency":"CNY","unit_price":129900}', 1, 129900, 129900),
    (1035, NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 11 DAY, NULL, 1, 0, 116, 1, 1066, '{"sku":"NOVA-1066","name":"智能手机 Nova 12","sku_id":1066,"quantity":1,"currency":"CNY","unit_price":249900}', 1, 249900, 249900),
    (1036, NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 11 DAY, NULL, 1, 0, 116, 1, 1102, '{"sku":"BAND-1102","name":"运动手环 Band 7","sku_id":1102,"quantity":1,"currency":"CNY","unit_price":24900}', 1, 24900, 24900),
    (1037, NOW() - INTERVAL 9 DAY,  NOW() - INTERVAL 9 DAY,  NULL, 1, 0, 117, 1, 1000, '{"sku":"LUMEN-1000","name":"智能灯泡 Lumen 二代","sku_id":1000,"quantity":3,"currency":"CNY","unit_price":5900}', 3, 5900, 17700),
    (1038, NOW() - INTERVAL 7 DAY,  NOW() - INTERVAL 7 DAY,  NULL, 1, 0, 118, 1, 1042, '{"sku":"USB-C-1042","name":"Type-C 扩展坞 8 合 1","sku_id":1042,"quantity":1,"currency":"CNY","unit_price":21900}', 1, 21900, 21900),
    (1039, NOW() - INTERVAL 7 DAY,  NOW() - INTERVAL 7 DAY,  NULL, 1, 0, 118, 1, 1036, '{"sku":"ECHO-1036","name":"便携蓝牙音箱 Echo Mini","sku_id":1036,"quantity":1,"currency":"CNY","unit_price":29900}', 1, 29900, 29900),
    (1040, NOW() - INTERVAL 5 DAY,  NOW() - INTERVAL 5 DAY,  NULL, 1, 0, 119, 1, 1108, '{"sku":"KIDSAFE-1108","name":"儿童电话手表 KidSafe","sku_id":1108,"quantity":1,"currency":"CNY","unit_price":89900}', 1, 89900, 89900),
    (1041, NOW() - INTERVAL 4 DAY,  NOW() - INTERVAL 4 DAY,  NULL, 1, 0, 120, 1, 1060, '{"sku":"PEBBLE-1060","name":"移动电源 Pebble 10000","sku_id":1060,"quantity":1,"currency":"CNY","unit_price":15900}', 1, 15900, 15900),
    (1042, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 1, 0, 121, 1, 1102, '{"sku":"BAND-1102","name":"运动手环 Band 7","sku_id":1102,"quantity":2,"currency":"CNY","unit_price":24900}', 2, 24900, 49800),
    (1043, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 1, 0, 121, 1, 1054, '{"sku":"AERO-1054","name":"磁吸手机支架 Aero","sku_id":1054,"quantity":1,"currency":"CNY","unit_price":6900}', 1, 6900, 6900),
    (1044, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 1, 0, 122, 1, 1090, '{"sku":"NOVA-1090","name":"智能手机 Nova 12 Ultra","sku_id":1090,"quantity":1,"currency":"CNY","unit_price":549900}', 1, 549900, 549900),
    (1045, NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 2 HOUR, NULL, 1, 0, 123, 1, 1012, '{"sku":"TRIO-1012","name":"三合一无线充电板","sku_id":1012,"quantity":1,"currency":"CNY","unit_price":39900}', 1, 39900, 39900),
    (1046, NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 2 HOUR, NULL, 1, 0, 123, 1, 1000, '{"sku":"LUMEN-1000","name":"智能灯泡 Lumen 二代","sku_id":1000,"quantity":1,"currency":"CNY","unit_price":5900}', 1, 5900, 5900);

-- 支付流水（已支付/已履约/已关闭订单对应；多支付方式轮换）
INSERT INTO mall_payment_transactions (id, created_at, updated_at, deleted_at, created_by, tenant_id, currency, business_ref_id, idempotency_key, payment_method, business_type, order_id, user_id, amount, status, gateway_trade_no)
VALUES
    (24, NOW() - INTERVAL 18 DAY,  NOW() - INTERVAL 18 DAY,  NULL, 1, 0, 'CNY', 'PAY-DEMO-1121', 'idem-pay-enrich-1121', 'WECHAT',  'BUSINESS_TYPE_CONSUME', 113, 1, 129900, 'SUCCEEDED', 'WX2026100212345678'),
    (25, NOW() - INTERVAL 15 DAY,  NOW() - INTERVAL 15 DAY,  NULL, 1, 0, 'CNY', 'PAY-DEMO-1122', 'idem-pay-enrich-1122', 'ALIPAY',  'BUSINESS_TYPE_CONSUME', 114, 1, 65700,  'SUCCEEDED', 'AL2026100512345678'),
    (26, NOW() - INTERVAL 13 DAY,  NOW() - INTERVAL 13 DAY,  NULL, 1, 0, 'CNY', 'PAY-DEMO-1123', 'idem-pay-enrich-1123', 'BALANCE', 'BUSINESS_TYPE_CONSUME', 115, 1, 129900, 'SUCCEEDED', 'BAL2026100712345678'),
    (27, NOW() - INTERVAL 11 DAY,  NOW() - INTERVAL 11 DAY,  NULL, 1, 0, 'CNY', 'PAY-DEMO-1124', 'idem-pay-enrich-1124', 'WECHAT',  'BUSINESS_TYPE_CONSUME', 116, 1, 274800, 'SUCCEEDED', 'WX2026100912345678'),
    (28, NOW() - INTERVAL 9 DAY,   NOW() - INTERVAL 9 DAY,   NULL, 1, 0, 'CNY', 'PAY-DEMO-1125', 'idem-pay-enrich-1125', 'ALIPAY',  'BUSINESS_TYPE_CONSUME', 117, 1, 17700,  'SUCCEEDED', 'AL2026101112345678'),
    (29, NOW() - INTERVAL 7 DAY,   NOW() - INTERVAL 7 DAY,   NULL, 1, 0, 'CNY', 'PAY-DEMO-1126', 'idem-pay-enrich-1126', 'WECHAT',  'BUSINESS_TYPE_CONSUME', 118, 1, 43800,  'SUCCEEDED', 'WX2026101312345678'),
    (30, NOW() - INTERVAL 5 DAY,   NOW() - INTERVAL 5 DAY,   NULL, 1, 0, 'CNY', 'PAY-DEMO-1127', 'idem-pay-enrich-1127', 'ALIPAY',  'BUSINESS_TYPE_CONSUME', 119, 1, 89900,  'SUCCEEDED', 'AL2026101512345678'),
    (31, NOW() - INTERVAL 3 DAY,   NOW() - INTERVAL 3 DAY,   NULL, 1, 0, 'CNY', 'PAY-DEMO-1128', 'idem-pay-enrich-1128', 'ALIPAY',  'BUSINESS_TYPE_CONSUME', 121, 1, 48700,  'SUCCEEDED', 'AL2026101712345678'),
    (32, NOW() - INTERVAL 1 DAY,   NOW() - INTERVAL 1 DAY,   NULL, 1, 0, 'CNY', 'PAY-DEMO-1129', 'idem-pay-enrich-1129', 'BALANCE', 'BUSINESS_TYPE_CONSUME', 122, 1, 549900, 'SUCCEEDED', 'BAL2026101912345678');

-- 退款：一笔已到账（七天无理由）、一笔处理中（部分退款补偿）
INSERT INTO mall_payment_refunds (id, created_at, updated_at, deleted_at, created_by, tenant_id, currency, business_ref_id, idempotency_key, transaction_id, user_id, amount, status, reason)
VALUES
    (10, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 2 DAY, NULL, 1, 0, 'CNY', 'REF-DEMO-1010', 'idem-refund-enrich-1010', 30, 1, 89900, 'SUCCEEDED', '七天无理由退货'),
    (11, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY, NULL, 1, 0, 'CNY', 'REF-DEMO-1011', 'idem-refund-enrich-1011', 31, 1, 8000,  'PENDING',   '满减活动少算差价补偿');

-- 发货单：已签收/运输中/待发货覆盖（含给存量订单 101 补的发货）
INSERT INTO mall_shipments (id, created_at, updated_at, deleted_at, created_by, tenant_id, order_id, user_id, carrier, tracking_number, status, tracking_events)
VALUES
    (6,  NOW() - INTERVAL 12 DAY, NOW() - INTERVAL 10 DAY, NULL, 1, 0, 115, 1, '顺丰速运', 'SF2026102100115', 'DELIVERED', '[{"time":"2026-10-21 14:00","event":"包裹已签收"},{"time":"2026-10-20 21:00","event":"运输中"},{"time":"2026-10-20 09:00","event":"已发货"}]'),
    (7,  NOW() - INTERVAL 10 DAY, NOW() - INTERVAL 8 DAY,  NULL, 1, 0, 116, 1, '京东物流', 'JD2026102300116', 'SHIPPED',   '[{"time":"2026-10-24 08:30","event":"运输中"},{"time":"2026-10-23 20:00","event":"已发货"}]'),
    (8,  NOW() - INTERVAL 8 DAY,  NOW() - INTERVAL 6 DAY,  NULL, 1, 0, 117, 1, '中通快递', 'ZT2026102500117', 'DELIVERED', '[{"time":"2026-10-26 11:20","event":"包裹已签收"},{"time":"2026-10-25 22:00","event":"运输中"},{"time":"2026-10-25 10:00","event":"已发货"}]'),
    (9,  NOW() - INTERVAL 6 DAY,  NOW() - INTERVAL 6 DAY,  NULL, 1, 0, 118, 1, '顺丰速运', NULL,              'PENDING',   '[]'),
    (10, NOW() - INTERVAL 16 DAY, NOW() - INTERVAL 14 DAY, NULL, 1, 0, 101, 1, '圆通速递', 'YT2026101800101', 'DELIVERED', '[{"time":"2026-10-19 15:00","event":"包裹已签收"},{"time":"2026-10-18 21:00","event":"运输中"},{"time":"2026-10-18 09:00","event":"已发货"}]'),
    (11, NOW() - INTERVAL 4 DAY,  NOW() - INTERVAL 4 DAY,  NULL, 1, 0, 119, 1, 'EMS',      'EMS2026102700119', 'DELIVERED', '[{"time":"2026-10-28 09:40","event":"包裹已签收"},{"time":"2026-10-27 20:00","event":"运输中"},{"time":"2026-10-27 08:00","event":"已发货"}]');

-- 商品评论：五星为主带 3-4 星与两条待审核，作者覆盖 5 个顾客；回复为商家（admin）
INSERT INTO mall_comments (id, created_at, updated_at, deleted_at, created_by, tenant_id, content_type, object_id, content, status, rating, parent_id)
VALUES
    (164, NOW() - INTERVAL 12 DAY,  NOW() - INTERVAL 12 DAY,  NULL, 2, 0, 'CONTENT_TYPE_PRODUCT', 116, 'Watch S3 续航真能打，重度使用六天还有电，表盘也好看。', 'STATUS_APPROVED', 5, NULL),
    (165, NOW() - INTERVAL 9 DAY,   NOW() - INTERVAL 9 DAY,   NULL, 3, 0, 'CONTENT_TYPE_PRODUCT', 116, '双频 GPS 轨迹比上一代稳太多了，越野跑飘移很小。', 'STATUS_APPROVED', 5, NULL),
    (166, NOW() - INTERVAL 6 DAY,   NOW() - INTERVAL 6 DAY,   NULL, 4, 0, 'CONTENT_TYPE_PRODUCT', 116, '功能没得挑，就是氟橡胶表带夏天戴有点闷，希望出编织款。', 'STATUS_APPROVED', 4, NULL),
    (167, NOW() - INTERVAL 2 DAY,   NOW() - INTERVAL 2 DAY,   NULL, 5, 0, 'CONTENT_TYPE_PRODUCT', 116, '给爸买的老花眼也能看清大字体模式，血氧监测很实用。', 'STATUS_APPROVED', 5, NULL),
    (168, NOW() - INTERVAL 10 DAY,  NOW() - INTERVAL 10 DAY,  NULL, 2, 0, 'CONTENT_TYPE_PRODUCT', 111, 'Nova 12 夜景很惊喜，夜景模式出片率比上代高不少。', 'STATUS_APPROVED', 5, NULL),
    (169, NOW() - INTERVAL 7 DAY,   NOW() - INTERVAL 7 DAY,   NULL, 3, 0, 'CONTENT_TYPE_PRODUCT', 111, '整体均衡，就是 200g 出头略压手，女生单手略累。', 'STATUS_APPROVED', 4, NULL),
    (170, NOW() - INTERVAL 3 DAY,   NOW() - INTERVAL 3 DAY,   NULL, 4, 0, 'CONTENT_TYPE_PRODUCT', 111, '80W 快充实测 35 分钟充满，出门终于不用带充电宝了。', 'STATUS_APPROVED', 5, NULL),
    (171, NOW() - INTERVAL 14 DAY,  NOW() - INTERVAL 14 DAY,  NULL, 3, 0, 'CONTENT_TYPE_PRODUCT', 100, '全屋联动很顺，人来灯亮人走灯灭，米家生态无缝接入。', 'STATUS_APPROVED', 4, NULL),
    (172, NOW() - INTERVAL 5 DAY,   NOW() - INTERVAL 5 DAY,   NULL, 5, 0, 'CONTENT_TYPE_PRODUCT', 100, '硬件没问题，但 app 首次配网转了半天，固件更新后好转。', 'STATUS_APPROVED', 3, NULL),
    (173, NOW() - INTERVAL 8 DAY,   NOW() - INTERVAL 8 DAY,   NULL, 2, 0, 'CONTENT_TYPE_PRODUCT', 102, '手机手表耳机三设备同时充，桌面终于只剩一根线。', 'STATUS_APPROVED', 5, NULL),
    (174, NOW() - INTERVAL 6 DAY,   NOW() - INTERVAL 6 DAY,   NULL, 4, 0, 'CONTENT_TYPE_PRODUCT', 106, '体积是够小，但低音下潜一般，听电子乐有点糊。', 'STATUS_APPROVED', 2, NULL),
    (175, NOW() - INTERVAL 4 DAY,   NOW() - INTERVAL 4 DAY,   NULL, 5, 0, 'CONTENT_TYPE_PRODUCT', 106, '户外随身放兜里就走，防尘做得不错，适合露营。', 'STATUS_APPROVED', 4, NULL),
    (176, NOW() - INTERVAL 13 DAY,  NOW() - INTERVAL 13 DAY,  NULL, 2, 0, 'CONTENT_TYPE_PRODUCT', 113, 'Fold X 展开追剧爽，铰链比上代紧实，折痕浅到几乎忽略。', 'STATUS_APPROVED', 5, NULL),
    (177, NOW() - INTERVAL 5 DAY,   NOW() - INTERVAL 5 DAY,   NULL, 3, 0, 'CONTENT_TYPE_PRODUCT', 118, 'KidSafe 定位准，电子围栏一出门就推送，娃上学放心多了。', 'STATUS_APPROVED', 5, NULL),
    (178, NOW() - INTERVAL 4 DAY,   NOW() - INTERVAL 4 DAY,   NULL, 2, 0, 'CONTENT_TYPE_PRODUCT', 3,   '耳机 Pro 降噪开一键世界安静，地铁通勤救星，单次续航够用。', 'STATUS_APPROVED', 5, NULL),
    (179, NOW() - INTERVAL 2 DAY,   NOW() - INTERVAL 2 DAY,   NULL, 4, 0, 'CONTENT_TYPE_PRODUCT', 3,   '音质对得起价格，入耳检测灵敏，眼镜党佩戴略有存在感。', 'STATUS_APPROVED', 4, NULL),
    (180, NOW() - INTERVAL 3 DAY,   NOW() - INTERVAL 3 DAY,   NULL, 3, 0, 'CONTENT_TYPE_PRODUCT', 4,   'X2 的轨迹导入导出很方便，骑行数据直接同步俱乐部平台。', 'STATUS_APPROVED', 5, NULL),
    (181, NOW() - INTERVAL 1 DAY,   NOW() - INTERVAL 1 DAY,   NULL, 2, 0, 'CONTENT_TYPE_PRODUCT', 106, '二单元比想象中均衡，人声突出，通勤够用。', 'STATUS_PENDING', 4, NULL),
    (182, NOW() - INTERVAL 12 HOUR, NOW() - INTERVAL 12 HOUR, NULL, 5, 0, 'CONTENT_TYPE_PRODUCT', 111, '首发购入，送了原厂壳膜，包装很扎实。', 'STATUS_PENDING', 5, NULL),
    (183, NOW() - INTERVAL 5 DAY,   NOW() - INTERVAL 5 DAY,   NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 106, '感谢反馈！低频表现我们会在后续固件 EQ 里优化，可关注更新。', 'STATUS_APPROVED', NULL, 174),
    (184, NOW() - INTERVAL 5 DAY,   NOW() - INTERVAL 5 DAY,   NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 116, '编织表带已在路上，下月上架，到时可凭订单享五折换购。', 'STATUS_APPROVED', NULL, 166),
    (185, NOW() - INTERVAL 8 DAY,   NOW() - INTERVAL 8 DAY,   NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 116, '感谢认可！轨迹功能建议搭配固件 2.1 的多星搜狗使用，精度更佳。', 'STATUS_APPROVED', NULL, 165),
    (186, NOW() - INTERVAL 11 DAY,  NOW() - INTERVAL 11 DAY,  NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 116, '欢迎加入 Watch 家族！长按表冠可一键唤出健康面板。', 'STATUS_APPROVED', NULL, 164),
    (187, NOW() - INTERVAL 4 DAY,   NOW() - INTERVAL 4 DAY,   NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 118, '感谢信任！开学季凭此评论截图可联系客服领取保护壳。', 'STATUS_APPROVED', NULL, 177),
    (188, NOW() - INTERVAL 3 DAY,   NOW() - INTERVAL 3 DAY,   NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 3,   '感谢支持！搭配充电盒总续航 32 小时，出差党可以放心。', 'STATUS_APPROVED', NULL, 178),
    (189, NOW() - INTERVAL 9 DAY,   NOW() - INTERVAL 9 DAY,   NULL, 1, 0, 'CONTENT_TYPE_PRODUCT', 111, '感谢分享！夜景模式长按可锁定超广角，试试星空模式。', 'STATUS_APPROVED', NULL, 168);

-- 站内信：新品/活动/维护/积分四类，一条已读（未读 badge 与状态列都有内容）
INSERT INTO internal_messages (id, created_at, updated_at, deleted_at, tenant_id, title, content, sender_id, category_id, status, type)
VALUES
    (35, NOW() - INTERVAL 5 DAY,  NOW() - INTERVAL 5 DAY,  NULL, 0, 'Nova 12 系列新品上架', 'Nova 12 / 12 Pro / 12 Ultra 全系开售，首发期下单享专属好礼。', 0, NULL, 'PUBLISHED', 'NOTIFICATION'),
    (36, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 0, '双十一预热开启',       '店铺满 500 减 80，Watch S3 限时直降 300，先到先得！', 0, NULL, 'PUBLISHED', 'NOTIFICATION'),
    (37, NOW() - INTERVAL 2 DAY,  NOW() - INTERVAL 2 DAY,  NULL, 0, '系统维护通知',         '本周日 02:00-04:00 系统升级，期间下单可能短暂受限。', 0, NULL, 'PUBLISHED', 'NOTIFICATION'),
    (38, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 0, '积分商城上新',         '积分兑换专区新增 Mesh 路由器与便携风扇，欢迎兑换。', 0, NULL, 'PUBLISHED', 'NOTIFICATION');

INSERT INTO internal_message_recipients (id, created_at, updated_at, deleted_at, tenant_id, message_id, recipient_user_id, status, received_at, read_at)
VALUES
    (31, NOW() - INTERVAL 5 DAY,  NOW() - INTERVAL 5 DAY,  NULL, 0, 35, 1, 'SENT', NOW() - INTERVAL 5 DAY,  NULL),
    (32, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 0, 36, 1, 'SENT', NOW() - INTERVAL 3 DAY,  NULL),
    (33, NOW() - INTERVAL 2 DAY,  NOW() - INTERVAL 2 DAY,  NULL, 0, 37, 1, 'SENT', NOW() - INTERVAL 2 DAY,  NULL),
    (34, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 0, 38, 1, 'READ', NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 20 HOUR);

-- 购物车：耳机 Pro / 手表 X2 进购物车（C端演示合并商品）
INSERT INTO mall_cart_items (id, created_at, updated_at, deleted_at, created_by, tenant_id, cart_id, user_id, sku_id, quantity)
VALUES
    (105, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 1, 0, 1, 1, 1, 1),
    (106, NOW() - INTERVAL 6 HOUR, NOW() - INTERVAL 6 HOUR, NULL, 1, 0, 1, 1, 2, 1);

-- 库存预警：两条 OPEN（对应现库存 1 与 3 的紧 SKU）
INSERT INTO mall_stock_alerts (id, created_at, updated_at, deleted_at, created_by, sku_id, stock_qty_at_trigger, threshold, status)
VALUES
    (5, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 1, 1060, 1, 5, 'OPEN'),
    (6, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 1, 1066, 3, 5, 'OPEN');

-- 库存操作流水：与新订单一一对应，期末值对齐 SKU 现存库存
INSERT INTO mall_stock_op_logs (id, created_at, updated_at, deleted_at, created_by, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, remark)
VALUES
    (52, NOW() - INTERVAL 18 DAY, NOW() - INTERVAL 18 DAY, NULL, 1, 1006, 101,  -1, 49, 48, 'ORDER_DEDUCT', 113, NULL),
    (53, NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY, NULL, 1, 1012, 102,  -1, 76, 75, 'ORDER_DEDUCT', 114, NULL),
    (54, NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY, NULL, 1, 1030, 105,  -2, 30, 28, 'ORDER_DEDUCT', 114, NULL),
    (55, NOW() - INTERVAL 13 DAY, NOW() - INTERVAL 13 DAY, NULL, 1, 1096, 116,  -1, 4,  3,  'ORDER_DEDUCT', 115, NULL),
    (56, NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 11 DAY, NULL, 1, 1066, 111,  -1, 4,  3,  'ORDER_DEDUCT', 116, NULL),
    (57, NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 11 DAY, NULL, 1, 1102, 117,  -1, 6,  5,  'ORDER_DEDUCT', 116, NULL),
    (58, NOW() - INTERVAL 9 DAY,  NOW() - INTERVAL 9 DAY,  NULL, 1, 1000, 100,  -3, 15, 12, 'ORDER_DEDUCT', 117, NULL),
    (59, NOW() - INTERVAL 7 DAY,  NOW() - INTERVAL 7 DAY,  NULL, 1, 1042, 107,  -1, 29, 28, 'ORDER_DEDUCT', 118, NULL),
    (60, NOW() - INTERVAL 7 DAY,  NOW() - INTERVAL 7 DAY,  NULL, 1, 1036, 106,  -1, 4,  3,  'ORDER_DEDUCT', 118, NULL),
    (61, NOW() - INTERVAL 5 DAY,  NOW() - INTERVAL 5 DAY,  NULL, 1, 1108, 118,  -1, 131, 130, 'ORDER_DEDUCT', 119, NULL),
    (62, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 1, 1102, 117,  -2, 5,  3,  'ORDER_DEDUCT', 121, NULL),
    (63, NOW() - INTERVAL 3 DAY,  NOW() - INTERVAL 3 DAY,  NULL, 1, 1054, 109,  -1, 47, 46, 'ORDER_DEDUCT', 121, NULL),
    (64, NOW() - INTERVAL 1 DAY,  NOW() - INTERVAL 1 DAY,  NULL, 1, 1090, 115,  -1, 6,  5,  'ORDER_DEDUCT', 122, NULL);


-- 提交事务
COMMIT;

SELECT '测试数据初始化完成' AS result;
