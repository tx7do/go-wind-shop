<template>
  <div class="workspace-page">
    <!-- 欢迎横幅 -->
    <el-card shadow="hover" class="welcome-card">
      <div class="welcome-body">
        <el-avatar :size="48" :src="userStore.userInfo?.avatar">
          {{ (userStore.userInfo?.nickname || userStore.userInfo?.username || "?").slice(0, 1) }}
        </el-avatar>
        <div class="welcome-text">
          <div class="welcome-greeting">{{ greetingText }}</div>
          <div class="welcome-date">{{ todayText }}</div>
        </div>
      </div>
    </el-card>

    <!-- 快捷入口 -->
    <el-card shadow="hover" class="mb-5" :body-style="{ paddingTop: '12px' }">
      <template #header>
        <span class="card-title">{{ $t("pages.dashboard.workspace.quickAccess.title") }}</span>
      </template>
      <el-row :gutter="8">
        <el-col v-for="link in QUICK_LINKS" :key="link.path" :xs="12" :sm="8" :md="6">
          <div
            class="quick-item"
            role="button"
            tabindex="0"
            @click="router.push(link.path)"
            @keydown.enter="router.push(link.path)"
          >
            <SvgIcon :icon="link.icon" :size="20" class="quick-item__icon" />
            <span class="quick-item__label">{{ $t(link.labelKey) }}</span>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="16">
      <!-- 我的站内信 / 我的活跃会话：两列等高——el-col 转 flex 让 el-card 纵向撑满
          （el-row 默认 stretch，两列高度本就相同，卡片不撑满才显参差）；md 起给共同
           min-height 整体拉高（站内信 7 行正好填满，断点与 :md="12" 一致），窄屏单列不设免留白 -->
      <el-col :xs="24" :md="12" class="pane-col mb-5-xs">
        <el-card shadow="hover" class="pane-card">
          <template #header>
            <div class="pane-header">
              <span class="card-title">{{ $t("pages.dashboard.workspace.myMessages.title") }}</span>
              <el-button
                link
                type="primary"
                @click="router.push('/internal-message/inbox')"
              >
                {{ $t("pages.dashboard.workspace.myMessages.viewAll") }}
                <template v-if="unreadCount > 0">({{ unreadCount }})</template>
              </el-button>
            </div>
          </template>
          <div v-if="inboxRecentQuery.isLoading.value" class="pane-loading">
            <el-skeleton :rows="4" animated />
          </div>
          <el-empty
            v-else-if="inboxRecentQuery.isError.value"
            :description="inboxRecentQuery.error.value?.message"
          />
          <el-empty
            v-else-if="inboxItems.length === 0"
            :description="$t('pages.dashboard.workspace.myMessages.empty')"
          />
          <template v-else>
            <div class="pane-list inbox-list">
              <div
                v-for="item in inboxItems"
                :key="item.id"
                class="inbox-item"
                @click="router.push('/internal-message/inbox')"
              >
                <span
                  class="inbox-item__dot"
                  :class="{ 'inbox-item__dot--unread': item.status === 'RECEIVED' }"
                />
                <span
                  class="inbox-item__title"
                  :class="{ 'inbox-item__title--unread': item.status === 'RECEIVED' }"
                >
                  {{ item.title || "-" }}
                </span>
                <span class="inbox-item__time">{{ formatDateTime(item.createdAt ?? "") }}</span>
              </div>
            </div>
            <div class="pane-pagination">
              <el-pagination
                v-model:current-page="inboxPage"
                size="small"
                background
                layout="total, prev, pager, next"
                :total="inboxTotal"
                :page-size="PANE_PAGE_SIZE"
                :pager-count="5"
              />
            </div>
          </template>
        </el-card>
      </el-col>

      <!-- 我的活跃会话 -->
      <el-col :xs="24" :md="12" class="pane-col">
        <el-card shadow="hover" class="pane-card">
          <template #header>
            <div class="pane-header">
              <span class="card-title">{{ $t("pages.dashboard.workspace.mySessions.title") }}</span>
              <el-button link type="primary" @click="router.push('/system/online-sessions')">
                {{ $t("pages.dashboard.workspace.mySessions.manage") }}
              </el-button>
            </div>
          </template>
          <div v-if="sessionsQuery.isLoading.value" class="pane-loading">
            <el-skeleton :rows="4" animated />
          </div>
          <el-empty
            v-else-if="sessionsQuery.isError.value"
            :description="sessionsQuery.error.value?.message"
          />
          <el-empty
            v-else-if="allSessions.length === 0"
            :description="$t('pages.dashboard.workspace.mySessions.empty')"
          />
          <template v-else>
            <div class="pane-list session-list">
              <div v-for="item in sessionItems" :key="item.jti" class="session-item">
                <div class="session-item__tags">
                  <el-tag :type="item.clientType === 'app' ? 'warning' : 'primary'" effect="light">
                    {{ item.clientType === "app"
                      ? $t("pages.online_session.clientApp")
                      : $t("pages.online_session.clientAdmin") }}
                  </el-tag>
                  <el-tag v-if="item.current" type="success" effect="light">
                    {{ $t("pages.online_session.currentSession") }}
                  </el-tag>
                  <span class="session-item__ip">{{ item.ipAddress || "-" }}</span>
                </div>
                <span class="session-item__time">{{ formatDateTime(item.loginAt ?? "") }}</span>
              </div>
            </div>
            <div class="pane-pagination">
              <el-pagination
                v-model:current-page="sessionPage"
                size="small"
                background
                layout="total, prev, pager, next"
                :total="sessionTotal"
                :page-size="PANE_PAGE_SIZE"
                :pager-count="5"
              />
            </div>
          </template>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { keepPreviousData } from "@tanstack/vue-query";
