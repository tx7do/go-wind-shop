<template>
  <ProModal
    v-model:visible="visible"
    :title="$t('pages.mall.groupon.groupsOf', { name: activityName })"
    :config="{ component: 'drawer', drawer: { size: DRAWER_WIDTH } }"
  >
    <div v-if="groupsQuery.isPending" class="p-8 text-center" v-loading="true" />
    <ElEmpty v-else-if="groups.length === 0" :description="$t('pages.mall.groupon.noGroups')" />
    <ElTable v-else :data="groups" border stripe>
      <ElTableColumn prop="id" label="ID" width="70" />
      <ElTableColumn prop="leaderOrderId" :label="$t('pages.mall.groupon.leaderOrder')" width="110" />
      <ElTableColumn :label="$t('pages.mall.groupon.paidProgress')" width="120">
        <template #default="{ row }">{{ row.paidCount ?? 0 }} / {{ row.targetSize ?? 0 }}</template>
      </ElTableColumn>
      <ElTableColumn :label="$t('pages.mall.groupon.groupStatus')" width="110">
        <template #default="{ row }">
          <ElTag :type="groupTagType(row.status)">
            {{ $t(`pages.mall.groupon.groupStatusLabel.${String(row.status).toLowerCase()}`) }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="expireAt" :label="$t('pages.mall.groupon.expireAt')" />
    </ElTable>
  </ProModal>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";

import ProModal from "@/components/Pro/ProModal/index.vue";
import { DRAWER_WIDTH } from "@/constants";
import { useListGrouponGroups } from "@/api/composables";
import { $t } from "@/core/i18n";

const visible = ref(false);
const activity = ref<any>(null);
const activityName = computed(() => activity.value?.name ?? "");

const groupsQuery = useListGrouponGroups(
  computed(() => ({ activityId: activity.value?.id ?? 0 })),
  { enabled: computed(() => visible.value && !!activity.value?.id) } as any
) as any;
const groups = computed(() => (groupsQuery.data.value as any)?.items ?? []);

function groupTagType(status: string) {
  switch (status) {
    case "SUCCESS": return "success";
    case "FAILED": return "danger";
    default: return "warning";
  }
}

function open(row: any) {
  activity.value = row;
  visible.value = true;
}

defineExpose({ open });
</script>
