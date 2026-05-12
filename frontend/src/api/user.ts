import request from '@/utils/request'

export const getUsers = (params?: { page?: number; page_size?: number }) =>
  request.get('/users', { params })

export const getUserById = (id: number) => request.get(`/users/${id}`)

export const getUserRoles = (id: number) => request.get(`/users/${id}/roles`)

export const createUser = (data: {
  username: string
  password: string
  email?: string
  phone?: string
}) => request.post('/users', data)

export const updateUser = (id: number, data: { email?: string; phone?: string }) =>
  request.put(`/users/${id}`, data)

export const deleteUser = (id: number) => request.delete(`/users/${id}`)

export const updateUserStatus = (id: number, status: number) =>
  request.put(`/users/${id}/status`, { status })
