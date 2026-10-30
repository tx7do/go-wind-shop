import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { useQuery } from '@tanstack/vue-query';
import { fetchListSkus } from './sku';
import { fetchListSkuPrices } from './sku-price';
import { useListActiveSeckills } from './seckill';
import { useListActiveGroupons } from './groupon';
import { useListActivePromotions } from './promotion';
import { getCurrentLocale } from '@/utils/locale';

/**
 * 商品网格（首页 / 分类 / 搜索 / 收藏）的价格与活动价来源。
 *
 * catalog.v1.Product 里**没有**价格字段：价格挂在 SkuPrice 上，按 SKU 存「分」。
 * 所以列表页要显示价格只有两条路：
 *   1) 逐商品取 SKU 再取价格 —— 12 个商品就是 24 次请求（详情页因为要选 SKU 才这么做）。
 *   2) 本文件：两次批量查询
 *        skus?productId__in=[ids]  →  sku-prices?skuId__in=[skuIds]
 *      再按商品取最低 SKU 价作为「起」价。
 * 过滤算子 `__in` 与 pageSize 生效性都已在 app 服务上实测（含 bogus id 返回 0 条的对照）。
 *
 * 单位一律「分」，不做提前换算：PriceTag / 活动价 / 划线价三处入参同标度，
 * 少一次 ×100 就少一处对不齐的地方。
 */

export type ProductPrice = {
  /** 最低价，单位分 */
  fen: number;
  currency: string;
};

/** 网格卡片上的活动角标：labelKey 交给页面 t()，本文件不依赖 i18n 实例 */
export type ProductPromo = {
  labelKey: 'mall.seckill.tag' | 'mall.groupon.tag' | 'mall.promotion.specialPriceTag';
  tone: 'danger' | 'warning' | 'primary';
  /** 活动价，单位分 */
  fen: number;
  /** 拼团成团人数，仅拼团有 */
  groupSize?: number;
};

const DISPLAY_CURRENCY = 'CNY';

// go-crud 的 paginator 在未传 limit 时默认只给 10 条，必须显式传大 pageSize。
// 网格页最多 48 个商品、每商品按 10 个 SKU 估算。
const BATCH_PAGE_SIZE = 500;

type SkuRow = { id?: number; productId?: number };
type PriceRow = { skuId?: number; currency?: string; amount?: string | number };

function rowsOf<T>(data: unknown): T[] {
  return ((data as any)?.items ?? []) as T[];
}

/** 同一 SKU 多币种时只比同币种：有 CNY 用 CNY，没有则整体退回第一出现的币种。 */
function minFenOfSku(rows: PriceRow[]): { fen: number; currency: string } | undefined {
  const usable = rows.filter((r) => r.currency === DISPLAY_CURRENCY);
  const candidates = usable.length > 0 ? usable : rows;
  let min = Number.POSITIVE_INFINITY;
  let currency = candidates[0]?.currency ?? DISPLAY_CURRENCY;
  for (const r of candidates) {
    // 线上 int64 是字符串（"12800"），必须先 Number 再比，否则是字典序
    const fen = Number(r.amount);
    if (!Number.isFinite(fen) || fen <= 0) continue;
    if (fen < min) {
      min = fen;
      currency = r.currency ?? currency;
    }
  }
  return Number.isFinite(min) ? { fen: min, currency } : undefined;
}

function buildPriceMap(skuRows: SkuRow[], priceRows: PriceRow[]) {
  const bySku = new Map<number, PriceRow[]>();
  for (const p of priceRows) {
    if (p.skuId === undefined) continue;
    const list = bySku.get(p.skuId);
    if (list) list.push(p);
    else bySku.set(p.skuId, [p]);
  }

  const price = new Map<number, ProductPrice>();
  const skuIds = new Map<number, number[]>();
  for (const sku of skuRows) {
    if (sku.id === undefined || sku.productId === undefined) continue;
    const ids = skuIds.get(sku.productId);
    if (ids) ids.push(sku.id);
    else skuIds.set(sku.productId, [sku.id]);

    const rows = bySku.get(sku.id);
    if (!rows || rows.length === 0) continue;
    const min = minFenOfSku(rows);
    if (!min) continue;
    const current = price.get(sku.productId);
    if (!current || min.fen < current.fen) {
      price.set(sku.productId, { fen: min.fen, currency: min.currency });
    }
  }
  return { price, skuIds };
}

