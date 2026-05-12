import request from '@/utils/request'

export const getMenus = (systemId: number) =>
  request.get(`/systems/${systemId}/menus`)

export const createMenu = (systemId: number, data: {
  parent_id?: number
  name: string
  path?: string
  icon?: string
  sort_order?: number
}) => request.post(`/systems/${systemId}/menus`, data)

export const updateMenu = (systemId: number, mid: number, data: {
  parent_id?: number
  name?: string
  path?: string
  icon?: string
  sort_order?: number
}) => request.put(`/systems/${systemId}/menus/${mid}`, data)

export const deleteMenu = (systemId: number, mid: number) =>
  request.delete(`/systems/${systemId}/menus/${mid}`)
