import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

interface MenuItem {
  id: number
  name: string
  path: string
  icon: string
  children?: MenuItem[]
}

export const useMenuStore = defineStore('menu', () => {
  const menus = ref<MenuItem[]>([])
  const currentSystemId = ref<number | null>(null)

  async function fetchMenus(systemId: number) {
    const data: any = await request.get(`/systems/${systemId}/menus`)
    menus.value = data
    currentSystemId.value = systemId
  }

  return { menus, currentSystemId, fetchMenus }
})
