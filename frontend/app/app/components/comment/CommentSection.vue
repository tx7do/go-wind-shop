<template>
  <LayoutSectionContainer class="!py-8">
    <div class="mx-auto w-full max-w-4xl">
      <h2 class="mb-6 text-xl font-bold text-foreground">
        {{ t('comment.sectionTitle') }}
      </h2>

      <div v-if="commentsLoading" class="py-8">
        <UiSkeleton class="h-24 w-full" />
      </div>

      <div v-else-if="topLevelComments.length === 0" class="py-12 text-center text-sm text-muted-foreground">
        {{ t('comment.empty') }}
      </div>

      <div v-else class="comment-list">
        <CommentItem
          v-for="c in topLevelComments"
          :key="c.id"
          :comment="c"
          :product-id="productId"
          :is-login="isLogin"
          :like-status-map="likeStatusMap"
          :like-count-map="likeCountMap"
        />
      </div>

      <div v-if="isLogin" class="mt-8 comment-post">
        <div class="comment-post__rating">
          <span class="text-xs text-muted-foreground">{{ t('mall.productRating.selectRating') }}</span>
          <UiRatingInput v-model="postRating" />
        </div>
        <UiTextarea
          v-model="postContent"
          :placeholder="t('comment.postPlaceholder')"
          class="comment-post__input"
          rows="3"
        />
        <UiButton
          variant="default"
          size="sm"
          :disabled="postMutation.isPending.value || !postContent.trim() || postRating < 1 || postRating > 5"
          @click="submitPost"
        >
          {{ postMutation.isPending.value ? t('common.submitting') : t('comment.submit') }}
        </UiButton>
      </div>

      <div v-else class="mt-8 py-4 text-center text-sm text-muted-foreground">
        {{ t('comment.loginToPost') }}
      </div>
    </div>
  </LayoutSectionContainer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { toast } from 'vue-sonner';
import {
  useListComments,
  useCreateComment,
  useGetInteractionStatus,
  useGetCounts,
} from '@/api/composables';
import { queryClient } from '@/plugins/vue-query';
import { useAccessStore } from '@/stores/modules/core/access.state';
import type {
  commentservicev1_Comment,
  commentservicev1_CreateCommentRequest,
} from '@/api/generated/app/service/v1';

const { t } = useI18n();
const localePath = useLocalePath();
const accessStore = useAccessStore();

const props = defineProps<{
  productId: number;
}>();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

const postContent = ref('');
const postRating = ref(0);

const commentQueryParams = computed(() => ({
  objectId: props.productId,
  contentType: 'CONTENT_TYPE_PRODUCT',
  page: 1,
  pageSize: 50,
  noPaging: true,
  sorting: [{ field: 'id', direction: 'DESC' }],
}));

const commentsQuery = useListComments(commentQueryParams);

const topLevelComments = computed<commentservicev1_Comment[]>(() => {
  const items = (commentsQuery.data?.value as any)?.items ?? [];
  return (items as commentservicev1_Comment[]).filter((c) => !c.parentId || c.parentId === 0);
});

const commentsLoading = computed(() => commentsQuery.isPending.value);

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

const statusQueryParams = computed(() => ({
  targetType: 'TARGET_TYPE_COMMENT',
  targetIds: allCommentIds.value,
}));

const countsQueryParams = computed(() => ({
  targetType: 'TARGET_TYPE_COMMENT',
  targetIds: allCommentIds.value,
  metrics: ['COUNTER_METRIC_LIKE'],
}));

const statusQuery = useGetInteractionStatus(statusQueryParams, {
  enabled: computed(() => isLogin.value && allCommentIds.value.length > 0),
});
const countsQuery = useGetCounts(countsQueryParams, {
  enabled: computed(() => allCommentIds.value.length > 0),
});

const likeStatusMap = computed<Record<number, boolean>>(() => {
  const map: Record<number, boolean> = {};
  const statuses = (statusQuery.data?.value as any)?.statuses;
  if (statuses) {
    for (const [k, v] of Object.entries(statuses)) {
      map[Number(k)] = (v as any)?.liked ?? false;
    }
  }
  return map;
});

const likeCountMap = computed<Record<number, number>>((() => {
  const map: Record<number, number> = {};
  const counts = (countsQuery.data?.value as any)?.counts;
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
}) as any);

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

async function submitPost() {
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
      objectId: props.productId,
      parentId: 0,
    } as any,
  };

  await postMutation.mutateAsync(req);
}
</script>

<style lang="scss" scoped>
.comment-list {
  display: flex;
  flex-direction: column;
}

.comment-post {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 32rem;

  &__input {
    width: 100%;
  }

  &__rating {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
}
</style>
