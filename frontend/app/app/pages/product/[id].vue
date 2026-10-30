<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { toast } from 'vue-sonner';
import { cn } from '@/lib/utils';
import { XIcon } from '@/plugins/xicon';
import {
  useGetProduct,
  useListProductAttributes,
  useListProductAttributeValues,
  useListSkus,
  fetchListSkuAttributeCombinationsStore,
  fetchListSkuPricesStore,
  useListCarts,
  useCreateCart,
  useCreateCartItem,
  useListWishlist,
  useCreateWishlist,
  useDeleteWishlist,
  useGetProductRating,
  useGetProductReviewSummary,
  useListActivePromotions,
  useListActiveSeckills,
  useListActiveGroupons,
  useListComments,
  useCreateComment,
  useGetInteractionStatus,
  useGetCounts,
  useLike,
  useUnlike,
} from '@/api/composables';
import { pickTranslation } from '@/utils/translation';
import { formatFen } from '@/utils/money';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';
import { queryClient } from '@/plugins/vue-query';
import type {
  commentservicev1_Comment,
  commentservicev1_CreateCommentRequest,
} from '@/api/generated/app/service/v1';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

const productId = computed(() => Number(route.params.id));

// ---------- 登录态 / 当前用户 ----------
const accessStore = useAccessStore();
const userStore = useUserStore();
const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
const currentUserId = computed(() => userStore.user?.id ?? 0);
const currentTenantId = computed(() => userStore.tenantId ?? 0);

// ---------- 类型 ----------
type ProductTranslation = {
  name?: string;
  shortDescription?: string;
  longDescription?: string;
  languageCode?: string;
};
type ProductEntity = {
  id?: number;
  imageUrl?: string;
  translations?: ProductTranslation[];
};
type AttributeTranslation = {
  name?: string;
  languageCode?: string;
};
type AttributeEntity = {
  id?: number;
  translations?: AttributeTranslation[];
};
type AttributeValueTranslation = {
  displayName?: string;
  languageCode?: string;
};
type AttributeValueEntity = {
  id?: number;
  attributeId?: number;
  translations?: AttributeValueTranslation[];
};
type SkuEntity = {
  id?: number;
  productId?: number;
  stockQty?: number;
};
type SkuPriceEntity = {
  skuId?: number;
  currency?: string;
  amount?: string;
};
type SkuCombinationEntity = {
  skuId?: number;
  attributeId?: number;
  attributeValueId?: number;
};

const DISPLAY_CURRENCY = 'CNY';

// ---------- 商品详情 ----------
const productQuery = useGetProduct(computed(() => productId.value), {
  enabled: computed(() => productId.value > 0),
});
const product = computed<ProductEntity | undefined>(() => {
  return productQuery.data?.value as ProductEntity | undefined;
});
const productLoading = computed(() => productQuery.isPending.value);
const productError = computed(() => productQuery.isError.value);
const productTranslation = computed(() => pickTranslation(product.value?.translations));

useHead({ title: () => productTranslation.value?.name || '' });

// ---------- SKU 列表 ----------
const skusQuery = useListSkus(
  computed(() => ({
    page: 1,
    pageSize: 100,
    noPaging: false,
    query: JSON.stringify({ productId: productId.value }),
  })),
  { enabled: computed(() => productId.value > 0) },
);
const skus = computed<SkuEntity[]>(() => {
  const items = (skusQuery.data?.value as any)?.items ?? [];
  return (items as SkuEntity[]) ?? [];
});

// ---------- 收藏状态 ----------
// 查询当前 viewer 对此商品是否已收藏。后端 UserPrivacy 自动注入 user_id=viewer，
// 故返回的 items 即当前用户对 productId 的收藏记录（至多 1 条，由唯一索引保证）。
const favoriteQuery = useListWishlist(
  computed(() => ({
    page: 1,
    pageSize: 10,
    noPaging: false,
    query: JSON.stringify({ productId: productId.value }),
  })),
  { enabled: computed(() => isLogin.value && productId.value > 0) },
);

// 当前用户对此商品的收藏记录 id（存在即已收藏）。
const favoriteRecordId = computed<number | undefined>(() => {
  const items = (favoriteQuery.data?.value as any)?.items ?? [];
  return (items as Array<{ id?: number }>)[0]?.id;
});
const isFavorited = computed(() => favoriteRecordId.value !== undefined);

const createFavoriteMutation = useCreateWishlist({
  onSuccess: () => {
    favoriteQuery.refetch();
  },
});
const deleteFavoriteMutation = useDeleteWishlist({
  onSuccess: () => {
    favoriteQuery.refetch();
  },
});

async function toggleFavorite() {
  if (!isLogin.value) {
    navigateTo(localePath('/login'));
    return;
  }
  if (isFavorited.value && favoriteRecordId.value !== undefined) {
    try {
      await deleteFavoriteMutation.mutateAsync(favoriteRecordId.value);
    } catch {
      toast.error(t('mall.wishlist.errors.removeFailed'));
    }
  } else {
    try {
      // 仅传 productId；user_id 由后端 UserPrivacy 强制覆盖为 viewer，不可伪造。
      await createFavoriteMutation.mutateAsync({ productId: productId.value } as any);
    } catch {
      // 忽略——状态由 query 体现
    }
  }
}

// ---------- 商品评分聚合（匿名可读）----------
const ratingQuery = useGetProductRating(
  computed(() => ({ productId: productId.value })),
  { enabled: computed(() => productId.value > 0) },
);
const ratingAverage = computed<number>(() => {
  const v = (ratingQuery.data?.value as any)?.average;
  return typeof v === 'number' ? v : 0;
});
const ratingCount = computed<number>(() => {
  const v = (ratingQuery.data?.value as any)?.count;
  return typeof v === 'number' ? v : 0;
});
const hasRating = computed(() => ratingCount.value > 0 && ratingAverage.value > 0);

// ---------- 评论 AI 摘要（匿名可读；LLM 生成，服务端缓存 6h） ----------
const reviewSummaryQuery = useGetProductReviewSummary(productId, {
  enabled: computed(() => productId.value > 0),
});
const reviewSummary = computed<string>(() => {
  const d = reviewSummaryQuery.data?.value as any;
  return d?.available && d?.summary ? String(d.summary) : '';
});

