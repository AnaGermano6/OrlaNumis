<template>
  <!--
    The guide is long-form prose with inline markup, so each locale keeps a full
    copy of the content instead of hundreds of catalog keys. Keep the two files
    in step: HelpSection.test.ts checks they have the same structure.
  -->
  <HelpContentPt v-if="locale === 'pt-PT'" />
  <HelpContentEn v-else />
</template>

<script setup lang="ts">
import { locale } from '@/i18n'
import HelpContentEn from '@/components/help/HelpContentEn.vue'
import HelpContentPt from '@/components/help/HelpContentPt.vue'
</script>
