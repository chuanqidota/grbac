import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getSystems } from '@/api/system'

interface System {
  id: number
  name: string
  code: string
  description: string
  status: number
  current_user_role?: string
}

export const useSystemStore = defineStore('system', () => {
  const systems = ref<System[]>([])
  const currentSystemId = ref<number | null>(null)
  const loading = ref(false)

  async function fetchSystems() {
    loading.value = true
    try {
      const data: any = await getSystems({ page: 1, page_size: 100 })
      systems.value = data.list || []
      if (systems.value.length > 0 && !currentSystemId.value) {
        currentSystemId.value = systems.value[0]?.id ?? null
      }
    } finally {
      loading.value = false
    }
  }

  function setCurrentSystem(id: number) {
    currentSystemId.value = id
  }

  function currentSystem() {
    return systems.value.find(s => s.id === currentSystemId.value) || null
  }

  return { systems, currentSystemId, loading, fetchSystems, setCurrentSystem, currentSystem }
})
