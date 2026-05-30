import { ref, type Ref } from 'vue'
import { ElMessageBox, ElNotification, ElMessage } from 'element-plus'

export function useConfirmDelete(
  deleteFn: (id: number) => Promise<any>,
  options?: {
    onSuccess?: () => void
    entityName?: string
  }
): {
  deleting: Ref<boolean>
  confirmDelete: (row: { id: number; name?: string }) => Promise<void>
  batchDelete: (ids: number[], callbacks?: {
    onProgress?: (done: number, total: number) => void
    onComplete?: (success: number, failed: number) => void
  }) => Promise<void>
} {
  const deleting = ref(false)
  const entityName = options?.entityName || '记录'

  async function confirmDelete(row: { id: number; name?: string }) {
    const name = row.name || `ID: ${row.id}`
    try {
      await ElMessageBox.confirm(`确定要删除${entityName} "${name}" 吗？`, '确认删除', { type: 'warning' })
      deleting.value = true
      await deleteFn(row.id)
      ElMessage.success('删除成功')
      options?.onSuccess?.()
    } catch (error: any) {
      if (error !== 'cancel') {
        ElMessage.error(error.message || '删除失败')
      }
    } finally {
      deleting.value = false
    }
  }

  async function batchDelete(ids: number[], callbacks?: {
    onProgress?: (done: number, total: number) => void
    onComplete?: (success: number, failed: number) => void
  }) {
    if (ids.length === 0) return

    try {
      await ElMessageBox.confirm(`确定要删除选中的 ${ids.length} 个${entityName}吗？`, '批量删除', { type: 'warning' })
    } catch {
      return
    }

    let success = 0
    let failed = 0

    for (let i = 0; i < ids.length; i++) {
      const id = ids[i]
      if (id === undefined) { failed++; continue }
      try {
        await deleteFn(id)
        success++
      } catch {
        failed++
      }
      callbacks?.onProgress?.(i + 1, ids.length)
    }

    if (failed === 0) {
      ElNotification.success({ title: '批量删除成功', message: `成功删除 ${success} 个${entityName}` })
    } else {
      ElNotification.warning({ title: '批量删除完成', message: `成功 ${success} 个，失败 ${failed} 个` })
    }

    callbacks?.onComplete?.(success, failed)
    options?.onSuccess?.()
  }

  return { deleting, confirmDelete, batchDelete }
}
