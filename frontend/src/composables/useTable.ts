import { ref, type Ref } from 'vue'
import { ElMessage } from 'element-plus'

export function useTable<T>(
  fetchFn: (params: { page: number; page_size: number; [key: string]: any }) => Promise<{ list: T[]; total: number }>,
  options?: {
    defaultPageSize?: number
    onError?: (error: Error) => void
  }
): {
  data: Ref<T[]>
  total: Ref<number>
  loading: Ref<boolean>
  skeleton: Ref<boolean>
  page: Ref<number>
  pageSize: Ref<number>
  refresh: () => Promise<void>
  handleSizeChange: (size: number) => void
  handleCurrentChange: (p: number) => void
} {
  const data = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const loading = ref(false)
  const skeleton = ref(false)
  const page = ref(1)
  const pageSize = ref(options?.defaultPageSize ?? 20)

  let isFirstLoad = true

  async function refresh() {
    if (isFirstLoad) {
      skeleton.value = true
    } else {
      loading.value = true
    }
    try {
      const result = await fetchFn({ page: page.value, page_size: pageSize.value })
      data.value = result.list || []
      total.value = result.total || 0
    } catch (error: any) {
      if (options?.onError) {
        options.onError(error)
      } else {
        ElMessage.error(error.message || '获取数据失败')
      }
    } finally {
      skeleton.value = false
      loading.value = false
      isFirstLoad = false
    }
  }

  function handleSizeChange(size: number) {
    pageSize.value = size
    page.value = 1
    refresh()
  }

  function handleCurrentChange(p: number) {
    page.value = p
    refresh()
  }

  return { data, total, loading, skeleton, page, pageSize, refresh, handleSizeChange, handleCurrentChange }
}
