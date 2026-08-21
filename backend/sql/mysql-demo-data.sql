-- Description: 初始化测试数据（租户/用户/组织架构/字典/站内信分类等）- MySQL版
-- Note: 表结构创建完成后执行，执行前备份数据；MySQL需5.7+（支持JSON/COALESCE）
-- 关闭外键检查，避免TRUNCATE关联表报错
SET FOREIGN_KEY_CHECKS = 0;
-- 开启事务，保证数据原子性
START TRANSACTION;

-- 一次性清理相关表并重置自增（MySQL无外键级联TRUNCATE，先关外键检查）
TRUNCATE TABLE sys_org_units AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_positions AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_tasks AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_login_policies AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_dict_types AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_dict_entries AUTO_INCREMENT = 1;
TRUNCATE TABLE internal_message_categories AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_languages AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_tenants AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_users AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_user_credentials AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_memberships AUTO_INCREMENT = 1;
TRUNCATE TABLE sys_membership_roles AUTO_INCREMENT = 1;

-- 测试租户
INSERT INTO sys_tenants(id, name, code, type, audit_status, status, admin_user_id, created_at)
VALUES (1, '测试租户', 'super', 'PAID', 'APPROVED', 'ON', 2, NOW());
-- 重置自增，后续新增从MAX(id)+1开始（COALESCE处理空表情况）
ALTER TABLE sys_tenants AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_tenants);

-- 插入租户管理员用户
INSERT INTO sys_users (id, tenant_id, username, nickname, realname, email, gender, created_at)
VALUES (2, 1, 'tenant_admin', '租户管理', '张管理员', 'tenant@company.com', 'MALE', NOW());
ALTER TABLE sys_users AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_users);

-- 插入租户管理员的登录凭证（密码：admin，与平台管理员哈希值一致）
INSERT INTO sys_user_credentials (tenant_id, user_id, identity_type, identifier, credential_type, credential, status,
                                  is_primary, created_at)
VALUES
    -- 租户管理员（对应users表id=2，tenant_id=1）
    (1, 2, 'USERNAME', 'tenant_admin', 'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a',
     'ENABLED', true, NOW()),
    (1, 2, 'EMAIL', 'tenant@company.com', 'PASSWORD_HASH',
     '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', 'ENABLED', false, NOW());
ALTER TABLE sys_user_credentials AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_user_credentials);

-- 组织架构单元
INSERT INTO sys_org_units (id, tenant_id, parent_id, type, name, code, description, path, sort_order, leader_id, status, created_at)
VALUES
    (1, 1, NULL, 'COMPANY', 'XX集团总部', 'HEADQUARTERS', '集团核心管理机构，统筹全集团战略规划、业务管控及资源调配', '/1', 1, 1, 'ON', NOW()),
    (2, 1, 1, 'DIVISION', '技术部', 'TECH', '负责集团整体技术架构规划、研发管理、系统运维及技术创新', '/1/2', 2, 5, 'ON', NOW()),
    (3, 1, 1, 'DIVISION', '财务部', 'FIN', '负责集团财务核算、资金管理、税务筹划、预算编制及财务风控', '/1/3', 3, 8, 'ON', NOW()),
    (4, 1, 1, 'DIVISION', '人事部', 'HR', '负责人力资源规划、招聘配置、薪酬绩效、员工培训及组织发展', '/1/4', 4, 9, 'ON', NOW()),
    (5, 1, 2, 'DEPARTMENT', '研发一部', 'DEV-1', '聚焦新能源领域产品研发、技术迭代及核心模块开发', '/1/2/5', 1, 6, 'ON', NOW()),
    (6, 1, 1, 'REGION', '华北大区', 'NORTH', '负责华北区域市场运营、客户维护、销售管理及本地化服务落地', '/1/6', 3, 12, 'ON', NOW()),
    (7, 1, 1, 'SUBSIDIARY', '广州分公司', 'GZ', '负责华南区域（广州及周边）业务拓展、客户服务及本地化运营', '/1/7', 5, 2, 'ON', NOW()),
    (8, 1, 1, 'SUBSIDIARY', '深圳子公司', 'SZ', '负责深圳区域市场开拓、科技创新业务落地及高端客户对接', '/1/8', 6, 4, 'ON', NOW()),
    (9, 1, 1, 'DIVISION', '销售部', 'SALES', '统筹集团整体销售策略制定、销售团队管理及业绩目标达成', '/1/9', 7, 16, 'ON', NOW()),
    (10, 1, 9, 'DEPARTMENT', '海外事业部', 'INTL', '负责海外市场拓展、国际客户合作、跨境业务管理及本地化运营', '/1/9/10', 1, 17, 'ON', NOW()),
    (11, 1, 10, 'TEAM', '海外销售组', 'INTL-SALES-1', '具体执行海外市场销售任务，跟进客户需求及订单落地', '/1/9/10/11', 1, 18, 'ON', NOW()),
    (12, 1, 5, 'PROJECT', '新能源项目组', 'NEO-PROJ', '专项负责新能源项目的研发、落地、运营及成果转化', '/1/2/5/12', 1, 6, 'ON', NOW()),
    (13, 1, 1, 'COMMITTEE', '审计委员会', 'AUDIT', '独立开展集团内部审计、风控检查、合规监督及问题整改跟进', '/1/13', 8, 12, 'ON', NOW()),
    (14, 1, 1, 'DEPARTMENT', '客服部', 'CS', '负责全集团客户咨询、投诉处理、售后服务及客户满意度提升', '/1/14', 9, 11, 'ON', NOW()),
    (15, 1, 14, 'TEAM', '客服一组', 'CS-1', '承接华南区域客户服务、售后问题处理及客户关系维护', '/1/14/15', 1, 20, 'ON', NOW());
