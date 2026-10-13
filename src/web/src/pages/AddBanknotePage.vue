<template>
  <div class="container">
    <div class="form-wrapper">
      <DesktopPageContext :title="t('banknotes.addTitle')">
        <template #actions>
          <AppIconButton :active="mode === 'manual'" :title="t('addCoin.manualMode')" @click="mode = 'manual'">
            <PencilLine :size="20" />
          </AppIconButton>
          <AppIconButton :active="mode === 'catalog'" :title="t('addCoin.catalogMode')" @click="openCatalog">
            <Library :size="20" />
          </AppIconButton>
        </template>
        <header class="page-header">
          <h1>{{ t('banknotes.addTitle') }}</h1>
          <div class="ml-auto flex shrink-0 gap-[0.35rem]">
            <button type="button" class="chip border border-border-subtle" :class="{ 'border-gold': mode === 'manual' }" data-testid="banknote-mode-manual" @click="mode = 'manual'">
              {{ t('addCoin.manualMode') }}
            </button>
            <button type="button" class="chip border border-border-subtle" :class="{ 'border-gold': mode === 'catalog' }" data-testid="banknote-mode-catalog" @click="openCatalog">
              {{ t('addCoin.catalogMode') }}
            </button>
          </div>
        </header>
      </DesktopPageContext>

      <NumistaCatalogBrowser v-if="catalogOpened" v-show="mode === 'catalog'" category="banknote" @use="applyCatalogSelection" />

      <template v-if="mode === 'manual'">
        <p v-if="catalogTitle" class="mb-4 text-body text-text-secondary" role="status" data-testid="banknote-from-catalog">
          {{ t('banknoteForm.fromCatalog', { title: catalogTitle }) }}
        </p>
        <BanknoteForm ref="formRef" :form="form" :submit-label="t('banknoteForm.save')" :loading="saving" @submit="save" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
// Adds a banknote: typed in, or started from a banknote type of the Numista
// catalogue. Saves it, then its photos (the user's own first, else the ones
// chosen in the catalogue), and opens it.
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { importCoinImage, listCatalogs, uploadImage } from '@/api/client'
import type { CatalogRegistry, Coin, CoinMutationPayload, CoinReference } from '@/types'
import { Library, PencilLine } from 'lucide-vue-next'
import BanknoteForm from '@/components/BanknoteForm.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import NumistaCatalogBrowser from '@/components/catalog/NumistaCatalogBrowser.vue'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { useDialog } from '@/composables/useDialog'
import { useCoinsStore } from '@/stores/coins'
import { t } from '@/i18n'
import { buildBanknotePrefill } from '@/utils/banknotePrefill'
import { photoCreditLine, type RealPhotoChoice } from '@/utils/coinPhotos'
import type { CatalogSelection } from '@/utils/numistaCatalogPrefill'

const router = useRouter()
const store = useCoinsStore()
const { showAlert } = useDialog()
const { categoryOptions, eraOptions, loadOptions } = useCoinOptions()

const mode = ref<'manual' | 'catalog'>('manual')
const catalogOpened = ref(false)
const saving = ref(false)
const formRef = ref<InstanceType<typeof BanknoteForm> | null>(null)
const catalogTitle = ref('')
const catalogPhotos = ref<RealPhotoChoice>({})

// Banknotes are modern; the configured category and era closest to that.
function configured(options: string[], wanted: string, fallback: string): string {
  return options.find((option) => option.toLowerCase() === wanted) ?? options.find((option) => option.toLowerCase() === fallback) ?? fallback
}

function emptyBanknote(): Partial<Coin> {
  return {
    kind: 'banknote',
    name: '',
    category: configured(categoryOptions.value, 'modern', 'other') as Coin['category'],
    era: eraOptions.value.find((era) => era.toLowerCase() === 'modern') ?? '',
    material: 'Paper',
    country: '',
    ruler: '',
    denomination: '',
    series: '',
    rarity: '',
    dateRange: '',
    year: null,
    serialNumber: '',
    signatures: '',
    noteVariant: '',
    grade: '',
    certifiedGrade: '',
    certificationNumber: '',
    printer: '',
    widthMm: null,
    heightMm: null,
    watermark: '',
    obverseInscription: '',
    reverseInscription: '',
    obverseDescription: '',
    reverseDescription: '',
    purchasePrice: null,
    currentValue: null,
    purchaseDate: null,
    purchaseLocation: '',
    notes: '',
    references: [],
    isWishlist: false,
  }
}

