<template>
  <div class="grid min-w-0 gap-3">
    <form class="flex min-w-0 gap-2" role="search" @submit.prevent="search">
      <input
        v-model="query"
        class="form-input min-w-0 flex-1"
        type="search"
        maxlength="200"
        :aria-label="t('photos.query')"
        :placeholder="t('photos.queryPlaceholder')"
      />
      <button type="submit" class="btn btn-secondary shrink-0" :disabled="searching || !query.trim()">
        {{ searching ? t('common.searching') : t('common.search') }}
      </button>
    </form>

    <p v-if="searching && !visiblePhotos.length" class="m-0 text-body text-text-secondary" role="status">{{ t('photos.searching') }}</p>
    <p v-else-if="failed" class="m-0 text-body text-text-secondary" role="status">{{ t('photos.searchFailed') }}</p>
    <p v-else-if="searched && !visiblePhotos.length" class="m-0 text-body text-text-secondary" role="status">{{ t('photos.noneFound') }}</p>

    <ul v-if="visiblePhotos.length" class="m-0 grid list-none grid-cols-2 gap-3 p-0 sm:grid-cols-3 lg:grid-cols-4" :aria-label="t('photos.results')">
      <li
        v-for="photo in visiblePhotos"
        :key="photo.imageUrl"
        class="flex min-w-0 flex-col gap-2 rounded-sm border p-2"
        :class="chosenSide(photo) ? 'border-gold bg-gold-glow' : 'border-border-subtle bg-card'"
      >
        <CatalogPhoto
          :url="photo.thumbnail"
          :via="photoLoadVia(photo)"
          :alt="photo.title || t('photos.photo')"
          :credit="photoCaption(photo)"
          fluid
          @unavailable="markUnavailable(photo)"
        />
        <div class="grid grid-cols-2 gap-1">
          <button
            v-for="side in SIDES"
            :key="side"
            type="button"
            class="btn btn-sm min-w-0 justify-center px-1"
            :class="isChosen(side, photo) ? 'btn-primary' : 'btn-secondary'"
            :aria-pressed="isChosen(side, photo)"
            @click="toggle(side, photo)"
          >
            {{ side === 'obverse' ? t('common.obverse') : t('common.reverse') }}
          </button>
        </div>
        <a v-if="photo.pageUrl" :href="photo.pageUrl" target="_blank" rel="noopener noreferrer" class="text-2xs text-gold underline">{{ t('photos.viewSource') }}</a>
      </li>
    </ul>
    <p v-if="visiblePhotos.length" class="m-0 text-chip text-text-muted">{{ t('photos.licenseHint') }}</p>
  </div>
</template>

<script setup lang="ts">
// Find real photos of a coin (official ECB and INCM images, Numista catalogue,
// Wikimedia Commons) and pick one for each side. Photos identified as this very
// coin come chosen already. Nothing is saved here; the parent imports the choice.
import { computed, reactive, ref, watch } from 'vue'
import { searchCoinPhotos, type CoinPhotoCandidate } from '@/api/client'
import { t } from '@/i18n'
import CatalogPhoto from '@/components/catalog/CatalogPhoto.vue'
import { numistaLang } from '@/utils/numistaCatalog'
import { photoCaption, photoLoadVia, type CoinPhotoDetails, type PhotoSide, type RealPhotoChoice } from '@/utils/coinPhotos'

const SIDES: PhotoSide[] = ['obverse', 'reverse']

const props = withDefaults(defineProps<{
  initialQuery: string
  numistaTypeId?: number | null
  /** Year, country… so a €2 commemorative's ECB image is found too. */
  details?: CoinPhotoDetails
  modelValue: RealPhotoChoice
}>(), { numistaTypeId: null, details: () => ({}) })

const emit = defineEmits<{
  'update:modelValue': [value: RealPhotoChoice]
}>()

const query = ref(props.initialQuery)
const photos = ref<CoinPhotoCandidate[]>([])
const unavailable = reactive(new Set<string>())
const searching = ref(false)
const searched = ref(false)
const failed = ref(false)
let searchSeq = 0

const visiblePhotos = computed(() => photos.value.filter((photo) => !unavailable.has(photo.imageUrl)))

function isChosen(side: PhotoSide, photo: CoinPhotoCandidate): boolean {
  return props.modelValue[side]?.imageUrl === photo.imageUrl
}

function chosenSide(photo: CoinPhotoCandidate): PhotoSide | null {
  return SIDES.find((side) => isChosen(side, photo)) ?? null
}

function toggle(side: PhotoSide, photo: CoinPhotoCandidate) {
  const next: RealPhotoChoice = { ...props.modelValue }
  if (isChosen(side, photo)) delete next[side]
  else next[side] = photo
  // One photo shows one side.
  const other: PhotoSide = side === 'obverse' ? 'reverse' : 'obverse'
  if (next[side] && next[other]?.imageUrl === photo.imageUrl) delete next[other]
  emit('update:modelValue', next)
}

function markUnavailable(photo: CoinPhotoCandidate) {
  unavailable.add(photo.imageUrl)
  const chosen = SIDES.filter((side) => isChosen(side, photo))
  if (!chosen.length) return
  const next: RealPhotoChoice = { ...props.modelValue }
  for (const side of chosen) delete next[side]
  emit('update:modelValue', next)
}

function chooseExactPhotos(found: CoinPhotoCandidate[]) {
  const next: RealPhotoChoice = { ...props.modelValue }
  for (const photo of found) {
    if (photo.match === 'exact' && photo.side && !next[photo.side]) next[photo.side] = photo
  }
  if (next.obverse !== props.modelValue.obverse || next.reverse !== props.modelValue.reverse) emit('update:modelValue', next)
}

async function search() {
  const text = query.value.trim()
  if (!text && !props.numistaTypeId) return
  const seq = ++searchSeq
  searching.value = true
  failed.value = false
  try {
    const res = await searchCoinPhotos({ q: text || undefined, numistaTypeId: props.numistaTypeId ?? undefined, lang: numistaLang(), ...props.details })
    if (seq !== searchSeq) return
    photos.value = res.data.photos
    failed.value = res.data.status !== 'ok' && !res.data.photos.length
    chooseExactPhotos(res.data.photos)
  } catch {
    if (seq === searchSeq) {
      photos.value = []
      failed.value = true
    }
  } finally {
    if (seq === searchSeq) {
      searching.value = false
      searched.value = true
    }
  }
}

watch(() => [props.initialQuery, props.numistaTypeId] as const, ([initial]) => {
  query.value = initial
  void search()
}, { immediate: true })
</script>