ALTER TABLE sys_org_units AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_org_units);

-- 岗位数据
INSERT INTO sys_positions (id, tenant_id, type, name, code, org_unit_id, reports_to_position_id, description, job_family, job_grade, level, headcount, is_key_position, status, sort_order, created_at)
VALUES
    (1, 1, 'LEADER', '技术总监', 'TECH-DIRECTOR-001', 2, NULL, '负责公司整体技术战略规划、团队管理及核心技术决策', 'TECH', 1, 1, 1, true, 'ON', 1, NOW()),
    (2, 1, 'MANAGER', '技术部经理', 'TECH-MANAGER-001', 2, 1, '负责技术部日常管理、项目排期及团队协作', 'TECH', 2, 2, 1, true, 'ON', 2, NOW()),
    (3, 1, 'MANAGER', '前端主管', 'TECH-FE-LEADER-001', 2, 2, '负责前端团队开发管理、技术方案评审及需求落地', 'TECH', 3, 3, 3, false, 'ON', 3, NOW()),
    (4, 1, 'MANAGER', '后端主管', 'TECH-BE-LEADER-001', 2, 2, '负责后端服务架构设计、数据库优化及接口开发管理', 'TECH', 4, 3, 3, false, 'ON', 4, NOW()),
    (5, 1, 'REGULAR', '前端开发专员', 'TECH-FE-DEV-001', 2, 3, '负责Web/移动端前端页面开发、交互实现及兼容性优化', 'TECH', 5, 4, 5, false, 'ON', 5, NOW()),
    (6, 1, 'REGULAR', '后端开发专员', 'TECH-BE-DEV-001', 2, 4, '负责后端接口开发、业务逻辑实现及系统稳定性维护', 'TECH', 6, 4, 5, false, 'ON', 6, NOW()),
    (7, 1, 'REGULAR', '测试工程师', 'TECH-TEST-001', 2, 2, '负责项目功能测试、性能测试及自动化测试脚本开发', 'TECH', 3, 4, 3, false, 'ON', 7, NOW()),
    (8, 1, 'LEADER', '人力总监', 'HR-DIRECTOR-001', 2, NULL, '负责人力资源战略规划、组织架构设计及人才梯队建设', 'HR', 1, 1, 1, true, 'ON', 1, NOW()),
    (9, 1, 'MANAGER', '招聘主管', 'HR-RECRUIT-LEADER-001', 2, 8, '负责公司各部门招聘需求对接、简历筛选及面试安排', 'HR', 2, 2, 1, false, 'ON', 2, NOW()),
    (10, 1, 'REGULAR', '薪酬绩效专员', 'HR-C&P-001', 2, 8, '负责员工薪酬核算、绩效考核制度落地及社保公积金管理', 'HR', 3, 2, 1, false, 'ON', 3, NOW()),
    (11, 1, 'REGULAR', 'HRBP', 'HR-BP-001', 2, 8, '对接业务部门，提供人力资源支持（入离职、员工关系等）', 'HR', 4, 2, 1, false, 'ON', 4, NOW()),
    (12, 1, 'LEADER', '财务总监', 'FIN-DIRECTOR-001', 2, NULL, '负责公司财务战略、预算管理及财务风险控制', 'FIN', 1, 1, 1, true, 'ON', 1, NOW()),
    (13, 1, 'MANAGER', '会计主管', 'FIN-ACCOUNT-LEADER-001', 2, 12, '负责账务处理、财务报表编制及税务申报管理', 'FIN', 2, 2, 1, false, 'ON', 2, NOW()),
    (14, 1, 'REGULAR', '出纳专员', 'FIN-CASHIER-001', 2, 13, '负责日常资金收付、银行对账及票据管理', 'FIN', 3, 3, 1, false, 'ON', 3, NOW()),
    (15, 1, 'REGULAR', '成本会计', 'FIN-COST-001', 2, 13, '负责成本核算、成本分析及成本控制方案制定', 'FIN', 4, 3, 1, false, 'ON', 4, NOW()),
    (16, 1, 'LEADER', '市场总监', 'MKT-DIRECTOR-001', 4, NULL, '负责市场战略规划、品牌建设及营销活动策划', 'MKT', 1, 1, 1, true, 'ON', 1, NOW()),
    (17, 1, 'MANAGER', '新媒体运营主管', 'MKT-NEWS-LEADER-001', 4, 16, '负责新媒体平台内容运营及用户增长', 'MKT', 2, 2, 1, false, 'ON', 2, NOW()),
    (18, 1, 'REGULAR', '活动策划专员', 'MKT-EVENT-001', 4, 16, '负责线下活动策划、执行及效果复盘', 'MKT', 3, 3, 1, false, 'ON', 3, NOW()),
    (19, 1, 'REGULAR', '市场调研专员', 'MKT-RESEARCH-001', 4, 16, '负责行业动态调研、竞品分析及市场趋势报告撰写', 'MKT', 4, 3, 1, false, 'ON', 4, NOW()),
    (20, 1, 'REGULAR', '行政助理', 'ADMIN-ASSIST-001', 2, 8, '负责办公用品采购、会议安排等行政工作（已合并至HRBP）', 'ADMIN', 5, 5, 1, false, 'OFF', 5, NOW());
