<script setup lang="ts">
import { useRoute } from 'vue-router'
import { usePreferencesStore } from '@/stores/preferences'
import { Button } from '@/components/ui/button'
import { SunMoon, KeyRound, FolderKanban } from '@lucide/vue'

const route = useRoute()
const prefs = usePreferencesStore()

const navItems = [
  { to: '/projects', label: 'Projects', icon: FolderKanban },
  { to: '/tokens', label: 'Tokens', icon: KeyRound },
]

function isActive(path: string) {
  return route.path.startsWith(path)
}
</script>

<template>
  <nav class="flex h-full flex-col">
    <div class="px-4 py-5">
      <h1 class="text-lg font-semibold tracking-tight">Token Admin</h1>
      <p class="text-xs text-muted-foreground mt-0.5">API Key Management</p>
    </div>

    <div class="flex-1 space-y-1 px-3">
      <a
        v-for="item in navItems"
        :key="item.to"
        :href="'#' + item.to"
        :class="[
          'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
          isActive(item.to)
            ? 'bg-primary text-primary-foreground'
            : 'text-muted-foreground hover:bg-accent hover:text-foreground',
        ]"
      >
        <component :is="item.icon" class="size-4" />
        {{ item.label }}
      </a>
    </div>

    <div class="border-t p-3">
      <Button variant="ghost" size="sm" class="w-full justify-start gap-2" @click="prefs.toggleDark()">
        <SunMoon class="size-4" />
        <span>{{ prefs.prefs.darkMode ? 'Light' : 'Dark' }} mode</span>
      </Button>
    </div>
  </nav>
</template>
