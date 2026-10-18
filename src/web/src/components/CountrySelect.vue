<template>
  <div ref="root" class="relative min-w-0" @keydown="onKeydown">
    <button
      ref="button"
      type="button"
      role="combobox"
      class="form-select flex h-10 w-full items-center gap-2 bg-card py-0 text-left text-body"
      aria-haspopup="listbox"
      :aria-label="label"
      :aria-expanded="open"
      :aria-controls="listId"
      :aria-activedescendant="open && active >= 0 ? optionId(active) : undefined"
      v-bind="$attrs"
      @click="toggle"
    >
      <CountryFlag v-if="selected?.country" :country="selected.country" class="shrink-0" />
      <span class="min-w-0 flex-1 truncate">{{ selected?.label ?? '' }}</span>
    </button>
    <Teleport to="body">
      <ul
        v-if="open"
        :id="listId"
        ref="list"
        role="listbox"
        :aria-label="label"
        class="fixed z-[500] m-0 max-h-72 list-none overflow-y-auto rounded-sm border border-border-subtle bg-card p-1 shadow-[0_8px_30px_var(--overlay-40)]"
        :style="position"
        data-testid="country-select-list"
      >
        <li
          v-for="(option, index) in options"
          :id="optionId(index)"
          :key="option.value"
          role="option"
          :aria-selected="option.value === modelValue"
          class="flex min-w-0 cursor-pointer items-center gap-2 rounded-[6px] px-2.5 py-1.5 text-body"
          :class="index === active ? 'bg-gold-glow text-gold' : 'text-text-primary'"
          data-testid="country-select-option"
          @mousedown.prevent
          @mousemove="active = index"
          @click="choose(index)"
        >
          <CountryFlag v-if="option.country" :country="option.country" class="shrink-0" />
          <span class="min-w-0 flex-1 truncate">{{ option.label }}</span>
          <Check v-if="option.value === modelValue" :size="14" class="shrink-0 text-gold" aria-hidden="true" />
        </li>
      </ul>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
// A country menu whose options show each country's flag: the emoji of a
// living country, the drawing of one that no longer exists (a browser's own
// menu shows text only). It works as a select: it opens on a click or the
// arrow keys, the arrows, Home and End move, a typed letter jumps, Enter or
// Space chooses, Escape and a click elsewhere close.
import { computed, nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'
import { Check } from 'lucide-vue-next'
import CountryFlag from '@/components/CountryFlag.vue'
import { foldCountryName } from '@/utils/worldCountries'

defineOptions({ inheritAttrs: false })

export interface CountrySelectOption {
  value: string
  label: string
  /** The country whose flag goes beside the label; none for "All countries". */
  country?: string
}

const props = defineProps<{
  modelValue: string
  options: CountrySelectOption[]
  /** The field's name, for screen readers. */
  label: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const listId = `country-select-${useId()}`
const root = ref<HTMLElement | null>(null)
const button = ref<HTMLElement | null>(null)
const list = ref<HTMLElement | null>(null)
const open = ref(false)
const active = ref(-1)
const position = ref<Record<string, string>>({})

const selected = computed(() => props.options.find((option) => option.value === props.modelValue) ?? props.options[0])

function optionId(index: number): string {
  return `${listId}-${index}`
}

// Below the button, or above it when there is no room below.
function place() {
  const rect = button.value?.getBoundingClientRect()
  if (!rect) return
  const below = window.innerHeight - rect.bottom
  const width = `${Math.max(rect.width, 224)}px`
  position.value = below >= 240 || below >= rect.top
    ? { top: `${rect.bottom + 4}px`, left: `${rect.left}px`, width }
    : { bottom: `${window.innerHeight - rect.top + 4}px`, left: `${rect.left}px`, width }
}

function show() {
  if (open.value) return
  place()
  open.value = true
  active.value = Math.max(0, props.options.findIndex((option) => option.value === props.modelValue))
  void nextTick(reveal)
}

function hide() {
  open.value = false
  active.value = -1
}

function toggle() {
  if (open.value) hide()
  else show()
}

function choose(index: number) {
  const option = props.options[index]
  if (!option) return
  if (option.value !== props.modelValue) emit('update:modelValue', option.value)
  hide()
  button.value?.focus()
}

function reveal() {
  list.value?.children[active.value]?.scrollIntoView?.({ block: 'nearest' })
}

function move(to: number) {
  active.value = Math.min(props.options.length - 1, Math.max(0, to))
  void nextTick(reveal)
}

// Letters typed within half a second make one word to jump to.
let typed = ''
let typedTimer: ReturnType<typeof setTimeout> | undefined
function jumpTo(letter: string) {
  typed += letter
  clearTimeout(typedTimer)
  typedTimer = setTimeout(() => { typed = '' }, 500)
  const wanted = foldCountryName(typed)
  const index = props.options.findIndex((option) => foldCountryName(option.label).startsWith(wanted))
  if (index >= 0) {
    if (!open.value) show()
    move(index)
  }
}

function onKeydown(event: KeyboardEvent) {
  const key = event.key
  if (!open.value) {
    if (key === 'ArrowDown' || key === 'ArrowUp' || key === 'Enter' || key === ' ') {
      event.preventDefault()
      show()
    } else if (key.length === 1 && /\p{L}/u.test(key)) {
      jumpTo(key)
    }
    return
  }
  switch (key) {
    case 'ArrowDown': event.preventDefault(); move(active.value + 1); break
    case 'ArrowUp': event.preventDefault(); move(active.value - 1); break
    case 'Home': event.preventDefault(); move(0); break
    case 'End': event.preventDefault(); move(props.options.length - 1); break
    case 'Enter':
    case ' ': event.preventDefault(); choose(active.value); break
    case 'Escape': event.preventDefault(); hide(); break
    case 'Tab': hide(); break
    default:
      if (key.length === 1 && /\p{L}/u.test(key)) jumpTo(key)
  }
}

// A click outside the button and its list closes it; scrolling or resizing moves it along.
function onPointerDown(event: Event) {
  const target = event.target as Node
  if (!root.value?.contains(target) && !list.value?.contains(target)) hide()
}

watch(open, (isOpen) => {
  if (isOpen) {
    document.addEventListener('mousedown', onPointerDown)
    window.addEventListener('scroll', place, true)
    window.addEventListener('resize', place)
  } else {
    document.removeEventListener('mousedown', onPointerDown)
    window.removeEventListener('scroll', place, true)
    window.removeEventListener('resize', place)
  }
})

onBeforeUnmount(() => {
  clearTimeout(typedTimer)
  hide()
  document.removeEventListener('mousedown', onPointerDown)
  window.removeEventListener('scroll', place, true)
  window.removeEventListener('resize', place)
})
</script>
