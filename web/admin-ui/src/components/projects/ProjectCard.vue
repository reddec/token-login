<script setup lang="ts">
import type { Project } from '@/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

defineProps<{
  project: Project
}>()

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<template>
  <a :href="'#/projects/' + project.id" class="block transition-colors hover:no-underline">
    <Card class="hover:border-primary/50 transition-colors cursor-pointer h-full">
      <CardHeader>
        <CardTitle class="text-base">{{ project.slug || '(default)' }}</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-sm text-muted-foreground line-clamp-2">
          {{ project.description || 'No description' }}
        </p>
        <p class="text-xs text-muted-foreground mt-3">
          Created {{ formatDate(project.createdAt) }}
        </p>
      </CardContent>
    </Card>
  </a>
</template>
