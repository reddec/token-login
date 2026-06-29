<script setup lang="ts">
import type { Component } from 'vue'
import { Button } from '@/components/ui/button'

defineProps<{
  icon?: Component
  title: string
  description?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <component v-if="icon" :is="icon" class="size-12 text-muted-foreground/50 mb-4" />
    <h2 class="text-lg font-medium">{{ title }}</h2>
    <p v-if="description" class="text-sm text-muted-foreground mt-1 max-w-sm">
      {{ description }}
    </p>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>
