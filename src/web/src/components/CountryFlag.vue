<template>
  <img
    v-if="image"
    :src="image"
    alt=""
    aria-hidden="true"
    class="inline-block h-[0.75em] w-auto rounded-[1px] align-[-0.02em] shadow-[0_0_0_1px_var(--border-subtle)]"
    data-testid="country-flag-image"
  />
  <span v-else-if="emoji" aria-hidden="true">{{ emoji }}</span>
</template>

<script setup lang="ts">
// A country's flag beside its name: the emoji of a living country (or of the
// one whose flag it flew), else the drawing of one that no longer exists.
// Decorative: the name says the country.
import { computed } from 'vue'
import { countryFlag } from '@/utils/countryFlags'
import { historicFlagImage } from '@/utils/historicFlags'

const props = defineProps<{ country: string | null | undefined }>()

const emoji = computed(() => countryFlag(props.country))
const image = computed(() => (emoji.value ? '' : historicFlagImage(props.country)))
</script>
