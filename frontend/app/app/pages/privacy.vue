<script setup lang="ts">
definePageMeta({ layout: 'default' })
const { t } = useI18n()
useHead({ title: t('legal.privacy.title') })
const items = [t('legal.privacy.item1'), t('legal.privacy.item2')]
</script>

<template>
  <LayoutLegalPage
    icon="carbon:document"
    :title="t('legal.privacy.title')"
    :description="t('legal.privacy.description')"
    :items="items"
  />
</template>
