import request from '@/utils/request'

export const getRoles = (systemId: number) =>
  request.get(`/systems/${systemId}/roles`)

export const createRole = (systemId: number, data: {
  name: string
  code: string
  description?: string
  is_default?: number
}) => request.post(`/systems/${systemId}/roles`, data)

export const updateRole = (systemId: number, rid: number, data: {
  name?: string
  description?: string
}) => request.put(`/systems/${systemId}/roles/${rid}`, data)

export const deleteRole = (systemId: number, rid: number) =>
  request.delete(`/systems/${systemId}/roles/${rid}`)

export const getRoleMenus = (systemId: number, rid: number) =>
  request.get(`/systems/${systemId}/roles/${rid}/menus`)

export const assignMenus = (systemId: number, rid: number, menu_ids: number[]) =>
  request.post(`/systems/${systemId}/roles/${rid}/menus`, { menu_ids })

export const getRolePermissions = (systemId: number, rid: number) =>
  request.get(`/systems/${systemId}/roles/${rid}/permissions`)

export const assignPermissions = (systemId: number, rid: number, permission_ids: number[]) =>
  request.post(`/systems/${systemId}/roles/${rid}/permissions`, { permission_ids })

export const getRoleUsers = (systemId: number, rid: number) =>
  request.get(`/systems/${systemId}/roles/${rid}/users`)

export const assignUsers = (systemId: number, rid: number, user_ids: number[]) =>
  request.post(`/systems/${systemId}/roles/${rid}/users`, { user_ids })

export const removeRoleUser = (systemId: number, rid: number, uid: number) =>
  request.delete(`/systems/${systemId}/roles/${rid}/users/${uid}`)
