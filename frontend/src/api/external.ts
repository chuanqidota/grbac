import axios from 'axios'

const externalRequest = axios.create({
  baseURL: '/api/external',
  timeout: 10000,
})

export interface ExternalApiResult {
  status: number
  statusText: string
  data: any
  time: number
}

export async function callExternalApi(
  path: string,
  headers: Record<string, string>,
  params?: Record<string, string>
): Promise<ExternalApiResult> {
  const start = Date.now()
  const response = await externalRequest.get(path, { headers, params })
  return {
    status: response.status,
    statusText: response.statusText,
    data: response.data,
    time: Date.now() - start,
  }
}
