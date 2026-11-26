import { computed, reactive, ref } from "vue";
import { ElMessage } from "element-plus";

import {
  fetchListSkus,
  fetchListSkuAttributeCombinations,
  fetchListSkuPrices,
  useCreateSku,
  useDeleteSku,
  useAdjustSkuStock,
  useCreateSkuPrice,
  useUpdateSkuPrice,
  useDeleteSkuPrice,
  useCreateSkuAttributeCombination,
  useDeleteSkuAttributeCombination,
} from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";
import { cartesianCombos, comboKey, type ComboCell } from "../utils/cartesian";

/**
 * SKU 矩阵状态机：销售规格 → 笛卡尔积行 → 与服务端 SKU/价格/组合的 diff 编排。
 *
 * 数据语义（与既有约束对齐）：
 * - SkuPrice 单金额（CNY），无市场价/成本价——界面只呈现「一口价」；
 * - 已有 SKU 的库存修改走 AdjustStock(delta) 保留流水审计；新建 SKU 直接带初始库存；
 * - 规格组合变化 = 删行（连带删价格/组合/SKU）+ 建行，不做部分更新；
 * - 后端 Create* 一律返回 Empty（拿不到新 ID）：新建 SKU 后按前端生成的唯一
 *   skuCode 反查回读 ID，再挂价格/组合；persist 由宿主在成功后整矩阵 reload，
 *   让 priceId/组合落库状态与服务端对齐（避免同会话二次保存重复建价）；
 * - BatchCreate 后端未实现，保存一律逐条顺序调用，失败按行收集、不中断其余。
 */

// 销售规格行：attributeId 锁定字典属性，valueIds 顺序即组合生成顺序
export interface SpecRow {
  attributeId: number;
  attributeName: string;
  valueIds: number[];
}

// SKU 明细行。skuId 为空 = 尚未落库的新行；被移除的已落库行走 pendingDeleteSkuIds
export interface SkuRow {
  key: string; // 组合签名（"attrId:valueId|..."），无规格行固定 "default"
  skuId?: number;
  skuCode: string;
  price: number | null; // 元；null = 未报价
  stock: number; // 界面库存；已有行保存时转 delta（基准 savedStock）
  savedStock: number; // 服务端当前库存（delta 基准，加载时记录）
  priceId?: number; // 已落库的价格记录 ID（更新 vs 新建）
  savedPrice: number | null; // 服务端当前价格（diff 基准）
  combos: ComboCell[];
  combosPersisted: boolean; // 组合是否已落库（新建行建完置 true）
}

// 规格值引用：供 SkuMatrix 渲染组合标签「属性:值」
export interface SpecValueRef {
  valueId: number;
  valueName: string;
}

const DEFAULT_ROW_KEY = "default";

function byFilter(formValues: Record<string, unknown>, pageSize = 100): PaginationQuery {
  return new PaginationQuery({ paging: { page: 1, pageSize }, formValues });
}

