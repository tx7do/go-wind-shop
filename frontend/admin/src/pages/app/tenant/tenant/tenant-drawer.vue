<template>
  <ProModal
    v-model:visible="visible"
    :title="title"
    :config="{ component: 'drawer', drawer: { size: DRAWER_WIDTH } }"
    :close-guard="guard.allowClose"
  >
    <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="120px">
      <!-- 基本信息 -->
      <ElFormItem :label="$t('pages.tenant.name')" prop="name" required>
        <ElInput v-model="formData.name" :placeholder="$t('common.placeholder.input')" clearable />
      </ElFormItem>

      <ElFormItem :label="$t('pages.tenant.code')" prop="code" required>
        <ElInput v-model="formData.code" :placeholder="$t('common.placeholder.input')" clearable />
      </ElFormItem>

      <ElFormItem :label="$t('pages.tenant.type')" prop="type" required>
        <ElSelect
          v-model="formData.type"
          :placeholder="$t('common.placeholder.select')"
          filterable
          class="w-full"
        >
          <ElOption
            v-for="item in tenantTypeList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.tenant.auditStatus')" prop="auditStatus" required>
        <ElSelect
          v-model="formData.auditStatus"
          :placeholder="$t('common.placeholder.select')"
          filterable
          class="w-full"
        >
          <ElOption
            v-for="item in tenantAuditStatusList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('common.table.status')" prop="status" required>
        <ElSelect
          v-model="formData.status"
          :placeholder="$t('common.placeholder.select')"
          filterable
          class="w-full"
        >
          <ElOption
            v-for="item in tenantStatusList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('common.table.remark')">
        <ElInput
          v-model="formData.remark"
          type="textarea"
          :rows="3"
          :placeholder="$t('common.placeholder.input')"
        />
      </ElFormItem>

      <!-- 【缓行】订阅套餐/到期时间：后端租户套餐域无下游实现，接入后恢复
      <ElFormItem :label="$t('pages.tenant.subscriptionPlan')" prop="subscriptionPlan">
        <ElSelect
          v-model="formData.subscriptionPlan"
          :placeholder="$t('common.placeholder.select')"
          filterable
          clearable
          class="w-full"
        >
          <ElOption
            v-for="item in planOptions"
            :key="item.id"
            :label="item.name"
            :value="item.id as number"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.tenant.expiredAt')" prop="expiredAt">
        <ElDatePicker
          v-model="formData.expiredAt"
          type="datetime"
          :placeholder="$t('common.placeholder.select')"
          class="w-full"
        />
      </ElFormItem>
      -->

      <!-- 【缓行】编辑模式下用量与配额 + 清理数据按钮：后端 GetUsage/CleanupData 未实现 -->

      <!-- 管理员设置（仅创建时显示） -->
      <ElDivider v-if="isCreate">{{ $t("pages.tenant.adminSetting") }}</ElDivider>

      <ElFormItem
        v-if="isCreate"
        :label="$t('pages.tenant.adminUserName')"
        prop="user.username"
        required
      >
        <ElInput
          v-model="formData.user.username"
          :placeholder="$t('common.placeholder.input')"
          clearable
        />
      </ElFormItem>

      <ElFormItem
        v-if="isCreate"
        :label="$t('pages.tenant.adminPassword')"
        prop="password"
        required
      >
        <ElInput
          v-model="formData.password"
          type="password"
          show-password
          :placeholder="$t('common.placeholder.input')"
        />
      </ElFormItem>

      <ElFormItem
        v-if="isCreate"
        :label="$t('pages.tenant.adminPasswordConfirm')"
        prop="passwordConfirm"
        required
      >
        <ElInput
          v-model="formData.passwordConfirm"
          type="password"
          show-password
          :placeholder="$t('common.placeholder.input')"
        />
      </ElFormItem>

      <ElFormItem
        v-if="isCreate"
        :label="$t('pages.tenant.adminMobile')"
        prop="user.mobile"
        required
      >
        <ElInput
          v-model="formData.user.mobile"
          :placeholder="$t('common.placeholder.input')"
          clearable
        />
      </ElFormItem>

      <ElFormItem v-if="isCreate" :label="$t('pages.tenant.adminEmail')" prop="user.email" required>
        <ElInput
          v-model="formData.user.email"
          :placeholder="$t('common.placeholder.input')"
          clearable
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="guard.requestClose(handleClose)">{{ $t("common.button.cancel") }}</ElButton>
        <ElButton type="primary" :loading="loading" @click="handleSubmit">
          {{ $t("common.button.confirm") }}
        </ElButton>
      </div>
    </template>
  </ProModal>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";

