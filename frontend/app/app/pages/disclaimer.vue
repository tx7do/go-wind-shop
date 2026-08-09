<script setup lang="ts">
definePageMeta({ layout: 'default' })
const { t } = useI18n()
useHead({ title: t('legal.disclaimer.title') })
const items = [t('legal.disclaimer.item1'), t('legal.disclaimer.item2')]
</script>

<template>
  <LayoutLegalPage
    icon="carbon:document"
    :title="t('legal.disclaimer.title')"
    :description="t('legal.disclaimer.description')"
    :items="items"
  />
</template>
