<script setup lang="ts">
import { useMediaQuery } from '@/composables/useMediaQuery'
import { usePreferencesStore } from '@/stores/preferences'
import AppSidebar from './AppSidebar.vue'
import { Sheet, SheetContent } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Menu } from '@lucide/vue'

const isDesktop = useMediaQuery('(min-width: 1024px)')
const prefs = usePreferencesStore()
</script>

<template>
  <!-- Desktop layout -->
  <div v-if="isDesktop" class="flex min-h-screen">
    <aside class="fixed inset-y-0 z-30 flex w-56 flex-col border-r bg-background">
      <AppSidebar />
    </aside>
    <main class="ml-56 flex-1">
      <div class="mx-auto max-w-6xl p-6">
        <router-view />
      </div>
    </main>
  </div>

  <!-- Mobile layout -->
  <div v-else class="min-h-screen">
    <header class="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4">
      <Sheet :open="prefs.prefs.sidebarOpen" @update:open="prefs.toggleSidebar()">
        <Button variant="ghost" size="icon" @click="prefs.toggleSidebar()">
          <Menu class="size-5" />
        </Button>
        <SheetContent side="left" class="w-56 p-0">
          <AppSidebar />
        </SheetContent>
      </Sheet>
      <span class="font-semibold">Token Admin</span>
    </header>
    <main class="p-4">
      <router-view />
    </main>
  </div>

  <Toaster rich-colors close-button />
</template>
