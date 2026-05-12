import axios from 'axios'
import { getToken, removeToken, removeRefreshToken } from './token'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code === 0) {
      return data
    }
    if (code === 10003 || code === 10004) {
      removeToken()
      removeRefreshToken()
      router.push('/login')
    }
    return Promise.reject(new Error(message))
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        removeToken()
        removeRefreshToken()
        router.push('/login')
        return Promise.reject(new Error(data?.message || 'Token无效，请重新登录'))
      }
      if (status === 403) {
        return Promise.reject(new Error(data?.message || '无权限访问'))
      }
      return Promise.reject(new Error(data?.message || `请求失败 (${status})`))
    }
    return Promise.reject(error)
  }
)

export default request
