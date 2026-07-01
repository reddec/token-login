<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Plus, Trash2 } from '@lucide/vue'
import type { Token, Project, NameValue } from '@/api'

interface FormConfig {
  label: string
  hosts: string[]
  paths: string[]
  projectId: number | undefined
  headers: NameValue[]
}

const props = defineProps<{
  config: FormConfig
  token?: Token | null
  projects: Project[]
  showProject: boolean
}>()

const emit = defineEmits<{
  'update:config': [value: FormConfig]
}>()

const customHeaders = reactive<{ name: string; value: string }[]>([])

// Local string refs for textareas — preserve raw input including trailing newlines.
// The parent config stores string[]; these refs hold the joined view.
const hostsText = ref(props.config.hosts.join('\n'))
const pathsText = ref(props.config.paths.join('\n'))

// Sync parent config → local text (initial load, external changes)
watch(
  () => props.config.hosts,
  (val) => {
    const text = val.join('\n')
    if (text !== hostsText.value) hostsText.value = text
  },
)
watch(
  () => props.config.paths,
  (val) => {
    const text = val.join('\n')
    if (text !== pathsText.value) pathsText.value = text
  },
)

function deriveArray(text: string): string[] {
  return text
    .split('\n')
    .map((s) => s.trim())
    .filter((s) => s !== '')
}

function onHostsInput(e: Event) {
  const text = (e.target as HTMLTextAreaElement).value
  hostsText.value = text
  const arr = deriveArray(text)
  if (arr.join('\n') !== props.config.hosts.join('\n')) {
    update('hosts', arr)
  }
}

function onPathsInput(e: Event) {
  const text = (e.target as HTMLTextAreaElement).value
  pathsText.value = text
  const arr = deriveArray(text)
  if (arr.join('\n') !== props.config.paths.join('\n')) {
    update('paths', arr)
  }
}

// Sync custom headers from config on changes
watch(
  () => props.config.headers,
  (val) => {
    customHeaders.length = 0
    if (val && val.length > 0) {
      customHeaders.push(...val.map((h) => ({ name: h.name, value: h.value })))
    }
  },
  { immediate: true },
)

function addHeader() {
  customHeaders.push({ name: '', value: '' })
}

function removeHeader(index: number) {
  customHeaders.splice(index, 1)
}

function syncHeaders() {
  props.config.headers = customHeaders
    .filter((h) => h.name.trim() !== '')
    .map((h) => ({ name: h.name, value: h.value }))
}

function update(key: keyof FormConfig, value: string | number | string[] | undefined) {
  ;(props.config as any)[key] = value
  emit('update:config', { ...props.config })
}
</script>

<template>
  <div class="space-y-5">
    <!-- Label -->
    <div class="space-y-2">
      <Label for="token-label">Label</Label>
      <Input
        id="token-label"
        :model-value="config.label"
        placeholder="My API Token"
        @update:model-value="update('label', $event)"
      />
      <p class="text-xs text-muted-foreground">
        A human-readable description to help identify this token.
      </p>
    </div>

    <!-- Host restriction -->
    <div class="space-y-2">
      <Label for="token-host">Host restriction</Label>
      <textarea
        id="token-host"
        class="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        :value="hostsText"
        placeholder="One glob per line. Leave empty to allow all."
        @input="onHostsInput"
      ></textarea>
      <p class="text-xs text-muted-foreground">
        Glob patterns supported (<code>*.example.com</code>, <code>**.com</code>). One per line. Empty means allow all hosts.
      </p>
    </div>

    <!-- Path restriction -->
    <div class="space-y-2">
      <Label for="token-path">Path restriction</Label>
      <textarea
        id="token-path"
        class="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        :value="pathsText"
        placeholder="One glob per line. Leave empty to allow all."
        @input="onPathsInput"
      ></textarea>
      <p class="text-xs text-muted-foreground">
        Glob patterns supported (<code>/api/**</code>, <code>/foo/*</code>). One per line. Empty means allow all paths.
      </p>
    </div>

    <!-- Project selector (create only) -->
    <div v-if="showProject" class="space-y-2">
      <Label for="token-project">Project</Label>
      <Select
        :model-value="config.projectId?.toString() ?? ''"
        @update:model-value="update('projectId', Number($event))"
      >
        <SelectTrigger id="token-project">
          <SelectValue placeholder="Select a project" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="p in projects" :key="p.id" :value="p.id.toString()">
            {{ p.slug || '(default)' }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <!-- Custom Headers -->
    <div class="space-y-2">
      <Label>Custom Headers</Label>
      <p class="text-xs text-muted-foreground">
        Headers injected when a request is authenticated by this token. Useful for forward
        authentication.
      </p>

      <div class="rounded-md border">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b bg-muted/50">
              <th class="px-3 py-2 text-left font-medium">Name</th>
              <th class="px-3 py-2 text-left font-medium">Value</th>
              <th class="w-10 px-3 py-2"></th>
            </tr>
          </thead>
          <tbody>
            <!-- Implicit headers (read-only) -->
            <tr v-if="token" class="border-b">
              <td class="px-3 py-2">
                <Badge variant="secondary">X-User</Badge>
              </td>
              <td class="px-3 py-2 text-muted-foreground">{{ token.user }}</td>
              <td class="px-3 py-2">
                <Badge variant="outline" class="text-xs">auto</Badge>
              </td>
            </tr>
            <tr v-if="token" class="border-b">
              <td class="px-3 py-2">
                <Badge variant="secondary">X-Token-Hint</Badge>
              </td>
              <td class="px-3 py-2 text-muted-foreground">{{ token.keyID }}</td>
              <td class="px-3 py-2">
                <Badge variant="outline" class="text-xs">auto</Badge>
              </td>
            </tr>

            <!-- Editable custom headers -->
            <tr v-for="(header, idx) in customHeaders" :key="idx" class="border-b">
              <td class="px-2 py-1.5">
                <Input
                  v-model="header.name"
                  placeholder="Header-Name"
                  class="h-8 text-sm"
                  @update:model-value="syncHeaders()"
                />
              </td>
              <td class="px-2 py-1.5">
                <Input
                  v-model="header.value"
                  placeholder="Value"
                  class="h-8 text-sm"
                  @update:model-value="syncHeaders()"
                />
              </td>
              <td class="px-2 py-1.5">
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-8"
                  @click="removeHeader(idx)"
                >
                  <Trash2 class="size-3.5" />
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <Button variant="outline" size="sm" class="gap-1" @click="addHeader()">
        <Plus class="size-3.5" />
        Add header
      </Button>
    </div>
  </div>
</template>