import SvgIcon from "@/components/SvgIcon/index.vue";
import { $t, i18n } from "@/core/i18n";
import { useAppUserStore } from "@/stores";
import { useListMyOnlineSessions } from "@/api/composables/online-session";
import { useListUserInbox } from "@/api/composables/internal-message";
import { PaginationQuery } from "@/core/transport/rest";
import { formatDateTime } from "@/utils/date";

/**
 * 工作台：操作入口聚合页（概览组的第二页，与分析页的"数据洞察"分工）。
 * 与 react 端同构：快捷入口 + 我的站内信 + 我的活跃会话。
 * labelKey 直接指到 routes 命名空间的既有菜单文案（ele 是嵌套键位），不新增同义 key。
 * 快捷入口不做权限过滤——与「菜单可见但禁止访问」口径一致，无权限由路由闸渲染 403。
 */
const QUICK_LINKS = [
  { path: "/opm/users", labelKey: "routes.opm.user", icon: "lucide:user" },
  { path: "/permission/roles", labelKey: "routes.permission.role", icon: "lucide:users" },
  { path: "/permission/menus", labelKey: "routes.permission.menu", icon: "lucide:menu" },
  { path: "/system/dict", labelKey: "routes.system.dict", icon: "lucide:book-open" },
  { path: "/system/tasks", labelKey: "routes.system.task", icon: "lucide:clock" },
  { path: "/system/files", labelKey: "routes.system.file", icon: "lucide:folder" },
  { path: "/system/online-sessions", labelKey: "routes.system.onlineSessions", icon: "lucide:globe" },
  { path: "/log/login-audit-logs", labelKey: "routes.log.loginAuditLog", icon: "lucide:file-search" },
];

const router = useRouter();
const userStore = useAppUserStore();

// 问候时段与日期：挂载/语言切换时计算一次，避免跨午夜渲染抖动
const greetingKey = computed(() => {
  const hour = new Date().getHours();
  if (hour < 6) return "night";
  if (hour < 12) return "morning";
  if (hour < 18) return "afternoon";
  return "evening";
});

const displayName = computed(
  () => userStore.userInfo?.nickname || userStore.userInfo?.username || ""
);

const greetingText = computed(() => {
  const base = `pages.dashboard.workspace.greeting.${greetingKey.value}`;
  return displayName.value
    ? $t(`${base}With`, { name: displayName.value })
    : $t(base);
});

const locale = computed(() => {
  const l = i18n.global.locale as unknown as { value?: string };
  return l?.value || "zh-CN";
});

const todayText = computed(() => {
  const date = new Intl.DateTimeFormat(locale.value, {
    year: "numeric",
    month: "long",
    day: "numeric",
    weekday: "long",
  }).format(new Date());
  return $t("pages.dashboard.workspace.today", { date });
});

// 站内信未读数与最近消息：queryKey 与收件箱页共用（PaginationQuery 结构相同），缓存天然共享。
// userInfo 未就绪时禁用查询：空串/undefined 会被 PaginationQuery 剥掉，
// 请求变成无 recipient 过滤的租户级查询，列出其他用户的收件记录。
const hasUserId = computed(() => !!userStore.userInfo?.id);
const unreadQuery = useListUserInbox(
  new PaginationQuery({
    paging: { page: 1, pageSize: 1 },
    formValues: {
      recipient_user_id: String(userStore.userInfo?.id ?? ""),
      status: "RECEIVED",
    },
  }),
  { enabled: hasUserId }
);
const unreadCount = computed(() => unreadQuery.data.value?.total ?? 0);

// ============================================================
// 两卡分页：站内信走服务端分页（PaginationQuery 按页换实例，vue-query
// 深解包 queryKey 后逐页分缓存）；活跃会话的 ListMyOnlineSession 后端
// 无分页参数（个人维度量少），取全量后客户端切片，UI 与站内信对齐。
// 每页行数与 .pane-list 固定高度（7 × 40px）联动，两卡等高不受
// 末页/单页数据量影响。
// ============================================================
const PANE_PAGE_SIZE = 7;

