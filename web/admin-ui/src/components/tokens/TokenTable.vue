<script setup lang="ts">
import type { Token } from '@/api'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Eye } from '@lucide/vue'

defineProps<{
  tokens: Token[]
  loading?: boolean
}>()

const SKELETON_ROWS = 5
</script>

<template>
  <div class="rounded-md border">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Label</TableHead>
          <TableHead>Key ID</TableHead>
          <TableHead>Project</TableHead>
          <TableHead>Host</TableHead>
          <TableHead>Path</TableHead>
          <TableHead class="text-right">Requests</TableHead>
          <TableHead class="w-10"></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <template v-if="loading">
          <TableRow v-for="i in SKELETON_ROWS" :key="i">
            <TableCell v-for="j in 7" :key="j">
              <Skeleton class="h-4 w-full" />
            </TableCell>
          </TableRow>
        </template>
        <template v-else-if="tokens.length === 0">
          <TableRow>
            <TableCell colspan="7" class="text-center text-muted-foreground py-8">
              No tokens found.
            </TableCell>
          </TableRow>
        </template>
        <template v-else>
          <TableRow v-for="token in tokens" :key="token.id">
            <TableCell class="font-medium">
              <a
                :href="'#/tokens/' + token.id"
                class="text-primary hover:underline"
              >
                {{ token.label || token.keyID }}
              </a>
            </TableCell>
            <TableCell>
              <Badge variant="secondary" class="font-mono text-xs">
                {{ token.keyID }}
              </Badge>
            </TableCell>
            <TableCell>
              <a v-if="token.projectId" :href="'#/projects/' + token.projectId">
                <Badge variant="outline" class="text-xs hover:bg-accent transition-colors">
                  {{ token.projectSlug || '(default)' }}
                </Badge>
              </a>
              <Badge v-else variant="outline" class="text-xs">none</Badge>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ token.host || 'Any' }}
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ token.path || 'Any' }}
            </TableCell>
            <TableCell class="text-right text-sm text-muted-foreground">
              {{ token.requests.toLocaleString() }}
            </TableCell>
            <TableCell>
              <Button variant="ghost" size="icon" as-child>
                <a :href="'#/tokens/' + token.id">
                  <Eye class="size-4" />
                </a>
              </Button>
            </TableCell>
          </TableRow>
        </template>
      </TableBody>
    </Table>
  </div>
</template>
