/**
 * API Composables 索引文件
 * 导出所有业务模块的 hooks、fetch 方法以及工具函数
 */

// 认证相关
export * from './auth';

// 用户资料
export * from './user-profile';

// 文件传输
export * from './file-transfer';

// 目录（只读：List + Get）
export * from './brand';
export * from './category';
export * from './product';
export * from './product-attribute';
export * from './product-attribute-value';
export * from './sku';
export * from './sku-price';
export * from './sku-attribute-combination';

// 交易（购物车 / 订单 / 支付 / 退款）
export * from './cart';
export * from './cart-item';
export * from './wishlist';
export * from './order';
export * from './payment-transaction';
export * from './payment-refund';
export * from './shipping-address';
export * from './shipment';
export * from './message';

// 优惠券（仅本人 List/Get + Quote 试算；领券中心浏览可领模板 + Claim 领取）
export * from './user-coupon';
export * from './coupon-template';

// 商品评论 + 点赞/计数
export * from './comment';
export * from './interaction';

// 发票查看（仅本人，只读）
export * from './invoice';
