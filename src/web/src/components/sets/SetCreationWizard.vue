<template>
  <div class="w-full">
    <!-- First, how the set is made. -->
    <div v-if="!path" class="flex flex-col gap-2" data-testid="set-paths">
      <button
        v-for="option in PATHS"
        :key="option.id"
        type="button"
        class="flex w-full items-center gap-3 rounded-md border border-border-subtle bg-card p-3 text-left transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :data-testid="`set-path-${option.id}`"
        @click="choose(option.id)"
      >
        <component :is="option.icon" :size="22" class="shrink-0 text-gold" />
        <span class="min-w-0 flex-1">
          <span class="block font-medium text-text-primary">{{ t(`sets.wizard.paths.${option.id}.title`) }}</span>
          <span class="block text-body text-text-secondary">{{ t(`sets.wizard.paths.${option.id}.body`) }}</span>
        </span>
        <ChevronRight :size="18" class="shrink-0 text-text-muted" />
      </button>
      <p class="mb-0 mt-2 text-chip text-text-muted">
        {{ t('sets.wizard.otherTypes') }}
        <button type="button" class="rounded-sm text-gold underline-offset-2 hover:underline" data-testid="set-path-goal" @click="choose('goal')">{{ t('sets.wizard.paths.goal.title') }}</button>
        ·
        <button type="button" class="rounded-sm text-gold underline-offset-2 hover:underline" data-testid="set-path-agentic" @click="choose('agentic')">{{ t('sets.wizard.paths.agentic.title') }}</button>
      </p>
      <div class="mt-4 flex justify-end">
        <button type="button" class="btn btn-secondary" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
      </div>
    </div>

    <form v-else @submit.prevent="submit">
      <button type="button" class="mb-3 inline-flex items-center gap-1 rounded-sm text-body text-text-secondary transition-colors hover:text-gold" data-testid="set-path-back" @click="path = null">
        <ArrowLeft :size="16" /> {{ t('common.back') }}
      </button>
      <h3 class="mb-3 mt-0 text-base font-semibold text-text-primary" data-testid="set-path-title">{{ pathTitle }}</h3>

      <!-- The catalogue: chosen there, with its coins in view. -->
      <div v-if="path === 'catalogue'" class="space-y-4">
        <p class="m-0 text-body text-text-secondary">{{ t('sets.wizard.catalogueHelp') }}</p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
          <RouterLink to="/sets/commemoratives" class="btn btn-primary" data-testid="set-open-catalogue" @click="$emit('cancel')">
            <BookOpenCheck :size="16" /> {{ t('sets.wizard.openCatalogue') }}
          </RouterLink>
        </div>
      </div>

      <template v-else>
        <div v-if="path === 'challenge'" class="mb-4 space-y-2" role="radiogroup" :aria-label="pathTitle">
          <label
            v-for="option in CHALLENGES"
            :key="option"
            class="flex cursor-pointer items-start gap-3 rounded-md border bg-card p-3 transition-colors"
            :class="challenge === option ? 'border-gold' : 'border-border-subtle hover:border-border-accent'"
            :data-testid="`challenge-${option}`"
          >
            <input v-model="challenge" type="radio" name="challenge" :value="option" class="mt-1 accent-[var(--accent-gold)]" />
            <span class="min-w-0">
              <span class="block font-medium text-text-primary">{{ t(`sets.challenges.${option}.title`) }}</span>
              <span class="block text-body text-text-secondary">{{ t(`sets.challenges.${option}.body`) }}</span>
            </span>
          </label>
          <div v-if="challenge === 'annualSeries'" class="grid gap-3 pt-2 sm:grid-cols-2">
            <div class="form-group m-0">
              <label for="challengeCountry" class="form-label mb-1 block">{{ t('sets.filters.country') }}</label>
              <select id="challengeCountry" v-model="seriesCountry" class="form-input w-full" required>
                <option value="" disabled>{{ t('sets.wizard.chooseCountry') }}</option>
                <option v-for="code in euroCountries" :key="code" :value="code">{{ releaseCountryName(code) }}</option>
              </select>
            </div>
            <div class="form-group m-0">
              <label for="challengeYear" class="form-label mb-1 block">{{ t('sets.wizard.year') }}</label>
              <input id="challengeYear" v-model.number="seriesYear" type="number" inputmode="numeric" :min="EURO_FIRST_YEAR" :max="thisYear + 1" class="form-input w-full" required />
            </div>
          </div>
          <div v-if="challenge === 'jointIssue'" class="pt-2">
            <p v-if="jointIssuesFailed" class="m-0 text-body text-text-secondary" role="alert">{{ t('sets.wizard.noJointIssues') }}</p>
            <p v-else-if="!jointIssues" class="m-0 text-body text-text-muted">{{ t('sets.wizard.loadingJointIssues') }}</p>
            <div v-else class="flex items-center gap-3">
              <img v-if="jointIssuePicture" :src="jointIssuePicture" alt="" class="h-16 w-16 shrink-0 rounded-full bg-input object-contain" data-testid="joint-issue-picture" />
              <div class="form-group m-0 min-w-0 flex-1">
                <label for="jointIssue" class="form-label mb-1 block">{{ t('sets.wizard.jointIssue') }}</label>
                <select id="jointIssue" v-model.number="jointIssueYear" class="form-input w-full" required>
                  <option v-for="issue in jointIssues" :key="issue.year" :value="issue.year">
                    {{ issue.year }} — {{ issue.title }} ({{ t('sets.wizard.jointIssueCountries', { count: issue.coins.length }) }})
                  </option>
                </select>
              </div>
            </div>
          </div>
          <div v-if="challenge === 'birthYear'" class="form-group m-0 pt-2">
            <label for="birthYear" class="form-label mb-1 block">{{ t('sets.wizard.birthYear') }}</label>
            <input id="birthYear" v-model.number="birthYear" type="number" inputmode="numeric" min="1" :max="thisYear" class="form-input w-full sm:w-40" required />
          </div>
        </div>

        <SetSimpleFilters v-if="path === 'filters'" v-model="filters" class="mb-4" />

        <p v-if="path === 'manual'" class="mb-4 mt-0 text-body text-text-secondary">{{ t('sets.wizard.manualHelp') }}</p>

        <div v-if="path === 'goal'" class="form-group mb-4 rounded-sm border border-border-subtle bg-card p-4">
          <span class="section-label">{{ t('sets.wizard.goalTitle') }}</span>
          <p class="mt-2 mb-0 text-body text-text-secondary">{{ t('sets.wizard.goalP1') }}</p>
          <p class="mt-2 mb-0 text-body text-text-secondary">{{ t('sets.wizard.goalP2') }}</p>
          <p class="mt-2 mb-0 text-chip text-text-muted">{{ t('sets.wizard.goalP3') }}</p>
        </div>
        <template v-if="path === 'agentic'">
          <div class="form-group mb-4 rounded-sm border border-border-subtle bg-card p-4">
            <span class="section-label">{{ t('sets.wizard.agenticTitle') }}</span>
            <p class="mt-2 mb-0 text-body text-text-secondary">{{ t('sets.wizard.agenticP1') }}</p>
            <p class="mt-2 mb-0 text-body text-text-secondary">{{ t('sets.wizard.agenticP2') }}</p>
            <p class="mt-2 mb-0 text-chip text-text-muted">{{ t('sets.wizard.agenticP3') }}</p>
          </div>
          <div class="form-group mb-4">
            <label for="agenticPrompt" class="form-label mb-2 block">{{ t('sets.wizard.agenticPrompt') }}</label>
            <textarea id="agenticPrompt" v-model="agenticPrompt" rows="3" maxlength="500" class="form-input w-full" :placeholder="t('sets.wizard.agenticPlaceholder')" required />
          </div>
        </template>

        <div class="form-group mb-4">
          <label for="setName" class="form-label mb-2 block">{{ t('common.name') }}</label>
          <input id="setName" v-model="name" type="text" required maxlength="80" class="form-input w-full" @input="nameEdited = true" />
        </div>
        <div class="form-group mb-4">
          <label for="setDescription" class="form-label mb-2 block">{{ t('common.description') }}</label>
          <textarea id="setDescription" v-model="description" rows="2" maxlength="2000" class="form-input w-full" />
        </div>
        <div class="form-group mb-4">
          <label for="setColor" class="form-label mb-2 block">{{ t('common.color') }}</label>
          <input id="setColor" v-model="color" type="color" class="h-10 w-full rounded-sm border border-border-subtle bg-input p-1 text-text-primary" />
        </div>
        <div class="mt-6 flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="!valid">{{ submitLabel || t('common.create') }}</button>
        </div>
      </template>
    </form>
  </div>
