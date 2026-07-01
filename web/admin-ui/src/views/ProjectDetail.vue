<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotifications } from '@/stores/notifications'
import { getErrorMessage } from '@/lib/api-error'
const { notify } = useNotifications()
import {
  uniqueNamesGenerator,
  adjectives,
  colors,
  animals,
} from 'unique-names-generator'

import { getProject, updateProject, deleteProject, listTokens, createToken } from '@/api'
import type { Project, ProjectConfig, ProjectPatch, Token, TokenConfig } from '@/api'
import { useCredentialStore } from '@/stores/credential'

import PageHeader from '@/components/shared/PageHeader.vue'
import ConfirmDialog from '@/components/shared/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import TokenTable from '@/components/tokens/TokenTable.vue'
import TokenEditForm from '@/components/tokens/TokenEditForm.vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ArrowLeft, Plus, Save, Trash2 } from '@lucide/vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const credentialStore = useCredentialStore()

const project = ref<Project | null>(null)
const loading = ref(true)
const notFound = ref(false)

const tokens = ref<Token[]>([])
const tokensLoading = ref(true)
const saving = ref(false)
const deleting = ref(false)
const showDeleteDialog = ref(false)

const createTokenOpen = ref(false)
const creatingToken = ref(false)
const newTokenConfig = reactive<TokenConfig>({
  label: '',
  hosts: [],
  paths: [],
  headers: [],
  projectId: 0,
})

const config = reactive<ProjectPatch>({
  description: '',
})

function initFromProject(p: Project) {
  config.description = p.description
}

