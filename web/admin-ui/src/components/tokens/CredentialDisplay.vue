<script setup lang="ts">
import { ref } from 'vue'
import type { Credential } from '@/api'
import { Button } from '@/components/ui/button'
import { Copy, Check, X } from '@lucide/vue'

const props = defineProps<{
  credential: Credential | null
}>()

const emit = defineEmits<{
  dismiss: []
}>()

const copied = ref(false)

async function copyKey() {
  if (!props.credential?.key) return
  try {
    await navigator.clipboard.writeText(props.credential.key)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 3000)
  } catch {
    // fallback: do nothing
  }
}
</script>

<template>
  <div v-if="props.credential" class="rounded-lg border border-destructive/50 bg-destructive/5 p-4">
    <div class="flex items-start gap-3">
      <div class="flex-1 space-y-2">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-destructive">Secret key — save it now!</p>
          <Button variant="ghost" size="icon" class="size-6 -mr-1" @click="emit('dismiss')">
            <X class="size-3.5" />
          </Button>
        </div>
        <p class="text-xs text-muted-foreground">
          This key is <strong>unrecoverable</strong> and shown only once. Copy it before you
          navigate away.
        </p>
        <div class="flex items-center gap-2">
          <pre
            class="flex-1 rounded bg-background px-3 py-2 text-sm font-mono break-all border"
          >{{ props.credential.key }}</pre>
          <Button
            size="icon"
            variant="outline"
            class="shrink-0"
            @click="copyKey()"
          >
            <Check v-if="copied" class="size-4 text-green-600" />
            <Copy v-else class="size-4" />
          </Button>
        </div>
        <p v-if="copied" class="text-xs text-green-600 font-medium">Copied!</p>
      </div>
    </div>
  </div>
</template>
