import request from '@/utils/request'

export const getWebhooks = (systemId: number) =>
  request.get(`/systems/${systemId}/webhooks`)

export const createWebhook = (systemId: number, data: {
  url: string
  events: string
}) => request.post(`/systems/${systemId}/webhooks`, data)

export const updateWebhook = (systemId: number, wid: number, data: {
  url?: string
  events?: string
  status?: number
}) => request.put(`/systems/${systemId}/webhooks/${wid}`, data)

export const deleteWebhook = (systemId: number, wid: number) =>
  request.delete(`/systems/${systemId}/webhooks/${wid}`)