const form = reactive<Partial<Coin>>(emptyBanknote())

onMounted(async () => {
  await loadOptions()
  // The configured options arrive after the form was made.
  if (!form.name) Object.assign(form, { category: emptyBanknote().category, era: emptyBanknote().era })
})

function openCatalog() {
  catalogOpened.value = true
  mode.value = 'catalog'
}

async function applyCatalogSelection(selection: CatalogSelection) {
  let registry: CatalogRegistry[] = []
  try {
    registry = await listCatalogs()
  } catch {
    // Without the registry only the Numista reference is saved; others go to notes.
  }
  const { fields, references } = buildBanknotePrefill(selection, { categories: categoryOptions.value, eras: eraOptions.value, registry })
  Object.assign(form, emptyBanknote(), fields, { references: references as CoinReference[] })
  catalogTitle.value = selection.type.title
  catalogPhotos.value = selection.photos ?? {}
  mode.value = 'manual'
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function text(value: string | null | undefined): string | undefined {
  return value?.trim() || undefined
}

function buildPayload(source: Partial<Coin>): CoinMutationPayload {
  return {
    kind: 'banknote',
    name: source.name?.trim() || t('banknoteForm.untitled'),
    category: source.category || 'Other',
    era: source.era?.trim() ?? '',
    material: source.material || 'Paper',
    country: text(source.country),
    ruler: text(source.ruler),
    denomination: text(source.denomination),
    series: text(source.series),
    rarity: source.rarity || undefined,
    dateRange: text(source.dateRange),
    year: source.year ?? null,
    serialNumber: text(source.serialNumber),
    signatures: text(source.signatures),
    noteVariant: source.noteVariant || '',
    grade: text(source.grade),
    certifiedGrade: text(source.certifiedGrade),
    certificationNumber: text(source.certificationNumber),
    // A graded note is sealed in its holder.
    presentation: source.certifiedGrade ? 'slab' : undefined,
    printer: text(source.printer),
    widthMm: source.widthMm ?? null,
    heightMm: source.heightMm ?? null,
    watermark: text(source.watermark),
    obverseInscription: text(source.obverseInscription),
    reverseInscription: text(source.reverseInscription),
    obverseDescription: text(source.obverseDescription),
    reverseDescription: text(source.reverseDescription),
    purchasePrice: source.purchasePrice ?? undefined,
    currentValue: source.currentValue ?? undefined,
    purchaseDate: source.purchaseDate || undefined,
    purchaseLocation: text(source.purchaseLocation),
    notes: text(source.notes),
    isWishlist: false,
    references: (source.references ?? []).map((reference) => ({ catalog: reference.catalog, number: reference.number, volume: reference.volume, uri: reference.uri })),
  }
}

async function save() {
  if (saving.value) return
  saving.value = true
  let banknoteId: number | null = null
  try {
    banknoteId = (await store.addCoin(buildPayload(form))).id
    // The user's own photos first; else the real photos chosen in the catalogue.
    const own = { obverse: formRef.value?.obverseFile ?? null, reverse: formRef.value?.reverseFile ?? null }
    for (const side of ['obverse', 'reverse'] as const) {
      const primary = side === 'obverse' || !(own.obverse || catalogPhotos.value.obverse)
      const file = own[side]
      const remote = catalogPhotos.value[side]
      if (file) await uploadImage(banknoteId, file, side, primary, false)
      else if (remote) await importCoinImage(banknoteId, { url: remote.imageUrl, imageType: side, isPrimary: primary, credit: photoCreditLine(side, remote) })
    }
    await router.push(`/coin/${banknoteId}`)
  } catch {
    if (banknoteId === null) {
      await showAlert(t('banknoteForm.saveFailed'), { title: t('common.error') })
    } else {
      // The note is saved; only a photo failed.
      await showAlert(t('banknoteForm.photoFailed'), { title: t('common.error') })
      await router.push(`/coin/${banknoteId}`)
    }
  } finally {
    saving.value = false
  }
}
</script>
