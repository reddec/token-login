import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { Credential } from '@/api'

export const useCredentialStore = defineStore('credential', () => {
  const credential = ref<Credential | null>(null)
  const tokenId = ref<number | null>(null)

  function set(creds: Credential) {
    credential.value = creds
    tokenId.value = creds.id
  }

  function consume(forTokenId: number): Credential | null {
    if (tokenId.value !== forTokenId) return null
    const c = credential.value
    credential.value = null
    tokenId.value = null
    return c
  }

  function clear() {
    credential.value = null
    tokenId.value = null
  }

  return { credential, tokenId, set, consume, clear }
})