ALTER TABLE sys_positions AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_positions);

-- 用户-租户关联关系
INSERT INTO sys_memberships (id, tenant_id, user_id, org_unit_id, position_id, role_id, is_primary, status)
VALUES (2, 1, 2, null, null, 2, true, 'ACTIVE');
ALTER TABLE sys_memberships AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_memberships);

-- 租户成员-角色关联关系（修复原脚本少分号、缺分隔符错误）
INSERT INTO sys_membership_roles (id, membership_id, tenant_id, role_id, is_primary, status)
VALUES (2, 2, 1, 2, true, 'ACTIVE');
ALTER TABLE sys_membership_roles AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_membership_roles);

-- 调度任务（JSON字段直接保留字符串，MySQL5.7+支持）
INSERT INTO sys_tasks(type, type_name, task_payload, cron_spec, enable, created_at)
VALUES ('PERIODIC', 'backup', '{ "name": "test"}', '0 * * * *', true, NOW());
INSERT INTO sys_tasks(type, type_name, task_payload, cron_spec, enable, created_at)
VALUES ('PERIODIC', 'coupon_expire_sweep', '{ "name": "sweep"}', '0 */6 * * *', true, NOW());
INSERT INTO sys_tasks(type, type_name, task_payload, cron_spec, enable, created_at)
VALUES ('PERIODIC', 'stock_alert', '{ "name": "scan"}', '0 2 * * *', true, NOW());
ALTER TABLE sys_tasks AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_tasks);

-- 登录策略
INSERT INTO sys_login_policies(id, target_id, type, method, value, reason, created_at)
VALUES
    (1, 1, 'BLACKLIST', 'IP', '127.0.0.1', '无理由', NOW()),
    (2, 1, 'WHITELIST', 'MAC', '00:1B:44:11:3A:B7 ', '无理由', NOW());
ALTER TABLE sys_login_policies AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_login_policies);

-- 字典类型（移除PG的序列重置指令，MySQL TRUNCATE已重置）
INSERT INTO sys_dict_types(id, type_code, type_name, sort_order, description, is_enabled, created_at)
VALUES
    (1, 'USER_STATUS', '用户状态', 10, '系统用户的状态管理，包括正常、冻结、注销', true, NOW()),
    (2, 'DEVICE_TYPE', '设备类型', 20, 'IoT平台接入的设备品类，新增需同步至设备接入模块', true, NOW()),
    (3, 'ORDER_STATUS', '订单状态', 30, '电商订单的全生命周期状态', true, NOW()),
    (4, 'GENDER', '性别', 40, '用户性别枚举，默认未知', true, NOW()),
    (5, 'PAYMENT_METHOD', '支付方式', 50, '支持的支付渠道，含第三方支付和自有渠道', true, NOW());
ALTER TABLE sys_dict_types AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_dict_types);