</template>

<script setup lang="ts">
// Making a set, one way of four: from the catalogue (chosen there), a
// ready-made challenge, simple filters, or coins picked by hand. Goal and
// AI-proposed sets stay a link away.
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowLeft, BookOpenCheck, ChevronRight, Hand, SlidersHorizontal, Trophy } from 'lucide-vue-next'
import { normalizeCoinSetType } from '@/types'
import type { ChecklistDefinition, CreateCoinSetRequest, SmartCriteriaGroup } from '@/types'
import { getJointIssues, type JointIssue } from '@/api/client'
import SetSimpleFilters from '@/components/sets/SetSimpleFilters.vue'
import { randomSetColor } from '@/utils/setColors'
import { releaseCountryName } from '@/utils/coinReleases'
import { criteriaToFilters, emptyFilters, EURO_COUNTRIES, filtersName, filtersToCriteria, type SimpleSetFilters } from '@/utils/setFilters'
import { intlLocale, t } from '@/i18n'

type Path = 'catalogue' | 'challenge' | 'filters' | 'manual' | 'goal' | 'agentic'
type Challenge = 'euroCountries' | 'jointIssue' | 'annualSeries' | 'birthYear'

const PATHS = [
  { id: 'catalogue', icon: BookOpenCheck },
  { id: 'challenge', icon: Trophy },
  { id: 'filters', icon: SlidersHorizontal },
  { id: 'manual', icon: Hand },
] as const
const CHALLENGES: Challenge[] = ['euroCountries', 'jointIssue', 'annualSeries', 'birthYear']
/** The challenges the server keeps as checklists, by their id there. */
const CHECKLIST_CHALLENGES: Partial<Record<Challenge, ChecklistDefinition['challenge']>> = {
  euroCountries: 'euro-countries',
  jointIssue: 'joint-issue',
  annualSeries: 'annual-series',
}
const EURO_FIRST_YEAR = 1999