const inboxPage = ref(1);
const inboxRecentQuery = useListUserInbox(
  computed(
    () =>
      new PaginationQuery({
        paging: { page: inboxPage.value, pageSize: PANE_PAGE_SIZE },
        formValues: { recipient_user_id: String(userStore.userInfo?.id ?? "") },
      })
  ),
  // 翻页取数期间保留上一页数据，列表不闪空态
  { enabled: hasUserId, placeholderData: keepPreviousData }
);
const inboxItems = computed(() => inboxRecentQuery.data.value?.items ?? []);
const inboxTotal = computed(() => Number(inboxRecentQuery.data.value?.total ?? 0));

const sessionsQuery = useListMyOnlineSessions();
const allSessions = computed(() => sessionsQuery.data.value?.items ?? []);
const sessionTotal = computed(() => allSessions.value.length);
const sessionPage = ref(1);
const sessionItems = computed(() => {
  const start = (sessionPage.value - 1) * PANE_PAGE_SIZE;
  return allSessions.value.slice(start, start + PANE_PAGE_SIZE);
});

// 总数收缩（会话在他处被吊销/过期）时把页码拉回合法范围
watch(sessionTotal, (total) => {
  const maxPage = Math.max(1, Math.ceil(total / PANE_PAGE_SIZE));
  if (sessionPage.value > maxPage) sessionPage.value = maxPage;
});
watch(inboxTotal, (total) => {
  const maxPage = Math.max(1, Math.ceil(total / PANE_PAGE_SIZE));
  if (inboxPage.value > maxPage) inboxPage.value = maxPage;
});
</script>

<style lang="scss" scoped>
.workspace-page {
  padding: 16px;
}

.mb-5-xs {
  margin-bottom: 16px;
}

@media (min-width: 992px) {
  .mb-5-xs {
    margin-bottom: 0;
  }
}

/* 两卡等高（见模板注释）：列转 flex 让卡片撑满列高，md 起再抬高到共同 min-height */
.pane-col {
  display: flex;
}

.pane-col .pane-card {
  width: 100%;
}

/* 卡体纵排：成功态 = 定高列表 + 贴底分页脚（margin-top:auto），两卡脚注对齐 */
.pane-col .pane-card :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
}

@media (min-width: 768px) {
  .pane-col .pane-card {
    min-height: 360px;
  }
}

.welcome-card {
  margin-bottom: 16px;
  border-radius: 12px;

  .welcome-body {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .welcome-greeting {
    font-size: 18px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .welcome-date {
    margin-top: 4px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.quick-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  margin-bottom: 8px;
  cursor: pointer;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  transition: background 0.2s;

  &:hover {
    background: var(--el-fill-color);
  }

  &__icon {
    flex-shrink: 0;
    color: var(--el-color-primary);
  }

  &__label {
    font-size: 14px;
    color: var(--el-text-color-regular);
  }
}

/* 列表区固定 7 行高（PANE_PAGE_SIZE × 行高 40px，行高压成等值保证两卡行对齐）；
   overflow hidden 兜底异常数据量，末页不满 7 行也不塌高度 */
.pane-list {
  height: 280px;
  overflow: hidden;
}

/* 两卡分页脚注：结构一致保证等高；贴底对齐。pager-count 压到 5 后正常
   单行放得下；el-pagination 自身 nowrap，这里放开 wrap + 卡体 auto 贴底，
   超窄卡片时 total 与页码换行而不裁剪 */
.pane-pagination {
  display: flex;
  justify-content: center;
  margin-top: auto;
  padding-top: 12px;

  :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
    row-gap: 4px;
  }
}

.inbox-item {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 4px;
  box-sizing: border-box;
  cursor: pointer;
  border-bottom: 1px solid var(--el-border-color-extra-light);

  &:last-child {
    border-bottom: none;
  }

  &__dot {
    flex-shrink: 0;
    width: 6px;
    height: 6px;
    background: transparent;
    border-radius: 50%;

    &--unread {
      background: var(--el-color-primary);
    }
  }

  &__title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    font-size: 13px;
    color: var(--el-text-color-regular);
    text-overflow: ellipsis;
    white-space: nowrap;

    &--unread {
      font-weight: 600;
      color: var(--el-text-color-primary);
    }
  }

  &__time {
    flex-shrink: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  height: 40px;
  padding: 0 4px;
  box-sizing: border-box;
  border-bottom: 1px solid var(--el-border-color-extra-light);

  &:last-child {
    border-bottom: none;
  }

  &__tags {
    display: flex;
    flex: 1;
    gap: 8px;
    align-items: center;
    min-width: 0;
  }

  &__ip {
    overflow: hidden;
    font-size: 13px;
    color: var(--el-text-color-regular);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__time {
    flex-shrink: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}
</style>
