import request from '@/utils/request'

export const login = (data: { username: string; password: string }) =>
  request.post('/auth/login', data)

export const logout = () => request.post('/auth/logout')

export const refreshToken = (data: { refresh_token: string }) =>
  request.post('/auth/refresh', data)

export const changePassword = (data: { old_password: string; new_password: string }) =>
  request.post('/auth/change-password', data)
