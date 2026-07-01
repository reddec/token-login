import { computed, ref, type Ref } from 'vue'
import type { Token } from '@/api'

/**
 * Shared substring search over token fields (label, key id, hosts, paths, project).
 * Used by both the tokens page and the per-project token list.
 */
export function useTokenSearch(tokens: Ref<Token[]>) {
  const searchQuery = ref('')

  const filtered = computed(() => {
    const q = searchQuery.value.toLowerCase().trim()
    if (!q) return tokens.value
    return tokens.value.filter(
      (t) =>
        t.label.toLowerCase().includes(q) ||
        t.keyID.toLowerCase().includes(q) ||
        t.hosts?.some((h) => h.toLowerCase().includes(q)) ||
        t.paths?.some((p) => p.toLowerCase().includes(q)) ||
        t.projectSlug?.toLowerCase().includes(q),
    )
  })

  return { searchQuery, filtered }
}