// ---------- 属性 & 属性值 ----------
const attributesQuery = useListProductAttributes({
  page: 1,
  pageSize: 100,
  noPaging: false,
});
const attributes = computed<AttributeEntity[]>(() => {
  const items = (attributesQuery.data?.value as any)?.items ?? [];
  return (items as AttributeEntity[]) ?? [];
});

const attributeValuesQuery = useListProductAttributeValues({
  page: 1,
  pageSize: 200,
  noPaging: false,
});
const allAttributeValues = computed<AttributeValueEntity[]>(() => {
  const items = (attributeValuesQuery.data?.value as any)?.items ?? [];
  return (items as AttributeValueEntity[]) ?? [];
});

const valuesByAttribute = computed(() => {
  const map = new Map<number, AttributeEntity[]>();
  for (const v of allAttributeValues.value) {
    if (v.attributeId === undefined) continue;
    const arr = map.get(v.attributeId) ?? [];
    arr.push(v);
    map.set(v.attributeId, arr);
  }
  return map;
});

// 可选属性：只保留确实挂了属性值的属性。属性表是全局数据，可能存在
// 「有属性名但该属性一个值都没配」的脏数据——这种属性渲染不出任何规格 chip，
// 用户永远无法选中它；若参与「是否选满」判断会让 matchedSku 永远为空，
// 整页价格隐藏、按钮禁用（购买链路死锁）。空属性应直接视为单 SKU 商品。
const selectableAttributes = computed<AttributeEntity[]>(() => {
  return attributes.value.filter(
    (attr) => attr.id !== undefined && (valuesByAttribute.value.get(attr.id)?.length ?? 0) > 0,
  );
});

// ---------- 用户选择 ----------
// 注意：reka-ui Select 使用字符串值，因此选择以字符串存储，匹配时转为数字。
const selections = reactive<Record<number, string | undefined>>({});

function clearSelections() {
  for (const k of Object.keys(selections)) {
    delete selections[Number(k)];
  }
}

// 为每个尚无选中值的属性设默认选中其第一个值，使页面初始即有完整选中态，
// 便于用户直观看到当前 SKU 与价格。
// 属性/属性值表为全局数据、与 productId 无关，故默认值与具体商品无关。
function applyDefaultSelections() {
  const attrs = attributes.value;
  const valsMap = valuesByAttribute.value;
  if (!attrs || attrs.length === 0 || !valsMap || valsMap.size === 0) return;
  for (const attr of attrs) {
    if (attr.id === undefined) continue;
    if (selections[attr.id] !== undefined) continue;
    const vals = valsMap.get(attr.id);
    if (!vals || vals.length === 0) continue;
    const first = vals[0];
    if (first?.id !== undefined) {
      selections[attr.id] = String(first.id);
    }
  }
}

// 属性/属性值数据加载完成后应用默认选中。
watch([attributes, valuesByAttribute], applyDefaultSelections, { immediate: true });

// ---------- 每个 SKU 的属性组合 & 价格（通过 fetch*Store 在 watch 中拉取） ----------
const skuCombinationsMap = reactive<Record<number, Map<number, number>>>({});
const skuPricesMap = reactive<Record<number, SkuPriceEntity[]>>({});

