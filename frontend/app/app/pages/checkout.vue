<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { toast } from 'vue-sonner';
import { cn } from '@/lib/utils';
import { XIcon } from '@/plugins/xicon';
import {
  useListCarts,
  useListCartItems,
  useCreateOrder,
  useCreatePaymentTransaction,
  fetchGetOrderByIdempotencyKey,
  fetchListSkuPricesStore,
  fetchListSkuAttributeCombinationsStore,
  useListProductAttributes,
  useListProductAttributeValues,
  fetchGetSkuStore,
  fetchGetProductStore,
  useListShippingAddresses,
} from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';
import { queryClient } from '@/plugins/vue-query';
import type {
  orderservicev1_Order,
  paymentservicev1_PaymentTransaction,
  paymentservicev1_PaymentMethod,
  paymentservicev1_BusinessType,
} from '@/api/generated/app/service/v1';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.checkout.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
const currentUserId = computed(() => userStore.user?.id ?? 0);

// ---------- 购物车快照（用于订单摘要） ----------
// 注意：下单所需 tenantId/userId 必须取自购物车实体本身（即购物车数据行的
// 真实归属租户与用户），而不是 userStore 里的登录态 tenantId/userId。
// 后端 Order.Create 用 (tenant_id, user_id) 复合查购物车，若两值与购物车
// 实际归属不一致即报 "cart is empty or not found"。
type CartEntity = { id?: number; userId?: number; tenantId?: number };
type CartItemEntity = { id?: number; skuId?: number; quantity?: number };

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
const cartId = computed(() => cart.value?.id);
// 购物车真实归属（下单时以此为准，避免与登录态租户错配）。
const cartUserId = computed(() => cart.value?.userId ?? 0);
const cartTenantId = computed(() => cart.value?.tenantId ?? 0);

const cartItemsQuery = useListCartItems(
  computed(() => ({
    page: 1,
    pageSize: 100,
    noPaging: false,
    query: cartId.value === undefined ? undefined : JSON.stringify({ cartId: cartId.value }),
  })),
  { enabled: computed(() => isLogin.value && cartId.value !== undefined) },
);
const cartItems = computed<CartItemEntity[]>(() => {
  const items = (cartItemsQuery.data?.value as any)?.items ?? [];
  return (items as CartItemEntity[]) ?? [];
});
const itemsLoading = computed(() => cartItemsQuery.isPending.value);
const itemsError = computed(() => cartItemsQuery.isError.value);

// ---------- 属性名 / 属性值 displayName 反查 map ----------
// 全量拉取属性与属性值，建立 id→name / id→displayName 映射，
// 用于把 SKU 组合的 attributeId/attributeValueId 翻译成人类可读的规格描述。
const currentLocale = computed(() => getCurrentLocale());

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined;
  const match = translations.find((tr) => tr.languageCode === currentLocale.value);
  return match ?? translations[0];
}

const attributesQuery = useListProductAttributes({
  page: 1,
  pageSize: 100,
  noPaging: false,
});
const attributeValuesQuery = useListProductAttributeValues({
  page: 1,
  pageSize: 200,
  noPaging: false,
});

// attributeNameMap: attributeId → 属性名（如 "颜色"）
const attributeNameMap = computed(() => {
  const items = ((attributesQuery.data?.value as any)?.items ?? []) as Array<{
    id?: number;
    translations?: Array<{ name?: string; languageCode?: string }>;
  }>;
  const map = new Map<number, string>();
  for (const a of items) {
    if (a.id === undefined) continue;
    const tr = pickTranslation(a.translations);
    if (tr?.name) map.set(a.id, tr.name);
  }
  return map;
});

// attributeValueDisplayNameMap: attributeValueId → 属性值 displayName（如 "黑色"）
const attributeValueDisplayNameMap = computed(() => {
  const items = ((attributeValuesQuery.data?.value as any)?.items ?? []) as Array<{
    id?: number;
    attributeId?: number;
    translations?: Array<{ displayName?: string; languageCode?: string }>;
  }>;
  const map = new Map<number, string>();
  for (const v of items) {
    if (v.id === undefined) continue;
    const tr = pickTranslation(v.translations);
    if (tr?.displayName) map.set(v.id, tr.displayName);
  }
  return map;
});

// ---------- 每个 SKU 的价格（通过 fetchListSkuPricesStore 在 watch 中拉取） ----------
// skuPricesMap: skuId → amount（string，CNY）
const skuPricesMap = reactive<Record<number, string>>({});