-- 字典条目
INSERT INTO sys_dict_entries(id, type_id, entry_value, entry_label, numeric_value, sort_order, description, is_enabled, created_at)
VALUES
    -- 用户状态
    (1, 1, 'NORMAL', '正常', 1, 1, '用户可正常登录和操作', true, NOW()),
    (2, 1, 'FROZEN', '冻结', 2, 2, '因违规被临时冻结，需管理员解冻', true, NOW()),
    (3, 1, 'CANCELED', '注销', 3, 3, '用户主动注销，数据保留但不可登录', true, NOW()),
    -- 设备类型
    (4, 2, 'TEMP_SENSOR', '温湿度传感器', 101, 1, '支持温度（-20~80℃）和湿度（0~100%RH）采集', true, NOW()),
    (5, 2, 'CURRENT_METER', '电流仪表', 102, 2, '交流/直流电流测量，精度0.5级', true, NOW()),
    (6, 2, 'GAS_DETECTOR', '气体探测器', 103, 3, '暂不支持，待硬件适配（2025Q4计划启用）', false, NOW()),
    -- 订单状态
    (7, 3, 'PENDING', '待支付', 1, 1, '下单后未支付，超时自动取消', true, NOW()),
    (8, 3, 'PAID', '已支付', 2, 2, '支付成功，等待发货', true, NOW()),
    (9, 3, 'SHIPPED', '已发货', 3, 3, '商品已出库，物流配送中', true, NOW()),
    (10, 3, 'COMPLETED', '已完成', 4, 4, '用户确认收货，订单结束', true, NOW()),
    (11, 3, 'CANCELED', '已取消', 5, 5, '用户或系统取消订单', true, NOW()),
    -- 性别
    (12, 4, 'MALE', '男', 1, 1, '', true, NOW()),
    (13, 4, 'FEMALE', '女', 2, 2, '', true, NOW()),
    (14, 4, 'UNKNOWN', '未知', 0, 3, '用户未填写时默认值', true, NOW()),
    -- 支付方式
    (15, 5, 'ALIPAY', '支付宝', 1, 1, '支持花呗、余额宝', true, NOW()),
    (16, 5, 'WECHAT', '微信支付', 2, 2, '需绑定微信', true, NOW()),
    (17, 5, 'UNIONPAY', '银联支付', 3, 3, '支持信用卡、储蓄卡', true, NOW()),
    (18, 5, 'CASH', '现金支付', 4, 4, '线下支付，已废弃（2025-01停用）', false, NOW());
ALTER TABLE sys_dict_entries AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_dict_entries);

-- 站内信分类（扁平结构，无树形层级）
INSERT INTO internal_message_categories (id, code, name, remark, sort_order, is_enabled, created_at)
VALUES
    -- 订单相关分类
    (1, 'order', '订单通知', '包含订单支付、发货、退款等全流程通知', 1, true, NOW()),
    (101, 'order_paid', '支付成功', '订单支付完成时触发的通知', 2, true, NOW()),
    (102, 'order_unpaid', '支付超时', '订单未在规定时间内支付的提醒', 3, true, NOW()),
    (103, 'order_shipped', '已发货', '商家发货后通知用户', 4, true, NOW()),
    (104, 'order_refunded', '已退款', '订单退款流程完成的通知', 5, true, NOW()),
    -- 系统相关分类
    (2, 'system', '系统通知', '系统公告、维护提醒、版本更新等平台级通知', 6, true, NOW()),
    (201, 'system_announcement', '系统公告', '平台规则更新、重要通知等', 7, true, NOW()),
    (202, 'system_maintenance', '维护通知', '系统计划内维护的时间提醒', 8, true, NOW()),
    (203, 'system_upgrade', '版本更新', '客户端或功能升级的提示', 9, true, NOW()),
    -- 活动相关分类
    (3, 'activity', '活动通知', '营销活动报名、开始、结束等提醒', 10, true, NOW()),
    (301, 'activity_signup', '报名成功', '用户报名活动后确认通知', 11, true, NOW()),
    (302, 'activity_start', '活动开始', '活动即将开始的倒计时提醒', 12, true, NOW()),
    (303, 'activity_end', '活动结束', '活动结束及结果公示通知', 13, true, NOW()),
    -- 用户相关分类
    (4, 'user', '用户通知', '账号安全、信息变更、权限调整等个人相关通知', 14, true, NOW()),
    (401, 'user_login_abnormal', '异地登录', '账号在陌生设备登录的安全提醒', 15, true, NOW()),
    (402, 'user_profile_updated', '资料变更', '用户手机号、邮箱等信息修改后通知', 16, true, NOW()),
    (403, 'user_permission_changed', '权限变更', '账号角色或功能权限调整通知', 17, true, NOW());
ALTER TABLE internal_message_categories AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM internal_message_categories);

-- ----------------------------
-- 语言注册表（与 postgresql-demo-data.sql 对齐；mall_*_translations 依赖此表存在对应行）
-- ----------------------------
INSERT INTO sys_languages (language_code, language_name, native_name, is_default, is_enabled, sort_order, created_at, updated_at)
VALUES
    ('zh-CN', '简体中文', '简体中文', true, true, 1, NOW(), NOW()),
    ('en-US', 'English (US)', 'English (US)', false, true, 2, NOW(), NOW());
