import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

interface UserInfo {
  id: number
  username: string
  chinese_name?: string
  email: string
  is_super_admin: boolean
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)

  async function fetchUserInfo() {
    const data: any = await request.get('/users/me')
    userInfo.value = data
    return data
  }

  function isSuperAdmin() {
    return userInfo.value?.is_super_admin ?? false
  }

  return { userInfo, fetchUserInfo, isSuperAdmin }
})
