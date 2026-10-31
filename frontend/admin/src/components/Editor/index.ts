/**
 * Editor 薄包装：re-export @go-wind/editor-vue 并注入应用 env。
 * 页面 import 路径不变（@/components/Editor），内部已切库消费（同 go-wind-admin vue-element 端）。
 * 图片上传 env 在 App.vue setup 里 provideEditorEnv 注入（组件级 provide 须在 setup 内）。
 */
export { Editor, EditorType } from "@go-wind/editor-vue";
export {
  provideEditorEnv,
  useEditorDark,
  setEditorLocale,
  isDarkNow as isDarkMode,
} from "@go-wind/editor-vue";
export type { EditorEnv } from "@go-wind/editor-vue";
import "@go-wind/editor-vue/style.css";

import { unref, watchEffect } from "vue";
import { i18n } from "@/core/i18n";
import { setEditorLocale } from "@go-wind/editor-vue";

// 编辑器文案跟随界面语言：包内置 zh-CN/en-US 两套，界面语言不在两档内时回落 en-US
watchEffect(() => {
  const current = String(unref(i18n.global.locale) || "zh-CN");
  setEditorLocale(current.toLowerCase().startsWith("zh") ? "zh-CN" : "en-US");
});