ALTER TABLE sys_languages AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_languages);

-- ====================================================================
-- 商城 · 目录域（catalog）种子数据
-- 说明：演示用。所有金额为最小货币单位（分）；翻译表成对写入 zh-CN + en-US。
-- ====================================================================

-- ----------------------------
-- 类目（树形：1 顶层 → 2 个二级）
-- ----------------------------
INSERT INTO mall_categories (id, parent_id, path, depth, sort_order, image_url, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, NULL, '/',      0, 1, '/demo-images/cat-1.png', NOW(), NOW(), 1, 1, NULL),
    (2, 1,    '/1/',    1, 1, '/demo-images/cat-2.png', NOW(), NOW(), 1, 1, NULL),
    (3, 1,    '/1/',    1, 2, '/demo-images/cat-3.png', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_categories AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_categories);

-- 类目翻译
INSERT INTO mall_category_translations (category_id, language_code, name, slug, description, full_path, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'zh-CN', '数码电子', 'digital', '数码电子产品类目', '/1',    NOW(), NOW(), 1, 1, NULL),
    (1, 'en-US', 'Digital',  'digital', 'Digital electronics', '/1', NOW(), NOW(), 1, 1, NULL),
    (2, 'zh-CN', '手机',     'phones',  '移动电话类目', '/1/2',     NOW(), NOW(), 1, 1, NULL),
    (2, 'en-US', 'Phones',   'phones',  'Mobile phones', '/1/2',    NOW(), NOW(), 1, 1, NULL),
    (3, 'zh-CN', '配件',     'accessories', '数码配件类目', '/1/3',  NOW(), NOW(), 1, 1, NULL),
    (3, 'en-US', 'Accessories','accessories','Digital accessories','/1/3', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_category_translations AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_category_translations);

-- ----------------------------
-- 品牌
-- ----------------------------
INSERT INTO mall_brands (id, logo_url, sort_order, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, '/demo-images/brand-1.png', 1, NOW(), NOW(), 1, 1, NULL),
    (2, '/demo-images/brand-2.png', 2, NOW(), NOW(), 1, 1, NULL),
    (3, '/demo-images/brand-3.png', 3, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_brands AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_brands);

-- 品牌翻译
INSERT INTO mall_brand_translations (brand_id, language_code, name, slug, description, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'zh-CN', '北辰',   'beichen',   '国产数码品牌',   NOW(), NOW(), 1, 1, NULL),
    (1, 'en-US', 'Beichen','beichen',   'Domestic digital brand', NOW(), NOW(), 1, 1, NULL),
    (2, 'zh-CN', '远洋',   'yuanyang',  '海洋电子品牌',   NOW(), NOW(), 1, 1, NULL),
    (2, 'en-US', 'Yuanyang','yuanyang', 'Marine electronics brand', NOW(), NOW(), 1, 1, NULL),
    (3, 'zh-CN', '极光',   'jiguang',   '消费配件品牌',   NOW(), NOW(), 1, 1, NULL),
    (3, 'en-US', 'Aurora', 'jiguang',   'Consumer accessory brand', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_brand_translations AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_brand_translations);

-- ----------------------------
-- 商品（挂类目+品牌，状态 ACTIVE）
-- ----------------------------
INSERT INTO mall_products (id, status, category_id, brand_id, sort_order, image_url, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'PRODUCT_STATUS_ACTIVE', 2, 1, 1, '/demo-images/prod-1.png', NOW(), NOW(), 1, 1, NULL),
    (2, 'PRODUCT_STATUS_ACTIVE', 2, 2, 2, '/demo-images/prod-2.png', NOW(), NOW(), 1, 1, NULL),
    (3, 'PRODUCT_STATUS_ACTIVE', 2, 1, 3, '/demo-images/prod-3.png', NOW(), NOW(), 1, 1, NULL),
    (4, 'PRODUCT_STATUS_ACTIVE', 3, 3, 1, '/demo-images/prod-4.png', NOW(), NOW(), 1, 1, NULL),
    (5, 'PRODUCT_STATUS_ACTIVE', 3, 3, 2, '/demo-images/prod-5.png', NOW(), NOW(), 1, 1, NULL),
    (6, 'PRODUCT_STATUS_ACTIVE', 3, 1, 3, '/demo-images/prod-6.png', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_products AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_products);