import { ElMessage } from "element-plus";
import type { FormInstance, FormRules } from "element-plus";

import {
  tenantAuditStatusList,
  tenantStatusList,
  tenantTypeList,
  useCreateTenantWithAdminUser,
  useUpdateTenant,
  fetchListTenants,
} from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
// 【缓行】租户套餐/用量/清理：后端 GetUsage/CleanupData/Plan 未实现
// import type { identityservicev1_Plan as Plan } from "@/api/generated/admin/service/v1";
import type { identityservicev1_Tenant as Tenant } from "@/api/generated/admin/service/v1";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import { injectProModalApi } from "@/components/Pro";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

// 通过 inject 获取列表页传入的 modalApi
const modalApi = injectProModalApi();

// 注意：不能使用 modalApi.getData()，其内部对 toRaw(store).sharedData 取值会丢失响应式，
// 导致多次 open（新增↔编辑切换、编辑不同租户）时 data/isCreate 停留在首次快照，造成串数据。
// 这里直接读取响应式的 store.sharedData。
const data = computed(() => modalApi.store.sharedData as { create?: boolean; row?: Tenant });
const isCreate = computed(() => !!data.value.create);

const visible = computed({
  get: () => modalApi.store.isOpen,
  set: (v) => {
    if (!v) modalApi.close();
  },
});

const { mutateAsync: createTenantWithAdminUserMut } = useCreateTenantWithAdminUser();
const { mutateAsync: updateTenantMut } = useUpdateTenant();

// 【缓行】编辑模式下租户用量/配额/清理数据：后端 GetUsage/CleanupData 未实现，整段停用。

// 加载状态
const loading = ref(false);

// 表单数据
const formData = ref({
  name: "",
  code: "",
  type: "PAID",
  auditStatus: "APPROVED",
  status: "ON",
  remark: "",
  user: {
    username: "",
    mobile: "",
    email: "",
  },
  password: "",
  passwordConfirm: "",
});

const formRef = ref<FormInstance>();

// 表单校验规则（补齐 required + 格式校验，与 React 版对齐）
const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validation.required"), trigger: "blur" }],
  code: [{ required: true, message: $t("common.validation.required"), trigger: "blur" }],
  type: [{ required: true, message: $t("common.validation.selectRequired"), trigger: "change" }],
  auditStatus: [
    { required: true, message: $t("common.validation.selectRequired"), trigger: "change" },
  ],
  status: [{ required: true, message: $t("common.validation.selectRequired"), trigger: "change" }],
  "user.username": [{ required: true, message: $t("common.validation.required"), trigger: "blur" }],
  "user.mobile": [
    { required: true, message: $t("common.validation.required"), trigger: "blur" },
    {
      pattern: /^1[3-9]\d{9}$/,
      message: $t("common.validation.phoneFormat"),
      trigger: "blur",
    },
  ],
  "user.email": [
    { required: true, message: $t("common.validation.required"), trigger: "blur" },
    { type: "email", message: $t("common.validation.email"), trigger: "blur" },
  ],
  password: [
    { required: true, message: $t("common.validation.required"), trigger: "blur" },
    { min: 6, message: $t("common.validation.passwordMin"), trigger: "blur" },
  ],
  passwordConfirm: [
    { required: true, message: $t("common.validation.required"), trigger: "blur" },
    { min: 6, message: $t("common.validation.passwordMin"), trigger: "blur" },
  ],
};

