<template>
  <div class="spec-editor">
    <!-- 规格行：上下移排序（顺序即组合笛卡尔积顺序） -->
    <div class="spec-editor__rows">
      <div v-for="(spec, index) in specs" :key="spec.attributeId" class="spec-row">
        <div class="spec-row__name">
          <span class="spec-row__name-text">{{ spec.attributeName }}</span>
          <span class="spec-row__name-order">{{ index + 1 }}</span>
        </div>

        <!-- 规格值：字典值多选（chips 视觉），勾选顺序即组合生成顺序 -->
        <ElSelect
          :model-value="spec.valueIds"
          multiple
          filterable
          :placeholder="$t('pages.mall.product.salesInfo.specValuePlaceholder')"
          class="spec-row__values"
          @update:model-value="(ids: number[]) => handleValuesChange(spec, ids)"
        >
          <ElOption
            v-for="v in valuesOf(spec.attributeId)"
            :key="v.id"
            :value="v.id"
            :label="v.name"
          />
        </ElSelect>

        <!-- 快捷新增字典值：创建后自动勾选（不动已选顺序） -->
        <ElButton
          size="small"
          :disabled="creatingValueFor === spec.attributeId"
          @click="handleAddValue(spec)"
        >
          <IconifyIcon icon="lucide:plus" class="mr-1 h-3.5 w-3.5" />
          {{ $t("pages.mall.product.salesInfo.addValue") }}
        </ElButton>

        <ElButton size="small" text :disabled="index === 0" @click="handleMove(index, -1)">
          <IconifyIcon icon="lucide:arrow-up" />
        </ElButton>
        <ElButton
          size="small"
          text
          :disabled="index === specs.length - 1"
          @click="handleMove(index, 1)"
        >
          <IconifyIcon icon="lucide:arrow-down" />
        </ElButton>

        <ElButton size="small" text type="danger" @click="handleRemoveSpec(index)">
          {{ $t("common.button.delete") }}
        </ElButton>
      </div>
    </div>

    <div v-if="specs.length === 0" class="spec-editor__empty">
      {{ $t("pages.mall.product.salesInfo.specEmptyTip") }}
    </div>

    <ElSelect
      :model-value="undefined"
      :placeholder="$t('pages.mall.product.salesInfo.addSpec')"
      filterable
      class="spec-editor__add"
      @update:model-value="handleAddSpec"
    >
      <ElOption
        v-for="attr in selectableAttributes"
        :key="attr.id"
        :value="attr.id"
        :label="attr.name"
      />
    </ElSelect>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Icon as IconifyIcon } from "@iconify/vue";

import {
  useCreateProductAttributeValue,
  useCreateProductAttributeValueTranslation,
  fetchListProductAttributeValues,
} from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t, useI18n } from "@/core/i18n";
import type { SpecRow } from "../composables/use-sku-matrix";

/**
 * 销售规格编辑器：从全局属性字典选属性成行，行内从该属性的字典值中勾选。
 *
 * - chips 上移除只是「本商品不再使用该值」，不动字典；
 * - 「新增值」会在字典里真实创建（属性值 + 当前语言翻译），创建后按
 *   attributeId + displayName 反查回读 ID（Create 返回 Empty），自动勾选。
 */
defineOptions({ name: "MallProductSpecEditor" });

const props = defineProps<{
  specs: SpecRow[];
  /** 预载字典：{ id, name, values: [{ id, name }] }[] */
  attributes: { id: number; name: string; values: { id: number; name: string }[] }[];
}>();

const emit = defineEmits<{ change: [] }>();

const { mutateAsync: createAttributeValue } = useCreateProductAttributeValue();
const { mutateAsync: createAttributeValueTranslation } =
  useCreateProductAttributeValueTranslation();
const { locale } = useI18n();

const creatingValueFor = ref<number>();

const selectableAttributes = computed(() =>
  props.attributes.filter((a) => !props.specs.some((s) => s.attributeId === a.id))
);

function valuesOf(attributeId: number) {
  return props.attributes.find((a) => a.id === attributeId)?.values ?? [];
}

function handleAddSpec(attributeId: number | undefined) {
  if (!attributeId) return;
  const attr = props.attributes.find((a) => a.id === attributeId);
  if (!attr) return;
  props.specs.push({ attributeId: attr.id, attributeName: attr.name, valueIds: [] });
  emit("change");
}

function handleRemoveSpec(index: number) {
  props.specs.splice(index, 1);
  emit("change");
}

function handleMove(index: number, dir: -1 | 1) {
  const target = index + dir;
  if (target < 0 || target >= props.specs.length) return;
  const [row] = props.specs.splice(index, 1);
  props.specs.splice(target, 0, row);
  emit("change");
}

function handleValuesChange(spec: SpecRow, ids: number[]) {
  spec.valueIds = ids;
  emit("change");
}

async function handleAddValue(spec: SpecRow) {
  try {
    const { value } = await ElMessageBox.prompt(
      $t("pages.mall.product.salesInfo.newValuePrompt", [spec.attributeName]),
      $t("pages.mall.product.salesInfo.addValue"),
      { inputPattern: /\S+/, inputErrorMessage: $t("common.validation.required") }
    );
    const displayName = String(value ?? "").trim();
    if (!displayName) return;

    creatingValueFor.value = spec.attributeId;
    // 属性值 + 当前语言翻译（Create 返回 Empty：随后按名反查 ID）
    await createAttributeValue({ attributeId: spec.attributeId } as Record<string, any>);
    const q = new PaginationQuery({
      paging: { page: 1, pageSize: 200 },
      formValues: { attributeId: spec.attributeId },
    });
    const resp: any = await fetchListProductAttributeValues(q);
    const items = (resp?.items ?? []) as any[];
    const knownIds = new Set(valuesOf(spec.attributeId).map((v) => v.id));
    const hit = items.find(
      (it) =>
        !knownIds.has(Number(it.id)) &&
        (it.translations ?? []).some((t: any) => t.displayName === displayName)
    );
    if (!hit) throw new Error("created value not found");
    const valueId = Number(hit.id);

    await createAttributeValueTranslation({
      data: {
        valueId: String(valueId),
        languageCode: locale.value || "zh-CN",
        displayName,
      },
    });

    // 本地字典补一条并自动勾选（不重拉整表）
    valuesOf(spec.attributeId).push({ id: valueId, name: displayName });
    spec.valueIds.push(valueId);
    emit("change");
    ElMessage.success($t("common.notification.createSuccess"));
  } catch (e) {
    if (e !== "cancel" && e !== false) {
      ElMessage.error($t("common.notification.createFailed"));
    }
  } finally {
    creatingValueFor.value = undefined;
  }
}
</script>

<style lang="scss" scoped>
.spec-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;

  &__rows {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  &__empty {
    padding: 8px 0;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  &__add {
    width: 240px;
  }
}

.spec-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: var(--el-fill-color-blank);

  &__drag {
    flex-shrink: 0;
    cursor: grab;
    color: var(--el-text-color-placeholder);

    &:hover {
      color: var(--el-color-primary);
    }
  }

  &__name {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 120px;

    &-text {
      font-weight: 600;
    }

    &-order {
      padding: 0 6px;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      background: var(--el-fill-color-light);
      border-radius: 4px;
    }
  }

  &__values {
    flex: 1;
  }
}
</style>
