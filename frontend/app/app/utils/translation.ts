import { getCurrentLocale } from '@/utils/locale';

/**
 * 多语言实体翻译取值。
 *
 * 后端所有可翻译实体（Product / Category / Sku / Coupon …）都返回
 * `translations: [{ name, languageCode, ... }]`，页面要先挑出当前语言那一条。
 * 这段挑拣逻辑此前在 10 个页面/组件里各抄了一份副本，
 * 副本之间已经开始分叉（有的缺失语言时退回第一条，有的直接显示 '—'）。
 * 统一到此：缺失语言时退回第一条，让内容可见优先于语言纯净。
 */
export function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined | null,
  locale: string = getCurrentLocale(),
): T | undefined {
  if (!translations || translations.length === 0) return undefined;
  return translations.find((tr) => tr.languageCode === locale) ?? translations[0];
}

/** 名称取当前语言译文，没有译文时给一个占位符而不是空白。 */
export function translationName(
  translations: Array<{ name?: string; languageCode?: string }> | undefined | null,
  fallback = '—',
  locale: string = getCurrentLocale(),
): string {
  return pickTranslation(translations, locale)?.name || fallback;
}
