import { reactive, watch, onMounted, type UnwrapNestedRefs } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'

export function useUrlState<T extends Record<string, any>>(
  defaults: T,
  options?: { debounce?: number }
): {
  state: UnwrapNestedRefs<T>
  reset: () => void
} {
  const route = useRoute()
  const router = useRouter()
  const debounceMs = options?.debounce ?? 300

  const state = reactive<T>({ ...defaults }) as T

  function restoreFromUrl(query: LocationQuery) {
    for (const key of Object.keys(defaults) as Array<keyof T>) {
      const val = query[key as string]
      if (val !== undefined && val !== null && val !== '') {
        const defaultVal = defaults[key]
        if (typeof defaultVal === 'number') {
          const num = Number(val)
          ;(state as any)[key] = isNaN(num) ? defaultVal : num
        } else if (typeof defaultVal === 'boolean') {
          (state as any)[key] = val === 'true' || val === '1'
        } else {
          (state as any)[key] = String(val)
        }
      } else {
        (state as any)[key] = defaults[key]
      }
    }
  }

  let syncTimer: ReturnType<typeof setTimeout> | null = null
  function syncToUrl() {
    if (syncTimer) clearTimeout(syncTimer)
    syncTimer = setTimeout(() => {
      const query: Record<string, string> = {}
      for (const key of Object.keys(defaults) as Array<keyof T>) {
        const val = state[key]
        const defaultVal = defaults[key]
        if (val !== defaultVal && val !== '' && val !== null && val !== undefined) {
          query[key as string] = String(val)
        }
      }
      router.replace({ query })
    }, debounceMs)
  }

  watch(
    () => ({ ...state }),
    () => syncToUrl(),
    { deep: true }
  )

  watch(
    () => route.query,
    (query) => restoreFromUrl(query),
    { deep: true }
  )

  onMounted(() => {
    restoreFromUrl(route.query)
  })

  function reset() {
    for (const key of Object.keys(defaults) as Array<keyof T>) {
      (state as any)[key] = defaults[key]
    }
    router.replace({ query: {} })
  }

  return { state: state as UnwrapNestedRefs<T>, reset }
}
