import { reactive, watch, toRaw } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY = 'token-admin-prefs'

interface Preferences {
  darkMode: boolean
  sidebarOpen: boolean
}

function load(): Preferences {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw) as Preferences
  } catch {
    /* ignore corrupt data */
  }
  return { darkMode: false, sidebarOpen: false }
}

export const usePreferencesStore = defineStore('preferences', () => {
  const prefs = reactive<Preferences>(load())

  watch(
    () => ({ ...prefs }),
    (val) => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(toRaw(val)))
    },
    { deep: true },
  )

  watch(
    () => prefs.darkMode,
    (val) => {
      document.documentElement.classList.toggle('dark', val)
    },
    { immediate: true },
  )

  function toggleDark() {
    prefs.darkMode = !prefs.darkMode
  }

  function toggleSidebar() {
    prefs.sidebarOpen = !prefs.sidebarOpen
  }

  return { prefs, toggleDark, toggleSidebar }
})