// 弹窗标题
const title = computed(() =>
  isCreate.value
    ? $t("common.modal.create", { moduleName: $t("pages.tenant.moduleName") })
    : $t("common.modal.update", { moduleName: $t("pages.tenant.moduleName") })
);

// 监听弹窗打开/关闭
// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
const guard = useDrawerCloseGuard(() => ({ ...formData.value }));

watch(visible, async (val) => {
  if (val) {
    if (!isCreate.value && data.value.row) {
      // 编辑模式
      const row = data.value.row;
      formData.value = {
        name: row.name || "",
        code: row.code || "",
        type: row.type || "PAID",
        auditStatus: row.auditStatus || "APPROVED",
        status: row.status || "ON",
        remark: row.remark || "",
        user: {
          username: "",
          mobile: "",
          email: "",
        },
        password: "",
        passwordConfirm: "",
      };

      // 【缓行】租户用量与配额加载：后端 GetUsage 未实现
    } else {
      // 创建模式
      resetForm();
    }

    // 【缓行】套餐下拉选项加载：后端 Plan 域未实现

    // 填充完成后记录基线快照（回填不算脏）
    guard.syncBaseline();
  } else {
    // ProModal 关闭时自动重置表单
    resetForm();
  }
});

// 重置表单
const resetForm = () => {
  formData.value = {
    name: "",
    code: "",
    type: "PAID",
    auditStatus: "APPROVED",
    status: "ON",
    remark: "",
    user: {
      username: "",
      mobile: "",
      email: "",
    },
    password: "",
    passwordConfirm: "",
  };
};

// 关闭弹窗
const handleClose = () => {
  modalApi.close();
  resetForm();
};

// 提交表单
const handleSubmit = async () => {
  try {
    loading.value = true;

    // 表单校验（required + 手机/邮箱格式 + 密码长度，由 formRules 驱动）
    if (!formRef.value) return;
    await formRef.value.validate();

    if (isCreate.value) {
      await createTenantWithAdminUser();
    } else {
      await updateTenant();
    }

    // 成功回调
    modalApi.close();
  } catch (error) {
    console.error("Submit error:", error);
  } finally {
    loading.value = false;
  }
};

// 创建租户和管理员用户
async function createTenantWithAdminUser() {
  // 检查密码和确认密码是否一致
  if (formData.value.password !== formData.value.passwordConfirm) {
    ElMessage.error($t("common.notification.password_mismatch"));
    return;
  }

  // 检查租户编码是否存在
  try {
    const result = await fetchListTenants(
      new PaginationQuery({ formValues: { code: formData.value.code } })
    );
    if (result.items && result.items.length > 0) {
      // 查询成功且存在重复，提示编码已存在
      ElMessage.error($t("pages.tenant.tenant_code_exists"));
      return;
    }
  } catch {
    // 查询本身失败（网络/500 等），不能误判为编码已存在
    ElMessage.error($t("common.operationFailed"));
    return;
  }

  await createTenantWithAdminUserMut({
    tenant: {
      name: formData.value.name,
      code: formData.value.code,
      type: formData.value.type as any,
      auditStatus: formData.value.auditStatus as any,
      status: formData.value.status as any,
      remark: formData.value.remark,
      // 【缓行】planId/expiredAt：租户套餐域未实现
    },
    user: formData.value.user as any,
    password: formData.value.password,
  });

  ElMessage.success($t("common.notification.create_success"));
}

// 更新租户
async function updateTenant() {
  if (!data.value.row?.id) {
    ElMessage.error($t("common.notification.update_failed"));
    return;
  }

  await updateTenantMut({
    id: data.value.row!.id,
    values: {
      name: formData.value.name,
      code: formData.value.code,
      type: formData.value.type,
      auditStatus: formData.value.auditStatus,
      status: formData.value.status,
      remark: formData.value.remark,
      // 【缓行】planId/expiredAt：租户套餐域未实现
    },
  });

  ElMessage.success($t("common.notification.update_success"));
}
</script>

<style scoped>
.drawer-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>
