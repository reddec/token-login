<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useNotifications } from '@/stores/notifications'
import { getErrorMessage } from '@/lib/api-error'
const { notify } = useNotifications()

import { createProject } from '@/api'
import type { Project, ProjectConfig } from '@/api'
import {
  uniqueNamesGenerator,
  adjectives,
  colors,
  animals,
} from 'unique-names-generator'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [project: Project]
}>()

const config = reactive<ProjectConfig>({
  slug: '',
  description: '',
})

const loading = ref(false)
const slugError = ref('')

async function handleCreate() {
  if (!config.slug.trim()) {
    slugError.value = 'Slug is required'
    return
  }
  slugError.value = ''
  loading.value = true
  try {
    const result = await createProject({ body: { ...config } })
    const project = result.data
    if (project) {
      notify('Project created', 'success')
      emit('created', project)
      emit('update:open', false)
      config.slug = ''
      config.description = ''
    }
  } catch (e) {
    slugError.value = getErrorMessage(e, 'Failed to create project')
  } finally {
    loading.value = false
  }
}

function handleOpenChange(val: boolean) {
  if (!val) {
    config.slug = ''
    config.description = ''
    slugError.value = ''
  } else {
    config.slug = uniqueNamesGenerator({
      dictionaries: [adjectives, colors, animals],
      separator: '-',
      length: 2,
    })
  }
  emit('update:open', val)
}
</script>

<template>
  <Dialog :open="props.open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Create Project</DialogTitle>
        <DialogDescription>Projects help organize your tokens by application or environment.</DialogDescription>
      </DialogHeader>
      <div class="space-y-4">
        <div class="space-y-2">
          <Label for="project-slug">Slug</Label>
          <Input
            id="project-slug"
            v-model="config.slug"
            placeholder="my-project"
            :disabled="loading"
          />
          <p v-if="slugError" class="text-xs text-destructive">{{ slugError }}</p>
          <p v-else class="text-xs text-muted-foreground">
            URL-friendly identifier (letters, numbers, hyphens, underscores).
          </p>
        </div>
        <div class="space-y-2">
          <Label for="project-desc">Description</Label>
          <Textarea
            id="project-desc"
            v-model="config.description"
            placeholder="Optional description..."
            :disabled="loading"
            class="resize-none"
          />
        </div>
      </div>
      <div class="flex justify-end gap-2 mt-4">
        <Button variant="outline" :disabled="loading" @click="handleOpenChange(false)">
          Cancel
        </Button>
        <Button :disabled="loading" @click="handleCreate()">
          {{ loading ? 'Creating...' : 'Create' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
