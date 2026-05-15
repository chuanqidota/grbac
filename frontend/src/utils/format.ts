/**
 * Format a date string to a localized display string.
 */
export function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

/**
 * Format a user's display name as "中文名(username)" or just "username"
 * when chinese_name is not set.
 */
export function formatUserDisplay(user: { username: string; chinese_name?: string }): string {
  if (user.chinese_name) {
    return `${user.chinese_name}(${user.username})`
  }
  return user.username
}

/**
 * Check whether a user matches a search query against both username and chinese_name.
 */
export function matchUserSearch(user: { username: string; chinese_name?: string }, query: string): boolean {
  const q = query.toLowerCase()
  if (user.username.toLowerCase().includes(q)) return true
  if (user.chinese_name && user.chinese_name.toLowerCase().includes(q)) return true
  return false
}
