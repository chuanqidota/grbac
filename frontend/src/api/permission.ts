import request from '@/utils/request'

export const getPermissions = (systemId: number, params?: { page?: number; page_size?: number; keyword?: string; method?: string }) =>
  request.get(`/systems/${systemId}/permissions`, { params })

export const createPermission = (systemId: number, data: {
  code: string
  name: string
  method: string
  path: string
  description?: string
}) => request.post(`/systems/${systemId}/permissions`, data)

export const updatePermission = (systemId: number, pid: number, data: {
  code?: string
  name?: string
  method?: string
  path?: string
  description?: string
}) => request.put(`/systems/${systemId}/permissions/${pid}`, data)

export const deletePermission = (systemId: number, pid: number) =>
  request.delete(`/systems/${systemId}/permissions/${pid}`)