watch(
  skus,
  async (list) => {
    for (const k of Object.keys(skuCombinationsMap)) delete skuCombinationsMap[Number(k)];
    for (const k of Object.keys(skuPricesMap)) delete skuPricesMap[Number(k)];
    if (!list || list.length === 0) return;
    await Promise.all(
      list.map(async (sku) => {
        if (sku.id === undefined) return;
        try {
          const combosResp: any = await fetchListSkuAttributeCombinationsStore({
            page: 1,
            pageSize: 100,
            noPaging: false,
            query: JSON.stringify({ skuId: sku.id }),
          });
          const combos = (combosResp?.items ?? []) as SkuCombinationEntity[];
          const map = new Map<number, number>();
          for (const c of combos) {
            if (c.attributeId !== undefined && c.attributeValueId !== undefined) {
              map.set(c.attributeId, c.attributeValueId);
            }
          }
          skuCombinationsMap[sku.id] = map;
        } catch (e) {
          // ignore
        }
        try {
          const pricesResp: any = await fetchListSkuPricesStore({
            page: 1,
            pageSize: 100,
            noPaging: false,
            query: JSON.stringify({ skuId: sku.id }),
          });
          const prices = (pricesResp?.items ?? []) as SkuPriceEntity[];
          skuPricesMap[sku.id] = prices;
        } catch (e) {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

const allAttributesSelected = computed(() => {
  // 无可选属性（单 SKU / 属性未配值）：无可选项，视为已满足选择前提。
  if (selectableAttributes.value.length === 0) return true;
  for (const attr of selectableAttributes.value) {
    if (attr.id === undefined) continue;
    if (selections[attr.id] === undefined) return false;
  }
  return true;
});

// ---------- 匹配 SKU ----------
const matchedSku = computed<SkuEntity | undefined>(() => {
  if (!allAttributesSelected.value) return undefined;

  // 无可选属性（单 SKU / 属性未配值）：无属性组合可匹配，
  // 优先取有库存的第一个 SKU，否则退回第一个 SKU，保证价格与购买可用。
  if (selectableAttributes.value.length === 0) {
    const inStock = skus.value.find((s) => (s.stockQty ?? 0) > 0) ?? skus.value[0];
    if (inStock && inStock.id !== undefined) return inStock;
    return undefined;
  }

  const selectionEntries: { attrId: number; valId: number }[] = [];
  for (const [attrIdStr, valIdStr] of Object.entries(selections)) {
    if (valIdStr === undefined) continue;
    const attrId = Number(attrIdStr);
    const valId = Number(valIdStr);
    if (Number.isNaN(attrId) || Number.isNaN(valId)) continue;
    selectionEntries.push({ attrId, valId });
  }
  if (selectionEntries.length === 0) return undefined;
  for (const sku of skus.value) {
    if (sku.id === undefined) continue;
    const combos = skuCombinationsMap[sku.id];
    if (!combos || combos.size === 0) continue;
    if (combos.size !== selectionEntries.length) continue;
    let matches = true;
    for (const { attrId, valId } of selectionEntries) {
      if (combos.get(attrId) !== valId) {
        matches = false;
        break;
      }
    }
    if (matches) return sku;
  }
  return undefined;
});

const matchedSkuPrice = computed<SkuPriceEntity | undefined>(() => {
  const sku = matchedSku.value;
  if (!sku || sku.id === undefined) return undefined;
  const prices = skuPricesMap[sku.id];
  if (!prices || prices.length === 0) return undefined;
  return prices.find((p) => p.currency === DISPLAY_CURRENCY) ?? prices[0];
});

const isOutOfStock = computed(() => {
  const sku = matchedSku.value;
  if (!sku) return false;
  return (sku.stockQty ?? 0) <= 0;
});

const canAddToCart = computed(() => {
  return matchedSku.value !== undefined && !isOutOfStock.value;
});

// 后端金额一律是「分」且经 protobuf 序列化成字符串：先 Number() 再比较/展示。
const basePriceFen = computed<number>(() => {
  const fen = Number(matchedSkuPrice.value?.amount ?? 0);
  return Number.isFinite(fen) && fen > 0 ? fen : 0;
});
const hasPrice = computed(() => basePriceFen.value > 0);

// ---------- 秒杀（最高优先级价格展示） ----------
// ListActiveSeckills 返回当前生效中且未售罄的秒杀，这里按当前商品的 SKU 匹配。
type SeckillEntity = {
  id?: number;
  name?: string;
  skuId?: number;
  seckillPrice?: number;
  totalStock?: number;
  soldCount?: number;
  validUntil?: string;
};

const seckillsQuery = useListActiveSeckills();
const activeSeckills = computed(() => {
  const items = ((seckillsQuery.data?.value as any)?.items ?? []) as SeckillEntity[];
  return items;
});

// ---------- 拼团（最高优先级价格展示） ----------
type GrouponEntity = {
  id?: number;
  name?: string;
  skuId?: number;
  groupPrice?: number;
  groupSize?: number;
};

const grouponsQuery = useListActiveGroupons();
const activeGroupons = computed(() => {
  const items = ((grouponsQuery.data?.value as any)?.items ?? []) as GrouponEntity[];
  return items;
});

// ---------- 促销活动（专价划线价/满减角标） ----------
// ListActivePromotions 返回当前租户全部生效活动，这里按当前商品过滤。
type PromotionEntity = {
  id?: number;
  name?: string;
  promoType?: string;
  productId?: number;
  specialPrice?: number;
  thresholdAmount?: number;
  reductionAmount?: number;
  priority?: number;
};

const promotionsQuery = useListActivePromotions();

const productPromotions = computed(() => {
  const items = ((promotionsQuery.data?.value as any)?.items ?? []) as PromotionEntity[];
  return items.filter((p) => p.productId === productId.value);
});

// 专价活动：priority 最大者优先；仅当专价低于现价时生效（与服务端同语义）。
const specialPricePromo = computed(() => {
  const list = productPromotions.value.filter(
    (p) => p.promoType === 'SPECIAL_PRICE' && Number(p.specialPrice ?? 0) > 0,
  );
  return list.reduce<PromotionEntity | undefined>(
    (best, p) => (!best || (p.priority ?? 0) > (best.priority ?? 0) ? p : best),
    undefined,
  );
});

const specialPriceFen = computed<number | null>(() => {
  const base = basePriceFen.value;
  const fen = Number(specialPricePromo.value?.specialPrice ?? 0);
  return base > 0 && fen > 0 && fen < base ? fen : null;
});

// 秒杀（最高优先级）：当前选中 SKU 的生效秒杀，秒杀价低于现价时覆盖展示
const activeSeckillForSku = computed(() => {
  const sid = matchedSku.value?.id;
  if (!sid) return undefined;
  return activeSeckills.value.find((k) => k.skuId === sid);
});
const seckillPriceFen = computed<number | null>(() => {
  const base = basePriceFen.value;
  const fen = Number(activeSeckillForSku.value?.seckillPrice ?? 0);
  return base > 0 && fen > 0 && fen < base ? fen : null;
});

// 拼团（最高优先级）：当前选中 SKU 的生效拼团，拼团价低于现价时覆盖展示
const activeGrouponForSku = computed(() => {
  const sid = matchedSku.value?.id;
  if (!sid) return undefined;
  return activeGroupons.value.find((g) => g.skuId === sid);
});
const grouponPriceFen = computed<number | null>(() => {
  const base = basePriceFen.value;
  const fen = Number(activeGrouponForSku.value?.groupPrice ?? 0);
  return base > 0 && fen > 0 && fen < base ? fen : null;
});

// 满减角标：取减让最大的店铺级活动展示「满 X 减 Y」。
const fullReductionPromo = computed(() => {
  const list = productPromotions.value.filter((p) => p.promoType === 'FULL_REDUCTION');
  return list.reduce<PromotionEntity | undefined>(
    (best, p) => (!best || (p.reductionAmount ?? 0) > (best.reductionAmount ?? 0) ? p : best),
    undefined,
  );
});

const effectivePriceFen = computed(
  () => grouponPriceFen.value ?? seckillPriceFen.value ?? specialPriceFen.value ?? basePriceFen.value,
);
const hasActivePrice = computed(() => hasPrice.value && effectivePriceFen.value < basePriceFen.value);

const grouponBadgeLabel = computed(() => {
  const g = activeGrouponForSku.value;
  if (!g) return '';
  const name = g.name || t('mall.groupon.tag');
  return g.groupSize ? `${name} · ${t('mall.groupon.tagSize', { size: g.groupSize })}` : name;
});
const seckillBadgeLabel = computed(() => activeSeckillForSku.value?.name || t('mall.seckill.tag'));
const specialBadgeLabel = computed(
  () => specialPricePromo.value?.name || t('mall.promotion.specialPriceTag'),
);
const fullReductionLabel = computed(() => {
  const p = fullReductionPromo.value;
  if (!p) return '';
  return t('mall.promotion.fullReductionTag', {
    threshold: formatFen(p.thresholdAmount),
    reduction: formatFen(p.reductionAmount),
  });
});

const servicePromises = computed(() => [
  { icon: 'lucide:shield-check', label: t('mall.product.serviceAuthentic') },
  { icon: 'lucide:truck', label: t('mall.product.serviceShipping') },
  { icon: 'lucide:rotate-ccw', label: t('mall.product.serviceReturn') },
]);

// ---------- 主图 / 缩略图 ----------
const galleryImages = computed<string[]>(() => {
  const url = product.value?.imageUrl;
  return url ? [url] : [];
});
const activeImageIndex = ref(0);
const activeImage = computed(() => galleryImages.value[activeImageIndex.value]);

// ---------- 当前用户的购物车（与 cart.vue 同样的按 userId 过滤逻辑） ----------
type CartEntity = { id?: number; userId?: number };

// 注：user_id 行级隔离由后端 UserPrivacy 策略强制（钉定为当前登录用户），
// 前端 userId 值会被服务端忽略。enabled 守卫避免未登录/未 hydrate 发请求。
const cartsQuery = useListCarts(
  computed(() => ({
    page: 1,
    pageSize: 1,
    noPaging: false,
    query: JSON.stringify({ userId: currentUserId.value }),
  })),
  { enabled: isLogin },
);
const cart = computed<CartEntity | undefined>(() => {
  const items = ((cartsQuery.data?.value as any)?.items ?? []) as CartEntity[];
  return items[0];
});

const createCartMutation = useCreateCart({
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['listCarts'] });
  },
  onError: (err: any) => {
    toast.error(err?.message || t('cart.errors.updateFailed'));
  },
});
const createCartItemMutation = useCreateCartItem({
  onSuccess: () => {
    // 购物车内容已变，刷新购物车项与购物车列表缓存
    queryClient.invalidateQueries({ queryKey: ['listCartItems'] });
    queryClient.invalidateQueries({ queryKey: ['listCarts'] });
    toast.success(t('mall.product.addToCart'));
  },
  onError: (err: any) => {
    toast.error(err?.message || t('cart.errors.updateFailed'));
  },
});

// ---------- 数量与库存 ----------
const quantity = ref(1);

// 当前匹配 SKU 的库存上限（无匹配 SKU 时回落 1，此时数量框禁用）
const matchedSkuStock = computed(() => {
  const sku = matchedSku.value;
  if (!sku || sku.id === undefined) return 1;
  const s = sku.stockQty ?? 1;
  return s < 1 ? 1 : s;
});

function incrementQty() {
  if (quantity.value < matchedSkuStock.value) {
    quantity.value++;
  }
}

function decrementQty() {
  if (quantity.value > 1) {
    quantity.value--;
  }
}

// 路由 id 变化（同组件复用 /product/1→/product/2）：清除上一商品的属性选择并
// 重置数量，再按当前（全局）属性表重新应用默认选中，避免旧选择残留导致
// matchedSku 误配、quantity 越过新 SKU 库存上限；同时滚回顶部，避免停留于
// 前一商品页底的滚动位置。
watch(productId, () => {
  clearSelections();
  applyDefaultSelections();
  quantity.value = 1;
  activeImageIndex.value = 0;
  closeReply();
  for (const k of Object.keys(likeState)) delete likeState[Number(k)];
  if (import.meta.client) scrollToTop();
});

// ---------- 添加购物车 ----------
async function addToCartWithQty(qty: number) {
  if (!canAddToCart.value) return;
  if (!isLogin.value) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  const sku = matchedSku.value;
  if (sku?.id === undefined) return;

  let targetCartId = cart.value?.id;

  // 新用户尚无购物车：先创建
  if (targetCartId === undefined) {
    try {
      const created: any = await createCartMutation.mutateAsync({
        userId: currentUserId.value,
        tenantId: currentTenantId.value,
      } as any);
      targetCartId = created?.id;
    } catch {
      return;
    }
  }

  if (targetCartId === undefined) {
    toast.error(t('cart.errors.updateFailed'));
    return;
  }

  await createCartItemMutation.mutateAsync({
    cartId: targetCartId,
    skuId: sku.id,
    quantity: qty,
    tenantId: currentTenantId.value,
  } as any);
}

// ---------- 立即购买：加购后跳转结算页（拼团单带拼团参数） ----------
async function buyNow() {
  if (!canAddToCart.value) return;
  await addToCartWithQty(quantity.value);
  const g = activeGrouponForSku.value;
  if (g?.id) {
    navigateTo(localePath('/checkout') + `?grouponActivityId=${g.id}`);
    return;
  }
  navigateTo(localePath('/checkout'));
}

// ---------- SKU 卡片选择器：键盘方向键在当前组内切换 ----------
function onRadioKeydown(e: KeyboardEvent, attrId: number) {
  const vals = valuesByAttribute.value.get(attrId) ?? [];
  if (vals.length === 0) return;
  const cur = selections[attrId];
  let idx = vals.findIndex((v) => v.id !== undefined && String(v.id) === cur);
  if (idx < 0) idx = -1;
  let next = idx;
  if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
    next = (idx + 1) % vals.length;
  } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
    next = (idx - 1 + vals.length) % vals.length;
  } else {
    return;
  }
  e.preventDefault();
  const nv = vals[next];
  if (nv?.id !== undefined) {
    selections[attrId] = String(nv.id);
  }
}

// ---------- 评价：评论列表 / 点赞 / 回复 / 发表 ----------
// 评论列表接口不在白名单，匿名请求会 401；未登录时只渲染空态，避免打红控制台。
const commentQueryParams = computed(() => ({
  objectId: productId.value,
  contentType: 'CONTENT_TYPE_PRODUCT',
  page: 1,
  pageSize: 50,
  noPaging: true,
  sorting: [{ field: 'id', direction: 'DESC' }],
}));
const commentsQuery = useListComments(commentQueryParams, {
  enabled: isLogin,
});
const commentsLoading = computed(() => isLogin.value && commentsQuery.isPending.value);
const commentsError = computed(() => isLogin.value && commentsQuery.isError.value);
const topLevelComments = computed<commentservicev1_Comment[]>(() => {
  const items = (commentsQuery.data?.value as any)?.items ?? [];
  return (items as commentservicev1_Comment[]).filter((c) => !c.parentId || c.parentId === 0);
});

const allCommentIds = computed<number[]>(() => {
  const ids: number[] = [];
  const walk = (list: commentservicev1_Comment[]) => {
    for (const c of list) {
      if (c.id) ids.push(c.id);
      if (c.children && c.children.length > 0) walk(c.children as commentservicev1_Comment[]);
    }
  };
  walk(topLevelComments.value);
  return ids;
});

const commentStatusQuery = useGetInteractionStatus(
  computed(() => ({
    targetType: 'TARGET_TYPE_COMMENT',
    targetIds: allCommentIds.value,
  })),
  {
    enabled: computed(() => isLogin.value && allCommentIds.value.length > 0),
    retry: false,
  },
);
const commentCountsQuery = useGetCounts(
  computed(() => ({
    targetType: 'TARGET_TYPE_COMMENT',
    targetIds: allCommentIds.value,
    metrics: ['COUNTER_METRIC_LIKE'],
  })),
  { enabled: computed(() => allCommentIds.value.length > 0) },
);

const likeStatusMap = computed<Record<number, boolean>>(() => {
  const map: Record<number, boolean> = {};
  const statuses = (commentStatusQuery.data?.value as any)?.statuses;
  if (statuses) {
    for (const [k, v] of Object.entries(statuses)) {
      map[Number(k)] = (v as any)?.liked ?? false;
    }
  }
  return map;
});

const likeCountMap = computed<Record<number, number>>(() => {
  const map: Record<number, number> = {};
  const counts = (commentCountsQuery.data?.value as any)?.counts;
  if (counts) {
    for (const [k, v] of Object.entries(counts)) {
      const targetId = Number(k);
      const cm = v as any;
      if (cm?.counts) {
        for (const mc of cm.counts) {
          if (mc.metric === 'COUNTER_METRIC_LIKE') {
            map[targetId] = Number(mc.count) || 0;
          }
        }
      }
    }
  }
  return map;
});

const likeState = reactive<Record<number, { liked: boolean; count: number }>>({});

function commentLiked(id?: number) {
  if (!id) return false;
  return likeState[id]?.liked ?? likeStatusMap.value[id] ?? false;
}
function commentLikeCount(id?: number) {
  if (!id) return 0;
  return likeState[id]?.count ?? likeCountMap.value[id] ?? 0;
}

const likeMutation = useLike();
const unlikeMutation = useUnlike();

async function toggleCommentLike(c: commentservicev1_Comment) {
  if (!isLogin.value) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  if (!c.id) return;
  const req = { targetId: c.id, targetType: 'TARGET_TYPE_COMMENT' } as any;
  try {
    const resp = commentLiked(c.id)
      ? await unlikeMutation.mutateAsync(req)
      : await likeMutation.mutateAsync(req);
    likeState[c.id] = { liked: !!resp?.liked, count: Number(resp?.likeCount ?? 0) };
    queryClient.invalidateQueries({ queryKey: ['interactionStatus'] });
    queryClient.invalidateQueries({ queryKey: ['interactionCounts'] });
  } catch {
    toast.error(t('common.operationFailed'));
  }
}

const replyToId = ref<number | null>(null);
const replyContent = ref('');
const replyMutation = useCreateComment();

function openReply(id?: number) {
  if (!id) return;
  replyToId.value = replyToId.value === id ? null : id;
  replyContent.value = '';
}
function closeReply() {
  replyToId.value = null;
  replyContent.value = '';
}

async function submitReply(parentId?: number) {
  if (parentId === undefined) return;
  if (!isLogin.value) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  if (!replyContent.value.trim()) return;
  const req: commentservicev1_CreateCommentRequest = {
    data: {
      content: replyContent.value,
      contentType: 'CONTENT_TYPE_PRODUCT' as any,
      objectId: productId.value,
      parentId,
    } as any,
  };
  try {
    await replyMutation.mutateAsync(req);
    toast.success(t('comment.pendingApproval'));
    closeReply();
    queryClient.invalidateQueries({ queryKey: ['listComments'] });
  } catch {
    toast.error(t('common.operationFailed'));
  }
}

const postContent = ref('');
const postRating = ref(0);
const postMutation = useCreateComment({
  onSuccess: () => {
    toast.success(t('comment.pendingApproval'));
    postContent.value = '';
    postRating.value = 0;
    queryClient.invalidateQueries({ queryKey: ['listComments'] });
  },
  onError: () => {
    toast.error(t('common.operationFailed'));
  },
});

async function submitComment() {
  if (!isLogin.value) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  if (!postContent.value.trim()) return;
  if (postRating.value < 1 || postRating.value > 5) return;

  const req: commentservicev1_CreateCommentRequest = {
    data: {
      content: postContent.value,
      rating: postRating.value,
      contentType: 'CONTENT_TYPE_PRODUCT' as any,
      objectId: productId.value,
      parentId: 0,
    } as any,
  };
  await postMutation.mutateAsync(req);
}

function formatCommentTime(ts: unknown): string {
  if (!ts) return '';
  const d = new Date(String(ts));
  if (Number.isNaN(d.getTime())) return '';
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

// 滚到页尾时收起吸底购买条，否则它会盖住站点页脚的链接。
// 商品数据是客户端到达的，哨兵节点晚于 mount 才渲染，所以观察器挂在 ref 上而不是 onMounted。
const endSentinel = ref<HTMLElement | null>(null);
const showStickyBar = ref(true);
let sentinelObserver: IntersectionObserver | undefined;

watch(endSentinel, (el) => {
  sentinelObserver?.disconnect();
  if (!el || !import.meta.client) return;
  sentinelObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (entry) showStickyBar.value = !entry.isIntersecting;
    },
    { rootMargin: '0px 0px 64px 0px' },
  );
  sentinelObserver.observe(el);
});