export function useProductPrices(productIds: MaybeRefOrGetter<number[]>) {
  const ids = computed(() =>
    Array.from(new Set(toValue(productIds).filter((n) => Number.isInteger(n) && n > 0))).sort(
      (a, b) => a - b,
    ),
  );

  const skusQuery = useQuery({
    queryKey: computed(() => ['listSkusByProductIds', ids.value, getCurrentLocale()]),
    queryFn: () =>
      fetchListSkus({
        page: 1,
        pageSize: BATCH_PAGE_SIZE,
        noPaging: false,
        query: JSON.stringify({ productId__in: ids.value }),
      }),
    enabled: computed(() => ids.value.length > 0),
    staleTime: 5 * 60 * 1000,
  });

  const skuRows = computed(() => rowsOf<SkuRow>(skusQuery.data?.value));
  const skuIds = computed(() => skuRows.value.map((s) => s.id).filter((id): id is number => id !== undefined));

  const pricesQuery = useQuery({
    queryKey: computed(() => ['listSkuPricesBySkuIds', skuIds.value, getCurrentLocale()]),
    queryFn: () =>
      fetchListSkuPrices({
        page: 1,
        pageSize: BATCH_PAGE_SIZE,
        noPaging: false,
        query: JSON.stringify({ skuId__in: skuIds.value }),
      }),
    enabled: computed(() => skuIds.value.length > 0),
    staleTime: 5 * 60 * 1000,
  });

  const built = computed(() => buildPriceMap(skuRows.value, rowsOf<PriceRow>(pricesQuery.data?.value)));

  return {
    /** 商品 → 最低 SKU 价（分） */
    priceByProductId: computed(() => built.value.price),
    /** 商品 → 它的 SKU 列表，活动角标按 SKU 匹配秒杀/拼团 */
    skuIdsByProduct: computed(() => built.value.skuIds),
    isLoading: computed(
      () => ids.value.length > 0 && (skusQuery.isLoading.value || pricesQuery.isLoading.value),
    ),
  };
}

type SeckillRow = { skuId?: number; seckillPrice?: number };
type GrouponRow = { skuId?: number; groupPrice?: number; groupSize?: number };
type PromotionRow = {
  productId?: number;
  promoType?: string;
  specialPrice?: number;
  priority?: number;
};

/**
 * 商品网格的活动角标。
 *
 * 优先级与商品详情页模板的渲染顺序保持一致（拼团 > 秒杀 > 专价），
 * 并且沿用同一道闸：活动价必须低于现价才展示——同一个商品在卡片和详情页不会互相打脸。
 * 秒杀/拼团按 SKU 命中，专价是商品级活动。
 *
 * 三个 ListActive* 都是匿名可读、无参数的整表查询，query key 固定，
 * 所以一张网格只发 3 次请求，且与详情页共用同一份缓存。
 */
export function useProductPromos(
  productIds: MaybeRefOrGetter<number[]>,
  basePrice: MaybeRefOrGetter<Map<number, ProductPrice>>,
  skuIdsByProduct: MaybeRefOrGetter<Map<number, number[]>>,
) {
  const seckills = useListActiveSeckills();
  const groupons = useListActiveGroupons();
  const promotions = useListActivePromotions();

  const promoByProductId = computed(() => {
    const ids = toValue(productIds);
    const priceMap = toValue(basePrice);
    const skuMap = toValue(skuIdsByProduct);

    const seckillBySku = new Map<number, number>();
    for (const s of rowsOf<SeckillRow>(seckills.data?.value)) {
      if (s.skuId === undefined) continue;
      const fen = Number(s.seckillPrice);
      if (Number.isFinite(fen) && fen > 0) seckillBySku.set(s.skuId, fen);
    }
    const grouponBySku = new Map<number, { fen: number; size?: number }>();
    for (const g of rowsOf<GrouponRow>(groupons.data?.value)) {
      if (g.skuId === undefined) continue;
      const fen = Number(g.groupPrice);
      if (Number.isFinite(fen) && fen > 0) grouponBySku.set(g.skuId, { fen, size: g.groupSize });
    }
    const specialByProduct = new Map<number, number>();
    for (const p of rowsOf<PromotionRow>(promotions.data?.value)) {
      if (p.productId === undefined || p.promoType !== 'SPECIAL_PRICE') continue;
      const fen = Number(p.specialPrice);
      if (!Number.isFinite(fen) || fen <= 0) continue;
      const current = specialByProduct.get(p.productId);
      if (current === undefined || fen < current) specialByProduct.set(p.productId, fen);
    }

    const out = new Map<number, ProductPromo>();
    for (const id of ids) {
      const base = priceMap.get(id);
      if (!base) continue;
      const skuIds = skuMap.get(id) ?? [];

      for (const sid of skuIds) {
        const g = grouponBySku.get(sid);
        if (g && g.fen < base.fen) {
          out.set(id, { labelKey: 'mall.groupon.tag', tone: 'warning', fen: g.fen, groupSize: g.size });
          break;
        }
      }
      if (out.has(id)) continue;

      for (const sid of skuIds) {
        const fen = seckillBySku.get(sid);
        if (fen !== undefined && fen < base.fen) {
          out.set(id, { labelKey: 'mall.seckill.tag', tone: 'danger', fen });
          break;
        }
      }
      if (out.has(id)) continue;

      const special = specialByProduct.get(id);
      if (special !== undefined && special < base.fen) {
        out.set(id, { labelKey: 'mall.promotion.specialPriceTag', tone: 'primary', fen: special });
      }
    }
    return out;
  });

  return { promoByProductId };
}
