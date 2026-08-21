<template>
  <div class="comment-item">
    <div class="comment-item__header">
      <UiAvatar :src="undefined" alt="avatar" class="comment-item__avatar" />
      <div class="comment-item__meta">
        <span class="comment-item__author">{{ t('comment.author') }} #{{ comment.createdBy ?? '?' }}</span>
        <span class="comment-item__time">{{ formatTime(comment.createdAt) }}</span>
      </div>
    </div>

    <div class="comment-item__content">{{ comment.content || '' }}</div>

    <div class="comment-item__actions">
      <button
        type="button"
        class="comment-item__like-btn"
        :class="{ 'comment-item__like-btn--active': isLiked }"
        @click="handleLike"
      >
        <XIcon icon="lucide:heart" :class="{ 'comment-item__like-icon--active': isLiked }" />
        <span>{{ likeCount }}</span>
      </button>

      <button
        v-if="isLogin"
        type="button"
        class="comment-item__reply-btn"
        @click="showReplyForm = !showReplyForm"
      >
        {{ t('comment.reply') }}
      </button>
    </div>

    <div v-if="showReplyForm" class="comment-item__reply-form">
      <UiTextarea
        v-model="replyContent"
        :placeholder="t('comment.replyPlaceholder')"
        class="comment-item__reply-input"
        rows="3"
      />
      <UiButton
        variant="default"
        size="sm"
        :disabled="replyMutation.isPending.value || !replyContent.trim()"
        @click="submitReply"
      >
        {{ replyMutation.isPending.value ? t('common.submitting') : t('comment.submit') }}
      </UiButton>
    </div>

    <div
      v-if="comment.children && comment.children.length > 0"
      class="comment-item__children"
    >
      <CommentItem
        v-for="child in comment.children"
        :key="child.id"
        :comment="child"
        :product-id="productId"
        :is-login="isLogin"
        :like-status-map="likeStatusMap"
        :like-count-map="likeCountMap"
        @like="handleLike"
        @post-comment="handlePostComment"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { toast } from 'vue-sonner';
import { useCreateComment, useLike, useUnlike } from '@/api/composables';
import { queryClient } from '@/plugins/vue-query';
import { useAccessStore } from '@/stores/modules/core/access.state';
import type {
  commentservicev1_Comment,
  commentservicev1_CreateCommentRequest,
  interactionservicev1_LikeRequest,
  interactionservicev1_LikeResponse,
} from '@/api/generated/app/service/v1';

const { t } = useI18n();
const localePath = useLocalePath();
const accessStore = useAccessStore();

const props = defineProps<{
  comment: commentservicev1_Comment;
  productId: number;
  isLogin: boolean;
  likeStatusMap: Record<number, boolean>;
  likeCountMap: Record<number, number>;
}>();

const emit = defineEmits<{
  (e: 'like', targetId: number): void;
  (e: 'post-comment', request: commentservicev1_CreateCommentRequest): void;
}>();

const showReplyForm = ref(false);
const replyContent = ref('');

const commentId = props.comment.id ?? 0;
const isLiked = ref(props.likeStatusMap[commentId] ?? false);
const likeCount = ref(props.likeCountMap[commentId] ?? 0);

const likeMutation = useLike();
const unlikeMutation = useUnlike();
const replyMutation = useCreateComment();

function formatTime(ts: any): string {
  if (!ts) return '';
  const d = ts instanceof Date ? ts : new Date(ts);
  if (isNaN(d.getTime())) return '';
  return d.toLocaleString();
}

async function handleLike() {
  if (!props.isLogin) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  const req: interactionservicev1_LikeRequest = {
    targetId: commentId,
    targetType: 'TARGET_TYPE_COMMENT' as any,
  };
  try {
    let resp: interactionservicev1_LikeResponse;
    if (isLiked.value) {
      resp = await unlikeMutation.mutateAsync(req);
    } else {
      resp = await likeMutation.mutateAsync(req);
    }
    isLiked.value = resp.liked ?? false;
    likeCount.value = resp.likeCount ?? 0;
    queryClient.invalidateQueries({ queryKey: ['interactionStatus'] });
    queryClient.invalidateQueries({ queryKey: ['interactionCounts'] });
  } catch {
    toast.error(t('common.operationFailed'));
  }
}

async function submitReply() {
  if (!props.isLogin) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  if (!replyContent.value.trim()) return;

  const req: commentservicev1_CreateCommentRequest = {
    data: {
      content: replyContent.value,
      contentType: 'CONTENT_TYPE_PRODUCT' as any,
      objectId: props.productId,
      parentId: commentId,
    } as any,
  };

  try {
    await replyMutation.mutateAsync(req);
    toast.success(t('comment.pendingApproval'));
    replyContent.value = '';
    showReplyForm.value = false;
    queryClient.invalidateQueries({ queryKey: ['listComments'] });
  } catch {
    toast.error(t('common.operationFailed'));
  }
}

// 局部刷新
function syncFromParent() {
  isLiked.value = props.likeStatusMap[commentId] ?? false;
  likeCount.value = props.likeCountMap[commentId] ?? 0;
}

defineExpose({ syncFromParent });
</script>

<style lang="scss" scoped>
.comment-item {
  padding: 1rem 0;
  border-bottom: 1px solid var(--el-border-color-lighter);

  &__header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
  }

  &__avatar {
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 50%;
    background: var(--el-fill-color-light);
  }

  &__meta {
    display: flex;
    flex-direction: column;
    font-size: 0.75rem;
    color: var(--el-text-color-secondary);
  }

  &__author {
    color: var(--el-text-color-secondary);
  }

  &__time {
    color: var(--el-text-color-placeholder);
  }

  &__content {
    font-size: 0.875rem;
    line-height: 1.5;
    color: var(--el-text-color-regular);
    margin-bottom: 0.5rem;
    white-space: pre-wrap;
    word-break: break-word;
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  &__like-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
    color: var(--el-text-color-secondary);
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;

    &:hover {
      background: var(--el-fill-color-light);
    }

    &--active {
      color: var(--el-color-danger);
    }
  }

  &__reply-btn {
    font-size: 0.75rem;
    color: var(--el-text-color-secondary);
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;

    &:hover {
      background: var(--el-fill-color-light);
      color: var(--el-color-primary);
    }
  }

  &__reply-form {
    margin-top: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 32rem;
  }

  &__children {
    margin-top: 0.5rem;
    margin-left: 2.5rem;
    border-left: 2px solid var(--el-border-color-lighter);
    padding-left: 1rem;
  }
}
</style>
