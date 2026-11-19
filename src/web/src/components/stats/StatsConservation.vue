<template>
  <section v-if="graded > 0" class="card" :aria-labelledby="`${id}-conservation-title`" data-testid="stats-conservation">
    <h2 :id="`${id}-conservation-title`" class="mb-5 text-lg">{{ t('stats.conservation.title') }}</h2>
    <StackedBar :parts="parts" :total="graded" :legend="false" separated />

    <!-- Every grade the collection has, best first: its pieces, their share and what it means. -->
    <ul class="m-0 mt-5 grid list-none grid-cols-2 gap-3 p-0 sm:grid-cols-3 lg:grid-cols-4 lg:gap-4">
      <li
        v-for="part in parts"
        :key="part.key"
        class="flex min-w-0 flex-col gap-1 rounded-md border border-border-subtle bg-input p-4"
        :data-testid="`conservation-${part.key}`"
      >
        <span class="flex items-center gap-2 text-md font-semibold text-text-primary">
          <span class="h-3 w-3 shrink-0 rounded-full" :style="{ background: part.color }" aria-hidden="true" />
          {{ part.label }}
        </span>
        <span class="text-body text-text-primary">{{ t('stats.conservation.pieces', { count: part.count }) }} · {{ formatPercent(part.count / graded) }}</span>
        <span v-if="part.hint" class="text-body text-text-secondary">{{ part.hint }}</span>
      </li>
    </ul>

    <!-- Those without a grade, and a way to the gallery showing only them, to grade them. -->
    <p v-if="ungraded > 0" class="m-0 mt-5 border-t border-border-subtle pt-4 text-body text-text-secondary" data-testid="conservation-ungraded">
      {{ t('stats.conservation.ungraded', { count: ungraded }) }} —
      <button
        type="button"
        class="cursor-pointer rounded-sm text-gold underline underline-offset-4 hover:no-underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        data-testid="conservation-grade-now"
        @click="emit('grade-now')"
      >
        {{ t('stats.conservation.gradeNow') }}
      </button>
    </p>
  </section>
</template>

<script setup lang="ts">
// The pieces (of the whole collection, its coins or its banknotes) by grade,
// best first: a bar of every grade they have, then each grade with its
// pieces, their share of those that have a grade and what the grade means;
// then how many have none, and a button asking for the gallery showing only
// them, to grade them (which gallery is the parent's to say).
import { computed } from 'vue'
import StackedBar, { type StackedBarPart } from '@/components/collection/StackedBar.vue'
import { t } from '@/i18n'
import { formatShare } from '@/utils/numberFormat'
import { COIN_FINISHES, COIN_GRADES } from '@/utils/coinAttributes'

const props = defineProps<{
  /** Unique on the page: its title's id starts with it. */
  id: string
  byGrade: { grade: string; count: number }[] | undefined
  /** Every piece, graded or not. */
  total: number
}>()

const emit = defineEmits<{ 'grade-now': [] }>()

/** Best first: the finishes, then the circulation grades from uncirculated down. */
const ORDER = [...[...COIN_FINISHES].reverse(), ...COIN_GRADES]
type KnownGrade = (typeof ORDER)[number]
const COLORS = [
  'var(--accent-gold)', 'var(--cat-modern)', 'var(--accent-bronze)', 'var(--cat-extra-1)', 'var(--cat-extra-2)',
  'var(--cat-extra-3)', 'var(--cat-extra-4)', 'var(--cat-extra-5)', 'var(--cat-extra-6)',
]

const grades = computed(() => (props.byGrade ?? []).filter((entry) => entry.count > 0))
const graded = computed(() => grades.value.reduce((sum, entry) => sum + entry.count, 0))
const ungraded = computed(() => Math.max(0, props.total - graded.value))

function isKnown(grade: string): grade is KnownGrade {
  return (ORDER as readonly string[]).includes(grade)
}

function rank(grade: string): number {
  return isKnown(grade) ? ORDER.indexOf(grade) : ORDER.length
}

// "Sem circulação (Soberba, FDC)"; nothing for a grade the app does not know.
function gradeName(grade: string): string {
  return isKnown(grade) ? t(`coinForm.grades.${grade}`) : ''
}

const parts = computed<StackedBarPart[]>(() => [...grades.value]
  .sort((a, b) => rank(a.grade) - rank(b.grade) || b.count - a.count)
  .map((entry, index) => ({
    key: entry.grade,
    label: entry.grade,
    count: entry.count,
    color: COLORS[index % COLORS.length]!,
    hint: gradeName(entry.grade),
  })))

function formatPercent(value: number): string {
  return formatShare(value)
}
</script>
