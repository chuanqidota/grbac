import { ref, reactive, computed, type Ref, type UnwrapNestedRefs } from 'vue'
import { ElMessageBox } from 'element-plus'

export function useFormDialog<T extends Record<string, any>>(
  submitFn: (data: T) => Promise<any>,
  options?: {
    onSuccess?: () => void
    onError?: (error: Error) => void
  }
): {
  visible: Ref<boolean>
  isEditing: Ref<boolean>
  formData: UnwrapNestedRefs<T>
  formChanged: Ref<boolean>
  submitting: Ref<boolean>
  editingId: Ref<number | null>
  open: (row?: T & { id?: number }) => void
  close: () => Promise<void>
  submit: () => Promise<void>
  forceClose: () => void
} {
  const visible = ref(false)
  const isEditing = ref(false)
  const submitting = ref(false)
  const editingId = ref<number | null>(null)

  const formData = reactive<T>({} as T) as UnwrapNestedRefs<T>
  const snapshot = ref<string>('')
  const formChanged = computed(() => JSON.stringify(formData) !== snapshot.value)

  function open(row?: T & { id?: number }) {
    if (row && row.id) {
      isEditing.value = true
      editingId.value = row.id
      const { id, ...rest } = row
      Object.assign(formData, JSON.parse(JSON.stringify(rest)))
    } else {
      isEditing.value = false
      editingId.value = null
      if (row) {
        Object.assign(formData, JSON.parse(JSON.stringify(row)))
      }
    }
    snapshot.value = JSON.stringify(formData)
    visible.value = true
  }

  async function close() {
    if (formChanged.value) {
      try {
        await ElMessageBox.confirm('表单已修改，确认放弃更改？', '提示', { type: 'warning' })
        forceClose()
      } catch {
        // user cancelled
      }
    } else {
      forceClose()
    }
  }

  function forceClose() {
    visible.value = false
  }

  async function submit() {
    submitting.value = true
    try {
      const data = { ...formData } as T
      if (isEditing.value && editingId.value) {
        (data as any).id = editingId.value
      }
      await submitFn(data)
      options?.onSuccess?.()
      forceClose()
    } catch (error: any) {
      if (options?.onError) {
        options.onError(error)
      }
    } finally {
      submitting.value = false
    }
  }

  return { visible, isEditing, formData, formChanged, submitting, editingId, open, close, submit, forceClose }
}
