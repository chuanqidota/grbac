import { onMounted, onUnmounted } from 'vue'

export function useKeyboard(handlers: {
  onSearch?: () => void
  onNew?: () => void
}) {
  function isInInput(): boolean {
    const el = document.activeElement
    if (!el) return false
    const tag = el.tagName.toLowerCase()
    return tag === 'input' || tag === 'textarea' || el.getAttribute('contenteditable') === 'true'
  }

  function handleKeydown(e: KeyboardEvent) {
    // Ctrl+Enter: submit the active dialog/drawer form
    if (e.ctrlKey && e.key === 'Enter') {
      const submitBtn = document.querySelector(
        '.el-dialog:not([style*="display: none"]) .el-button--primary, .el-drawer:not([style*="display: none"]) .el-button--primary'
      ) as HTMLButtonElement | null
      if (submitBtn && !submitBtn.disabled) {
        submitBtn.click()
        e.preventDefault()
      }
      return
    }

    // Skip shortcuts when user is typing in an input
    if (isInInput()) return

    // `/`: focus search input
    if (e.key === '/') {
      e.preventDefault()
      const searchInput = document.querySelector(
        '.search-bar .el-input__inner, .filter-bar .el-input__inner'
      ) as HTMLInputElement | null
      searchInput?.focus()
      return
    }

    // `n`: trigger new/create action
    if (e.key === 'n' && !e.ctrlKey && !e.metaKey && !e.altKey) {
      e.preventDefault()
      handlers.onNew?.()
      return
    }
  }

  onMounted(() => document.addEventListener('keydown', handleKeydown))
  onUnmounted(() => document.removeEventListener('keydown', handleKeydown))
}
