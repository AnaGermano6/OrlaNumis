<template>
  <section class="space-y-4" :aria-label="t('sets.album.title')" data-testid="set-album">
    <!-- How far the set is. -->
    <div class="rounded-md border border-border-subtle bg-card p-4">
      <div class="mb-2 flex flex-wrap items-baseline justify-between gap-2">
        <span class="text-base font-medium text-text-primary" data-testid="album-progress">{{ t('sets.album.progress', { owned: ownedCount, total: slots.length }) }}</span>
        <span class="text-xl font-semibold text-gold">{{ percent }}%</span>
      </div>
      <div class="h-2 overflow-hidden rounded-full bg-input" role="progressbar" :aria-valuenow="percent" aria-valuemin="0" aria-valuemax="100" :aria-label="t('sets.album.title')">
        <div class="h-full rounded-full bg-gold transition-[width] duration-300" :style="{ width: `${percent}%` }"></div>
      </div>
    </div>

    <div v-if="missingCount > 0 || groupOptions.length > 1" class="flex flex-wrap items-center gap-2">
      <template v-if="missingCount > 0 && ownedCount > 0">
        <button
          v-for="option in SHOW"
          :key="option"
          type="button"
          class="chip border"
          :class="option === show ? 'border-gold text-gold' : 'border-border-subtle'"
          :aria-pressed="option === show"
          :data-testid="`album-show-${option}`"
          @click="show = option"
        >
          {{ t(`sets.album.show.${option}`) }} · {{ option === 'all' ? slots.length : option === 'owned' ? ownedCount : missingCount }}
        </button>
      </template>
      <template v-if="groupOptions.length > 1">
        <span v-if="missingCount > 0 && ownedCount > 0" class="mx-1 h-5 w-px bg-border-subtle" aria-hidden="true"></span>
        <button
          v-for="option in groupOptions"
          :key="option"
          type="button"
          class="chip border"
          :class="option === group ? 'border-gold text-gold' : 'border-border-subtle'"
          :aria-pressed="option === group"
          :data-testid="`album-group-${option}`"
          @click="group = option"
        >
          {{ t(`sets.album.group.${option}`) }}
        </button>
      </template>
    </div>

    <p v-if="!shown.length" class="text-body text-text-secondary" data-testid="album-empty">{{ t('sets.album.nothing') }}</p>
    <section v-for="section in sections" :key="section.id" :aria-label="section.title || undefined">
      <h3 v-if="section.title" class="mb-2 mt-0 text-base font-medium text-text-secondary" data-testid="album-section">
        {{ section.title }}
        <span class="text-chip text-text-muted">· {{ t('sets.catalogue.progress', { owned: section.owned, total: section.slots.length }) }}</span>
      </h3>
      <ul class="m-0 grid list-none grid-cols-2 gap-3 p-0 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6">
        <li
          v-for="slot in section.slots"
          :key="slot.key"
          class="flex min-w-0 flex-col overflow-hidden rounded-md"
          :class="slot.owned ? 'border border-border-subtle bg-card' : 'border-2 border-dashed border-border-subtle bg-transparent'"
          data-testid="album-slot"
          :data-owned="slot.owned"
        >
          <component
            :is="slotLink(slot).tag"
            v-bind="slotLink(slot).attrs"
            class="flex min-w-0 flex-1 flex-col text-left no-underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            @click="slotLink(slot).tag === 'button' ? emit('select', slot) : undefined"
          >
            <span class="relative block aspect-square" :class="slot.owned ? 'bg-input' : ''">
              <AuthenticatedImage v-if="slot.coinImage" :media-path="slot.coinImage" :alt="slot.title" loading="lazy" class="h-full w-full object-contain p-2">
                <template #fallback><span class="flex h-full w-full items-center justify-center text-text-primary/30"><Coins :size="40" :stroke-width="1" /></span></template>
              </AuthenticatedImage>
              <img
                v-else-if="slot.imageUrl"
                :src="slot.imageUrl"
                :alt="slot.title"
                loading="lazy"
                class="h-full w-full object-contain p-2"
                :class="slot.owned ? '' : 'opacity-40 grayscale'"
              />
              <span v-else class="flex h-full w-full items-center justify-center text-text-primary/30" data-testid="album-no-image">
                <Coins :size="40" :stroke-width="1" />
              </span>
              <span class="absolute left-2 top-2">
                <span v-if="slot.owned" class="inline-flex items-center gap-1 rounded-full bg-gain px-2 py-[0.1rem] text-chip font-semibold text-surface" data-testid="album-owned">
                  <Check :size="12" /> {{ t('collectionOverview.releases.owned') }}
                </span>
                <span v-else-if="slot.wishlist" class="rounded-full border border-border-subtle bg-card px-2 py-[0.1rem] text-chip font-semibold text-text-secondary">{{ t('sets.album.wishlist') }}</span>
                <span v-else class="rounded-full border border-border-subtle bg-card px-2 py-[0.1rem] text-chip font-semibold text-text-muted" data-testid="album-missing">{{ t('sets.album.missing') }}</span>
              </span>
            </span>
            <span class="flex min-w-0 flex-1 flex-col gap-1 p-3">
              <span v-if="slot.detail" class="truncate text-label font-semibold uppercase tracking-[0.08em] text-text-muted" :title="slot.detail">{{ slot.detail }}</span>
              <span class="line-clamp-2 text-body font-medium leading-snug" :class="slot.owned ? 'text-text-primary' : 'text-text-secondary'" :title="slot.title" data-testid="album-title">{{ slot.title }}</span>
              <span v-if="slot.year" class="text-chip text-gold">{{ slot.year }}</span>
            </span>
          </component>
        </li>
      </ul>
    </section>
  </section>