export function useSkuMatrix() {
  const { mutateAsync: createSku } = useCreateSku();
  const { mutateAsync: deleteSku } = useDeleteSku();
  const { mutateAsync: adjustStock } = useAdjustSkuStock();
  const { mutateAsync: createSkuPrice } = useCreateSkuPrice();
  const { mutateAsync: updateSkuPrice } = useUpdateSkuPrice();
  const { mutateAsync: deleteSkuPrice } = useDeleteSkuPrice();
  const { mutateAsync: createCombination } = useCreateSkuAttributeCombination();
  const { mutateAsync: deleteCombination } = useDeleteSkuAttributeCombination();

  // ── 状态 ──
  const specs = reactive<SpecRow[]>([]);
  const rows = reactive<SkuRow[]>([]);
  const valueRefs = ref(new Map<number, SpecValueRef>());
  const pendingDeleteSkuIds = ref<number[]>([]);
  const saving = ref(false);

  function makeCode(): string {
    // 自动编码：SKU + 时间戳尾 10 位 + 2 位随机；允许用户改，创建后按它反查 ID
    const base = Date.now().toString().slice(-10);
    const rand = Math.floor(Math.random() * 90 + 10);
    return `SKU${base}${rand}`;
  }

  function emptyRow(combos: ComboCell[], key: string): SkuRow {
    return {
      key,
      skuCode: makeCode(),
      price: null,
      stock: 0,
      savedStock: 0,
      savedPrice: null,
      combos,
      combosPersisted: false,
    };
  }

  // ── 目标组合集：specs → 笛卡尔积（无规格时 [默认行]）──
  const targetCombos = computed<ComboCell[][]>(() => {
    if (specs.length === 0) return [[]];
    return cartesianCombos(
      specs.map((s) =>
        s.valueIds.map((vid) => ({ attributeId: s.attributeId, attributeValueId: vid }))
      )
    );
  });

  /** 规格变化后同步行集合：新增组合补行、消失组合移行（已落库的进待删）。 */
  function syncRowsToSpecs() {
    const keySet = new Set(targetCombos.value.map((c) => comboKey(c)));

    for (let i = rows.length - 1; i >= 0; i--) {
      if (!keySet.has(rows[i].key)) {
        if (rows[i].skuId) pendingDeleteSkuIds.value.push(rows[i].skuId!);
        rows.splice(i, 1);
      }
    }
    for (const combos of targetCombos.value) {
      const key = comboKey(combos);
      if (!rows.some((r) => r.key === key)) {
        rows.push(emptyRow(combos, key));
      }
    }
  }

  /** 组合标签文案：颜色:白色 × 尺码:M；无规格行显示「默认规格」。 */
  function rowLabel(row: SkuRow): string {
    if (row.combos.length === 0) return $t("pages.mall.product.salesInfo.defaultSpec");
    return row.combos
      .map((c) => {
        const attr = specs.find((s) => s.attributeId === c.attributeId);
        const val = valueRefs.value.get(c.attributeValueId);
        return `${attr?.attributeName ?? c.attributeId}:${val?.valueName ?? c.attributeValueId}`;
      })
      .join(" × ");
  }

  // ── 加载回显：ListSkus(productId) → 每 sku 拉 combinations + prices 拼行 ──
  async function load(productId: number, specRows?: SpecRow[]) {
    rows.length = 0;
    pendingDeleteSkuIds.value = [];
    specs.length = 0;
    valueRefs.value.clear();

    if (specRows) {
      specs.push(...specRows);
    }

    const skuResp: any = await fetchListSkus(byFilter({ productId }, 200));
    const skus = (skuResp?.items ?? []) as any[];

    for (const sku of skus) {
      if (sku.id === undefined) continue;
      const [comboResp, priceResp] = await Promise.all([
        fetchListSkuAttributeCombinations(byFilter({ skuId: sku.id })),
        fetchListSkuPrices(byFilter({ skuId: sku.id })),
      ]);

      const combos: ComboCell[] = ((comboResp as any)?.items ?? []).map((c: any) => ({
        attributeId: Number(c.attributeId),
        attributeValueId: Number(c.attributeValueId),
      }));
      const prices = ((priceResp as any)?.items ?? []) as any[];
      // 币种取 CNY，未命中回落第一条（当前业务单币种）
      const price =
        prices.find((p) => p.currency === "CNY") ?? (prices.length > 0 ? prices[0] : undefined);

      rows.push({
        key: combos.length > 0 ? comboKey(combos) : DEFAULT_ROW_KEY,
        skuId: Number(sku.id),
        skuCode: sku.skuCode ?? "",
        price: price ? Number(price.amount) : null,
        stock: Number(sku.stockQty ?? 0),
        savedStock: Number(sku.stockQty ?? 0),
        priceId: price ? Number(price.id) : undefined,
        savedPrice: price ? Number(price.amount) : null,
        combos,
        combosPersisted: true,
      });
    }
  }

  /** 注入字典（attributeId→name、valueId→name），供标签渲染。 */
  function injectDictionaries(attrNames: Map<number, string>, valNames: Map<number, string>) {
    for (const spec of specs) {
      const n = attrNames.get(spec.attributeId);
      if (n && !spec.attributeName) spec.attributeName = n;
    }
    for (const [vid, name] of valNames) {
      if (!valueRefs.value.has(vid)) {
        valueRefs.value.set(vid, { valueId: vid, valueName: name });
      }
    }
  }

  /** Create 返回 Empty：按唯一 skuCode 反查新 SKU 的 ID。 */
  async function resolveCreatedSkuId(skuCode: string): Promise<number> {
    const resp: any = await fetchListSkus(byFilter({ skuCode }, 10));
    const hit = ((resp?.items ?? []) as any[]).find((s) => s.skuCode === skuCode);
    if (!hit?.id) throw new Error(`created sku not found by code ${skuCode}`);
    return Number(hit.id);
  }

  // ── 保存编排：删 → 增/改，逐行容错 ──
  /** @returns 是否全部成功（false = 有行失败，已用 message 汇总）。成功后宿主应整矩阵 reload 对齐基线。 */
  async function persist(productId: number): Promise<boolean> {
    if (saving.value) return false;
    saving.value = true;
    const errors: string[] = [];

    try {
      // 0. 待删行：价格 → 组合 → SKU
      for (const skuId of [...pendingDeleteSkuIds.value]) {
        try {
          const [comboResp, priceResp] = await Promise.all([
            fetchListSkuAttributeCombinations(byFilter({ skuId })),
            fetchListSkuPrices(byFilter({ skuId })),
          ]);
          for (const p of (priceResp as any)?.items ?? []) {
            await deleteSkuPrice({ id: String(p.id) });
          }
          for (const c of (comboResp as any)?.items ?? []) {
            await deleteCombination({ id: String(c.id) });
          }
          await deleteSku({ id: String(skuId) });
          pendingDeleteSkuIds.value = pendingDeleteSkuIds.value.filter((id) => id !== skuId);
        } catch {
          errors.push(`SKU#${skuId}`);
        }
      }

      for (let i = 0; i < rows.length; i++) {
        const row = rows[i];
        const label = rowLabel(row);
        try {
          if (!row.skuId) {
            // 新增：CreateSku（带初始库存）→ 按唯一 skuCode 反查 ID → 价格 → 组合
            await createSku({
              productId,
              skuCode: row.skuCode,
              stockQty: row.stock,
            });
            row.skuId = await resolveCreatedSkuId(row.skuCode);
            row.savedStock = row.stock;
          } else if (row.stock !== row.savedStock) {
            // 更新：库存变了走 AdjustStock（流水审计）；编码允许改但组合不变
            await adjustStock({
              id: row.skuId,
              values: {
                delta: row.stock - row.savedStock,
                remark: $t("pages.mall.product.salesInfo.adjustRemark"),
              },
            });
            row.savedStock = row.stock;
          }

          // 价格：有价才落；已有价且清空则跳过（不提供删价）
          if (row.price !== null && row.price !== undefined && Number.isFinite(row.price)) {
            const amount = String(row.price);
            if (row.priceId) {
              if (amount !== String(row.savedPrice)) {
                await updateSkuPrice({ id: row.priceId, values: { amount } });
                row.savedPrice = row.price;
              }
            } else {
              await createSkuPrice({ skuId: row.skuId, currency: "CNY", amount });
            }
          }

          // 组合：仅新行需要建（已有行的组合变化 = 删行+建行，不在此路径）
          if (!row.combosPersisted && row.combos.length > 0) {
            for (const c of row.combos) {
              await createCombination({
                skuId: row.skuId,
                attributeId: c.attributeId,
                attributeValueId: c.attributeValueId,
              });
            }
            row.combosPersisted = true;
          }
        } catch {
          errors.push(label);
        }
      }

      if (errors.length > 0) {
        ElMessage.error(`${$t("pages.mall.product.salesInfo.errorRows")}: ${errors.join("、")}`);
        return false;
      }
      return true;
    } finally {
      saving.value = false;
    }
  }

  /** 脏判断快照（宿主路由守卫用）。 */
  function snapshot(): string {
    return JSON.stringify({
      specs: specs.map((s) => ({ a: s.attributeId, v: [...s.valueIds] })),
      rows: rows.map((r) => ({
        k: r.key,
        id: r.skuId,
        c: r.skuCode,
        p: r.price,
        s: r.stock,
      })),
      del: [...pendingDeleteSkuIds.value],
    });
  }

  const hasPendingDelete = computed(() => pendingDeleteSkuIds.value.length > 0);

  return {
    specs,
    rows,
    valueRefs,
    pendingDeleteSkuIds,
    hasPendingDelete,
    saving,
    makeCode,
    syncRowsToSpecs,
    rowLabel,
    load,
    injectDictionaries,
    persist,
    snapshot,
  };
}
