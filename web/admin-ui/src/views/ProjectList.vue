<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useNotifications } from '@/stores/notifications'
import { getErrorMessage } from '@/lib/api-error'
const { notify } = useNotifications()

import { listProjects } from '@/api'
import type { Project } from '@/api'

import PageHeader from '@/components/shared/PageHeader.vue'
import EmptyState from '@/components/shared/EmptyState.vue'
import CreateProjectDialog from '@/components/projects/CreateProjectDialog.vue'
import ProjectCard from '@/components/projects/ProjectCard.vue'
import { Button } from '@/components/ui/button'
import { FolderKanban, Plus, RotateCw } from '@lucide/vue'

const projects = ref<Project[]>([])
const loading = ref(true)
const error = ref(false)
const createOpen = ref(false)

async function load() {
  loading.value = true
  error.value = false
  try {
    const result = await listProjects()
    projects.value = result.data ?? []
  } catch (e) {
    error.value = true
    notify(getErrorMessage(e, 'Failed to load projects'), 'error')
  } finally {
    loading.value = false
  }
}

function onCreated() {
  createOpen.value = false
  load()
}

onMounted(load)
</script>

<template>
  <PageHeader title="Projects" description="Organize tokens by application or environment">
    <template #actions>
      <Button variant="outline" size="icon" :disabled="loading" @click="load()">
        <RotateCw :class="['size-4', loading && 'animate-spin']" />
      </Button>
      <Button @click="createOpen = true">
        <Plus class="size-4" />
        Create Project
      </Button>
    </template>
  </PageHeader>

  <!-- Error state -->
  <div v-if="error && !loading" class="text-center py-12">
    <p class="text-sm text-muted-foreground mb-4">Failed to load projects.</p>
    <Button variant="outline" @click="load()">Retry</Button>
  </div>

  <!-- Empty state -->
  <EmptyState
    v-else-if="!loading && projects.length === 0"
    :icon="FolderKanban"
    title="No projects yet"
    description="Create your first project to organize tokens."
  >
    <template #action>
      <Button @click="createOpen = true">
        <Plus class="size-4" />
        Create Project
      </Button>
    </template>
  </EmptyState>

  <!-- Loading skeleton -->
  <div v-else-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
    <div v-for="i in 6" :key="i" class="rounded-lg border bg-card p-5 space-y-3">
      <div class="h-5 w-1/2 bg-muted rounded animate-pulse" />
      <div class="h-4 w-3/4 bg-muted rounded animate-pulse" />
      <div class="h-3 w-1/3 bg-muted rounded animate-pulse" />
    </div>
  </div>

  <!-- Project grid -->
  <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
    <ProjectCard v-for="project in projects" :key="project.id" :project="project" />
  </div>

  <CreateProjectDialog v-model:open="createOpen" @created="onCreated" />
</template>