-- 商品翻译
INSERT INTO mall_product_translations (product_id, language_code, name, slug, short_description, long_description, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'zh-CN', '北辰 X1 智能手机', 'beichen-x1', '入门级智能手机', '北辰 X1 是一款面向入门市场的智能手机，配备基础通信与娱乐功能。', NOW(), NOW(), 1, 1, NULL),
    (1, 'en-US', 'Beichen X1 Smartphone', 'beichen-x1', 'Entry-level smartphone', 'Beichen X1 is an entry-level smartphone with basic communication and entertainment features.', NOW(), NOW(), 1, 1, NULL),
    (2, 'zh-CN', '远洋 M2 平板', 'yuanyang-m2', '中端娱乐平板', '远洋 M2 平板定位于影音娱乐，配备较大屏幕与长续航。', NOW(), NOW(), 1, 1, NULL),
    (2, 'en-US', 'Yuanyang M2 Tablet', 'yuanyang-m2', 'Mid-range entertainment tablet', 'Yuanyang M2 tablet is positioned for audio-visual entertainment with a large screen and long battery life.', NOW(), NOW(), 1, 1, NULL),
    (3, 'zh-CN', '北辰 T3 老人机', 'beichen-t3', '大按键老人机', '北辰 T3 面向老年用户，大按键大字体，操作简便。', NOW(), NOW(), 1, 1, NULL),
    (3, 'en-US', 'Beichen T3 Feature Phone', 'beichen-t3', 'Large-button feature phone', 'Beichen T3 targets elderly users with large buttons and fonts for ease of use.', NOW(), NOW(), 1, 1, NULL),
    (4, 'zh-CN', '极光 USB-C 数据线', 'aurora-usb-c-cable', 'USB-C 编织数据线', '极光 USB-C 数据线采用编织外被，支持快充与数据传输。', NOW(), NOW(), 1, 1, NULL),
    (4, 'en-US', 'Aurora USB-C Cable', 'aurora-usb-c-cable', 'USB-C braided cable', 'Aurora USB-C cable features a braided jacket and supports fast charging and data transfer.', NOW(), NOW(), 1, 1, NULL),
    (5, 'zh-CN', '极光无线充电板', 'aurora-wireless-charger', 'Qi 协议无线充电板', '极光无线充电板兼容 Qi 协议，提供基础无线充电体验。', NOW(), NOW(), 1, 1, NULL),
    (5, 'en-US', 'Aurora Wireless Charger', 'aurora-wireless-charger', 'Qi-compatible wireless charging pad', 'Aurora wireless charging pad is Qi-compatible and provides basic wireless charging.', NOW(), NOW(), 1, 1, NULL),
    (6, 'zh-CN', '北辰蓝牙耳机', 'beichen-bluetooth-headset', '入门蓝牙耳机', '北辰蓝牙耳机提供基础无线音频连接。', NOW(), NOW(), 1, 1, NULL),
    (6, 'en-US', 'Beichen Bluetooth Headset', 'beichen-bluetooth-headset', 'Entry-level Bluetooth headset', 'Beichen Bluetooth headset provides basic wireless audio connectivity.', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_product_translations AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_product_translations);

-- ----------------------------
-- 商品属性（颜色 / 容量）+ 属性值 + 翻译
-- ----------------------------
INSERT INTO mall_product_attributes (id, sort_order, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, NOW(), NOW(), 1, 1, NULL),
    (2, 2, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_product_attributes AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_product_attributes);

-- 属性翻译
INSERT INTO mall_product_attribute_translations (attribute_id, language_code, name, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'zh-CN', '颜色', NOW(), NOW(), 1, 1, NULL),
    (1, 'en-US', 'Color', NOW(), NOW(), 1, 1, NULL),
    (2, 'zh-CN', '容量', NOW(), NOW(), 1, 1, NULL),
    (2, 'en-US', 'Capacity', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_product_attribute_translations AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_product_attribute_translations);

