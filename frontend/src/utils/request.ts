import axios, { type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import { getToken, getRefreshToken, setToken, setRefreshToken, removeToken, removeRefreshToken } from './token'
import router from '@/router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

NProgress.configure({ showSpinner: false, minimum: 0.2 })

let nprogressTimer: ReturnType<typeof setTimeout> | null = null
let activeRequests = 0

function startProgress() {
  activeRequests++
  if (!nprogressTimer) {
    nprogressTimer = setTimeout(() => {
      NProgress.start()
      nprogressTimer = null
    }, 300)
  }
}

function doneProgress() {
  activeRequests--
  if (nprogressTimer) {
    clearTimeout(nprogressTimer)
    nprogressTimer = null
  }
  if (activeRequests <= 0) {
    activeRequests = 0
    NProgress.done()
  }
}

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// Token refresh state
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    startProgress()
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { code, message, data } = response.data
    if (code === 0) {
      doneProgress()
      return data
    }
    if (code === 10003 || code === 10004) {
      removeToken()
      removeRefreshToken()
      router.push('/login')
    }
    doneProgress()
    return Promise.reject(new Error(message))
  },
  async (error) => {
    const originalConfig = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status === 401 && !originalConfig._retry) {
      // If already refreshing, queue this request
      if (isRefreshing) {
        doneProgress()
        return new Promise((resolve) => {
          pendingRequests.push((token: string) => {
            originalConfig.headers.Authorization = `Bearer ${token}`
            resolve(request(originalConfig))
          })
        })
      }

      originalConfig._retry = true
      isRefreshing = true

      try {
        const refreshToken = getRefreshToken()
        if (!refreshToken) {
          throw new Error('No refresh token')
        }

        // Use raw axios to avoid interceptor loop
        const resp = await axios.post('/api/auth/refresh', { refresh_token: refreshToken })
        const { code, data } = resp.data

        if (code !== 0) {
          throw new Error('Refresh failed')
        }

        // Store new tokens
        setToken(data.access_token)
        setRefreshToken(data.refresh_token)

        // Retry original request
        originalConfig.headers.Authorization = `Bearer ${data.access_token}`

        // Execute queued requests
        pendingRequests.forEach(cb => cb(data.access_token))
        pendingRequests = []

        doneProgress()
        return request(originalConfig)
      } catch {
        // Refresh failed — clear everything and redirect
        removeToken()
        removeRefreshToken()
        pendingRequests = []
        router.push('/login')
        doneProgress()
        return Promise.reject(new Error('Token已过期，请重新登录'))
      } finally {
        isRefreshing = false
      }
    }

    if (error.response) {
      const { status, data } = error.response
      if (status === 403) {
        doneProgress()
        return Promise.reject(new Error(data?.message || '无权限访问'))
      }
      doneProgress()
      return Promise.reject(new Error(data?.message || `请求失败 (${status})`))
    }
    doneProgress()
    return Promise.reject(error)
  }
)

export default request
