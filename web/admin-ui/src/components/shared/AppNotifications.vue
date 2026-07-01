<script setup lang="ts">
import { useNotifications } from '@/stores/notifications'
import { X, CheckCircle, AlertCircle, Info } from '@lucide/vue'

const { notifications, dismiss } = useNotifications()

const icons: Record<string, typeof CheckCircle> = {
  success: CheckCircle,
  error: AlertCircle,
  info: Info,
}
</script>

<template>
  <div class="fixed bottom-4 right-4 z-[9999] flex flex-col-reverse gap-2 max-w-sm w-full pointer-events-none">
    <div
      v-for="n in notifications"
      :key="n.id"
      :class="[
        'pointer-events-auto flex items-start gap-3 rounded-lg border p-4 shadow-lg text-sm animate-in slide-in-from-right-full fade-in duration-300',
        n.kind === 'error'
          ? 'border-destructive/50 bg-destructive/10 text-destructive'
          : n.kind === 'success'
            ? 'border-green-500/50 bg-green-500/10 text-green-700 dark:text-green-400'
            : 'border-border bg-popover text-popover-foreground',
      ]"
    >
      <component :is="icons[n.kind]" class="size-4 shrink-0 mt-0.5" />
      <p class="flex-1">{{ n.message }}</p>
      <button class="shrink-0 opacity-50 hover:opacity-100" @click="dismiss(n.id)">
        <X class="size-4" />
      </button>
    </div>
  </div>
</template>
