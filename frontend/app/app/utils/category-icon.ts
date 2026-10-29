/**
 * 分类图标兜底映射：分类未配置图片时按 id 取模取图标，纯视觉用途。
 * QuickNav 与分类索引页共用，保证同一分类在各处图标一致。
 */
export const CATEGORY_ICON_SET = [
  'carbon:shopping-bag',
  'carbon:product',
  'carbon:category',
  'carbon:settings',
  'carbon:home',
  'carbon:headset',
  'carbon:shopping-cart',
  'carbon:gift',
] as const

export function categoryIconFor(id?: number): string {
  const n = id ?? 0
  return CATEGORY_ICON_SET[Math.abs(n) % CATEGORY_ICON_SET.length]
}
