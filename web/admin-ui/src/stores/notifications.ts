import { ref } from 'vue'

export interface Notification {
  id: number
  message: string
  kind: 'success' | 'error' | 'info'
}

let nextId = 1
const notifications = ref<Notification[]>([])

export function useNotifications() {
  function notify(message: string, kind: Notification['kind'] = 'info') {
    const id = nextId++
    notifications.value.push({ id, message, kind })
    setTimeout(() => dismiss(id), 5000)
  }

  function dismiss(id: number) {
    notifications.value = notifications.value.filter((n) => n.id !== id)
  }

  return { notifications, notify, dismiss }
}
