/**
 * 多语言翻译字段配置（供 TranslationTabs 及各翻译抽屉共用）
 */
export interface TranslationFieldConfig {
  prop: string;
  label: string;
  type: "input" | "textarea" | "input-number";
}
