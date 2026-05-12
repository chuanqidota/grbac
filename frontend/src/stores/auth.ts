import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'
import { setToken, setRefreshToken, removeToken, removeRefreshToken, getToken, getRefreshToken } from '@/utils/token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())

  async function login(username: string, password: string) {
    const data: any = await request.post('/auth/login', { username, password })
    token.value = data.access_token
    setToken(data.access_token)
    setRefreshToken(data.refresh_token)
    return data
  }

  async function logout() {
    try {
      await request.post('/auth/logout')
    } finally {
      token.value = null
      removeToken()
      removeRefreshToken()
    }
  }

  async function refreshToken() {
    const rt = getRefreshToken()
    if (!rt) throw new Error('No refresh token')

    const data: any = await request.post('/auth/refresh', { refresh_token: rt })
    token.value = data.access_token
    setToken(data.access_token)
    setRefreshToken(data.refresh_token)
    return data
  }

  return { token, login, logout, refreshToken }
})
