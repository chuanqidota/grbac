import request from '@/utils/request'

export const getAuditLogs = (params?: {
  page?: number
  page_size?: number
  system_id?: number
}) => request.get('/audit-logs', { params })