const props = withDefaults(defineProps<{
  initialValue?: Partial<CreateCoinSetRequest>
  submitLabel?: string
}>(), {
  submitLabel: '',
})

const emit = defineEmits<{
  submit: [value: CreateCoinSetRequest, csv?: string]
  cancel: []
}>()

const thisYear = new Date().getFullYear()
const euroCountries = computed(() => [...EURO_COUNTRIES].sort((a, b) => releaseCountryName(a).localeCompare(releaseCountryName(b), intlLocale())))

const path = ref<Path | null>(null)
const name = ref('')
const nameEdited = ref(false)
const description = ref('')
const color = ref(randomSetColor())
const challenge = ref<Challenge>('euroCountries')
const seriesCountry = ref('')
const seriesYear = ref<number>(thisYear)
const birthYear = ref<number | null>(null)
const filters = ref<SimpleSetFilters>(emptyFilters())
const agenticPrompt = ref('')
/** The euro area's joint issues, once read; the year chosen. */
const jointIssues = ref<JointIssue[] | null>(null)
const jointIssuesFailed = ref(false)
const jointIssueYear = ref<number | null>(null)
const jointIssue = computed(() => jointIssues.value?.find((issue) => issue.year === jointIssueYear.value) ?? null)
// The issue's Portuguese version, else its first.
const jointIssuePicture = computed(() => {
  const coins = jointIssue.value?.coins ?? []
  return (coins.find((coin) => coin.country === 'PT') ?? coins[0])?.imageUrl ?? ''
})

