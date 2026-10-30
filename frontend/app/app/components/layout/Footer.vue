<script setup lang="ts">
type FooterLinkKey = 'about_us' | 'contact_us' | 'non_responsibility' | 'privacy_agreement' | 'terms_of_service'

const { t } = useI18n()
const localePath = useLocalePath()

const footerLinks: { key: FooterLinkKey; labelKey: string }[] = [
  { key: 'about_us', labelKey: 'ui.button.about_us' },
  { key: 'contact_us', labelKey: 'ui.button.contact_us' },
  { key: 'non_responsibility', labelKey: 'ui.button.non_responsibility' },
  { key: 'privacy_agreement', labelKey: 'ui.button.privacy_agreement' },
  { key: 'terms_of_service', labelKey: 'ui.button.terms_of_service' },
]

const footerRouteMap: Record<FooterLinkKey, string> = {
  about_us: '/about',
  contact_us: '/contact',
  non_responsibility: '/disclaimer',
  privacy_agreement: '/privacy',
  terms_of_service: '/terms',
}

const handleFooterLinkClick = (key: FooterLinkKey) => {
  navigateTo(localePath(footerRouteMap[key]))
}
</script>

<template>
  <footer class="w-full border-t border-border bg-card py-6">
    <div class="mx-auto flex w-full max-w-(--layout-max-width) flex-wrap items-center justify-between gap-x-6 gap-y-3 px-4 md:px-8">
      <nav class="flex flex-wrap items-center gap-x-5 gap-y-2" aria-label="Footer links">
        <button
          v-for="link in footerLinks"
          :key="link.key"
          type="button"
          class="text-sm text-muted-foreground transition-colors hover:text-foreground"
          @click="handleFooterLinkClick(link.key)"
        >
          {{ t(link.labelKey) }}
        </button>
      </nav>

      <span class="text-xs text-muted-foreground">
        {{ t('ui.copyright') }}
      </span>
    </div>
  </footer>
</template>
