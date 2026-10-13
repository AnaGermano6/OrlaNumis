<template>
  <li class="flex min-w-0 flex-col overflow-hidden rounded-md border border-border-subtle bg-card" data-testid="release">
    <!-- Without a picture, a lower space keeps the list easy to scan. -->
    <a
      :href="release.pageUrl"
      target="_blank"
      rel="noopener noreferrer"
      class="relative block bg-input"
      :class="release.imageUrl ? 'aspect-square' : 'aspect-[2/1]'"
      :title="t('collectionOverview.releases.open')"
    >
      <img v-if="release.imageUrl" :src="release.imageUrl" :alt="release.title" loading="lazy" class="h-full w-full object-contain p-2" />
      <span v-else class="flex h-full w-full items-center justify-center text-text-primary/30" data-testid="release-no-image">
        <Coins :size="40" :stroke-width="1" />
      </span>
      <span class="absolute left-2 top-2 flex flex-col items-start gap-1">
        <span v-if="release.upcoming" class="rounded-full bg-gold px-2 py-[0.1rem] text-chip font-semibold text-surface" data-testid="release-upcoming">{{ t('collectionOverview.releases.upcoming') }}</span>
        <span v-if="release.owned" class="inline-flex items-center gap-1 rounded-full bg-gain px-2 py-[0.1rem] text-chip font-semibold text-surface" data-testid="release-owned">
          <Check :size="12" /> {{ t('collectionOverview.releases.owned') }}
        </span>
      </span>
    </a>
    <div class="flex min-w-0 flex-1 flex-col gap-1 p-3">
      <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">
        <template v-if="showCountry">{{ releaseCountryName(release.country) }} · </template>{{ releaseFaceValues(release) }}
      </span>
      <span class="line-clamp-3 text-body font-medium leading-snug text-text-primary" :title="release.title">{{ release.title }}</span>
      <span class="text-chip text-gold">{{ releasePeriod(release) }}</span>
      <span v-if="release.variants?.length" class="text-chip text-text-muted">{{ release.variants.join(' · ') }}</span>
      <span v-if="release.volume" class="text-chip text-text-muted">{{ t('collectionOverview.releases.volume', { volume: formatCount(release.volume) }) }}</span>
    </div>
  </li>
</template>

<script setup lang="ts">
// A coin the ECB lists or a mint sells: its picture (opening the source's
// page), country, face values, title, date, variants and mintage, marked when
// it is to come or the collection has it.
import { Check, Coins } from 'lucide-vue-next'
import type { CoinRelease } from '@/api/client'
import { t } from '@/i18n'
import { formatCount, releaseCountryName, releaseFaceValues, releasePeriod } from '@/utils/coinReleases'

withDefaults(defineProps<{
  release: CoinRelease
  /** Off where the country is already said, as in a country's catalogue. */
  showCountry?: boolean
}>(), { showCountry: true })
</script>
