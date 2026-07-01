<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import {
  uniqueNamesGenerator,
  adjectives,
  colors,
  animals,
} from 'unique-names-generator'
import { useNotifications } from '@/stores/notifications'
import { getErrorMessage } from '@/lib/api-error'
const { notify } = useNotifications()

import { listTokens, createToken, listProjects } from '@/api'
import type { Token, Project, TokenConfig } from '@/api'
import { useCredentialStore } from '@/stores/credential'

import PageHeader from '@/components/shared/PageHeader.vue'
import EmptyState from '@/components/shared/EmptyState.vue'
import TokenTable from '@/components/tokens/TokenTable.vue'
import TokenCard from '@/components/tokens/TokenCard.vue'
import TokenEditForm from '@/components/tokens/TokenEditForm.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { KeyRound, Plus, RotateCw } from '@lucide/vue'

const router = useRouter()
const credentialStore = useCredentialStore()

const tokens = ref<Token[]>([])
const projects = ref<Project[]>([])
const loading = ref(true)
const error = ref(false)
const searchQuery = ref('')

const createDialogOpen = ref(false)
const creating = ref(false)

const newConfig = reactive<TokenConfig>({
  label: '',
  hosts: [],
  paths: [],
  headers: [],
  projectId: 0,
})

async function load() {
  loading.value = true
  error.value = false
  try {
    const [tokensRes, projectsRes] = await Promise.all([
      listTokens(),
      listProjects(),
    ])
    tokens.value = tokensRes.data ?? []
    projects.value = projectsRes.data ?? []
  } catch (e) {
    error.value = true
    notify(getErrorMessage(e, 'Failed to load tokens'), 'error')
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  const q = searchQuery.value.toLowerCase().trim()
  if (!q) return tokens.value
  return tokens.value.filter(
    (t) =>
      t.label.toLowerCase().includes(q) ||
      t.keyID.toLowerCase().includes(q) ||
      t.hosts?.some(h => h.toLowerCase().includes(q)) ||
      t.paths?.some(p => p.toLowerCase().includes(q)) ||
      t.projectSlug?.toLowerCase().includes(q),
  )
})

function openCreateDialog() {
  newConfig.label = uniqueNamesGenerator({
    dictionaries: [adjectives, colors, animals],
    separator: '-',
    length: 2,
  })
  newConfig.hosts = []
  newConfig.paths = []
  newConfig.headers = []
  newConfig.projectId = projects.value[0]?.id ?? 0
  createDialogOpen.value = true
}

async function handleCreate() {
  if (!newConfig.projectId) return
  creating.value = true
  try {
    const result = await createToken({ body: { ...newConfig } })
    const cred = result.data
    if (cred) {
      credentialStore.set(cred)
      notify('Token created', 'success')
      createDialogOpen.value = false
      router.push({ name: 'token-detail', params: { id: String(cred.id) } })
    }
  } catch (e) {
    notify(getErrorMessage(e, 'Failed to create token'), 'error')
  } finally {
    creating.value = false
  }
}

onMounted(load)
</script>

<template>
  <PageHeader title="Tokens" description="Manage your API access tokens">
    <template #actions>
      <Button variant="outline" size="icon" :disabled="loading" @click="load()">
        <RotateCw :class="['size-4', loading && 'animate-spin']" />
      </Button>
      <Button @click="openCreateDialog()">
        <Plus class="size-4" />
        Create Token
      </Button>
    </template>
  </PageHeader>

  <!-- Error state -->
  <div v-if="error && !loading" class="text-center py-12">
    <p class="text-sm text-muted-foreground mb-4">Failed to load tokens.</p>
    <Button variant="outline" @click="load()">Retry</Button>
  </div>

  <!-- Empty state -->
  <EmptyState
    v-else-if="!loading && tokens.length === 0"
    :icon="KeyRound"
    title="No tokens yet"
    description="Create your first API token to get started."
  >
    <template #action>
      <Button @click="openCreateDialog()">
        <Plus class="size-4" />
        Create Token
      </Button>
    </template>
  </EmptyState>

  <!-- Populated -->
  <template v-else>
    <!-- Search -->
    <div class="mb-4">
      <Input
        v-model="searchQuery"
        placeholder="Search tokens by label, key, host, path, or project..."
        class="max-w-md"
      />
    </div>

    <!-- Desktop table -->
    <div class="hidden md:block">
      <TokenTable :tokens="filtered" :loading="loading" />
    </div>

    <!-- Mobile cards -->
    <div class="md:hidden space-y-3">
      <template v-if="loading">
        <div v-for="i in 3" :key="i" class="rounded-lg border bg-card p-4 space-y-2">
          <div class="h-4 w-2/3 bg-muted rounded animate-pulse" />
          <div class="h-3 w-1/3 bg-muted rounded animate-pulse" />
        </div>
      </template>
      <TokenCard v-for="token in filtered" :key="token.id" :token="token" />
      <p v-if="filtered.length === 0 && !loading" class="text-center text-muted-foreground py-8 text-sm">
        No tokens match your search.
      </p>
    </div>
  </template>

  <!-- Create Dialog -->
  <Dialog v-model:open="createDialogOpen">
    <DialogContent class="sm:max-w-lg max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Create Token</DialogTitle>
        <DialogDescription>Set up a new API token with host and path restrictions.</DialogDescription>
      </DialogHeader>
      <TokenEditForm
        :config="(newConfig as any)"
        :projects="projects"
        :show-project="true"
      />
      <div class="flex justify-end gap-2 mt-4">
        <Button variant="outline" :disabled="creating" @click="createDialogOpen = false">
          Cancel
        </Button>
        <Button :disabled="creating || !newConfig.projectId" @click="handleCreate()">
          {{ creating ? 'Creating...' : 'Create' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
