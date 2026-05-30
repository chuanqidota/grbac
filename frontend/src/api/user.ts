import request from '@/utils/request'

export const getUsers = (params?: { page?: number; page_size?: number; is_super_admin?: number }) =>
  request.get('/users', { params })

export const getUserById = (id: number) => request.get(`/users/${id}`)

export const getUserRoles = (id: number) => request.get(`/users/${id}/roles`)

export const createUser = (data: {
  username: string
  chinese_name?: string
  password: string
  email?: string
  phone?: string
}) => request.post('/users', data)

export const updateUser = (id: number, data: {
  chinese_name?: string
  email?: string
  phone?: string
}) => request.put(`/users/${id}`, data)

export const deleteUser = (id: number) => request.delete(`/users/${id}`)

export const updateUserStatus = (id: number, status: number) =>
  request.put(`/users/${id}/status`, { status })

export const updateUserSuperAdmin = (id: number, is_super_admin: number) =>
  request.put(`/users/${id}/super-admin`, { is_super_admin })

export const resetUserPassword = (id: number, new_password: string) =>
  request.put(`/users/${id}/reset-password`, { new_password })
