<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotifications } from '@/stores/notifications'
import { getErrorMessage, getResponseStatus } from '@/lib/api-error'
const { notify } = useNotifications()

import { getToken, updateToken, deleteToken, refreshToken, listProjects } from '@/api'
import type { Token, Project, TokenPatch } from '@/api'
import { useCredentialStore } from '@/stores/credential'

import PageHeader from '@/components/shared/PageHeader.vue'
import ConfirmDialog from '@/components/shared/ConfirmDialog.vue'
import TokenEditForm from '@/components/tokens/TokenEditForm.vue'
import CredentialDisplay from '@/components/tokens/CredentialDisplay.vue'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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
import { ArrowLeft, Key, RefreshCw, Save, Trash2 } from '@lucide/vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const credentialStore = useCredentialStore()

const token = ref<Token | null>(null)
const projects = ref<Project[]>([])
const loading = ref(true)
const notFound = ref(false)
const savedCredential = ref(credentialStore.consume(Number(props.id)))

const saving = ref(false)
const regenerating = ref(false)
const deleting = ref(false)
const showDeleteDialog = ref(false)

const config = reactive<TokenPatch>({
  label: '',
  hosts: [],
  paths: [],
  headers: [],
})

function initFromToken(t: Token) {
  config.label = t.label
  config.hosts = t.hosts ?? []
  config.paths = t.paths ?? []
  config.headers = t.headers ? [...t.headers] : []
}

async function load() {
  loading.value = true
  notFound.value = false
  try {
    const tokenId = Number(props.id)
    const [tokenRes, projectsRes] = await Promise.all([
      getToken({ path: { token: tokenId } }),
      listProjects(),
    ])
    const t = tokenRes.data
    if (t) {
      token.value = t
      initFromToken(t)
    } else {
      notFound.value = true
    }
    projects.value = projectsRes.data ?? []
  } catch (e) {
    if (getResponseStatus(e) === 404) {
      notFound.value = true
    } else {
      notify(getErrorMessage(e, 'Failed to load token'), 'error')
    }
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await updateToken({
      path: { token: Number(props.id) },
      body: { ...config },
    })
    notify('Token updated', 'success')
    await load()
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to update token'), 'error')
  } finally {
    saving.value = false
  }
}

async function handleRegenerate() {
  regenerating.value = true
  try {
    const result = await refreshToken({ path: { token: Number(props.id) } })
    const cred = result.data
    if (cred) {
      credentialStore.set(cred)
      savedCredential.value = cred
      notify('Secret regenerated', 'success')
      await load()
    }
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to regenerate secret'), 'error')
  } finally {
    regenerating.value = false
  }
}

async function handleDelete() {
  deleting.value = true
  try {
    await deleteToken({ path: { token: Number(props.id) } })
    notify('Token deleted', 'success')
    router.push({ name: 'tokens' })
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to delete token'), 'error')
    deleting.value = false
    showDeleteDialog.value = false
  }
}

function formatDate(iso: string | undefined): string {
  if (!iso) return 'Never'
  return new Date(iso).toLocaleString()
}

const breadcrumbLabel = computed(() => {
  if (!token.value) return ''
  return `${token.value.label || 'Token #' + token.value.id} (${token.value.keyID})`
})

const projectUrl = computed(() => {
  if (!token.value) return '#/projects'
  return `#/projects/${token.value.projectId}`
})

const projectLabel = computed(() => {
  if (!token.value) return ''
  return token.value.projectSlug || `Project #${token.value.projectId}`
})

onMounted(load)

onUnmounted(() => {
  credentialStore.clear()
})
</script>

<template>
  <!-- Loading skeleton -->
  <template v-if="loading">
    <div class="space-y-4">
      <Skeleton class="h-6 w-48" />
      <Skeleton class="h-32 w-full" />
      <Skeleton class="h-64 w-full" />
    </div>
  </template>

  <!-- Not found -->
  <div v-else-if="notFound" class="text-center py-16">
    <h2 class="text-xl font-semibold">Token not found</h2>
    <p class="text-sm text-muted-foreground mt-2">This token may have been deleted.</p>
    <Button variant="outline" class="mt-4" @click="router.push({ name: 'tokens' })">
      Back to tokens
    </Button>
  </div>

  <!-- Token loaded -->
  <template v-else-if="token">
    <!-- Credential display (one-time) -->
    <CredentialDisplay
      :credential="savedCredential"
      class="mb-6"
      @dismiss="savedCredential = null"
    />

    <Breadcrumb class="mb-6">
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink href="#/projects">Projects</BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbLink :href="projectUrl">{{ projectLabel }}</BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbPage>{{ breadcrumbLabel }}</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>

    <PageHeader :title="breadcrumbLabel">
      <template #icon>
        <div class="size-10 rounded-lg bg-muted flex items-center justify-center">
          <Key class="size-5" />
        </div>
      </template>
      <template #actions>
        <Button variant="outline" size="sm" @click="router.push({ name: 'tokens' })">
          <ArrowLeft class="size-4" />
          Back
        </Button>
      </template>
    </PageHeader>

    <!-- Metadata grid -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Key ID</p>
        <Badge variant="secondary" class="font-mono text-xs">{{ token.keyID }}</Badge>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Created</p>
        <p class="text-sm">{{ formatDate(token.createdAt) }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Updated</p>
        <p class="text-sm">{{ formatDate(token.updatedAt) }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Last access</p>
        <p class="text-sm">{{ formatDate(token.lastAccessAt) }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Requests</p>
        <p class="text-sm font-medium">{{ token.requests.toLocaleString() }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">User</p>
        <p class="text-sm">{{ token.user }}</p>
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted-foreground">Project</p>
        <a
          v-if="token.projectId"
          :href="'#/projects/' + token.projectId"
          class="text-sm text-primary hover:underline"
        >
          {{ token.projectSlug || '(default)' }}
        </a>
        <span v-else class="text-sm text-muted-foreground">None</span>
      </div>
    </div>

    <Separator class="my-6" />

    <!-- Edit form -->
    <h3 class="text-lg font-medium mb-4">Configuration</h3>
    <TokenEditForm
      :config="(config as any)"
      :token="token"
      :projects="projects"
      :show-project="false"
    />

    <Separator class="my-6" />

    <!-- Actions -->
    <div class="flex flex-wrap gap-3">
      <Button :disabled="saving" @click="handleSave()">
        <Save class="size-4" />
        {{ saving ? 'Saving...' : 'Save Changes' }}
      </Button>
      <Button variant="outline" :disabled="regenerating" @click="handleRegenerate()">
        <RefreshCw class="size-4" />
        {{ regenerating ? 'Regenerating...' : 'Regenerate Secret' }}
      </Button>
      <Button
        variant="destructive"
        :disabled="deleting"
        @click="showDeleteDialog = true"
      >
        <Trash2 class="size-4" />
        Delete Token
      </Button>
    </div>
  </template>

  <!-- Delete confirmation -->
  <ConfirmDialog
    v-model:open="showDeleteDialog"
    title="Delete Token"
    :description="`Are you sure you want to delete this token? This action cannot be undone and all applications using this key will lose access.`"
    confirm-text="Delete"
    variant="destructive"
    :loading="deleting"
    @confirm="handleDelete()"
  />
</template>
