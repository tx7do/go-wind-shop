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
      label-width="140px"
      class="drawer-form"
    >
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElFormItem :label="$t('pages.mall.comment.content')">
        <div class="readonly-content">{{ formData.content || "—" }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.comment.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption label="待审核" value="STATUS_PENDING" />
          <ElOption label="已批准" value="STATUS_APPROVED" />
          <ElOption label="已拒绝" value="STATUS_REJECTED" />
          <ElOption label="垃圾" value="STATUS_SPAM" />
        </ElSelect>
      </ElFormItem>

      <ElDivider content-position="left">{{ $t("pages.mall.comment.replySection") }}</ElDivider>

      <!-- 已有回复（商家/其他买家的树形子评论） -->
      <div v-if="replies.length > 0" class="replies-list">
        <div v-for="r in replies" :key="r.id" class="reply-item">
          <div class="reply-item__meta">
            <span>#{{ r.id }}</span>
            <span class="reply-item__author">{{ r.createdBy ? `用户 ${r.createdBy}` : "—" }}</span>
            <span v-if="r.createdAt" class="reply-item__time">
              {{ String(r.createdAt).slice(0, 16).replace("T", " ") }}
            </span>
          </div>
          <div class="reply-item__content">{{ r.content }}</div>
        </div>
      </div>
      <div v-else class="replies-empty">{{ $t("pages.mall.comment.noReplies") }}</div>

      <!-- 商家回复表单 -->
      <ElFormItem :label="$t('pages.mall.comment.replyContent')" prop="replyContent">
        <ElInput
          v-model="formData.replyContent"
          type="textarea"
          :rows="3"
          :maxlength="500"
          show-word-limit
          :placeholder="$t('pages.mall.comment.replyPlaceholder')"
        />
        <div class="form-tip">{{ $t("pages.mall.comment.replyTip") }}</div>
      </ElFormItem>

      <ElFormItem>
        <ElButton
          type="primary"
          plain
          :loading="replyLoading"
          :disabled="!formData.replyContent.trim()"
          @click="handleReply"
        >
          {{ $t("pages.mall.comment.replySubmit") }}
        </ElButton>
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
  useUpdateComment,
  useCreateComment,
  fetchGetComment,
  fetchListComments,
} from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: updateComment } = useUpdateComment();
const { mutateAsync: createComment } = useCreateComment();

const visible = ref(false);
const submitLoading = ref(false);
const replyLoading = ref(false);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

// 回复需要落库的父评论上下文
const parentContext = reactive<{
  objectId: number | undefined;
  contentType: string;
}>({
  objectId: undefined,
  contentType: "CONTENT_TYPE_PRODUCT",
});

const replies = ref<
  Array<{ id?: number; content?: string; createdBy?: number; createdAt?: string }>
>([]);

const formData = reactive<{
  content: string;
  status: string;
  replyContent: string;
}>({
  content: "",
  status: "STATUS_PENDING",
  replyContent: "",
});

const formRules: FormRules = {
  status: [{ required: true, message: $t("common.validation.required"), trigger: "change" }],
};

const title = computed(() => $t("pages.mall.comment.button.update"));

function loadReplies(commentId: number) {
  fetchListComments(
    new PaginationQuery({
      paging: { page: 1, pageSize: 50 },
      formValues: { parentId: commentId },
    })
  )
    .then((resp: any) => {
      replies.value = (resp.items ?? []) as Array<{
        id?: number;
        content?: string;
        createdBy?: number;
        createdAt?: string;
      }>;
    })
    .catch(() => {
      replies.value = [];
    });
}

// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
// replyContent 一并在 getModel 内——回复框打了字未发送，关闭时也要确认
const guard = useDrawerCloseGuard(() => ({ ...formData }));

function open(row?: any) {
  visible.value = true;

  if (row) {
    currentId.value = row.id;
    parentContext.objectId = row.objectId;
    parentContext.contentType = row.contentType || "CONTENT_TYPE_PRODUCT";
    loadReplies(row.id);
    fetchGetComment({ id: row.id })
      .then((resp: any) => {
        formData.content = resp.content ?? "";
        formData.status = resp.status ?? "STATUS_PENDING";
        parentContext.objectId = resp.objectId ?? parentContext.objectId;
        guard.syncBaseline();
      })
      .catch(() => {
        ElMessage.error($t("common.message.error"));
      });
  } else {
    currentId.value = undefined;
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
  formData.content = "";
  formData.status = "STATUS_PENDING";
  formData.replyContent = "";
  replies.value = [];
  parentContext.objectId = undefined;
  formRef.value?.clearValidate();
}

// 商家回复：以管理员身份创建子评论（parentId 指向当前评论），直接置 APPROVED 上墙
async function handleReply() {
  const content = formData.replyContent.trim();
  if (!content || !currentId.value) return;

  try {
    replyLoading.value = true;
    await createComment({
      objectId: parentContext.objectId,
      contentType: parentContext.contentType,
      parentId: currentId.value,
      content,
      status: "STATUS_APPROVED",
    } as any);
    ElMessage.success($t("pages.mall.comment.replySuccess"));
    formData.replyContent = "";
    loadReplies(currentId.value);
    emit("success");
  } catch {
    ElMessage.error($t("pages.mall.comment.replyFailed"));
  } finally {
    replyLoading.value = false;
  }
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    await updateComment({ id: currentId.value!, values: { status: formData.status } });
    ElMessage.success($t("common.notification.updateSuccess"));

    emit("success");
    handleClose();
  } catch (error) {
    if (error !== false) {
      ElMessage.error($t("common.notification.updateFailed"));
    }
  } finally {
    submitLoading.value = false;
  }
}

watch(visible, (val) => {
  if (!val) resetForm();
});

defineExpose({
  open,
});
</script>

<style lang="scss" scoped>
.drawer-form {
  padding-right: 10px;
}

.readonly-content {
  width: 100%;
  max-height: 200px;
  overflow-y: auto;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
}

.replies-list {
  width: 100%;
  max-height: 220px;
  overflow-y: auto;
  margin-bottom: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 4px 0;
}

.reply-item {
  padding: 8px 12px;

  & + & {
    border-top: 1px solid var(--el-border-color-extra-light);
  }

  &__meta {
    display: flex;
    gap: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  &__content {
    margin-top: 4px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    white-space: pre-wrap;
    word-break: break-word;
  }
}

.replies-empty {
  width: 100%;
  margin-bottom: 16px;
  padding: 12px;
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.form-tip {
  width: 100%;
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--el-text-color-secondary);
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
