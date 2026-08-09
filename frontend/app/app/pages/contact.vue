<script setup lang="ts">
definePageMeta({ layout: 'default' })
const { t } = useI18n()
useHead({ title: t('legal.contact.title') })
const items = [t('legal.contact.item1'), t('legal.contact.item2')]
</script>

<template>
  <LayoutLegalPage
    icon="carbon:document"
    :title="t('legal.contact.title')"
    :description="t('legal.contact.description')"
    :items="items"
  />
</template>