-- 属性值（属性1=颜色: 黑/白；属性2=容量: 64/128）
INSERT INTO mall_product_attribute_values (id, attribute_id, sort_order, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (2, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (3, 2, 1, NOW(), NOW(), 1, 1, NULL),
    (4, 2, 2, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_product_attribute_values AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_product_attribute_values);

-- 属性值翻译
INSERT INTO mall_product_attribute_value_translations (value_id, language_code, display_name, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'zh-CN', '黑色', NOW(), NOW(), 1, 1, NULL),
    (1, 'en-US', 'Black', NOW(), NOW(), 1, 1, NULL),
    (2, 'zh-CN', '白色', NOW(), NOW(), 1, 1, NULL),
    (2, 'en-US', 'White', NOW(), NOW(), 1, 1, NULL),
    (3, 'zh-CN', '64GB', NOW(), NOW(), 1, 1, NULL),
    (3, 'en-US', '64GB', NOW(), NOW(), 1, 1, NULL),
    (4, 'zh-CN', '128GB', NOW(), NOW(), 1, 1, NULL),
    (4, 'en-US', '128GB', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_product_attribute_value_translations AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_product_attribute_value_translations);

-- ----------------------------
-- SKU（每商品 2 个，颜色×容量组合）
-- ----------------------------
INSERT INTO mall_skus (id, product_id, sku_code, stock_qty, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, 'BC-X1-BLK-64',  100, NOW(), NOW(), 1, 1, NULL),
    (2, 1, 'BC-X1-WHT-128', 100, NOW(), NOW(), 1, 1, NULL),
    (3, 2, 'YY-M2-BLK-64',  100, NOW(), NOW(), 1, 1, NULL),
    (4, 2, 'YY-M2-WHT-128', 100, NOW(), NOW(), 1, 1, NULL),
    (5, 3, 'BC-T3-BLK-64',  100, NOW(), NOW(), 1, 1, NULL),
    (6, 3, 'BC-T3-WHT-128', 100, NOW(), NOW(), 1, 1, NULL),
    (7, 4, 'AU-CBL-BLK-64', 100, NOW(), NOW(), 1, 1, NULL),
    (8, 4, 'AU-CBL-WHT-128',100, NOW(), NOW(), 1, 1, NULL),
    (9, 5, 'AU-WC-BLK-64',  100, NOW(), NOW(), 1, 1, NULL),
    (10, 5, 'AU-WC-WHT-128',100, NOW(), NOW(), 1, 1, NULL),
    (11, 6, 'BC-BT-BLK-64', 100, NOW(), NOW(), 1, 1, NULL),
    (12, 6, 'BC-BT-WHT-128',100, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_skus AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_skus);

-- SKU 价格（每 SKU 一行 CNY，amount 为字符串）
INSERT INTO mall_sku_prices (sku_id, currency, amount, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 'CNY', '199900',  NOW(), NOW(), 1, 1, NULL),
    (2, 'CNY', '229900',  NOW(), NOW(), 1, 1, NULL),
    (3, 'CNY', '159900',  NOW(), NOW(), 1, 1, NULL),
    (4, 'CNY', '189900',  NOW(), NOW(), 1, 1, NULL),
    (5, 'CNY', '99900',   NOW(), NOW(), 1, 1, NULL),
    (6, 'CNY', '109900',  NOW(), NOW(), 1, 1, NULL),
    (7, 'CNY', '1900',    NOW(), NOW(), 1, 1, NULL),
    (8, 'CNY', '1900',    NOW(), NOW(), 1, 1, NULL),
    (9, 'CNY', '9900',    NOW(), NOW(), 1, 1, NULL),
    (10, 'CNY', '9900',   NOW(), NOW(), 1, 1, NULL),
    (11, 'CNY', '5900',   NOW(), NOW(), 1, 1, NULL),
    (12, 'CNY', '5900',   NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_sku_prices AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_sku_prices);

-- SKU 属性组合（每 SKU 关联颜色值+容量值）
INSERT INTO mall_sku_attribute_combinations (sku_id, attribute_id, attribute_value_id, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (1, 2, 3, NOW(), NOW(), 1, 1, NULL),
    (2, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (2, 2, 4, NOW(), NOW(), 1, 1, NULL),
    (3, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (3, 2, 3, NOW(), NOW(), 1, 1, NULL),
    (4, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (4, 2, 4, NOW(), NOW(), 1, 1, NULL),
    (5, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (5, 2, 3, NOW(), NOW(), 1, 1, NULL),
    (6, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (6, 2, 4, NOW(), NOW(), 1, 1, NULL),
    (7, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (7, 2, 3, NOW(), NOW(), 1, 1, NULL),
    (8, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (8, 2, 4, NOW(), NOW(), 1, 1, NULL),
    (9, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (9, 2, 3, NOW(), NOW(), 1, 1, NULL),
    (10, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (10, 2, 4, NOW(), NOW(), 1, 1, NULL),
    (11, 1, 1, NOW(), NOW(), 1, 1, NULL),
    (11, 2, 3, NOW(), NOW(), 1, 1, NULL),
    (12, 1, 2, NOW(), NOW(), 1, 1, NULL),
    (12, 2, 4, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_sku_attribute_combinations AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_sku_attribute_combinations);

-- ====================================================================
-- 商城 · 顾客用户（交易域 user_id 依赖）
-- 注：shopper(id=3) 仅作演示用普通顾客，凭证密码同为 admin。
-- sys_roles 由运行时 createDefaultRoles 生成，此处不重复 seed。
-- ====================================================================
INSERT INTO sys_users (id, tenant_id, username, nickname, realname, email, gender, created_at)
VALUES (3, 1, 'shopper', '测试顾客', '李顾客', 'shopper@company.com', 'MALE', NOW());
ALTER TABLE sys_users AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_users);

INSERT INTO sys_user_credentials (tenant_id, user_id, identity_type, identifier, credential_type, credential, status, is_primary, created_at)
VALUES
    (1, 3, 'USERNAME', 'shopper', 'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', 'ENABLED', true, NOW()),
    (1, 3, 'EMAIL', 'shopper@company.com', 'PASSWORD_HASH', '$2a$10$yajZDX20Y40FkG0Bu4N19eXNqRizez/S9fK63.JxGkfLq.RoNKR/a', 'ENABLED', false, NOW());
ALTER TABLE sys_user_credentials AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_user_credentials);

INSERT INTO sys_memberships (id, tenant_id, user_id, org_unit_id, position_id, role_id, is_primary, status)
VALUES (3, 1, 3, null, null, 2, true, 'ACTIVE');
ALTER TABLE sys_memberships AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_memberships);

INSERT INTO sys_membership_roles (id, membership_id, tenant_id, role_id, is_primary, status)
VALUES (3, 3, 1, 2, true, 'ACTIVE');
ALTER TABLE sys_membership_roles AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM sys_membership_roles);

-- ====================================================================
-- 商城 · 交易域种子数据（演示用）
-- UserPrivacy 行级隔离在应用层经 viewer context 生效，直 SQL 不受影响。
-- tenant_id 显式写 1；所有金额单位为分。
-- ====================================================================

-- 购物车（shopper 的购物车）
INSERT INTO mall_carts (id, user_id, tenant_id, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES (1, 3, 1, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_carts AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_carts);

-- 购物车项
INSERT INTO mall_cart_items (id, cart_id, sku_id, quantity, tenant_id, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, 7,  2, 1, NOW(), NOW(), 1, 1, NULL),
    (2, 1, 11, 1, 1, NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_cart_items AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_cart_items);

-- 订单（三种状态各一：待支付 / 已支付 / 已履约）
INSERT INTO mall_orders (id, user_id, tenant_id, total_amount, status, recipient_name, recipient_phone, shipping_address, business_ref_id, idempotency_key, currency, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 3, 1, 3800,  'PENDING_PAYMENT', '张三', '13800000001', '北京市朝阳区演示街1号', 'ORD-DEMO-1001', 'idem-order-1001', 'CNY', NOW(), NOW(), 1, 1, NULL),
    (2, 3, 1, 5900,  'PAID',            '李四', '13800000002', '上海市浦东新区演示路2号', 'ORD-DEMO-1002', 'idem-order-1002', 'CNY', NOW(), NOW(), 1, 1, NULL),
    (3, 3, 1, 9900,  'FULFILLED',       '王五', '13800000003', '广州市天河区演示道3号', 'ORD-DEMO-1003', 'idem-order-1003', 'CNY', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_orders AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_orders);

-- 订单项
INSERT INTO mall_order_items (id, order_id, sku_id, sku_snapshot, quantity, unit_price, subtotal, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, 7,  '{"sku":"AU-CBL-BLK-64","name":"Aurora USB-C Cable"}', 2, 1900, 3800,  NOW(), NOW(), 1, 1, NULL),
    (2, 2, 11, '{"sku":"BC-BT-BLK-64","name":"Beichen Bluetooth Headset"}', 1, 5900, 5900, NOW(), NOW(), 1, 1, NULL),
    (3, 3, 9,  '{"sku":"AU-WC-BLK-64","name":"Aurora Wireless Charger"}', 1, 9900, 9900,  NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_order_items AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_order_items);

-- 支付流水（仅对应已支付订单）
INSERT INTO mall_payment_transactions (id, order_id, user_id, tenant_id, amount, status, business_ref_id, idempotency_key, currency, payment_method, business_type, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 2, 3, 1, 5900, 'SUCCEEDED', 'PAY-DEMO-2001', 'idem-pay-2001', 'CNY', 'ALIPAY', 'BUSINESS_TYPE_CONSUME', NOW(), NOW(), 1, 1, NULL),
    (2, 3, 3, 1, 9900, 'SUCCEEDED', 'PAY-DEMO-2002', 'idem-pay-2002', 'CNY', 'WECHAT',  'BUSINESS_TYPE_CONSUME', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_payment_transactions AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_payment_transactions);

-- 退款（对应第二笔支付，状态 PENDING，演示退款流程）
INSERT INTO mall_payment_refunds (id, transaction_id, tenant_id, amount, status, business_ref_id, idempotency_key, currency, created_at, updated_at, created_by, updated_by, deleted_by)
VALUES
    (1, 1, 1, 5900, 'PENDING', 'REF-DEMO-3001', 'idem-refund-3001', 'CNY', NOW(), NOW(), 1, 1, NULL);
ALTER TABLE mall_payment_refunds AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM mall_payment_refunds);

-- 提交事务+恢复外键检查
COMMIT;
SET FOREIGN_KEY_CHECKS = 1;

SELECT '测试数据初始化完成' AS result;