</template>

<script setup lang="ts">
// A set as an album: a place for each coin, the collection's with its
// photo and "In my collection", the missing ones dashed. Filters show what
// the set has or still lacks, and the places can be grouped by year or
// country.
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { Check, Coins } from 'lucide-vue-next'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { intlLocale, t } from '@/i18n'
import type { AlbumSlot } from '@/utils/setAlbum'

const SHOW = ['all', 'owned', 'missing'] as const
type Grouping = 'none' | 'year' | 'country'

const props = withDefaults(defineProps<{
  slots: AlbumSlot[]
  /** How the places are grouped at first. */
  defaultGroup?: Grouping
  /** Slots the page acts on (an agentic set's) instead of opening a page. */
  selectable?: boolean
}>(), { defaultGroup: 'none', selectable: false })

const emit = defineEmits<{ select: [slot: AlbumSlot] }>()

const show = ref<(typeof SHOW)[number]>('all')
const group = ref<Grouping>(props.defaultGroup)

const ownedCount = computed(() => props.slots.filter((slot) => slot.owned).length)
const missingCount = computed(() => props.slots.length - ownedCount.value)
const percent = computed(() => props.slots.length ? Math.round((ownedCount.value / props.slots.length) * 100) : 0)

// Grouping by year or country only when the places differ in it.
const groupOptions = computed((): Grouping[] => {
  const options: Grouping[] = ['none']
  if (new Set(props.slots.map((slot) => slot.year || 0)).size > 1) options.push('year')
  if (new Set(props.slots.map((slot) => slot.country ?? '')).size > 1) options.push('country')
  return options
})
watch(groupOptions, (options) => {
  if (!options.includes(group.value)) group.value = options.includes(props.defaultGroup) ? props.defaultGroup : 'none'
}, { immediate: true })

const shown = computed(() => props.slots.filter((slot) => show.value === 'all' || (show.value === 'owned') === slot.owned))

// The places in sections: one without a title, or one per year (newest
// first) or country (by name), the places without one last.
const sections = computed(() => {
  if (group.value === 'none') return shown.value.length ? [{ id: 'all', title: '', slots: shown.value, owned: ownedOf(shown.value) }] : []
  const byKey = new Map<string, AlbumSlot[]>()
  for (const slot of shown.value) {
    const key = group.value === 'year' ? String(slot.year || '') : (slot.country ?? '')
    byKey.set(key, [...(byKey.get(key) ?? []), slot])
  }
  const keys = [...byKey.keys()].sort((a, b) => {
    if (!a || !b) return a ? -1 : b ? 1 : 0
    return group.value === 'year' ? Number(b) - Number(a) : a.localeCompare(b, intlLocale())
  })
  return keys.map((key) => ({ id: key || 'other', title: key || t('sets.album.noGroup'), slots: byKey.get(key) ?? [], owned: ownedOf(byKey.get(key) ?? []) }))
})

function ownedOf(slots: AlbumSlot[]): number {
  return slots.filter((slot) => slot.owned).length
}

// What a place opens: the collection's coin, else the source's page of the
// coin to find; an agentic set's places are the page's to act on.
function slotLink(slot: AlbumSlot): { tag: unknown; attrs: Record<string, unknown> } {
  if (props.selectable) return { tag: 'button', attrs: { type: 'button' } }
  if (slot.coinId) return { tag: RouterLink, attrs: { to: { name: 'coin-detail', params: { id: slot.coinId } } } }
  if (slot.pageUrl) return { tag: 'a', attrs: { href: slot.pageUrl, target: '_blank', rel: 'noopener noreferrer', title: t('collectionOverview.releases.open') } }
  return { tag: 'div', attrs: {} }
}
</script>
