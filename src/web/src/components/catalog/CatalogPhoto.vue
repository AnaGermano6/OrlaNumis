<template>
  <figure v-if="url && !failed" class="m-0 flex min-w-0 flex-col gap-1" :class="{ 'w-full': fluid }" :style="fluid ? undefined : { width: size }">
    <div
      class="flex items-center justify-center overflow-hidden rounded-sm border border-border-subtle bg-card"
      :class="{ 'aspect-square w-full': fluid }"
      :style="fluid ? undefined : { width: size, height: size }"
    >
      <img v-if="src" :src="src" :alt="alt" class="h-full w-full object-contain" @error="onError" />
    </div>
    <figcaption v-if="credit" class="text-2xs leading-tight text-text-muted [overflow-wrap:anywhere]">{{ credit }}</figcaption>
  </figure>
</template>

<script setup lang="ts">
// A reference photo of a coin. With via="direct" (Numista, whose host the CSP
// allows) the browser loads it straight away; Numista's bot protection can
// still refuse it, and then the image proxy is tried. With via="proxy" (other
// public sources) it goes through the proxy from the start. If nothing works
// the photo is left out rather than shown as a broken box. Photos belong to
// their authors: pass the credit whenever one is known.
import { ref, watch } from 'vue'
import { loadCatalogPhoto } from '@/utils/catalogPhotoCache'

const props = withDefaults(defineProps<{
  url?: string
  alt: string
  credit?: string
  size?: string
  /** Fill the parent's width with a square frame instead of a fixed size. */
  fluid?: boolean
  via?: 'direct' | 'proxy'
}>(), { url: undefined, credit: '', size: '4rem', fluid: false, via: 'direct' })

const emit = defineEmits<{
  /** Neither the source nor the proxy could serve the photo; it is no longer shown. */
  unavailable: []
}>()

const src = ref<string | null>(null)
const failed = ref(false)
let triedProxy = false

watch(() => [props.url, props.via] as const, ([url, via]) => {
  failed.value = false
  triedProxy = false
  src.value = via === 'proxy' ? null : url ?? null
  if (url && via === 'proxy') void onError()
}, { immediate: true })

async function onError() {
  const url = props.url
  if (!url || triedProxy) {
    failed.value = true
    emit('unavailable')
    return
  }
  triedProxy = true
  const proxied = await loadCatalogPhoto(url)
  if (props.url !== url) return
  if (proxied) {
    src.value = proxied
  } else {
    failed.value = true
    emit('unavailable')
  }
}
</script>