async function load(showLoading = true) {
  if (showLoading) {
    loading.value = true
    notFound.value = false
  }
  try {
    const projectId = Number(props.id)
    const [result, tokensResult] = await Promise.all([
      getProject({ path: { project: projectId } }),
      listTokens({ query: { project: projectId } }),
    ])
    const p = result.data
    if (p) {
      project.value = p
      initFromProject(p)
    } else if (showLoading) {
      notFound.value = true
    }
    tokens.value = tokensResult.data ?? []
  } catch (e: any) {
    if (e?.response?.status === 404 && showLoading) {
      notFound.value = true
    } else if (showLoading) {
      notify(getErrorMessage(e, 'Failed to load project'), 'error')
    }
  } finally {
    loading.value = false
    tokensLoading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await updateProject({
      path: { project: Number(props.id) },
      body: { description: config.description },
    })
    notify('Project updated', 'success')
    await load(false)
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to update project'), 'error')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  deleting.value = true
  try {
    await deleteProject({ path: { project: Number(props.id) } })
    notify('Project deleted', 'success')
    router.push({ name: 'projects' })
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to delete project'), 'error')
    deleting.value = false
    showDeleteDialog.value = false
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

function openCreateTokenDialog() {
  newTokenConfig.label = uniqueNamesGenerator({
    dictionaries: [adjectives, colors, animals],
    separator: '-',
    length: 2,
  })
  newTokenConfig.hosts = []
  newTokenConfig.paths = []
  newTokenConfig.headers = []
  newTokenConfig.projectId = Number(props.id)
  createTokenOpen.value = true
}

async function handleCreateToken() {
  creatingToken.value = true
  try {
    const result = await createToken({ body: { ...newTokenConfig } })
    const cred = result.data
    if (cred) {
      credentialStore.set(cred)
      notify('Token created', 'success')
      createTokenOpen.value = false
      router.push({ name: 'token-detail', params: { id: String(cred.id) } })
    }
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to create token'), 'error')
  } finally {
    creatingToken.value = false
  }
}

const breadcrumbLabel = computed(() => {
  return project.value?.slug ?? 'Project'
})

onMounted(load)
</script>

<template>
  <!-- Loading -->
  <template v-if="loading">
    <div class="space-y-4">
      <Skeleton class="h-6 w-36" />
      <Skeleton class="h-24 w-full" />
      <Skeleton class="h-40 w-full" />
    </div>
  </template>

  <!-- Not found -->
  <div v-else-if="notFound" class="text-center py-16">
    <h2 class="text-xl font-semibold">Project not found</h2>
    <p class="text-sm text-muted-foreground mt-2">This project may have been deleted.</p>
    <Button variant="outline" class="mt-4" @click="router.push({ name: 'projects' })">
      Back to projects
    </Button>
  </div>

  <!-- Loaded -->
  <template v-else-if="project">
    <Breadcrumb class="mb-6">
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink href="#/projects">Projects</BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbPage>{{ breadcrumbLabel }}</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>

    <PageHeader :title="project.slug || '(default)'" :description="project.description || undefined">
      <template #actions>
        <Button variant="outline" size="sm" @click="router.push({ name: 'projects' })">
          <ArrowLeft class="size-4" />
          Back
        </Button>
      </template>
    </PageHeader>

    <!-- Metadata -->
    <div class="grid grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">ID</p>
        <p class="text-sm font-mono">{{ project.id }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Created</p>
        <p class="text-sm">{{ formatDate(project.createdAt) }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Updated</p>
        <p class="text-sm">{{ formatDate(project.updatedAt) }}</p>
      </div>
    </div>

    <!-- Edit form -->
    <div class="space-y-4 mb-4">
      <div class="space-y-2">
        <Label for="edit-slug">Slug</Label>
        <Input id="edit-slug" :model-value="project.slug || '(default)'" disabled />
        <p class="text-xs text-muted-foreground">Slug cannot be changed after creation.</p>
      </div>
      <div class="space-y-2">
        <Label for="edit-desc">Description</Label>
        <Textarea
          id="edit-desc"
          v-model="config.description"
          placeholder="Optional description..."
          :disabled="saving"
          class="resize-none"
        />
      </div>
    </div>

    <!-- Actions -->
    <div class="flex flex-wrap gap-3">
      <Button :disabled="saving" @click="handleSave()">
        <Save class="size-4" />
        {{ saving ? 'Saving...' : 'Save Changes' }}
      </Button>
      <Button
        variant="destructive"
        :disabled="deleting"
        @click="showDeleteDialog = true"
      >
        <Trash2 class="size-4" />
        Delete Project
      </Button>
    </div>

    <Separator class="mt-6 mb-4" />

    <!-- Tokens in this project -->
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">
        Tokens
        <span v-if="!tokensLoading" class="text-sm text-muted-foreground font-normal ml-1">
          ({{ tokens.length }})
        </span>
      </h3>
      <Button size="sm" @click="openCreateTokenDialog()">
        <Plus class="size-4" />
        Create Token
      </Button>
    </div>
    <TokenTable :tokens="tokens" :loading="tokensLoading" />
    <div v-if="!tokensLoading && tokens.length === 0" class="text-center text-sm text-muted-foreground py-4">
      No tokens in this project yet.
    </div>
  </template>

  <ConfirmDialog
    v-model:open="showDeleteDialog"
    title="Delete Project"
    :description="`Are you sure you want to delete this project? Tokens in this project will be unlinked (not deleted). This action cannot be undone.`"
    confirm-text="Delete"
    variant="destructive"
    :loading="deleting"
    @confirm="handleDelete()"
  />

  <!-- Create Token Dialog -->
  <Dialog v-model:open="createTokenOpen">
    <DialogContent class="sm:max-w-lg max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Create Token</DialogTitle>
        <DialogDescription>
          Create a new token for the <strong>{{ project?.slug || '(default)' }}</strong> project.
        </DialogDescription>
      </DialogHeader>
      <TokenEditForm
        :config="(newTokenConfig as any)"
        :projects="[]"
        :show-project="false"
      />
      <div class="flex justify-end gap-2 mt-4">
        <Button variant="outline" :disabled="creatingToken" @click="createTokenOpen = false">
          Cancel
        </Button>
        <Button :disabled="creatingToken" @click="handleCreateToken()">
          {{ creatingToken ? 'Creating...' : 'Create' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
