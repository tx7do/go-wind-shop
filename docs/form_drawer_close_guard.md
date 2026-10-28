# 表单抽屉关闭守卫规范（Form Drawer Close Guard）

> 三端唯一的表单抽屉关闭行为权威。改遮罩/ESC/关闭确认相关代码前先读本文。
> 2026-10-21 定稿并三端落地。

## 1. 行为规范

**表单类 Drawer（新增 / 编辑，即内嵌可编辑表单的抽屉）**：`maskClosable = true`（允许点击遮罩关闭），但必须带脏数据判断：

1. 表单**没有修改内容（无脏数据）**：点击遮罩 → 直接关闭 drawer；
2. 表单**已经填写 / 修改（存在脏数据）**：点击遮罩 → 不关闭，弹出确认框。

> 文案：表单内容已修改，是否放弃本次编辑？【取消】【确定放弃】

**配套统一**：ESC 快捷键、标题栏关闭按钮（X）、底部取消按钮与遮罩点击行为**一模一样**——四条关闭路径全部走同一守卫，有脏数据一律先确认。

| 关闭路径 | 无脏数据 | 有脏数据 |
|---|---|---|
| 点击遮罩 | 直接关闭 | 确认框，确定后关闭 |
| ESC | 直接关闭 | 确认框，确定后关闭 |
| 标题栏 X | 直接关闭 | 确认框，确定后关闭 |
| 底部取消按钮 | 直接关闭 | 确认框，确定后关闭 |

**推论（禁改项）**：

- 表单抽屉**禁止**再设 `maskClosable: false` / `close-on-click-modal: false` / `closeOnClickModal: false`——遮罩是合法关闭入口，安全由守卫保证；
- 表单抽屉保持 `keyboard: true` / `close-on-press-escape: true`（各端默认值，不显式关闭）；
- **提交成功后的关闭不走守卫**（提交本身就是"保存"意图，成功后直接关闭并重置脏状态）。

### 脏数据定义

抽屉打开并完成初始化（新增默认值填充 / 编辑回填）之后，**用户**改动过任何受控内容即为脏；程序化回填不算脏。提交成功后脏状态重置。

三端判定机制（语义差异仅一处，方向均为保守侧）：

| 端 | 机制 | 「改了又改回」 |
|---|---|---|
| react | `onValuesChange` 置脏（仅用户修改触发） | 视为脏（保守） |
| vue-element | 打开填充完成后 JSON 基线快照，关闭时对比 | 视为不脏 |
| vue-vben | 同 vue-element（`formApi.getValues()` 快照对比） | 视为不脏 |

### 适用与例外

**接入**：所有内嵌可编辑表单的抽屉（新增/编辑/带表单的工具抽屉）。

**例外（不接入）**：

- **查看/详情类抽屉**（无表单或纯只读：收件箱详情、五类审计日志详情、脚本运行日志等）——没有可丢失的编辑态；
- **偏好设置面板**（PreferencesPanel）——设置即时生效，无"未保存"概念；
- **画布编辑器**（AI Agent 编排画布等）——非表单形态，自身另有保存语义；
- **对话框/Modal 表单**——本规范限 Drawer；Modal 表单如需同样守卫另行立项。

> 端内特有：vue-element 有 40 个抽屉页使用**自定义 footer**（footer 在页面内而非 ProModal 默认），这些页面的取消按钮需手动路由守卫；带表单但内容在 `formData` 之外的编辑态（如勾选树 `setCheckedKeys`）必须纳入脏判定（见 §2.2 getModel）。

## 2. 三端实现约定

### 2.1 react（行为基准）

共享 Hook：`src/components/common/use-drawer-close-guard.ts`

```tsx
const { guardClose, handleValuesChange } = useDrawerCloseGuard(open, onClose);
```

ProComponents `DrawerForm` 页面接入四件事：

```tsx
<DrawerForm
  open={open}
  onValuesChange={handleValuesChange}          // ① 用户修改置脏
  drawerProps={{
    destroyOnHidden: true,                     // ② 关闭销毁重建，重开即净表单
    size: 600,
    onClose: guardClose,                       // ③ 遮罩/ESC/X 统一守卫入口
  }}
  submitter={{
    ...,
    resetButtonProps: { onClick: guardClose }, // ④ 取消按钮同样守卫
  }}
>
```

