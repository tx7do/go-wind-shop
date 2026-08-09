<script setup lang="ts">
definePageMeta({ layout: 'default' })
const { t } = useI18n()
useHead({ title: t('legal.about.title') })
const items = [t('legal.about.item1'), t('legal.about.item2')]
</script>

<template>
  <LayoutLegalPage
    icon="carbon:document"
    :title="t('legal.about.title')"
    :description="t('legal.about.description')"
    :items="items"
  />
</template>
