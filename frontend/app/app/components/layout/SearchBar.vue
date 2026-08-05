<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
const searchQuery = ref('')

// 回车或点击搜索时跳转到搜索结果页，携带 ?q= 关键字。
// 空关键字不跳转，避免进入无意义结果页。
const handleSearch = () => {
  const q = searchQuery.value.trim()
  if (!q) return
  navigateTo(localePath('/search') + `?q=${encodeURIComponent(q)}`)
}
</script>

<template>
  <div class="mx-2 hidden h-11 max-w-80 flex-1 items-center md:flex lg:max-w-80">
    <div class="relative w-full">
      <XIcon icon="lucide:search" width="16" height="16" class="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
      <UiInput
        class="h-full w-full pl-8"
        v-model="searchQuery"
        @keyup.enter="handleSearch"
        :placeholder="t('navbar.top.search_placeholder')"
      />
    </div>
  </div>
</template>
