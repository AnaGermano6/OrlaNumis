<template>
  <nav :aria-label="t('app.nav.label')" data-testid="desktop-nav">
    <ul class="m-0 flex list-none items-center gap-0.5 p-0">
      <li v-for="item in items" :key="item.id" :data-id="item.id">
        <!-- With pages under it: a list that opens as the pointer passes. -->
        <NavDropdown
          v-if="item.children?.length"
          :label="item.label"
          :button-class="[ITEM, isActive(item) ? ACTIVE : '']"
          hover
        >
          <template #button>
            <component :is="item.icon" :size="18" aria-hidden="true" />
            <span class="hidden lg:inline">{{ item.label }}</span>
            <ChevronDown :size="14" class="shrink-0 text-text-muted" aria-hidden="true" />
          </template>
          <router-link
            v-for="child in item.children"
            :key="child.id"
            :to="child.to"
            class="block px-4 py-2 text-body text-text-secondary no-underline transition-colors hover:bg-gold-glow hover:text-gold"
            :class="child.id === activeChild(item)?.id ? 'text-gold' : ''"
            :data-testid="`desktop-nav-${child.id}`"
          >
            {{ child.label }}
          </router-link>
        </NavDropdown>
        <!-- One that does something rather than open a page: the agent. -->
        <button
          v-else-if="item.action"
          type="button"
          :class="ITEM"
          :aria-label="item.label"
          :title="item.label"
          @click="$emit('action', item)"
        >
          <component :is="item.icon" :size="18" aria-hidden="true" />
          <span class="hidden lg:inline">{{ item.label }}</span>
        </button>
        <router-link
          v-else-if="item.to"
          :to="item.to"
          :class="[ITEM, isActive(item) ? ACTIVE : '']"
          :aria-label="item.label"
          :title="item.label"
        >
          <component :is="item.icon" :size="18" aria-hidden="true" />
          <span class="hidden lg:inline">{{ item.label }}</span>
        </router-link>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
// The app's menu in the top bar, on a wider screen: each item in a row, in
// the order chosen for the side menu; those with pages under them open a
// list. Between 768 and 1023 pixels only their icons show.
import type { Component } from 'vue'
import { useRoute } from 'vue-router'
import { ChevronDown } from 'lucide-vue-next'
import NavDropdown from '@/components/nav/NavDropdown.vue'
import { t } from '@/i18n'

export interface DesktopNavItem {
  id: string
  label: string
  icon: Component
  to?: string
  action?: () => void
  children?: { id: string; label: string; to: string }[]
}

defineProps<{ items: DesktopNavItem[] }>()
defineEmits<{ action: [item: DesktopNavItem] }>()

const ITEM = 'flex items-center gap-1.5 rounded-sm border-0 bg-transparent px-2.5 py-2 text-body font-medium whitespace-nowrap text-text-secondary no-underline transition-colors hover:bg-gold-glow hover:text-gold cursor-pointer'
const ACTIVE = 'bg-gold-glow text-gold'

const route = useRoute()

// A page and the pages under it ("/stats" and "/stats/timeline"); the gallery's "/" alone.
function matches(to: string): boolean {
  return route.path === to || (to !== '/' && route.path.startsWith(`${to}/`))
}

// The page open in a list: the most specific link that leads to it ("/stats/timeline" rather than "/stats").
function activeChild(item: DesktopNavItem) {
  return item.children?.filter((child) => matches(child.to)).sort((a, b) => b.to.length - a.to.length)[0]
}

function isActive(item: DesktopNavItem): boolean {
  return Boolean((item.to && matches(item.to)) || item.children?.some((child) => matches(child.to)))
}
</script>
