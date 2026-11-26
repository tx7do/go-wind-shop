<template>
  <div class="sales-info-step">
    <template v-if="productId">
      <!-- 属性与规格 -->
      <ElCard shadow="never" class="sales-info-step__card">
        <template #header>
          <span>{{ $t("pages.mall.product.salesInfo.specSection") }}</span>
        </template>
        <SpecEditor
          :attributes="dictionary"
          :specs="matrix.specs"
          @change="matrix.syncRowsToSpecs"
        />
      </ElCard>

      <!-- SKU 明细 -->
      <ElCard shadow="never" class="sales-info-step__card">
        <template #header>
          <span>{{ $t("pages.mall.product.salesInfo.skuSection") }}</span>
        </template>
        <SkuMatrix
          :rows="matrix.rows"
          :specs="matrix.specs"
          :value-refs="matrix.valueRefs.value"
          @remove="handleRemoveRow"
        />
      </ElCard>
    </template>

    <ElAlert
      v-else
      :title="$t('pages.mall.product.salesInfo.needProductFirst')"
      type="info"
      :closable="false"
      show-icon
    />
  </div>
</template>

<script lang="ts" setup>
import { onMounted, ref, watch } from "vue";

import { fetchListProductAttributes, fetchListProductAttributeValues } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t, useI18n } from "@/core/i18n";
import { useSkuMatrix, type SpecRow } from "../composables/use-sku-matrix";
import SpecEditor from "../components/SpecEditor.vue";
import SkuMatrix from "../components/SkuMatrix.vue";

/**
 * 销售信息步骤：属性规格选择 + SKU 明细。
 * 字典（属性+值）预载并按 UI locale 取名；编辑模式从已落库组合反推规格行回显。
 */
defineOptions({ name: "MallProductStepSalesInfo" });

const props = defineProps<{ productId?: number }>();

const emit = defineEmits<{ loaded: [] }>();

const { locale } = useI18n();
const matrix = useSkuMatrix();

// 字典：{ id, name, values: [{id, name}] }[]，按 UI locale 取名，未命中回落 #id
const dictionary = ref<{ id: number; name: string; values: { id: number; name: string }[] }[]>([]);

function translationName(row: any, field: string): string {
  const list = Array.isArray(row?.translations) ? row.translations : [];
  const hit = list.find((t: any) => t?.languageCode === locale.value);
  return hit?.[field] ?? "";
}

async function loadDictionary() {
  try {
    const attrResp: any = await fetchListProductAttributes(
      new PaginationQuery({ paging: { page: 1, pageSize: 200 } })
    );
    const attrs = (attrResp?.items ?? []) as any[];

    const dict = await Promise.all(
      attrs.map(async (attr) => {
        const valResp: any = await fetchListProductAttributeValues(
          new PaginationQuery({
            paging: { page: 1, pageSize: 200 },
            formValues: { attributeId: attr.id },
          })
        );
        const values = ((valResp?.items ?? []) as any[]).map((v) => ({
          id: Number(v.id),
          name: translationName(v, "displayName") || `#${v.id}`,
        }));
        return {
          id: Number(attr.id),
          name: translationName(attr, "name") || `#${attr.id}`,
          values,
        };
      })
    );
    dictionary.value = dict;
  } catch {
    dictionary.value = [];
  }
}

/** 编辑模式：从已落库组合反推规格行（valueId 集合按属性归组，顺序随字典）。 */
function deriveSpecsFromRows(): SpecRow[] {
  const byAttr = new Map<number, Set<number>>();
  for (const row of matrix.rows) {
    for (const c of row.combos) {
      if (!byAttr.has(c.attributeId)) byAttr.set(c.attributeId, new Set());
      byAttr.get(c.attributeId)!.add(c.attributeValueId);
    }
  }
  return [...byAttr.entries()].map(([attributeId, ids]) => ({
    attributeId,
    attributeName:
      matrix.specs.find((s) => s.attributeId === attributeId)?.attributeName ||
      dictionary.value.find((a) => a.id === attributeId)?.name ||
      String(attributeId),
    valueIds: [...ids],
  }));
}

async function loadSalesInfo() {
  if (!props.productId) return;
  await matrix.load(props.productId);
  // 反推规格并注入字典名（供标签渲染），再按规格对齐行集合
  const derived = deriveSpecsFromRows();
  matrix.injectDictionaries(
    new Map(dictionary.value.map((a) => [a.id, a.name])),
    new Map(
      dictionary.value.flatMap((a) => a.values.map((v) => [v.id, v.name] as [number, string]))
    )
  );
  matrix.specs.length = 0;
  matrix.specs.push(...derived);
  matrix.syncRowsToSpecs();
}

function handleRemoveRow(row: any) {
  const index = matrix.rows.findIndex((r) => r.key === row.key);
  if (index >= 0) {
    if (matrix.rows[index].skuId) {
      matrix.pendingDeleteSkuIds.value.push(matrix.rows[index].skuId!);
    }
    matrix.rows.splice(index, 1);
  }
}

/** 宿主保存编排入口：persist 后整矩阵 reload 对齐基线（Create 返回 Empty，priceId 等靠回读）。 */
async function persistAll(): Promise<boolean> {
  if (!props.productId) return true;
  const ok = await matrix.persist(props.productId);
  if (ok) {
    await loadSalesInfo();
  }
  return ok;
}

async function init() {
  await loadDictionary();
  await loadSalesInfo();
  emit("loaded");
}

onMounted(init);

// 新建商品在第一步建档后 productId 就地出现：触发销售信息首次加载
watch(
  () => props.productId,
  (id, old) => {
    if (id && !old) init();
  }
);

defineExpose({
  persistAll,
  snapshot: matrix.snapshot,
  hasPendingDelete: matrix.hasPendingDelete,
});
</script>

<style lang="scss" scoped>
.sales-info-step {
  display: flex;
  flex-direction: column;
  gap: 16px;

  &__card {
    margin-bottom: 0;
  }
}
</style>
