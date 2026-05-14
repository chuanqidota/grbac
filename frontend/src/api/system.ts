import request from '@/utils/request'

export const getSystems = (params?: { page?: number; page_size?: number }) =>
  request.get('/systems', { params })

export const getSystemById = (id: number) => request.get(`/systems/${id}`)

export const createSystem = (data: {
  name: string
  description?: string
}) => request.post('/systems', data)

export const updateSystem = (id: number, data: { name?: string; description?: string }) =>
  request.put(`/systems/${id}`, data)

export const deleteSystem = (id: number) => request.delete(`/systems/${id}`)

export const getSystemMembers = (id: number) => request.get(`/systems/${id}/members`)

export const getMemberUsers = (id: number) => request.get(`/systems/${id}/members/users`)

export const addSystemMember = (id: number, data: { user_id: number; role: string }) =>
  request.post(`/systems/${id}/members`, data)

export const removeSystemMember = (id: number, uid: number) =>
  request.delete(`/systems/${id}/members/${uid}`)

export const getMemberRoles = (systemId: number, userId: number) =>
  request.get(`/systems/${systemId}/members/${userId}/roles`)

export const getMemberMenus = (systemId: number, userId: number) =>
  request.get(`/systems/${systemId}/members/${userId}/menus`)

export const getMemberPermissions = (systemId: number, userId: number) =>
  request.get(`/systems/${systemId}/members/${userId}/permissions`)
