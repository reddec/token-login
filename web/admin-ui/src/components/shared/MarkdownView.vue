<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps<{
  source: string
}>()

// marked runs synchronously by default; sanitize the output before rendering.
const html = computed(() => {
  const parsed = marked.parse(props.source, { async: false })
  return DOMPurify.sanitize(typeof parsed === 'string' ? parsed : '')
})
</script>

<template>
  <div class="prose prose-sm dark:prose-invert max-w-none" v-html="html" />
</template>