- **删掉** `onOpenChange` 里的无条件关闭/`resetFields`：DrawerForm 内部收到遮罩关闭时先 `setOpen(false)`（受控下仅回调 `onOpenChange(false)`，不改渲染），再调 `drawerProps.onClose`——守卫单点必须只在 `drawerProps.onClose`；在 `onOpenChange(false)` 里关抽屉/重置表单会绕过守卫或误清用户输入。
- 确认框用 `App.useApp()` 的 `modal.confirm`（勿用静态 `Modal.confirm`，拿不到主题上下文）。
- 自绘 `<Drawer>` + `Form` 的页面：`Drawer onClose={guardClose}` + `Form onValuesChange={handleValuesChange}`，同一 hook。

### 2.2 vue-element

两层改造：

1. **ProModal**（`src/components/Pro/ProModal/index.vue`）新增可选 prop `beforeClose?: () => boolean | Promise<boolean>`（返回 true 才允许关闭）：
   - 绑定容器 `before-close`（ElDrawer/ElDialog 的遮罩/ESC/X 统一入口）；
   - 底部默认取消按钮同路由；
   - `visible` 已为 false 时直接放行（EP close 事件去重，防二次弹确认）；
   - 提交成功路径 `visible = false` 直关，不经守卫。
2. **共享组合式** `src/components/Pro/composables/use-drawer-close-guard.ts`：

```ts
const guard = useDrawerCloseGuard(getModel);   // getModel: () => 编辑态对象
await guard.syncBaseline();                    // 打开填充完成后调用（useDrawerForm 自动调）
guard.beforeClose(done);                       // 接 ProModal :before-close
guard.requestClose(close);                     // 自定义 footer 的取消按钮用这个
```

- `useDrawerForm` 内建 guard，`open()` 末尾自动 `syncBaseline()`，页面从 `drawer.guard` 取；
- 自包含抽屉页（不用 useDrawerForm 的）自行创建 guard，并在自身 `open()` 末尾 `syncBaseline()`；
- 编辑态超出 `formData` 的（勾选树等），getModel 里合并：`() => ({ ...formData, menuIds: menuTreeRef.value?.getCheckedKeys() })`；
- 原来设了 `close-on-click-modal: false` 的表单抽屉一律删除该配置。

### 2.3 vue-vben

1. **框架**（in-repo fork）：`DrawerApi.close()` 的 `onBeforeClose` 升级支持 `boolean | Promise<boolean>`——同步 false 阻止关闭，Promise 以 resolve 值决定。遮罩（interactOutside）、ESC（escapeKeyDown）、X（SheetClose）、底部取消（onCancel→close）四路径本就统一汇入 `close()`，无需逐路径拦截。
2. **共享组合式** `src/composables/use-form-close-guard.ts`：

```ts
const guard = useFormCloseGuard(baseFormApi);
const [Drawer, drawerApi] = useVbenDrawer({
  onBeforeClose: guard.onBeforeClose,   // 遮罩/ESC/X/取消 全走这里
  onCancel() { drawerApi.close(); },
  async onConfirm() {
    // ...提交成功后：
    guard.markClean();                  // 先清脏再关，否则提交成功还会弹确认
    drawerApi.close();
  },
  async onOpenChange(isOpen) {
    if (isOpen) {
      // ...setValues 填充完成后：
      await guard.syncBaseline();
    }
  },
});
```

- 注意：vben 表单的 `handleValuesChange` 连程序化 `setValues` 都会触发（deep watch + 300ms debounce），所以 vben 不用事件置脏，用快照对比；
- 确认框用 `ant-design-vue` 的 `Modal.confirm`（与 TableActions 同惯例）。

## 3. i18n 文案

| key | zh-CN | en-US |
|---|---|---|
| `common.drawer.unsavedMessage`（vben 为 `ui.drawer.unsavedMessage`） | 表单内容已修改，是否放弃本次编辑？ | The form has been modified. Do you want to discard this edit? |
| `common.drawer.discardButton`（vben 为 `ui.drawer.discardButton`） | 确定放弃 | Discard |
| 取消按钮 | 复用既有 `common.button.cancel` | — |

## 4. 验收口径

每端至少抽一个表单抽屉实测 4 条路径 × 2 种脏状态：

1. 打开新增抽屉不填写 → 遮罩/ESC → 直接关闭；
2. 填写任一字段 → 遮罩/ESC/X/取消 → 弹确认框；点「取消」抽屉保持且内容还在；点「确定放弃」关闭；
3. 编辑抽屉打开后不动 → 直接关闭（回填不误判为脏）；
4. 提交成功 → 直接关闭、不弹确认。