onBeforeUnmount(() => sentinelObserver?.disconnect());
</script>

<template>
  <LayoutSectionContainer class="!pt-4 md:!pt-6">
    <nav class="mb-4 flex items-center gap-2 overflow-hidden text-sm text-muted-foreground md:mb-5">
      <NuxtLink :to="localePath('/')" class="shrink-0 transition-colors hover:text-foreground">
        {{ t('mall.category.breadcrumbHome') }}
      </NuxtLink>
      <span class="shrink-0" aria-hidden="true">/</span>
      <span class="truncate text-foreground">
        {{ productTranslation?.name || t('mall.product.notFound') }}
      </span>
    </nav>

    <div v-if="productLoading" class="grid gap-6 md:grid-cols-12 md:gap-8">
      <div class="md:col-span-5">
        <UiSkeleton class="aspect-square w-full rounded-xl" />
      </div>
      <div class="flex flex-col gap-3 md:col-span-7">
        <UiSkeleton class="h-24 w-full rounded-xl" />
        <UiSkeleton class="h-8 w-3/4" />
        <UiSkeleton class="h-4 w-1/2" />
        <UiSkeleton class="h-16 w-full rounded-xl" />
        <UiSkeleton class="h-40 w-full rounded-xl" />
      </div>
    </div>

    <UiAppEmpty v-else-if="productError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="productQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiAppEmpty v-else-if="!productTranslation" :description="t('mall.product.notFound')">
      <template #action>
        <UiButton variant="outline" size="sm" @click="navigateTo(localePath('/'))">
          {{ t('authentication.login.back_home') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <div v-else class="grid gap-6 md:grid-cols-12 md:gap-8">
      <div class="md:col-span-5">
        <div class="flex flex-col gap-3 md:sticky md:top-24">
          <div class="overflow-hidden rounded-xl border border-border bg-card shadow-card">
            <UiImage
              :src="activeImage"
              :alt="productTranslation.name || ''"
              class="aspect-square w-full object-cover"
            />
          </div>
          <div v-if="galleryImages.length > 0" class="flex flex-wrap gap-2">
            <button
              v-for="(img, i) in galleryImages"
              :key="img + i"
              type="button"
              class="size-16 shrink-0 overflow-hidden rounded-lg border bg-card transition-colors"
              :class="i === activeImageIndex ? 'border-primary' : 'border-border hover:border-primary'"
              :aria-label="`${productTranslation.name || ''} ${i + 1}`"
              :aria-current="i === activeImageIndex"
              @click="activeImageIndex = i"
            >
              <UiImage :src="img" alt="" class="size-full object-cover" />
            </button>
          </div>
        </div>
      </div>

      <div class="flex flex-col md:col-span-7">
        <div
          v-if="hasPrice"
          class="rounded-xl border border-border bg-card px-4 py-3.5 shadow-card md:px-5"
        >
          <UiPriceTag
            :amount="effectivePriceFen"
            :old-amount="hasActivePrice ? basePriceFen : null"
            size="lg"
          />
          <div class="mt-2 flex flex-wrap items-center gap-1.5">
            <UiStatusBadge v-if="activeGrouponForSku" tone="warning" size="sm">
              {{ grouponBadgeLabel }}
            </UiStatusBadge>
            <UiStatusBadge v-else-if="activeSeckillForSku" tone="danger" size="sm">
              {{ seckillBadgeLabel }}
            </UiStatusBadge>
            <UiStatusBadge v-else-if="specialPricePromo" tone="primary" size="sm">
              {{ specialBadgeLabel }}
            </UiStatusBadge>
            <UiStatusBadge v-else-if="fullReductionPromo" tone="primary" size="sm">
              {{ fullReductionLabel }}
            </UiStatusBadge>
          </div>
        </div>

        <div class="mt-4 flex items-start justify-between gap-3">
          <h1 class="min-w-0 text-xl font-bold leading-snug text-foreground md:text-2xl">
            {{ productTranslation.name }}
          </h1>
          <button
            type="button"
            class="flex size-10 shrink-0 items-center justify-center rounded-full border bg-card transition-colors"
            :class="isFavorited
              ? 'border-primary text-primary'
              : 'border-border text-muted-foreground hover:border-primary hover:text-primary'"
            :aria-label="isFavorited ? t('mall.wishlist.table.action') : t('mall.wishlist.title')"
            :title="isFavorited ? t('mall.wishlist.table.action') : t('mall.wishlist.title')"
            @click="toggleFavorite"
          >
            <XIcon icon="lucide:heart" :size="18" :class="isFavorited ? 'fill-current' : ''" />
          </button>
        </div>

        <p
          v-if="productTranslation.shortDescription"
          class="mt-2 text-sm leading-relaxed text-muted-foreground"
        >
          {{ productTranslation.shortDescription }}
        </p>

        <a href="#reviews" class="mt-3 flex w-fit flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
          <UiRatingStars v-if="hasRating" :value="ratingAverage" />
          <span v-if="hasRating">
            {{ t('mall.productRating.average') }}
            <span class="font-semibold tabular-nums text-foreground">{{ ratingAverage.toFixed(1) }}</span>
            · {{ t('mall.productRating.count') }}
            <span class="tabular-nums text-foreground">{{ ratingCount }}</span>
          </span>
          <span v-else>{{ t('mall.productRating.noRating') }}</span>
        </a>

        <div class="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 border-y border-border py-3">
          <span
            v-for="service in servicePromises"
            :key="service.label"
            class="flex items-center gap-1.5 text-xs text-muted-foreground"
          >
            <XIcon :icon="service.icon" :size="14" class="text-primary" />
            {{ service.label }}
          </span>
        </div>

        <div class="mt-4 rounded-xl border border-border bg-card p-4 shadow-card md:p-5">
          <p class="text-sm font-medium text-foreground">{{ t('mall.product.selectSku') }}</p>

          <div v-if="selectableAttributes.length > 0" class="mt-3 flex flex-col gap-3">
            <div v-for="attr in selectableAttributes" :key="attr.id" class="flex flex-col gap-2">
              <span class="text-xs text-muted-foreground">
                {{ pickTranslation(attr.translations)?.name || '—' }}
              </span>
              <div
                role="radiogroup"
                :aria-label="pickTranslation(attr.translations)?.name || ''"
                class="flex flex-wrap gap-2"
                @keydown="onRadioKeydown($event, attr.id!)"
              >
                <button
                  v-for="val in valuesByAttribute.get(attr.id!) ?? []"
                  :key="val.id"
                  type="button"
                  role="radio"
                  :aria-checked="selections[attr.id!] === String(val.id)"
                  :class="cn(
                    'min-w-16 rounded-lg border px-4 py-2 text-sm transition-colors',
                    selections[attr.id!] === String(val.id)
                      ? 'border-primary bg-primary-soft font-medium text-primary-soft-foreground'
                      : 'border-border bg-background text-foreground hover:border-primary',
                  )"
                  @click="selections[attr.id!] = String(val.id)"
                >
                  {{ pickTranslation(val.translations)?.displayName || '—' }}
                </button>
              </div>
            </div>
          </div>

          <p v-if="allAttributesSelected && !matchedSku" class="mt-3 text-sm text-destructive">
            {{ t('mall.product.skuUnavailable') }}
          </p>
          <p v-else-if="matchedSku && isOutOfStock" class="mt-3 text-sm text-destructive">
            {{ t('mall.product.outOfStock') }}
          </p>

          <div class="mt-4 flex flex-wrap items-center gap-3">
            <div class="flex items-center rounded-lg border border-border bg-background">
              <button
                type="button"
                class="flex size-10 items-center justify-center text-foreground transition-colors hover:text-primary disabled:opacity-40"
                :disabled="quantity <= 1"
                aria-label="decrease quantity"
                @click="decrementQty"
              >
                <XIcon icon="carbon:subtract" :size="16" />
              </button>
              <input
                v-model.number="quantity"
                type="number"
                class="h-10 w-12 bg-transparent text-center text-sm tabular-nums text-foreground outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none"
                :min="1"
                :max="matchedSkuStock"
                readonly
              >
              <button
                type="button"
                class="flex size-10 items-center justify-center text-foreground transition-colors hover:text-primary disabled:opacity-40"
                :disabled="quantity >= matchedSkuStock"
                aria-label="increase quantity"
                @click="incrementQty"
              >
                <XIcon icon="carbon:add" :size="16" />
              </button>
            </div>
          </div>

          <div class="mt-4 hidden items-center gap-3 md:flex">
            <UiButton
              variant="outline"
              size="lg"
              class="flex-1"
              :disabled="!canAddToCart"
              @click="addToCartWithQty(quantity)"
            >
              <XIcon icon="carbon:shopping-cart" :size="16" class="mr-2" />
              {{ t('mall.product.addToCart') }}
            </UiButton>
            <UiButton size="lg" class="flex-1" :disabled="!canAddToCart" @click="buyNow">
              {{ t('mall.product.buyNow') }}
            </UiButton>
          </div>
        </div>
      </div>
    </div>
  </LayoutSectionContainer>

  <LayoutSectionContainer v-if="productTranslation?.longDescription">
    <UiSectionHeader :title="t('mall.product.description')" />
    <div class="rounded-xl border border-border bg-card p-4 shadow-card md:p-6">
      <ClientOnly>
        <ContentViewer :content="productTranslation.longDescription" type="markdown" />
        <template #fallback>
          <UiSkeleton class="h-24 w-full" />
        </template>
      </ClientOnly>
    </div>
  </LayoutSectionContainer>

  <LayoutSectionContainer
    v-if="productTranslation"
    id="reviews"
    class="scroll-mt-24 border-t border-border"
  >
    <UiSectionHeader
      :title="t('comment.sectionTitle')"
      :description="hasRating
        ? `${t('mall.productRating.average')} · ${t('mall.productRating.count')} ${ratingCount}`
        : t('mall.productRating.noRating')"
    />

    <div
      v-if="hasRating || reviewSummary"
      class="flex flex-col gap-4 rounded-xl border border-border bg-card p-4 shadow-card md:flex-row md:items-center md:gap-6 md:p-5"
    >
      <div v-if="hasRating" class="flex items-center gap-3 md:min-w-56">
        <span class="flex items-baseline gap-1 text-foreground">
          <span class="text-3xl font-bold tabular-nums">{{ ratingAverage.toFixed(1) }}</span>
          <span class="text-sm text-muted-foreground">/ 5</span>
        </span>
        <span class="flex flex-col gap-1">
          <UiRatingStars :value="ratingAverage" />
          <span class="text-xs text-muted-foreground">
            {{ t('mall.productRating.average') }} · {{ t('mall.productRating.count') }} {{ ratingCount }}
          </span>
        </span>
      </div>

      <div
        v-if="reviewSummary"
        class="flex items-start gap-2 rounded-lg bg-secondary px-3 py-2.5 md:flex-1"
      >
        <XIcon icon="lucide:sparkles" :size="14" class="mt-0.5 shrink-0 text-primary" />
        <p class="text-xs leading-relaxed text-muted-foreground">
          <span class="font-medium text-foreground">{{ t('mall.productRating.aiSummary') }}：</span>{{ reviewSummary }}
        </p>
      </div>
    </div>

    <div v-if="commentsLoading" class="mt-3">
      <UiListSkeleton :rows="3" :header="false" />
    </div>

    <div v-else-if="commentsError" class="mt-3 rounded-xl border border-border bg-card">
      <UiAppEmpty variant="error" class="py-8">
        <template #action>
          <UiButton variant="outline" size="sm" @click="commentsQuery.refetch()">
            {{ t('ui.button.retry') }}
          </UiButton>
        </template>
      </UiAppEmpty>
    </div>

    <div v-else-if="topLevelComments.length === 0" class="mt-3 rounded-xl border border-border bg-card">
      <UiAppEmpty variant="noData" :description="t('comment.empty')" class="py-9" />
    </div>

    <ul v-else class="mt-3 divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
      <li v-for="c in topLevelComments" :key="c.id" class="p-4 md:p-5">
        <div class="flex items-start gap-3">
          <span class="flex size-9 shrink-0 items-center justify-center rounded-full bg-secondary text-muted-foreground">
            <XIcon icon="lucide:user" :size="16" />
          </span>
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
              <span class="text-sm font-medium text-foreground">
                {{ t('comment.author') }} #{{ c.createdBy ?? '?' }}
              </span>
              <span class="text-xs text-muted-foreground">{{ formatCommentTime(c.createdAt) }}</span>
            </div>

            <div v-if="c.rating && c.rating > 0" class="mt-1">
              <UiRatingStars :value="c.rating" :size="12" />
            </div>

            <p class="mt-1.5 whitespace-pre-wrap break-words text-sm leading-relaxed text-foreground">
              {{ c.content || '' }}
            </p>

            <div class="mt-2 flex items-center gap-4">
              <button
                type="button"
                class="inline-flex items-center gap-1 text-xs transition-colors"
                :class="commentLiked(c.id) ? 'text-destructive' : 'text-muted-foreground hover:text-primary'"
                :aria-pressed="commentLiked(c.id)"
                @click="toggleCommentLike(c)"
              >
                <XIcon icon="lucide:heart" :size="14" :class="commentLiked(c.id) ? 'fill-current' : ''" />
                <span class="tabular-nums">{{ commentLikeCount(c.id) }}</span>
              </button>
              <button
                v-if="isLogin"
                type="button"
                class="text-xs text-muted-foreground transition-colors hover:text-primary"
                @click="openReply(c.id)"
              >
                {{ t('comment.reply') }}
              </button>
            </div>

            <div v-if="replyToId === c.id" class="mt-3 flex flex-col gap-2">
              <UiTextarea v-model="replyContent" :placeholder="t('comment.replyPlaceholder')" rows="2" />
              <div class="flex justify-end gap-2">
                <UiButton variant="ghost" size="sm" @click="closeReply">
                  {{ t('common.cancel') }}
                </UiButton>
                <UiButton
                  size="sm"
                  :disabled="replyMutation.isPending.value || !replyContent.trim()"
                  @click="submitReply(c.id)"
                >
                  {{ replyMutation.isPending.value ? t('common.submitting') : t('comment.submit') }}
                </UiButton>
              </div>
            </div>

            <ul
              v-if="c.children && c.children.length > 0"
              class="mt-3 flex flex-col gap-3 border-l border-border pl-3 md:pl-4"
            >
              <li v-for="child in c.children" :key="child.id" class="min-w-0">
                <div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                  <span class="text-xs font-medium text-foreground">
                    {{ t('comment.author') }} #{{ child.createdBy ?? '?' }}
                  </span>
                  <span class="text-xs text-muted-foreground">{{ formatCommentTime(child.createdAt) }}</span>
                </div>
                <p class="mt-1 whitespace-pre-wrap break-words text-sm leading-relaxed text-foreground">
                  {{ child.content || '' }}
                </p>
                <button
                  type="button"
                  class="mt-1.5 inline-flex items-center gap-1 text-xs transition-colors"
                  :class="commentLiked(child.id) ? 'text-destructive' : 'text-muted-foreground hover:text-primary'"
                  :aria-pressed="commentLiked(child.id)"
                  @click="toggleCommentLike(child)"
                >
                  <XIcon icon="lucide:heart" :size="14" :class="commentLiked(child.id) ? 'fill-current' : ''" />
                  <span class="tabular-nums">{{ commentLikeCount(child.id) }}</span>
                </button>
              </li>
            </ul>
          </div>
        </div>
      </li>
    </ul>

    <div
      v-if="isLogin"
      class="mt-4 rounded-xl border border-border bg-card p-4 shadow-card md:p-5"
    >
      <div class="flex flex-wrap items-center gap-3">
        <span class="text-xs text-muted-foreground">{{ t('mall.productRating.selectRating') }}</span>
        <UiRatingInput v-model="postRating" />
      </div>
      <UiTextarea
        v-model="postContent"
        :placeholder="t('comment.postPlaceholder')"
        class="mt-3"
        rows="3"
      />
      <div class="mt-3 flex justify-end">
        <UiButton
          size="sm"
          :disabled="postMutation.isPending.value || !postContent.trim() || postRating < 1 || postRating > 5"
          @click="submitComment"
        >
          {{ postMutation.isPending.value ? t('common.submitting') : t('comment.submit') }}
        </UiButton>
      </div>
    </div>

    <div
      v-else
      class="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-card px-4 py-4"
    >
      <p class="text-sm text-muted-foreground">{{ t('comment.loginToPost') }}</p>
      <UiButton variant="outline" size="sm" @click="navigateTo(localePath('/login'))">
        {{ t('authentication.login.login') }}
      </UiButton>
    </div>
  </LayoutSectionContainer>

  <div v-if="productTranslation" ref="endSentinel" class="h-px w-full md:hidden" aria-hidden="true" />

  <UiStickyBar v-if="productTranslation && showStickyBar" above-tab-bar class="md:hidden">
    <button
      type="button"
      class="flex size-11 shrink-0 items-center justify-center rounded-full border bg-card transition-colors"
      :class="isFavorited
        ? 'border-primary text-primary'
        : 'border-border text-muted-foreground active:border-primary active:text-primary'"
      :aria-label="isFavorited ? t('mall.wishlist.table.action') : t('mall.wishlist.title')"
      @click="toggleFavorite"
    >
      <XIcon icon="lucide:heart" :size="18" :class="isFavorited ? 'fill-current' : ''" />
    </button>
    <UiButton
      variant="outline"
      class="flex-1"
      :disabled="!canAddToCart"
      @click="addToCartWithQty(quantity)"
    >
      <XIcon icon="carbon:shopping-cart" :size="16" class="mr-2" />
      {{ t('mall.product.addToCart') }}
    </UiButton>
    <UiButton class="flex-1" :disabled="!canAddToCart" @click="buyNow">
      {{ t('mall.product.buyNow') }}
    </UiButton>
  </UiStickyBar>
</template>
