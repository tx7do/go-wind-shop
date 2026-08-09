<script setup lang="ts">
definePageMeta({ layout: 'default' })
const { t } = useI18n()
useHead({ title: t('legal.terms.title') })
const items = [t('legal.terms.item1'), t('legal.terms.item2')]
</script>

<template>
  <LayoutLegalPage
    icon="carbon:document"
    :title="t('legal.terms.title')"
    :description="t('legal.terms.description')"
    :items="items"
  />
</template>
