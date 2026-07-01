<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useNotifications } from '@/stores/notifications'
import { getErrorMessage, getResponseStatus } from '@/lib/api-error'
const { notify } = useNotifications()
import {
  uniqueNamesGenerator,
  adjectives,
  colors,
  animals,
} from 'unique-names-generator'

import { getProject, updateProject, deleteProject, listTokens, createToken } from '@/api'
import type { Project, Token, TokenConfig } from '@/api'
import { useCredentialStore } from '@/stores/credential'
import { useTokenSearch } from '@/composables/useTokenSearch'

import PageHeader from '@/components/shared/PageHeader.vue'
import MarkdownView from '@/components/shared/MarkdownView.vue'
import ConfirmDialog from '@/components/shared/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Folder, Pencil, Plus, Save, Trash2 } from '@lucide/vue'
import SetupTab from '@/components/projects/SetupTab.vue'

const props = defineProps<{ id: string; tab?: string }>()
const router = useRouter()
const route = useRoute()
const credentialStore = useCredentialStore()

const project = ref<Project | null>(null)
const loading = ref(true)
const notFound = ref(false)

const tokens = ref<Token[]>([])
const tokensLoading = ref(true)

const { searchQuery, filtered } = useTokenSearch(tokens)
const saving = ref(false)
const deleting = ref(false)
const showDeleteDialog = ref(false)

const validTabs = ['overview', 'setup'] as const
type Tab = (typeof validTabs)[number]

function resolveTab(raw: unknown): Tab {
  if (typeof raw === 'string' && (validTabs as readonly string[]).includes(raw)) {
    return raw as Tab
  }
  return 'overview'
}

const activeTab = ref<Tab>(resolveTab(route.params.tab))

watch(() => route.params.tab, (raw) => {
  const t = resolveTab(raw)
  if (t !== activeTab.value) {
    activeTab.value = t
  }
})

watch(activeTab, (tab) => {
  if (route.params.tab !== tab) {
    router.replace({ name: 'project-detail', params: { id: props.id, tab } })
  }
})


const createTokenOpen = ref(false)
const creatingToken = ref(false)
const newTokenConfig = reactive<TokenConfig>({
  label: '',
  hosts: [],
  paths: [],
  headers: [],
  projectId: 0,
})

const editDescriptionOpen = ref(false)
const descriptionDraft = ref('')

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
    } else if (showLoading) {
      notFound.value = true
    }
    tokens.value = tokensResult.data ?? []
  } catch (e) {
    if (getResponseStatus(e) === 404 && showLoading) {
      notFound.value = true
    } else if (showLoading) {
      notify(getErrorMessage(e, 'Failed to load project'), 'error')
    }
  } finally {
    loading.value = false
    tokensLoading.value = false
  }
}

function openEditDescription() {
  descriptionDraft.value = project.value?.description ?? ''
  editDescriptionOpen.value = true
}

async function handleSaveDescription() {
  saving.value = true
  try {
    await updateProject({
      path: { project: Number(props.id) },
      body: { description: descriptionDraft.value },
    })
    notify('Description updated', 'success')
    editDescriptionOpen.value = false
    await load(false)
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to update description'), 'error')
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

    <Tabs v-model="activeTab" default-value="overview">
      <TabsList class="mb-6">
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="setup">Setup</TabsTrigger>
      </TabsList>

      <TabsContent value="overview">
        <PageHeader :title="project.slug || '(default)'">
          <template #icon>
            <div class="size-10 rounded-lg bg-muted flex items-center justify-center">
              <Folder class="size-5" />
            </div>
          </template>
          <template #actions>
            <Button variant="outline" size="sm" @click="openEditDescription()">
              <Pencil class="size-4" />
              Edit
            </Button>
            <Button variant="destructive" size="sm" :disabled="deleting" @click="showDeleteDialog = true">
              <Trash2 class="size-4" />
              Delete
            </Button>
          </template>
        </PageHeader>

        <!-- Description (rendered markdown) -->
        <div class="mb-6 min-h-[1.5rem]">
          <MarkdownView v-if="project.description" :source="project.description" />
          <p v-else class="text-sm text-muted-foreground italic">No description yet.</p>
        </div>

        <!-- Metadata -->
        <div class="grid grid-cols-2 gap-4 mb-8">
          <div class="space-y-1">
            <p class="text-xs text-muted-foreground">Created</p>
            <p class="text-sm">{{ formatDate(project.createdAt) }}</p>
          </div>
          <div class="space-y-1">
            <p class="text-xs text-muted-foreground">Updated</p>
            <p class="text-sm">{{ formatDate(project.updatedAt) }}</p>
          </div>
        </div>

        <!-- Tokens -->
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-medium">
            Tokens
            <span v-if="!tokensLoading" class="text-sm text-muted-foreground font-normal ml-1">
              ({{ tokens.length }})
            </span>
          </h3>
          <Button size="sm" @click="openCreateTokenDialog()">
            <Plus class="size-4" />
            Create
          </Button>
        </div>
        <div v-if="tokens.length" class="mb-4">
          <Input
            v-model="searchQuery"
            placeholder="Search tokens by label, key, host, or path..."
          />
        </div>
        <TokenTable :tokens="filtered" :loading="tokensLoading" :show-project="false" />
        <div v-if="!tokensLoading && tokens.length === 0" class="text-center text-sm text-muted-foreground py-4">
          No tokens in this project yet.
        </div>

      </TabsContent>

      <TabsContent value="setup">
        <SetupTab :project="project" />
      </TabsContent>
    </Tabs>
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

  <!-- Edit Description Dialog -->
  <Dialog v-model:open="editDescriptionOpen">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>Edit Description</DialogTitle>
        <DialogDescription>Markdown is supported.</DialogDescription>
      </DialogHeader>
      <Tabs default-value="write">
        <TabsList class="grid w-full grid-cols-2">
          <TabsTrigger value="write">Write</TabsTrigger>
          <TabsTrigger value="preview">Preview</TabsTrigger>
        </TabsList>
        <TabsContent value="write">
          <Textarea
            v-model="descriptionDraft"
            class="min-h-[200px] resize-none font-mono text-sm"
            placeholder="Describe this project…"
          />
        </TabsContent>
        <TabsContent value="preview">
          <div class="min-h-[200px] rounded-md border p-4">
            <MarkdownView v-if="descriptionDraft" :source="descriptionDraft" />
            <p v-else class="text-sm text-muted-foreground">Nothing to preview.</p>
          </div>
        </TabsContent>
      </Tabs>
      <DialogFooter>
        <Button variant="outline" :disabled="saving" @click="editDescriptionOpen = false">
          Cancel
        </Button>
        <Button :disabled="saving" @click="handleSaveDescription()">
          <Save class="size-4" />
          {{ saving ? 'Saving...' : 'Save' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

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