function reset(value: Partial<CreateCoinSetRequest> | undefined) {
  name.value = value?.name ?? ''
  nameEdited.value = Boolean(value?.name)
  description.value = value?.description ?? ''
  color.value = value?.color ?? randomSetColor()
  agenticPrompt.value = value?.agenticPrompt ?? ''
  filters.value = criteriaToFilters(value?.smartCriteria as SmartCriteriaGroup | undefined) ?? emptyFilters()
  // A set of a known type opens its own form; a new set starts with the choice.
  const type = value?.setType ? normalizeCoinSetType(value.setType) : 'standard'
  path.value = type === 'goal' ? 'goal' : type === 'agentic' ? 'agentic' : type === 'smart' ? 'filters' : null
}
reset(props.initialValue)
watch(() => props.initialValue, reset)

function choose(next: Path) {
  path.value = next
  if (next === 'challenge') void loadJointIssues()
}

async function loadJointIssues() {
  if (jointIssues.value) return
  jointIssuesFailed.value = false
  try {
    jointIssues.value = (await getJointIssues()).data.issues
    jointIssueYear.value = jointIssues.value[0]?.year ?? null
  } catch {
    jointIssuesFailed.value = true
  }
}

const pathTitle = computed(() => path.value ? t(`sets.wizard.paths.${path.value}.title`) : '')

// A name for the set, until one is typed.
const suggestedName = computed(() => {
  if (path.value === 'filters') return filtersName(filters.value)
  if (path.value !== 'challenge') return ''
  switch (challenge.value) {
    case 'annualSeries':
      return seriesCountry.value ? t('sets.challenges.annualSeries.named', { country: releaseCountryName(seriesCountry.value), year: seriesYear.value || '' }) : t('sets.challenges.annualSeries.title')
    case 'birthYear':
      return birthYear.value ? t('sets.challenges.birthYear.named', { year: birthYear.value }) : t('sets.challenges.birthYear.title')
    case 'jointIssue':
      return jointIssue.value ? t('sets.challenges.jointIssue.named', { year: jointIssue.value.year, title: jointIssue.value.title }) : t('sets.challenges.jointIssue.title')
    default:
      return t(`sets.challenges.${challenge.value}.title`)
  }
})
watch(suggestedName, (value) => {
  if (!nameEdited.value) name.value = value
}, { immediate: true })

function validYear(value: number | null | undefined, min: number, max: number): value is number {
  return typeof value === 'number' && Number.isInteger(value) && value >= min && value <= max
}

const valid = computed(() => {
  if (!name.value.trim()) return false
  if (path.value === 'agentic') return Boolean(agenticPrompt.value.trim())
  if (path.value !== 'challenge') return true
  if (challenge.value === 'annualSeries') return Boolean(seriesCountry.value) && validYear(seriesYear.value, EURO_FIRST_YEAR, thisYear + 1)
  if (challenge.value === 'birthYear') return validYear(birthYear.value, 1, thisYear)
  if (challenge.value === 'jointIssue') return jointIssue.value !== null
  return true
})

// What the server makes of the choice: a checklist, a smart set of
// filters, or a set of the coins picked.
function request(): Omit<CreateCoinSetRequest, 'name' | 'description' | 'color'> {
  switch (path.value) {
    case 'challenge': {
      const checklist = CHECKLIST_CHALLENGES[challenge.value]
      if (!checklist) {
        const year = birthYear.value ?? thisYear
        return { setType: 'smart', smartCriteria: filtersToCriteria({ ...emptyFilters(), yearFrom: year, yearTo: year }) as unknown as Record<string, unknown> }
      }
      const definition: ChecklistDefinition = { source: 'challenge', challenge: checklist }
      if (checklist === 'annual-series') Object.assign(definition, { country: seriesCountry.value, year: seriesYear.value })
      if (checklist === 'joint-issue') definition.year = jointIssueYear.value ?? undefined
      return { setType: 'checklist', checklist: definition }
    }
    case 'filters':
      return { setType: 'smart', smartCriteria: filtersToCriteria(filters.value) as unknown as Record<string, unknown> }
    case 'goal':
      return { setType: 'goal' }
    case 'agentic':
      return { setType: 'agentic', agenticPrompt: agenticPrompt.value.trim() || undefined }
    default:
      return { setType: 'standard' }
  }
}

function submit() {
  if (!valid.value) return
  emit('submit', { name: name.value.trim(), description: description.value.trim(), color: color.value, ...request() }, undefined)
}
</script>