watch(
  cartItems,
  async (items) => {
    for (const k of Object.keys(skuPricesMap)) delete skuPricesMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const skuId = item.skuId;
        if (skuId === undefined) return;
        try {
          const resp: any = await fetchListSkuPricesStore({
            page: 1,
            pageSize: 10,
            noPaging: false,
            query: JSON.stringify({ skuId }),
          });
          const prices = (resp?.items ?? []) as Array<{
            skuId?: number;
            currency?: string;
            amount?: string;
          }>;
          const cny = prices.find((p) => p.currency === 'CNY') ?? prices[0];
          if (cny?.amount) {
            skuPricesMap[skuId] = cny.amount;
          }
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

// ---------- 每个 SKU 的规格组合（用于订单摘要显示规格描述） ----------
// skuCombosMap: skuId → [{attrId, valId}, ...]
const skuCombosMap = reactive<Record<number, Array<{ attrId: number; valId: number }>>>({});

watch(
  cartItems,
  async (items) => {
    for (const k of Object.keys(skuCombosMap)) delete skuCombosMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const skuId = item.skuId;
        if (skuId === undefined) return;
        try {
          const resp: any = await fetchListSkuAttributeCombinationsStore({
            page: 1,
            pageSize: 100,
            noPaging: false,
            query: JSON.stringify({ skuId }),
          });
          const combos = (resp?.items ?? []) as Array<{
            skuId?: number;
            attributeId?: number;
            attributeValueId?: number;
          }>;
          const list: Array<{ attrId: number; valId: number }> = [];
          for (const c of combos) {
            if (c.attributeId !== undefined && c.attributeValueId !== undefined) {
              list.push({ attrId: c.attributeId, valId: c.attributeValueId });
            }
          }
          if (list.length > 0) skuCombosMap[skuId] = list;
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

// ---------- 商品名 / 商品图片：skuId → SKU → productId → Product ----------
// productInfoMap: skuId → { name, imageUrl }
const productInfoMap = reactive<Record<number, { name: string; imageUrl: string }>>({});

watch(
  cartItems,
  async (items) => {
    for (const k of Object.keys(productInfoMap)) delete productInfoMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const skuId = item.skuId;
        if (skuId === undefined) return;
        try {
          const sku: any = await fetchGetSkuStore(skuId);
          const productId = sku?.productId;
          if (productId === undefined) return;
          const product: any = await fetchGetProductStore(productId);
          const imageUrl = product?.imageUrl ?? '';
          const tr = pickTranslation(
            product?.translations as Array<{ name?: string; languageCode?: string }> | undefined,
          );
          const name = tr?.name ?? '';
          productInfoMap[skuId] = { name, imageUrl };
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

function productName(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  return productInfoMap[skuId]?.name ?? '';
}
function productImageUrl(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  return productInfoMap[skuId]?.imageUrl ?? '';
}

// 将单个 SKU 的规格组合翻译成人类可读描述，如 "颜色: 黑色，容量: 64GB"。
// 任何一项属性名或属性值查不到时返回空串，避免显示残缺信息。
function describeSku(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  const combos = skuCombosMap[skuId];
  if (!combos || combos.length === 0) return '';
  const parts: string[] = [];
  for (const { attrId, valId } of combos) {
    const attrName = attributeNameMap.value.get(attrId);
    const valName = attributeValueDisplayNameMap.value.get(valId);
    if (!attrName || !valName) return '';
    parts.push(`${attrName}: ${valName}`);
  }
  return parts.join('，');
}

// ---------- 合计金额：累加每个 cart item 的 SKU 价格 × 数量 ----------
// 后端 SKU 价格 amount 以「分」为单位整型存储，展示时需 /100 转为元。
const totalAmount = computed(() => {
  let sum = 0;
  for (const item of cartItems.value) {
    const skuId = item.skuId;
    if (skuId === undefined) continue;
    const amountStr = skuPricesMap[skuId];
    if (!amountStr) continue;
    const priceCents = parseFloat(amountStr);
    if (Number.isNaN(priceCents)) continue;
    const priceYuan = priceCents / 100;
    const qty = item.quantity ?? 0;
    sum += priceYuan * qty;
  }
  return sum.toFixed(2);
});
const totalLabel = computed(() => `${t('mall.product.currencyCny')}${totalAmount.value}`);

// ---------- 收货表单 ----------
const form = reactive({
  recipientName: '',
  recipientPhone: '',
  shippingAddress: '',
});

const formValid = computed(() => {
  return (
    form.recipientName.trim().length > 0 &&
    form.recipientPhone.trim().length > 0 &&
    form.shippingAddress.trim().length > 0
  );
});

// ---------- 地址簿快捷填充 ----------
// 下单时可从已存地址簿选择一条，自动填充收货表单。无地址时该区域不显示。
type SavedAddress = {
  id?: number;
  recipientName?: string;
  recipientPhone?: string;
  region?: string;
  detailAddress?: string;
  isDefault?: boolean;
};
const savedAddressesQuery = useListShippingAddresses(
  computed(() => ({
    page: 1,
    pageSize: 20,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
  { enabled: isLogin },
);
const savedAddresses = computed<SavedAddress[]>(() => {
  const items = (savedAddressesQuery.data?.value as any)?.items ?? [];
  return (items as SavedAddress[]) ?? [];
});
const sortedSavedAddresses = computed(() => {
  return [...savedAddresses.value].sort((a, b) => {
    if (a.isDefault && !b.isDefault) return -1;
    if (!a.isDefault && b.isDefault) return 1;
    return 0;
  });
});

function useSavedAddress(addr: SavedAddress) {
  form.recipientName = addr.recipientName ?? '';
  form.recipientPhone = addr.recipientPhone ?? '';
  // region + detailAddress 拼成结构化地址文本（与订单 shipping_address 字段一致）
  const parts = [addr.region, addr.detailAddress].filter((s) => s && s.trim()).map((s) => s!.trim());
  form.shippingAddress = parts.join(' ');
}

function addressLabel(addr: SavedAddress): string {
  const region = addr.region ? addr.region + ' ' : '';
  return `${addr.recipientName ?? '—'} · ${region}${addr.detailAddress ?? ''}`;
}

// ---------- 下单 + 支付 ----------
const orderMutation = useCreateOrder({
  onSuccess: () => {
    toast.success(t('checkout.orderCreated'));
  },
  onError: (err: any) => {
    toast.error(err?.message || t('checkout.errors.orderFailed'));
  },
});
const paymentMutation = useCreatePaymentTransaction({
  onError: (err: any) => {
    toast.error(err?.message || t('checkout.errors.paymentFailed'));
  },
});

async function placeOrder() {
  if (!isLogin.value) {
    navigateTo(localePath('/login'));
    return;
  }
  if (!formValid.value) {
    toast.error(t('checkout.errors.invalidForm'));
    return;
  }
  if (cartItems.value.length === 0) {
    toast.error(t('checkout.errors.emptyCart'));
    return;
  }

  // 为本次下单生成一对幂等键与业务单号。
  // idempotency_key：订单与支付各一个，防止重放导致重复下单/重复扣款。
  // business_ref_id：跨域对账键，订单与支付共用同一值，便于后续对账。
  // 直接通过 crypto.randomUUID() 调用，保持方法在 crypto 对象上的 this 绑定，
  // 避免取出方法引用裸调导致 "Illegal invocation"。
  const cryptoObj = (globalThis as unknown as { crypto?: { randomUUID?: () => string } }).crypto;
  const orderIdempotencyKey = cryptoObj?.randomUUID?.() ?? '';
  const paymentIdempotencyKey = cryptoObj?.randomUUID?.() ?? '';
  const businessRefId = cryptoObj?.randomUUID?.() ?? '';
  if (!orderIdempotencyKey || !paymentIdempotencyKey || !businessRefId) {
    toast.error(t('checkout.errors.orderFailed'));
    return;
  }

  const orderData: orderservicev1_Order = {
    userId: cartUserId.value,
    tenantId: cartTenantId.value,
    recipientName: form.recipientName,
    recipientPhone: form.recipientPhone,
    shippingAddress: form.shippingAddress,
    currency: 'CNY',
    totalAmount: 0,
    idempotencyKey: orderIdempotencyKey,
    businessRefId: businessRefId,
  } as orderservicev1_Order;

  try {
    await orderMutation.mutateAsync(orderData);
  } catch {
    return;
  }

  // 订单创建后回查以拿到 orderId 与后端计算的真实 totalAmount。
  // Order.Create 返回 empty，只能通过 (idempotency_key, tenant_id) 反查。
  let createdOrder: orderservicev1_Order | null = null;
  try {
    createdOrder = (await fetchGetOrderByIdempotencyKey(
      orderIdempotencyKey,
      cartTenantId.value,
    )) as orderservicev1_Order;
  } catch {
    toast.error(t('checkout.errors.orderFailed'));
    return;
  }

  const orderId = createdOrder?.id;
  const realAmount = createdOrder?.totalAmount ?? 0;
  if (!orderId || realAmount <= 0) {
    toast.error(t('checkout.errors.orderFailed'));
    return;
  }

  // 余额支付（暂固定为 BALANCE，待后端支持多支付方式后扩展选择 UI）
  const paymentData: paymentservicev1_PaymentTransaction = {
    userId: cartUserId.value,
    tenantId: cartTenantId.value,
    orderId: orderId,
    amount: realAmount,
    currency: 'CNY',
    paymentMethod: 'BALANCE' as paymentservicev1_PaymentMethod,
    businessType: 'BUSINESS_TYPE_CONSUME' as paymentservicev1_BusinessType,
    idempotencyKey: paymentIdempotencyKey,
    businessRefId: businessRefId,
  } as paymentservicev1_PaymentTransaction;

  try {
    await paymentMutation.mutateAsync(paymentData);
    // 订单事务已在后端清空购物车（扣库存+清 cart_item 同事务），
    // 此处仅需刷新前端缓存以反映空车状态。
    queryClient.invalidateQueries({ queryKey: ['listCarts'] });
    queryClient.invalidateQueries({ queryKey: ['listCartItems'] });
    // 刷新订单列表缓存，使跳转 /orders 后立即显示新订单（默认 staleTime 会导致
    // 5 分钟内不重拉，需显式 invalidate）。
    queryClient.invalidateQueries({ queryKey: ['listOrders'] });
    toast.success(t('checkout.paymentSuccess'));
    navigateTo(localePath('/orders'));
  } catch {
    // 错误已由 onError 处理
  }
}
</script>

<template>
  <!-- 交易进度条：1.购物车 → 2.核对订单(当前) → 3.线上支付 → 4.完成 -->
  <LayoutSectionContainer class="!py-4">
    <ol class="flex items-center justify-center gap-2 md:gap-4">
      <li
        v-for="(step, idx) in [
          { key: 'cart', label: t('checkout.steps.cart') },
          { key: 'confirm', label: t('checkout.steps.confirm') },
          { key: 'pay', label: t('checkout.steps.pay') },
          { key: 'done', label: t('checkout.steps.done') },
        ]"
        :key="step.key"
        class="flex items-center gap-2 md:gap-4"
      >
        <!-- 步骤节点：序号圆 + 标签 -->
        <div class="flex flex-col items-center gap-1">
          <span
            :class="cn(
              'flex h-7 w-7 items-center justify-center rounded-full border text-xs font-bold transition-colors',
              idx === 1
                ? 'border-primary bg-primary/10 text-primary dark:border-green-500 dark:bg-green-500/10 dark:text-green-400'
                : 'border-border text-muted-foreground',
            )"
          >
            {{ idx + 1 }}
          </span>
          <span
            :class="cn(
              'text-[10px] font-medium transition-colors',
              idx === 1
                ? 'text-primary dark:text-green-400'
                : 'text-muted-foreground',
            )"
          >
            {{ step.label }}
          </span>
        </div>
        <!-- 连接线（最后一步不渲染） -->
        <span
          v-if="idx < 3"
          :class="cn(
            'h-px w-8 md:w-16 transition-colors',
            idx < 1
              ? 'bg-primary dark:bg-green-500'
              : 'bg-border',
          )"
        ></span>
      </li>
    </ol>
  </LayoutSectionContainer>

  <LayoutSectionContainer>
    <!-- 未登录 -->
    <div
      v-if="!isLogin"
      class="flex flex-col items-center gap-6 rounded-2xl border border-border bg-card p-12 text-center"
    >
      <XIcon icon="carbon:locked" :size="48" class="text-muted-foreground" />
      <p class="text-lg text-muted-foreground">{{ t('authentication.login.please_login') }}</p>
      <UiButton @click="navigateTo(localePath('/login'))">
        {{ t('navbar.user.login') }}
      </UiButton>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-[1fr_380px]">
      <!-- 收货信息 -->
      <div class="rounded-2xl border border-border bg-card p-6">
        <h2 class="mb-1 text-xl font-bold text-foreground">{{ t('checkout.recipientTitle') }}</h2>
        <p class="mb-6 text-sm text-muted-foreground">{{ t('checkout.recipientDesc') }}</p>

        <!-- 地址簿快捷选择区：有地址则列按钮可点击填充；无地址则显示空态提示 -->
        <h3 class="mb-3 text-sm font-semibold text-foreground">{{ t('checkout.useSavedAddress') }}</h3>
        <div v-if="sortedSavedAddresses.length > 0" class="mb-5 flex flex-wrap gap-2">
          <button
            v-for="addr in sortedSavedAddresses"
            :key="addr.id"
            type="button"
            class="flex items-center gap-2 rounded-lg border border-border bg-background/40 px-3 py-2 text-xs text-foreground transition-colors hover:border-primary/60"
            @click="useSavedAddress(addr)"
          >
            <span
              v-if="addr.isDefault"
              class="rounded-full bg-primary/15 px-1.5 py-0.5 text-[9px] font-medium text-primary"
            >
              {{ t('addresses.defaultBadge') }}
            </span>
            <span>{{ addressLabel(addr) }}</span>
          </button>
        </div>
        <UiAppEmpty
          v-else
          variant="noData"
          :description="t('checkout.noSavedAddress')"
        />

        <div class="flex flex-col gap-5">
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('checkout.recipientName') }}</UiLabel>
            <UiInput
              v-model="form.recipientName"
              type="text"
              autocomplete="name"
              :placeholder="t('checkout.recipientNamePlaceholder')"
            />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('checkout.recipientPhone') }}</UiLabel>
            <UiInput
              v-model="form.recipientPhone"
              type="tel"
              autocomplete="tel"
              :placeholder="t('checkout.recipientPhonePlaceholder')"
            />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('checkout.shippingAddress') }}</UiLabel>
            <UiTextarea
              v-model="form.shippingAddress"
              autocomplete="street-address"
              :placeholder="t('checkout.shippingAddressPlaceholder')"
              class="min-h-[96px]"
            />
          </div>
        </div>
      </div>

      <!-- 订单摘要 -->
      <div class="rounded-2xl border border-border bg-card p-6 lg:sticky lg:top-24 lg:self-start">
        <h2 class="mb-4 text-xl font-bold text-foreground">{{ t('checkout.summary') }}</h2>

        <div v-if="itemsLoading" class="flex flex-col gap-3">
          <div v-for="i in 3" :key="i" class="flex items-center gap-3 rounded-md border border-border bg-background/40 p-3">
            <UiSkeleton class="h-10 w-10 shrink-0 rounded" />
            <UiSkeleton class="h-4 flex-1" />
            <UiSkeleton class="h-4 w-6" />
          </div>
        </div>

        <UiAppEmpty
          v-else-if="itemsError"
          variant="error"
        >
          <template #action>
            <UiButton variant="outline" size="sm" @click="cartItemsQuery.refetch()">
              {{ t('ui.button.retry') }}
            </UiButton>
          </template>
        </UiAppEmpty>

        <template v-else>
          <UiAppEmpty
            v-if="cartItems.length === 0"
            variant="noData"
            :description="t('cart.empty')"
          >
            <template #action>
              <UiButton variant="outline" size="sm" @click="navigateTo(localePath('/'))">
                {{ t('cart.continueShopping') }}
              </UiButton>
            </template>
          </UiAppEmpty>

          <ul v-else class="flex flex-col gap-3">
            <li
              v-for="item in cartItems"
              :key="item.id"
              class="flex items-center gap-3 rounded-md border border-border bg-background/40 p-3"
            >
              <img
                v-if="productImageUrl(item.skuId)"
                :src="productImageUrl(item.skuId)"
                :alt="productName(item.skuId)"
                class="h-10 w-10 shrink-0 rounded object-cover"
              />
              <UiProductPlaceholder
                v-else
                :seed="item.skuId ?? 0"
                class="h-10 w-10 shrink-0 rounded text-muted-foreground"
              />
              <div class="min-w-0 flex-1">
                <p class="line-clamp-1 text-xs text-foreground">
                  {{ productName(item.skuId) || '—' }}
                </p>
                <p
                  v-if="describeSku(item.skuId)"
                  class="mt-0.5 line-clamp-1 text-[10px] text-muted-foreground"
                >
                  {{ describeSku(item.skuId) }}
                </p>
              </div>
              <span class="text-xs tabular-nums text-muted-foreground">×{{ item.quantity ?? 0 }}</span>
            </li>
          </ul>
        </template>

        <UiSeparator class="my-5" />

        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">{{ t('cart.total') }}</span>
          <span class="text-xl font-bold text-primary">{{ totalLabel }}</span>
        </div>
        <p class="mt-1 text-[10px] text-muted-foreground">{{ t('cart.totalNote') }}</p>

        <UiButton
          class="mt-5 w-full"
          size="lg"
          :disabled="!formValid || cartItems.length === 0 || orderMutation.isPending.value || paymentMutation.isPending.value"
          @click="placeOrder"
        >
          {{ t('checkout.placeOrder') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
