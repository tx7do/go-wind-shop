<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed, reactive, ref } from 'vue';
import { toast } from 'vue-sonner';
import { XIcon } from '@/plugins/xicon';
import {
  useListShippingAddresses,
  useCreateShippingAddress,
  useUpdateShippingAddress,
  useDeleteShippingAddress,
  invalidateShippingAddresses,
} from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';

const { t } = useI18n();

useHead({ title: t('addresses.title') });

const accessStore = useAccessStore();
const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

type AddressEntity = {
  id?: number;
  recipientName?: string;
  recipientPhone?: string;
  region?: string;
  detailAddress?: string;
  postalCode?: string;
  tag?: string;
  isDefault?: boolean;
};

// 列表查询：后端 UserPrivacy 策略 + 网关注入 userId 做行级隔离，
// 前端 query 无需带 userId，仅返回当前用户自己的地址。
// 注：user_id 行级隔离由后端 UserPrivacy 策略 + 网关注入 recipientUserId 强制，
// 前端无需带 userId。enabled 守卫：未登录时不发请求，避免预 hydrate 闪空。
const addressesQuery = useListShippingAddresses(
  computed(() => ({
    page: 1,
    pageSize: 50,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
  { enabled: isLogin },
);
const addresses = computed<AddressEntity[]>(() => {
  const items = (addressesQuery.data?.value as any)?.items ?? [];
  return (items as AddressEntity[]) ?? [];
});
const loading = computed(() => addressesQuery.isPending.value);
const loadError = computed(() => addressesQuery.isError.value);

// 默认地址置顶显示
const sortedAddresses = computed(() => {
  return [...addresses.value].sort((a, b) => {
    if (a.isDefault && !b.isDefault) return -1;
    if (!a.isDefault && b.isDefault) return 1;
    return 0;
  });
});

// ---------- 表单（新增/编辑） ----------
const showForm = ref(false);
const editingId = ref<number | null>(null);

const emptyForm = () => ({
  recipientName: '',
  recipientPhone: '',
  region: '',
  detailAddress: '',
  postalCode: '',
  tag: '',
  isDefault: false,
});
const form = reactive(emptyForm());

const formTitle = computed(() =>
  editingId.value === null ? t('addresses.form.titleAdd') : t('addresses.form.titleEdit'),
);

const formValid = computed(() => {
  const name = form.recipientName.trim();
  const phone = form.recipientPhone.trim();
  const addr = form.detailAddress.trim();
  const postal = form.postalCode.trim();

  // 基本非空 + 长度上限（防超长写入）
  if (name.length === 0 || name.length > 32) return false;
  if (addr.length === 0 || addr.length > 200) return false;

  // 中国大陆手机号：11 位、1 开头
  if (!/^1[3-9]\d{9}$/.test(phone)) return false;

  // 邮编可选；若填则必须 6 位数字
  if (postal.length > 0 && !/^\d{6}$/.test(postal)) return false;

  return true;
});

function openCreate() {
  editingId.value = null;
  Object.assign(form, emptyForm());
  showForm.value = true;
}

function openEdit(addr: AddressEntity) {
  editingId.value = addr.id ?? null;
  Object.assign(form, emptyForm(), {
    recipientName: addr.recipientName ?? '',
    recipientPhone: addr.recipientPhone ?? '',
    region: addr.region ?? '',
    detailAddress: addr.detailAddress ?? '',
    postalCode: addr.postalCode ?? '',
    tag: addr.tag ?? '',
    isDefault: addr.isDefault ?? false,
  });
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
}

const createMutation = useCreateShippingAddress({
  onSuccess: () => {
    toast.success(t('addresses.result.created'));
    invalidateShippingAddresses();
    closeForm();
  },
  onError: (err: any) => toast.error(err?.message || t('addresses.errors.createFailed')),
});

const updateMutation = useUpdateShippingAddress({
  onSuccess: () => {
    toast.success(t('addresses.result.updated'));
    invalidateShippingAddresses();
    closeForm();
  },
  onError: (err: any) => toast.error(err?.message || t('addresses.errors.updateFailed')),
});

const deleteMutation = useDeleteShippingAddress({
  onSuccess: () => {
    toast.success(t('addresses.result.deleted'));
    invalidateShippingAddresses();
  },
  onError: (err: any) => toast.error(err?.message || t('addresses.errors.deleteFailed')),
});

const setDefaultMutation = useUpdateShippingAddress({
  onSuccess: () => {
    toast.success(t('addresses.result.defaultSet'));
    invalidateShippingAddresses();
  },
  onError: (err: any) => toast.error(err?.message || t('addresses.errors.updateFailed')),
});

function submitForm() {
  if (!formValid.value) {
    toast.error(t('addresses.errors.invalidForm'));
    return;
  }
  const payload = {
    recipientName: form.recipientName.trim(),
    recipientPhone: form.recipientPhone.trim(),
    region: form.region.trim(),
    detailAddress: form.detailAddress.trim(),
    postalCode: form.postalCode.trim(),
    tag: form.tag.trim(),
    isDefault: form.isDefault,
  };
  if (editingId.value === null) {
    createMutation.mutate(payload as any);
  } else {
    updateMutation.mutate({ id: editingId.value, values: payload });
  }
}

function handleDelete(addr: AddressEntity) {
  if (!addr.id) return;
  if (!window.confirm(t('addresses.confirm.delete'))) return;
  deleteMutation.mutate(addr.id);
}

function handleSetDefault(addr: AddressEntity) {
  if (!addr.id || addr.isDefault) return;
  // 后端会自动清理同用户其它默认地址（互斥），前端只更新这一条。
  setDefaultMutation.mutate({ id: addr.id, values: { isDefault: true } });
}

const anyPending = computed(
  () =>
    createMutation.isPending.value ||
    updateMutation.isPending.value ||
    deleteMutation.isPending.value ||
    setDefaultMutation.isPending.value,
);

// 手机号脱敏（与订单详情一致）
function maskPhone(phone: string | undefined): string {
  if (!phone) return '—';
  const s = String(phone);
  if (s.length < 7) return s.replace(/\d/g, '*');
  return s.slice(0, 3) + '****' + s.slice(-4);
}
</script>

<template>
  <LayoutPageHero
    :title="t('addresses.title')"
    :description="t('addresses.subtitle')"
    icon="carbon:location"
    size="sm"
  />

  <LayoutSectionContainer>
    <UiAuthGate v-if="!isLogin" />

    <div v-else class="flex flex-col gap-4">
      <!-- 工具栏：空态时只保留空态里的新增入口，避免两个同名按钮 -->
      <div v-if="sortedAddresses.length > 0" class="flex justify-end">
        <UiButton @click="openCreate">
          <XIcon icon="carbon:add" :size="16" />
          {{ t('addresses.add') }}
        </UiButton>
      </div>

      <!-- 加载中 -->
      <UiListSkeleton v-if="loading" :rows="3" :header="false" />

      <!-- 错误态 -->
      <UiAppEmpty
        v-else-if="loadError"
        variant="error"
      >
        <template #action>
          <UiButton variant="outline" size="sm" @click="addressesQuery.refetch()">
            {{ t('ui.button.retry') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <!-- 空列表 -->
      <UiAppEmpty
        v-else-if="sortedAddresses.length === 0"
        variant="noData"
        :description="t('addresses.empty')"
      >
        <template #action>
          <UiButton variant="outline" @click="openCreate">
            {{ t('addresses.add') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <!-- 地址列表 -->
      <div v-else class="grid gap-3 md:grid-cols-2 md:gap-4">
        <article
          v-for="addr in sortedAddresses"
          :key="addr.id"
          class="flex flex-col gap-3 rounded-lg border bg-card p-4 shadow-card md:p-5"
          :class="addr.isDefault ? 'border-primary/40' : 'border-border'"
        >
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-sm font-semibold text-foreground">
              {{ addr.recipientName || '—' }}
            </span>
            <span class="text-xs tabular-nums text-muted-foreground">
              {{ maskPhone(addr.recipientPhone) }}
            </span>
            <UiStatusBadge
              v-if="addr.isDefault"
              tone="primary"
              size="sm"
              :label="t('addresses.defaultBadge')"
            />
            <UiStatusBadge
              v-if="addr.tag"
              tone="neutral"
              size="sm"
              :label="addr.tag"
            />
          </div>

          <p class="text-sm leading-relaxed text-foreground">
            <span v-if="addr.region" class="text-muted-foreground">{{ addr.region }}</span>
            {{ addr.detailAddress }}
          </p>
          <p v-if="addr.postalCode" class="text-[11px] tabular-nums text-muted-foreground">
            {{ t('addresses.form.postalCode') }} {{ addr.postalCode }}
          </p>

          <!-- 操作 -->
          <div class="mt-auto flex flex-wrap items-center justify-end gap-2 border-t border-border pt-3">
            <UiButton
              v-if="!addr.isDefault"
              variant="ghost"
              size="sm"
              :disabled="anyPending"
              @click="handleSetDefault(addr)"
            >
              {{ t('addresses.setDefault') }}
            </UiButton>
            <UiButton variant="outline" size="sm" :disabled="anyPending" @click="openEdit(addr)">
              {{ t('addresses.edit') }}
            </UiButton>
            <UiButton
              variant="ghost"
              size="sm"
              class="text-destructive hover:text-destructive"
              :disabled="anyPending"
              @click="handleDelete(addr)"
            >
              {{ t('addresses.delete') }}
            </UiButton>
          </div>
        </article>
      </div>
    </div>

    <!-- 表单弹层：统一走 UiSheet 基元 -->
    <UiSheet :open="showForm" @update:open="(v: boolean) => (showForm = v)">
      <UiSheetContent
        side="right"
        class="flex w-full flex-col gap-0 overflow-y-auto p-0 sm:max-w-md"
      >
        <UiSheetHeader class="border-b border-border px-5 py-4 text-left md:px-6">
          <UiSheetTitle class="text-lg font-bold text-foreground">{{ formTitle }}</UiSheetTitle>
        </UiSheetHeader>

        <div class="flex flex-col gap-4 px-5 py-5 md:px-6">
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('addresses.form.recipientName') }}</UiLabel>
            <UiInput v-model="form.recipientName" :placeholder="t('addresses.form.recipientNamePlaceholder')" />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('addresses.form.recipientPhone') }}</UiLabel>
            <UiInput v-model="form.recipientPhone" :placeholder="t('addresses.form.recipientPhonePlaceholder')" />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('addresses.form.region') }}</UiLabel>
            <UiInput v-model="form.region" :placeholder="t('addresses.form.regionPlaceholder')" />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('addresses.form.detailAddress') }}</UiLabel>
            <UiTextarea v-model="form.detailAddress" :placeholder="t('addresses.form.detailAddressPlaceholder')" class="min-h-[80px]" />
          </div>
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div class="flex flex-col gap-2">
              <UiLabel class="text-xs text-foreground">{{ t('addresses.form.postalCode') }}</UiLabel>
              <UiInput v-model="form.postalCode" :placeholder="t('addresses.form.postalCodePlaceholder')" />
            </div>
            <div class="flex flex-col gap-2">
              <UiLabel class="text-xs text-foreground">{{ t('addresses.form.tag') }}</UiLabel>
              <UiInput v-model="form.tag" :placeholder="t('addresses.form.tagPlaceholder')" />
            </div>
          </div>
          <div class="flex items-center justify-between gap-3 rounded-lg border border-border bg-muted p-3">
            <span class="text-sm text-foreground">{{ t('addresses.form.isDefault') }}</span>
            <UiSwitch v-model="form.isDefault" />
          </div>
        </div>

        <div class="mt-auto flex justify-end gap-3 border-t border-border px-5 py-4 md:px-6">
          <UiButton variant="outline" :disabled="anyPending" @click="closeForm">
            {{ t('addresses.form.cancel') }}
          </UiButton>
          <UiButton :disabled="!formValid || anyPending" @click="submitForm">
            {{ t('addresses.form.save') }}
          </UiButton>
        </div>
      </UiSheetContent>
    </UiSheet>
  </LayoutSectionContainer>
</template>
