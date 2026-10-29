<template>
  <ProModal
    v-model:visible="visible"
    :title="title"
    :config="{ component: 'drawer', drawer: { size: DRAWER_WIDTH } }"
    :close-guard="guard.allowClose"
  >
    <ElForm
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="160px"
      class="drawer-form"
    >
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElAlert
        v-if="isCreate && formData.channel === 'PAYMENT_CHANNEL_MOCK'"
        type="info"
        :title="$t('pages.mall.paymentChannel.mockHint')"
        show-icon
        :closable="false"
        class="mb-4"
      />

      <ElFormItem :label="$t('pages.mall.paymentChannel.name')" prop="name">
        <ElInput v-model="formData.name" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.paymentChannel.channel')" prop="channel">
        <ElSelect
          v-model="formData.channel"
          :disabled="!isCreate"
          :placeholder="$t('common.placeholder.select')"
        >
          <ElOption
            v-for="opt in CHANNEL_OPTIONS"
            :key="opt.value"
            :label="$t(`pages.mall.paymentChannel.channelType.${opt.labelKey}`)"
            :value="opt.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem
        v-if="formData.channel !== 'PAYMENT_CHANNEL_MOCK'"
        :label="$t('pages.mall.paymentChannel.appId')"
        prop="appId"
      >
        <ElInput v-model="formData.appId" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem
        v-if="formData.channel === 'PAYMENT_CHANNEL_WECHAT_NATIVE' || formData.channel === 'PAYMENT_CHANNEL_WECHAT_H5'"
        :label="$t('pages.mall.paymentChannel.mchId')"
        prop="mchId"
      >
        <ElInput v-model="formData.mchId" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElDivider content-position="left">{{ $t("pages.mall.paymentChannel.sectionCredentials") }}</ElDivider>

      <template v-if="formData.channel === 'PAYMENT_CHANNEL_WECHAT_NATIVE' || formData.channel === 'PAYMENT_CHANNEL_WECHAT_H5'">
        <ElFormItem :label="$t('pages.mall.paymentChannel.mchCertSerialNo')" prop="mchCertSerialNo">
          <ElInput v-model="credentials.mchCertSerialNo" :placeholder="$t('common.placeholder.input')" />
        </ElFormItem>
        <ElFormItem :label="$t('pages.mall.paymentChannel.apiV3Key')" prop="apiV3Key">
          <ElInput
            v-model="credentials.apiV3Key"
            type="password"
            show-password
            :placeholder="credentialsPlaceholder"
          />
        </ElFormItem>
        <ElFormItem :label="$t('pages.mall.paymentChannel.privateKeyPem')" prop="privateKeyPem">
          <ElInput
            v-model="credentials.privateKeyPem"
            type="textarea"
            :rows="5"
            :placeholder="credentialsPlaceholder"
          />
        </ElFormItem>
      </template>

      <template v-else-if="formData.channel === 'PAYMENT_CHANNEL_ALIPAY_PAGE'">
        <ElFormItem :label="$t('pages.mall.paymentChannel.alipayPrivateKey')" prop="privateKey">
          <ElInput
            v-model="credentials.privateKey"
            type="textarea"
            :rows="5"
            :placeholder="credentialsPlaceholder"
          />
        </ElFormItem>
        <ElFormItem :label="$t('pages.mall.paymentChannel.alipayPublicKey')" prop="alipayPublicKey">
          <ElInput
            v-model="credentials.alipayPublicKey"
            type="textarea"
            :rows="3"
            :placeholder="credentialsPlaceholder"
          />
        </ElFormItem>
      </template>

      <template v-else>
        <ElFormItem :label="$t('pages.mall.paymentChannel.credentials')">
          <span class="text-xs text-gray-400">{{ $t("pages.mall.paymentChannel.mockNoCredentials") }}</span>
        </ElFormItem>
      </template>

      <ElDivider content-position="left">{{ $t("common.section.status") }}</ElDivider>

      <ElFormItem :label="$t('pages.mall.paymentChannel.sandbox')" prop="sandbox">
        <ElSwitch v-model="formData.sandbox" :disabled="formData.channel.startsWith('PAYMENT_CHANNEL_WECHAT')" />
        <span v-if="formData.channel === 'PAYMENT_CHANNEL_WECHAT_NATIVE'" class="ml-2 text-xs text-gray-400">
          {{ $t("pages.mall.paymentChannel.wechatNoSandbox") }}
        </span>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.paymentChannel.enabled')" prop="enabled">
        <ElSwitch v-model="formData.enabled" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.paymentChannel.remark')" prop="remark">
        <ElInput v-model="formData.remark" type="textarea" :rows="2" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="guard.requestClose(handleClose)">
          {{ $t("common.button.cancel") }}
        </ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ $t("common.button.confirm") }}
        </ElButton>
      </div>
    </template>
  </ProModal>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from "vue";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";

import {
  useCreatePaymentChannelConfig,
  useUpdatePaymentChannelConfig,
  fetchGetPaymentChannelConfig,
} from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createChannel } = useCreatePaymentChannelConfig();
const { mutateAsync: updateChannel } = useUpdatePaymentChannelConfig();

const CHANNEL_OPTIONS = [
  { value: "PAYMENT_CHANNEL_MOCK", labelKey: "mock" },
  { value: "PAYMENT_CHANNEL_WECHAT_NATIVE", labelKey: "wechatNative" },
  { value: "PAYMENT_CHANNEL_WECHAT_H5", labelKey: "wechatH5" },
  { value: "PAYMENT_CHANNEL_ALIPAY_PAGE", labelKey: "alipayPage" },
] as const;

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const credentialsConfigured = ref(false);
const formRef = ref<FormInstance>();

