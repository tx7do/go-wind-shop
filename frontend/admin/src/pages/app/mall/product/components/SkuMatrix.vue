<template>
  <div class="sku-matrix">
    <!-- 批量设置条：一键应用到全部行 -->
    <div class="sku-matrix__toolbar">
      <span class="sku-matrix__toolbar-label">
        {{ $t("pages.mall.product.salesInfo.batchApply") }}
      </span>
      <ElInputNumber
        v-model="batchPrice"
        :min="0"
        :precision="2"
        :controls="false"
        :placeholder="$t('pages.mall.product.salesInfo.batchPrice')"
        class="sku-matrix__batch-input"
      />
      <ElInputNumber
        v-model="batchStock"
        :min="0"
        :controls="false"
        :placeholder="$t('pages.mall.product.salesInfo.batchStock')"
        class="sku-matrix__batch-input"
      />
      <ElButton size="small" type="primary" plain @click="applyBatch">
        {{ $t("pages.mall.product.salesInfo.applyAll") }}
      </ElButton>
      <ElButton size="small" text type="danger" @click="clearAllPrices">
        {{ $t("pages.mall.product.salesInfo.clear") }}
      </ElButton>
    </div>

    <ElTable :data="rows" border size="default" class="sku-matrix__table">
      <ElTableColumn :label="$t('pages.mall.product.salesInfo.specCombo')" min-width="200">
        <template #default="{ row }">
          <span class="sku-matrix__combo">{{ rowLabel(row) }}</span>
        </template>
      </ElTableColumn>

      <ElTableColumn :label="$t('pages.mall.product.salesInfo.price')" width="160">
        <template #default="{ row }">
          <ElInputNumber
            v-model="row.price"
            :min="0"
            :precision="2"
            :controls="false"
            :placeholder="$t('pages.mall.product.salesInfo.pricePlaceholder')"
            class="w-full"
          />
        </template>
      </ElTableColumn>

      <ElTableColumn :label="$t('pages.mall.product.salesInfo.stock')" width="140">
        <template #default="{ row }">
          <ElInputNumber v-model="row.stock" :min="0" :controls="false" class="w-full" />
        </template>
      </ElTableColumn>

      <ElTableColumn :label="$t('pages.mall.product.salesInfo.skuCode')" min-width="180">
        <template #default="{ row }">
          <!-- 已落库行编码只读（编码是反查身份键，改了会与库表脱钩） -->
          <ElInput
            v-if="!row.skuId"
            v-model="row.skuCode"
            :placeholder="$t('pages.mall.product.salesInfo.skuCodePlaceholder')"
          />
          <span v-else class="sku-matrix__code">{{ row.skuCode }}</span>
        </template>
      </ElTableColumn>

      <ElTableColumn :label="$t('common.table.action')" width="80" fixed="right">
        <template #default="{ row }">
          <ElButton text type="danger" size="small" @click="emit('remove', row)">
            {{ $t("common.button.delete") }}
          </ElButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <div class="sku-matrix__tip">
      {{ $t("pages.mall.product.salesInfo.matrixTip") }}
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { ElInput, ElInputNumber, ElTable, ElTableColumn } from "element-plus";

import { $t } from "@/core/i18n";
import type { SkuRow, SpecRow, SpecValueRef } from "../composables/use-sku-matrix";

/**
 * SKU 明细表：行 = 规格组合笛卡尔积（由 use-sku-matrix 维护），列内联编辑。
 * 价格写 SkuPrice(CNY)；已有行库存保存时走 AdjustStock(delta)（use-sku-matrix 编排）。
 */
defineOptions({ name: "MallProductSkuMatrix" });

const props = defineProps<{
  rows: SkuRow[];
  specs: SpecRow[];
  valueRefs: Map<number, SpecValueRef>;
}>();

const emit = defineEmits<{ remove: [row: SkuRow] }>();

const batchPrice = ref<number | undefined>();
const batchStock = ref<number | undefined>();

function rowLabel(row: SkuRow): string {
  if (row.combos.length === 0) return $t("pages.mall.product.salesInfo.defaultSpec");
  return row.combos
    .map((c) => {
      const attr = props.specs.find((s) => s.attributeId === c.attributeId);
      const val = props.valueRefs.get(c.attributeValueId);
      return `${attr?.attributeName ?? c.attributeId}:${val?.valueName ?? c.attributeValueId}`;
    })
    .join(" × ");
}

function applyBatch() {
  for (const row of props.rows) {
    if (batchPrice.value !== undefined && batchPrice.value !== null) {
      row.price = batchPrice.value;
    }
    if (batchStock.value !== undefined && batchStock.value !== null) {
      row.stock = batchStock.value;
    }
  }
}

function clearAllPrices() {
  for (const row of props.rows) {
    row.price = null;
  }
  batchPrice.value = undefined;
}
</script>

<style lang="scss" scoped>
.sku-matrix {
  display: flex;
  flex-direction: column;
  gap: 12px;

  &__toolbar {
    display: flex;
    align-items: center;
    gap: 10px;

    &-label {
      font-size: 13px;
      color: var(--el-text-color-secondary);
    }
  }

  &__batch-input {
    width: 140px;
  }

  &__combo {
    font-size: 13px;
  }

  &__code {
    font-family: var(--el-font-family-mono, monospace);
    font-size: 12px;
    color: var(--el-text-color-regular);
  }

  &__tip {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
  }
}
</style>
