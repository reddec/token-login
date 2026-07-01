<script setup lang="ts">
import type { Token } from '@/api'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ChevronRight } from '@lucide/vue'

defineProps<{
  token: Token
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
  <Card>
    <CardContent class="p-4">
      <div class="flex items-start justify-between gap-2">
        <div class="min-w-0 flex-1 space-y-1.5">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="font-medium truncate">{{ token.label || token.keyID }}</span>
            <Badge variant="secondary" class="font-mono text-xs shrink-0">
              {{ token.keyID }}
            </Badge>
          </div>
          <div class="flex items-center gap-2 text-xs text-muted-foreground flex-wrap">
            <a v-if="token.projectId" :href="'#/projects/' + token.projectId">
              <Badge variant="outline" class="text-xs hover:bg-accent transition-colors">
                {{ token.projectSlug || '(default)' }}
              </Badge>
            </a>
            <Badge v-else variant="outline" class="text-xs">no project</Badge>
            <span>{{ token.hosts?.join(', ') || 'Any host' }}</span>
            <span>{{ token.paths?.join(', ') || 'Any path' }}</span>
          </div>
          <div class="flex items-center gap-3 text-xs text-muted-foreground">
            <span>{{ token.requests.toLocaleString() }} requests</span>
            <span>Created {{ formatDate(token.createdAt) }}</span>
          </div>
        </div>
        <Button variant="ghost" size="icon" as-child class="shrink-0">
          <a :href="'#/tokens/' + token.id">
            <ChevronRight class="size-4" />
          </a>
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