// 渠道凭据按渠道类型分别录入，提交时打包为 credentials JSON（写-only，读取恒空）。
const formData = reactive<{
  name: string;
  channel: string;
  appId: string;
  mchId: string;
  sandbox: boolean;
  enabled: boolean;
  remark: string;
}>({
  name: "",
  channel: "PAYMENT_CHANNEL_MOCK",
  appId: "",
  mchId: "",
  sandbox: true,
  enabled: false,
  remark: "",
});

const credentials = reactive({
  mchCertSerialNo: "",
  apiV3Key: "",
  privateKeyPem: "",
  privateKey: "",
  alipayPublicKey: "",
});

const credentialsPlaceholder = computed(() =>
  credentialsConfigured.value
    ? $t("pages.mall.paymentChannel.credentialsKeepPlaceholder")
    : $t("common.placeholder.input")
);

const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  channel: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
};

const title = computed(() =>
  isCreate.value
    ? $t("pages.mall.paymentChannel.button.create")
    : $t("pages.mall.paymentChannel.button.update")
);

// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
const guard = useDrawerCloseGuard(() => ({ ...formData, ...credentials }));

function resetCredentials() {
  credentials.mchCertSerialNo = "";
  credentials.apiV3Key = "";
  credentials.privateKeyPem = "";
  credentials.privateKey = "";
  credentials.alipayPublicKey = "";
}

function buildCredentialsJson(): string | undefined {
  switch (formData.channel) {
    case "PAYMENT_CHANNEL_WECHAT_NATIVE":
    case "PAYMENT_CHANNEL_WECHAT_H5": {
      if (
        !credentials.mchCertSerialNo &&
        !credentials.apiV3Key &&
        !credentials.privateKeyPem
      ) {
        return undefined;
      }
      return JSON.stringify({
        mchCertSerialNo: credentials.mchCertSerialNo,
        apiV3Key: credentials.apiV3Key,
        privateKeyPem: credentials.privateKeyPem,
      });
    }
    case "PAYMENT_CHANNEL_ALIPAY_PAGE": {
      if (!credentials.privateKey && !credentials.alipayPublicKey) {
        return undefined;
      }
      return JSON.stringify({
        privateKey: credentials.privateKey,
        alipayPublicKey: credentials.alipayPublicKey,
      });
    }
    default:
      return undefined;
  }
}

function open(row?: any) {
  visible.value = true;

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    fetchGetPaymentChannelConfig({ id: row.id })
      .then((resp: any) => {
        formData.name = resp.name ?? "";
        formData.channel = resp.channel ?? "PAYMENT_CHANNEL_MOCK";
        formData.appId = resp.appId ?? "";
        formData.mchId = resp.mchId ?? "";
        formData.sandbox = resp.sandbox ?? true;
        formData.enabled = resp.enabled ?? false;
        formData.remark = resp.remark ?? "";
        credentialsConfigured.value = resp.credentialsConfigured ?? false;
        guard.syncBaseline();
      })
      .catch(() => {
        ElMessage.error($t("common.message.error"));
      });
  } else {
    isCreate.value = true;
    currentId.value = undefined;
    credentialsConfigured.value = false;
    resetForm();
  }

  // 填充完成后记录基线快照（回填不算脏；编辑异步回填在 then 里重记）
  guard.syncBaseline();
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.name = "";
  formData.channel = "PAYMENT_CHANNEL_MOCK";
  formData.appId = "";
  formData.mchId = "";
  formData.sandbox = true;
  // 安全默认：新渠道一律停用 + 沙箱，运营显式开启/切换生产
  formData.enabled = false;
  formData.remark = "";
  resetCredentials();
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    // undefined 不进 updateMask（避免「mask 声明了字段但数据缺失」被服务端清值）
    const raw: Record<string, any> = {
      name: formData.name,
      channel: formData.channel,
      sandbox: formData.sandbox,
      enabled: formData.enabled,
    };
    if (formData.appId) raw.appId = formData.appId;
    if (formData.mchId) raw.mchId = formData.mchId;
    if (formData.remark) raw.remark = formData.remark;
    const values: Record<string, any> = raw;

    const credentialsJson = buildCredentialsJson();
    if (isCreate.value) {
      if (credentialsJson) {
        values.credentials = credentialsJson;
      }
      await createChannel(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      if (credentialsJson) {
        // 仅在提交了新凭据时才携带 credentials 字段（否则保留库中原密文）
        values.credentials = credentialsJson;
      } else {
        delete values.credentials;
      }
      await updateChannel({ id: currentId.value!, values });
      ElMessage.success($t("common.notification.updateSuccess"));
    }

    emit("success");
    handleClose();
  } catch (error) {
    if (error !== false) {
      ElMessage.error(
        isCreate.value
          ? $t("common.notification.createFailed")
          : $t("common.notification.updateFailed")
      );
    }
  } finally {
    submitLoading.value = false;
  }
}

watch(visible, (val) => {
  if (!val) resetForm();
});

watch(
  () => formData.channel,
  () => resetCredentials()
);

defineExpose({
  open,
});
</script>

<style lang="scss" scoped>
.drawer-form {
  padding-right: 10px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
