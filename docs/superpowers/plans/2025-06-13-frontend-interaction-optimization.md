# 前端交互优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 对 grbac 前端进行全面交互体验优化，解决状态不持久、表单体验差、反馈不及时、数据查找低效四大痛点。

**Architecture:** 三阶段混合方案——第一阶段快速胜利（低成本全页面覆盖），第二阶段构建共享 composables 统一重复模式，第三阶段深层优化（向导、排序、快捷键等）。

**Tech Stack:** Vue 3 + TypeScript + Element Plus + Pinia + NProgress

---

## 第一阶段：快速胜利

### Task 1: 安装 NProgress 并添加全局加载指示器

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/src/utils/request.ts`

- [ ] **Step 1: 安装依赖**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npm install nprogress && npm install -D @types/nprogress
```

- [ ] **Step 2: 在 request.ts 中添加 NProgress**

在 `frontend/src/utils/request.ts` 顶部添加 import，在请求/响应拦截器中集成 NProgress：

```typescript
// 在第 1 行之前添加
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

NProgress.configure({ showSpinner: false, minimum: 0.2 })

// 用于延迟启动 NProgress，避免快速请求闪烁
let nprogressTimer: ReturnType<typeof setTimeout> | null = null
let activeRequests = 0

function startProgress() {
  activeRequests++
  if (!nprogressTimer) {
    nprogressTimer = setTimeout(() => {
      NProgress.start()
      nprogressTimer = null
    }, 300)
  }
}

function doneProgress() {
  activeRequests--
  if (nprogressTimer) {
    clearTimeout(nprogressTimer)
    nprogressTimer = null
  }
  if (activeRequests <= 0) {
    activeRequests = 0
    NProgress.done()
  }
}
```

在请求拦截器的 `return config` 之前添加 `startProgress()`：

```typescript
request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    startProgress()
    return config
  },
  (error) => {
    doneProgress()
    return Promise.reject(error)
  }
)
```

在响应拦截器的成功和错误路径末尾添加 `doneProgress()`。在成功路径 `return data` 之前添加，在错误路径的每个 `return Promise.reject(...)` 之前添加。

- [ ] **Step 3: 验证**

启动前端开发服务器，确认请求时顶部出现细线进度条。

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npm run dev
```

- [ ] **Step 4: 提交**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/utils/request.ts
git commit -m "feat(frontend): 添加 NProgress 全局加载指示器"
```

---

### Task 2: 表单体验增强 — 自动聚焦与 dirty 关闭确认

**Files:**
- Modify: `frontend/src/views/user/UserView.vue`
- Modify: `frontend/src/views/system/SystemView.vue`
- Modify: `frontend/src/views/system/RoleView.vue`
- Modify: `frontend/src/views/system/MenuView.vue`
- Modify: `frontend/src/views/system/PermissionView.vue`
- Modify: `frontend/src/views/system/WebhookView.vue`
- Modify: `frontend/src/views/system/MemberView.vue`
- Modify: `frontend/src/views/profile/ChangePasswordView.vue`

- [ ] **Step 1: 为 UserView 创建对话框添加 autofocus**

在 `frontend/src/views/user/UserView.vue` 中，找到创建/编辑对话框的第一个 `el-input`（用户名输入框，约第 112 行），添加 `autofocus` 属性：

```html
<el-input
  v-model="form.username"
  :disabled="isEditing"
  placeholder="请输入用户名（英文名）"
  autofocus
/>
```

- [ ] **Step 2: 为 UserView 添加 dirty 关闭确认**

在 `frontend/src/views/user/UserView.vue` 的 `<script setup>` 中，添加表单原始数据追踪和关闭确认逻辑。

在 `const formRef = ref<FormInstance>()` 之后添加：

```typescript
const originalForm = ref<string>('')
const formDirty = computed(() => JSON.stringify(form.value) !== originalForm.value)
```

修改 `showCreateDialog` 函数（约第 295 行）：

```typescript
function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { username: '', chinese_name: '', password: '', email: '', phone: '' }
  originalForm.value = JSON.stringify(form.value)
  dialogVisible.value = true
}
```

修改 `showEditDialog` 函数（约第 307 行），在末尾添加：

```typescript
originalForm.value = JSON.stringify(form.value)
```

在 `<el-dialog>` 标签上添加 `before-close` 处理：

```html
<el-dialog
  v-model="dialogVisible"
  :title="isEditing ? '编辑用户' : '创建用户'"
  width="500px"
  :before-close="handleDialogClose"
>
```

在 script 中添加 `handleDialogClose` 函数：

```typescript
async function handleDialogClose(done: () => void) {
  if (formDirty.value) {
    try {
      await ElMessageBox.confirm('表单已修改，确认放弃更改？', '提示', { type: 'warning' })
      done()
    } catch {
      // 用户取消关闭
    }
  } else {
    done()
  }
}
```

- [ ] **Step 3: 为其余 7 个页面应用相同的 autofocus + dirty 模式**

对以下每个页面重复 Step 1-2 的模式（每个页面的对话框第一个 `el-input` 添加 `autofocus`，添加 `originalForm`/`formDirty`/`handleDialogClose`）：

**SystemView.vue** — 系统名称输入框 autofocus，添加 dirty 检测
**RoleView.vue** — 角色名称输入框 autofocus，添加 dirty 检测
**MenuView.vue** — 菜单名称输入框 autofocus，添加 dirty 检测（注意：MenuView 的第一个输入是 `el-cascader`，autofocus 加在菜单名称 `el-input` 上）
**PermissionView.vue** — 权限编码输入框 autofocus，添加 dirty 检测
**WebhookView.vue** — URL 输入框 autofocus，添加 dirty 检测
**MemberView.vue** — 分配角色对话框的用户选择 `el-select` 不支持 autofocus，跳过自动聚焦，仅添加 dirty 检测
**ChangePasswordView.vue** — 旧密码输入框 autofocus（无对话框，此页面跳过 dirty 检测）

- [ ] **Step 4: 验证**

逐个页面测试：打开对话框 → 确认第一个输入框自动聚焦 → 修改表单 → 点击取消/关闭 → 确认弹出确认对话框。

- [ ] **Step 5: 提交**

```bash
git add frontend/src/views/
git commit -m "feat(frontend): 表单对话框自动聚焦与 dirty 关闭确认"
```

---

### Task 3: URL 状态同步 — UserView

**Files:**
- Modify: `frontend/src/views/user/UserView.vue`

- [ ] **Step 1: 添加 URL 状态同步逻辑**

在 `frontend/src/views/user/UserView.vue` 的 `<script setup>` 中，添加 `useRouter`/`useRoute` 导入和 URL 同步逻辑。

替换 import 行（约第 189 行）：

```typescript
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
```

在 `const formRef = ref<FormInstance>()` 之前，添加路由相关逻辑：

```typescript
const route = useRoute()
const router = useRouter()

// 从 URL 恢复筛选状态
function restoreFromUrl() {
  const q = route.query
  if (q.keyword) searchUsername.value = String(q.keyword)
  if (q.page) currentPage.value = Number(q.page) || 1
  if (q.pageSize) pageSize.value = Number(q.pageSize) || 20
  if (q.superAdmin !== undefined && q.superAdmin !== '') filterSuperAdmin.value = Number(q.superAdmin)
}

// 同步筛选状态到 URL
function syncToUrl() {
  const query: Record<string, string> = {}
  if (searchUsername.value) query.keyword = searchUsername.value
  if (currentPage.value > 1) query.page = String(currentPage.value)
  if (pageSize.value !== 20) query.pageSize = String(pageSize.value)
  if (filterSuperAdmin.value !== null) query.superAdmin = String(filterSuperAdmin.value)
  router.replace({ query })
}

// 搜索关键字防抖同步到 URL
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(searchUsername, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    syncToUrl()
    fetchUsers()
  }, 300)
})
```

- [ ] **Step 2: 修改 handleSearch、handleSizeChange、handleCurrentChange**

```typescript
function handleSearch() {
  currentPage.value = 1
  syncToUrl()
  fetchUsers()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  syncToUrl()
  fetchUsers()
}

function handleCurrentChange(page: number) {
  currentPage.value = page
  syncToUrl()
  fetchUsers()
}
```

- [ ] **Step 3: 修改 onMounted 恢复 URL 状态**

```typescript
onMounted(() => {
  restoreFromUrl()
  fetchUsers()
})
```

- [ ] **Step 4: 修改搜索框模板**

将 `@clear="handleSearch"` 和 `@keyup.enter="handleSearch"` 从搜索框移除（因为 watch 已自动处理），保留 `clearable`：

```html
<el-input
  v-model="searchUsername"
  placeholder="搜索用户名"
  clearable
  style="width: 300px"
>
```

超管筛选的 `@change="handleSearch"` 改为：

```html
<el-select
  v-model="filterSuperAdmin"
  placeholder="超管筛选"
  clearable
  style="width: 150px"
  @change="handleFilterSuperAdmin"
>
```

添加：

```typescript
function handleFilterSuperAdmin() {
  currentPage.value = 1
  syncToUrl()
  fetchUsers()
}
```

- [ ] **Step 5: 验证**

在 UserView 页面：输入搜索 → 刷新浏览器 → 确认搜索条件保留；切换页码 → 复制 URL 在新标签页打开 → 确认页码正确。

- [ ] **Step 6: 提交**

```bash
git add frontend/src/views/user/UserView.vue
git commit -m "feat(frontend): UserView URL 状态同步"
```

---

### Task 4: URL 状态同步 — PermissionView

**Files:**
- Modify: `frontend/src/views/system/PermissionView.vue`

- [ ] **Step 1: 添加 URL 状态同步**

在 `frontend/src/views/system/PermissionView.vue` 中，添加与 Task 3 相同的模式，但额外包含 `filterMethod` 参数。

在 `<script setup>` 中添加：

```typescript
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

function restoreFromUrl() {
  const q = route.query
  if (q.keyword) searchKeyword.value = String(q.keyword)
  if (q.page) currentPage.value = Number(q.page) || 1
  if (q.pageSize) pageSize.value = Number(q.pageSize) || 20
  if (q.method) filterMethod.value = String(q.method)
}

function syncToUrl() {
  const query: Record<string, string> = {}
  if (searchKeyword.value) query.keyword = searchKeyword.value
  if (currentPage.value > 1) query.page = String(currentPage.value)
  if (pageSize.value !== 20) query.pageSize = String(pageSize.value)
  if (filterMethod.value) query.method = filterMethod.value
  router.replace({ query })
}

let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(searchKeyword, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    syncToUrl()
    fetchPermissions()
  }, 300)
})
```

修改 `handleSearch`、`handleMethodChange`、`handleSizeChange`、`handleCurrentChange` 添加 `syncToUrl()` 调用（与 Task 3 相同模式）。

修改 `onMounted` 先调用 `restoreFromUrl()`。

- [ ] **Step 2: 验证并提交**

```bash
git add frontend/src/views/system/PermissionView.vue
git commit -m "feat(frontend): PermissionView URL 状态同步"
```

---

### Task 5: URL 状态同步 — AuditLogView

**Files:**
- Modify: `frontend/src/views/audit/AuditLogView.vue`

- [ ] **Step 1: 添加 URL 状态同步**

在 `frontend/src/views/audit/AuditLogView.vue` 中，同步 `filterSystemId`、`page`、`pageSize` 到 URL。

```typescript
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

function restoreFromUrl() {
  const q = route.query
  if (q.page) currentPage.value = Number(q.page) || 1
  if (q.pageSize) pageSize.value = Number(q.pageSize) || 20
  if (q.systemId) filterSystemId.value = Number(q.systemId) || null
}

function syncToUrl() {
  const query: Record<string, string> = {}
  if (currentPage.value > 1) query.page = String(currentPage.value)
  if (pageSize.value !== 20) query.pageSize = String(pageSize.value)
  if (filterSystemId.value) query.systemId = String(filterSystemId.value)
  router.replace({ query })
}
```

修改 `handleFilter`、`handleSizeChange`、`handleCurrentChange` 添加 `syncToUrl()`。

修改 `onMounted` 先调用 `restoreFromUrl()`。

- [ ] **Step 2: 验证并提交**

```bash
git add frontend/src/views/audit/AuditLogView.vue
git commit -m "feat(frontend): AuditLogView URL 状态同步"
```

---

### Task 6: URL 状态同步 — RoleView 和 MenuView

**Files:**
- Modify: `frontend/src/views/system/RoleView.vue`
- Modify: `frontend/src/views/system/MenuView.vue`

- [ ] **Step 1: 为 RoleView 添加 URL 状态同步**

RoleView 使用客户端过滤（`computed`），只需同步 `searchKeyword` 到 URL（不涉及分页）。

在 `frontend/src/views/system/RoleView.vue` 中添加：

```typescript
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

function restoreFromUrl() {
  const q = route.query
  if (q.keyword) searchKeyword.value = String(q.keyword)
}

function syncToUrl() {
  const query: Record<string, string> = {}
  if (searchKeyword.value) query.keyword = searchKeyword.value
  router.replace({ query })
}

watch(searchKeyword, () => {
  syncToUrl()
})
```

修改 `onMounted` 先调用 `restoreFromUrl()`。

- [ ] **Step 2: 为 MenuView 添加 URL 状态同步**

MenuView 同样使用客户端过滤，只需同步 `searchKeyword`。

与 RoleView 相同的模式。

- [ ] **Step 3: 验证并提交**

```bash
git add frontend/src/views/system/RoleView.vue frontend/src/views/system/MenuView.vue
git commit -m "feat(frontend): RoleView 和 MenuView URL 状态同步"
```

---

### Task 7: 骨架屏统一

**Files:**
- Modify: `frontend/src/views/user/UserView.vue`
- Modify: `frontend/src/views/audit/AuditLogView.vue`
- Modify: `frontend/src/views/system/WebhookView.vue`
- Modify: `frontend/src/views/system/MemberView.vue`
- Modify: `frontend/src/views/system/MemberDetailDrawer.vue`

- [ ] **Step 1: 为 UserView 添加骨架屏**

在 `frontend/src/views/user/UserView.vue` 中，用 `el-skeleton` 包裹 `el-table`。

将 `<el-table :data="users" v-loading="loading" border stripe>` 替换为：

```html
<el-skeleton :loading="loading" animated :count="5">
  <template #template>
    <el-skeleton-item variant="text" style="width: 40%; height: 32px; margin-bottom: 16px;" />
    <div v-for="i in 5" :key="i" style="display: flex; gap: 16px; margin-bottom: 12px;">
      <el-skeleton-item variant="text" style="width: 3%;" />
      <el-skeleton-item variant="text" style="width: 8%;" />
      <el-skeleton-item variant="text" style="width: 12%;" />
      <el-skeleton-item variant="text" style="width: 10%;" />
      <el-skeleton-item variant="text" style="width: 18%;" />
      <el-skeleton-item variant="text" style="width: 12%;" />
      <el-skeleton-item variant="text" style="width: 8%;" />
      <el-skeleton-item variant="text" style="width: 8%;" />
      <el-skeleton-item variant="text" style="width: 18%;" />
      <el-skeleton-item variant="text" style="width: 15%;" />
    </div>
  </template>
  <template #default>
    <el-table :data="users" border stripe>
      <!-- 原有列保持不变 -->
    </el-table>
  </template>
</el-skeleton>
```

注意：移除 `el-table` 上的 `v-loading="loading"`（骨架屏已替代此功能）。

- [ ] **Step 2: 为 AuditLogView 添加骨架屏**

同上模式，用 `el-skeleton` 包裹 `el-table`。

- [ ] **Step 3: 为 WebhookView 添加骨架屏**

同上模式。

- [ ] **Step 4: 为 MemberView 添加骨架屏**

同上模式。

- [ ] **Step 5: 为 MemberDetailDrawer 的角色表格添加骨架屏**

在 `frontend/src/views/system/MemberDetailDrawer.vue` 中，为"已分配角色"表格添加骨架屏。

- [ ] **Step 6: 验证并提交**

```bash
git add frontend/src/views/
git commit -m "feat(frontend): 统一骨架屏加载状态"
```

---

### Task 8: 空状态优化

**Files:**
- Modify: `frontend/src/views/user/UserView.vue`
- Modify: `frontend/src/views/audit/AuditLogView.vue`
- Modify: `frontend/src/views/system/WebhookView.vue`
- Modify: `frontend/src/views/system/MemberView.vue`
- Modify: `frontend/src/views/system/MenuView.vue`

- [ ] **Step 1: 为 UserView 添加空状态**

在骨架屏的 `<template #default>` 中，添加条件判断：

```html
<template #default>
  <el-table v-if="users.length > 0" :data="users" border stripe>
    <!-- 原有列 -->
  </el-table>
  <el-empty v-else description="暂无用户">
    <el-button type="primary" @click="showCreateDialog">创建用户</el-button>
  </el-empty>
</template>
```

- [ ] **Step 2: 为其余页面添加空状态**

为 AuditLogView、WebhookView、MemberView、MenuView 添加类似的 `el-empty` 组件，每个页面的引导按钮不同：
- AuditLogView：无引导按钮（只读页面）
- WebhookView：引导按钮"添加 Webhook"
- MemberView：引导按钮"分配角色"
- MenuView：引导按钮"添加菜单"

- [ ] **Step 3: 验证并提交**

```bash
git add frontend/src/views/
git commit -m "feat(frontend): 统一空状态组件与引导操作"
```

---

### Task 9: 批量操作消息反馈优化

**Files:**
- Modify: `frontend/src/views/system/RoleView.vue`
- Modify: `frontend/src/views/system/PermissionView.vue`

- [ ] **Step 1: 修改 RoleView 的批量删除反馈**

在 `frontend/src/views/system/RoleView.vue` 中，将 `handleBatchDelete` 函数的 `ElMessage.success` 替换为 `ElNotification`。

添加 `ElNotification` 到 import：

```typescript
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
```

修改 `handleBatchDelete`（约第 278 行）：

```typescript
async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 个角色吗？`, '批量删除', { type: 'warning' })
    const results = await Promise.allSettled(selectedIds.value.map(id => deleteRole(systemId.value, id)))
    const successCount = results.filter(r => r.status === 'fulfilled').length
    const failCount = results.filter(r => r.status === 'rejected').length
    if (failCount === 0) {
      ElNotification.success({ title: '批量删除成功', message: `成功删除 ${successCount} 个角色` })
    } else {
      ElNotification.warning({ title: '批量删除完成', message: `成功 ${successCount} 个，失败 ${failCount} 个` })
    }
    selectedIds.value = []
    fetchRoles()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量删除失败')
  }
}
```

- [ ] **Step 2: 修改 PermissionView 的批量删除反馈**

同上模式，修改 `frontend/src/views/system/PermissionView.vue` 的 `handleBatchDelete`。

- [ ] **Step 3: 验证并提交**

```bash
git add frontend/src/views/system/RoleView.vue frontend/src/views/system/PermissionView.vue
git commit -m "feat(frontend): 批量操作使用 ElNotification 反馈"
```

---

## 第二阶段：共享基础设施

### Task 10: 创建 useUrlState composable

**Files:**
- Create: `frontend/src/composables/useUrlState.ts`

- [ ] **Step 1: 创建 composables 目录**

```bash
mkdir -p /Users/zqqzqq/05_github/grbac/frontend/src/composables
```

- [ ] **Step 2: 创建 useUrlState.ts**

```typescript
import { reactive, watch, onMounted, type UnwrapNestedRefs } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'

/**
 * 将响应式状态双向绑定到 URL query string。
 * - 状态变化 → URL 自动更新（防抖）
 * - URL 变化（浏览器前进/后退）→ 状态自动恢复
 * - 提供 reset() 清空所有筛选条件
 */
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

  // 从 URL query 恢复状态
  function restoreFromUrl(query: LocationQuery) {
    for (const key of Object.keys(defaults) as Array<keyof T>) {
      const val = query[key as string]
      if (val !== undefined && val !== null && val !== '') {
        const defaultVal = defaults[key]
        if (typeof defaultVal === 'number') {
          (state as any)[key] = Number(val) || defaultVal
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

  // 同步状态到 URL
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

  // 监听状态变化同步到 URL
  watch(
    () => ({ ...state }),
    () => syncToUrl(),
    { deep: true }
  )

  // 监听 URL 变化恢复状态（浏览器前进/后退）
  watch(
    () => route.query,
    (query) => restoreFromUrl(query),
    { deep: true }
  )

  // 初始化：从 URL 恢复
  onMounted(() => {
    restoreFromUrl(route.query)
  })

  // 重置为默认值
  function reset() {
    for (const key of Object.keys(defaults) as Array<keyof T>) {
      (state as any)[key] = defaults[key]
    }
    router.replace({ query: {} })
  }

  return { state, reset }
}
```

- [ ] **Step 3: 验证类型检查**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/composables/useUrlState.ts
git commit -m "feat(frontend): 创建 useUrlState composable"
```

---

### Task 11: 创建 useTable composable

**Files:**
- Create: `frontend/src/composables/useTable.ts`

- [ ] **Step 1: 创建 useTable.ts**

```typescript
import { ref, type Ref } from 'vue'
import { ElMessage } from 'element-plus'

/**
 * 统一表格的加载、分页、搜索、刷新逻辑。
 */
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
      const result = await fetchFn({
        page: page.value,
        page_size: pageSize.value,
      })
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

  return {
    data,
    total,
    loading,
    skeleton,
    page,
    pageSize,
    refresh,
    handleSizeChange,
    handleCurrentChange,
  }
}
```

- [ ] **Step 2: 验证并提交**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
git add frontend/src/composables/useTable.ts
git commit -m "feat(frontend): 创建 useTable composable"
```

---

### Task 12: 创建 useFormDialog composable

**Files:**
- Create: `frontend/src/composables/useFormDialog.ts`

- [ ] **Step 1: 创建 useFormDialog.ts**

```typescript
import { ref, reactive, watch, computed, type Ref, type UnwrapNestedRefs } from 'vue'
import { ElMessageBox } from 'element-plus'

/**
 * 统一创建/编辑对话框的状态管理。
 */
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

  // 创建默认值的深拷贝函数
  function getDefaultData(): T {
    return {} as T  // 调用方通过 open() 时传入初始数据
  }

  const formData = reactive<T>(getDefaultData()) as UnwrapNestedRefs<T>
  const snapshot = ref<string>('')
  const formChanged = computed(() => JSON.stringify(formData) !== snapshot.value)

  function open(row?: T & { id?: number }) {
    if (row && row.id) {
      // 编辑模式
      isEditing.value = true
      editingId.value = row.id
      const { id, ...rest } = row
      Object.assign(formData, JSON.parse(JSON.stringify(rest)))
    } else {
      // 创建模式
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
        // 用户取消
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

  return {
    visible,
    isEditing,
    formData,
    formChanged,
    submitting,
    editingId,
    open,
    close,
    submit,
    forceClose,
  }
}
```

- [ ] **Step 2: 验证并提交**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
git add frontend/src/composables/useFormDialog.ts
git commit -m "feat(frontend): 创建 useFormDialog composable"
```

---

### Task 13: 创建 useConfirmDelete composable

**Files:**
- Create: `frontend/src/composables/useConfirmDelete.ts`

- [ ] **Step 1: 创建 useConfirmDelete.ts**

```typescript
import { ref } from 'vue'
import { ElMessageBox, ElNotification, ElMessage } from 'element-plus'

/**
 * 统一删除操作的确认、loading、成功回调。
 */
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
      return // 用户取消
    }

    let success = 0
    let failed = 0

    for (let i = 0; i < ids.length; i++) {
      try {
        await deleteFn(ids[i])
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
```

- [ ] **Step 2: 验证并提交**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
git add frontend/src/composables/useConfirmDelete.ts
git commit -m "feat(frontend): 创建 useConfirmDelete composable"
```

---

### Task 14: 重构 UserView 使用 composables

**Files:**
- Modify: `frontend/src/views/user/UserView.vue`

- [ ] **Step 1: 使用 useUrlState 替代手动 URL 同步**

重写 `frontend/src/views/user/UserView.vue` 的 `<script setup>` 部分，使用 `useUrlState`、`useTable`、`useFormDialog`、`useConfirmDelete` 替代手动逻辑。

关键变更：
1. 移除手动的 `restoreFromUrl`/`syncToUrl`/`watch(searchUsername)` 逻辑，改用 `useUrlState`
2. 移除手动的 `loading`/`total`/`currentPage`/`pageSize`/`fetchUsers`/`handleSizeChange`/`handleCurrentChange`，改用 `useTable`
3. 移除手动的 `dialogVisible`/`isEditing`/`submitting`/`editingId`/`form`/`originalForm`/`formDirty`/`handleDialogClose`，改用 `useFormDialog`
4. 移除手动的 `handleDelete` 中的 `ElMessageBox.confirm`，改用 `useConfirmDelete`

```typescript
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getUsers, createUser, updateUser, deleteUser, updateUserStatus, updateUserSuperAdmin, resetUserPassword } from '@/api/user'
import { formatDate } from '@/utils/format'
import UserDetailDrawer from './UserDetailDrawer.vue'
import { useUrlState } from '@/composables/useUrlState'
import { useTable } from '@/composables/useTable'
import { useFormDialog } from '@/composables/useFormDialog'
import { useConfirmDelete } from '@/composables/useConfirmDelete'

interface User {
  id: number
  username: string
  chinese_name?: string
  email?: string
  phone?: string
  status: number
  is_super_admin: number
  created_at: string
}

// URL 状态
const { state: urlState, reset: resetUrl } = useUrlState({
  keyword: '',
  page: 1,
  pageSize: 20,
  superAdmin: '' as string,
})

// 表格
const { data: users, total, loading, skeleton, page, pageSize, refresh: fetchUsers, handleSizeChange, handleCurrentChange } = useTable<User>(
  () => {
    const params: any = { page: urlState.page, page_size: urlState.pageSize }
    if (urlState.keyword) params.keyword = urlState.keyword
    if (urlState.superAdmin !== '') params.is_super_admin = Number(urlState.superAdmin)
    return getUsers(params)
  },
  { defaultPageSize: 20 }
)

// 搜索关键字防抖同步
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(() => urlState.keyword, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    urlState.page = 1
    fetchUsers()
  }, 300)
})

watch(() => urlState.superAdmin, () => {
  urlState.page = 1
  fetchUsers()
})

watch(page, (p) => { urlState.page = p })
watch(pageSize, (s) => { urlState.pageSize = s })

// 表单对话框
const formRef = ref<FormInstance>()
const { visible: dialogVisible, isEditing, formData: form, submitting, editingId, open: openDialog, close: closeDialog, submit: handleSubmit, forceClose } = useFormDialog(
  async (data) => {
    if (isEditing.value && editingId.value) {
      await updateUser(editingId.value, {
        chinese_name: data.chinese_name,
        email: data.email,
        phone: data.phone
      })
      ElMessage.success('更新成功')
    } else {
      await createUser({
        username: data.username,
        chinese_name: data.chinese_name,
        password: data.password,
        email: data.email,
        phone: data.phone
      })
      ElMessage.success('创建成功')
    }
  },
  { onSuccess: () => fetchUsers() }
)

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 位', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

async function handleFormSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  await handleSubmit()
}

function showCreateDialog() {
  openDialog({ username: '', chinese_name: '', password: '', email: '', phone: '' } as any)
}

function showEditDialog(user: User) {
  openDialog({
    username: user.username,
    chinese_name: user.chinese_name || '',
    password: '',
    email: user.email || '',
    phone: user.phone || ''
  } as any)
}

// 删除
const { confirmDelete } = useConfirmDelete(
  (id) => deleteUser(id),
  { onSuccess: () => fetchUsers(), entityName: '用户' }
)

// 详情
const detailVisible = ref(false)
const selectedUser = ref<User | null>(null)

function showDetail(user: User) {
  selectedUser.value = user
  detailVisible.value = true
}

// 重置密码
const resetPasswordVisible = ref(false)
const resetSubmitting = ref(false)
const resetPasswordUser = ref<User | null>(null)
const resetFormRef = ref<FormInstance>()
const resetForm = ref({ newPassword: '' })
const resetRules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 位', trigger: 'blur' }
  ]
}

function showResetPasswordDialog(user: User) {
  resetPasswordUser.value = user
  resetForm.value = { newPassword: '' }
  resetPasswordVisible.value = true
}

async function handleResetPassword() {
  if (!resetFormRef.value || !resetPasswordUser.value) return
  await resetFormRef.value.validate(async (valid) => {
    if (!valid) return
    resetSubmitting.value = true
    try {
      await resetUserPassword(resetPasswordUser.value!.id, resetForm.value.newPassword)
      ElMessage.success('密码重置成功')
      resetPasswordVisible.value = false
    } catch (error: any) {
      ElMessage.error(error.message || '密码重置失败')
    } finally {
      resetSubmitting.value = false
    }
  })
}

// 状态切换
async function handleStatusChange(user: User) {
  try {
    await updateUserStatus(user.id, user.status)
    ElMessage.success('状态更新成功')
  } catch (error: any) {
    user.status = user.status === 1 ? 0 : 1
    ElMessage.error(error.message || '状态更新失败')
  }
}

async function handleSuperAdminChange(user: User) {
  try {
    await ElMessageBox.confirm(
      user.is_super_admin === 1
        ? `确定要将 "${user.username}" 设为超管吗？超管可以管理所有系统。`
        : `确定要取消 "${user.username}" 的超管权限吗？`,
      '确认操作',
      { type: 'warning' }
    )
    await updateUserSuperAdmin(user.id, user.is_super_admin)
    ElMessage.success(user.is_super_admin === 1 ? '已设为超管' : '已取消超管')
  } catch (error: any) {
    if (error !== 'cancel') {
      user.is_super_admin = user.is_super_admin === 1 ? 0 : 1
      ElMessage.error(error.message || '操作失败')
    } else {
      user.is_super_admin = user.is_super_admin === 1 ? 0 : 1
    }
  }
}

onMounted(() => {
  fetchUsers()
})
```

- [ ] **Step 2: 更新模板绑定**

确保模板中的 `v-model`、事件绑定与新的 composable 返回值对齐。主要变更：
- 搜索框 `v-model="urlState.keyword"`（替代 `searchUsername`）
- 超管筛选 `v-model="urlState.superAdmin"`（替代 `filterSuperAdmin`）
- 分页组件使用 `page`/`pageSize`（替代 `currentPage`/`pageSize`）
- 对话框使用 `:before-close="closeDialog"`
- 删除按钮 `@click="confirmDelete(row)"`

- [ ] **Step 3: 验证**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
```

手动测试所有功能：创建、编辑、删除、搜索、分页、状态切换、超管切换、重置密码、详情。

- [ ] **Step 4: 提交**

```bash
git add frontend/src/views/user/UserView.vue
git commit -m "refactor(frontend): UserView 使用 composables 重构"
```

---

### Task 15: 重构 RoleView 使用 composables

**Files:**
- Modify: `frontend/src/views/system/RoleView.vue`

- [ ] **Step 1: 重写 RoleView**

与 Task 14 相同的模式，但 RoleView 有以下特殊点：
- 使用客户端过滤（`computed` 的 `filteredRoles`），`useTable` 的 `fetchFn` 不带 keyword 参数
- `searchKeyword` 保持客户端过滤，同步到 URL
- 有批量删除功能，使用 `useConfirmDelete` 的 `batchDelete`
- 有 `RoleAssignDrawer` 的交互

```typescript
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole } from '@/api/role'
import RoleAssignDrawer from './RoleAssignDrawer.vue'
import { useUrlState } from '@/composables/useUrlState'
import { useFormDialog } from '@/composables/useFormDialog'
import { useConfirmDelete } from '@/composables/useConfirmDelete'

interface Role {
  id: number
  name: string
  code: string
  description?: string
  is_default: number
  created_at: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

// URL 状态
const { state: urlState } = useUrlState({ keyword: '' })

// 数据
const roles = ref<Role[]>([])
const loading = ref(false)
const selectedIds = ref<number[]>([])

const filteredRoles = computed(() => {
  const kw = urlState.keyword.toLowerCase().trim()
  if (!kw) return roles.value
  return roles.value.filter(r =>
    r.name.toLowerCase().includes(kw) ||
    r.code.toLowerCase().includes(kw) ||
    (r.description && r.description.toLowerCase().includes(kw))
  )
})

async function fetchRoles() {
  loading.value = true
  try {
    const data: any = await getRoles(systemId.value)
    roles.value = data.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// 表单对话框
const formRef = ref<FormInstance>()
const { visible: dialogVisible, isEditing, formData: form, submitting, open: openDialog, close: closeDialog, submit: handleSubmit, forceClose } = useFormDialog(
  async (data) => {
    if (isEditing.value) {
      await updateRole(systemId.value, data.id!, {
        name: data.name,
        description: data.description
      })
      ElMessage.success('更新成功')
    } else {
      await createRole(systemId.value, {
        name: data.name,
        code: data.code,
        description: data.description,
        is_default: data.is_default ? 1 : 0
      })
      ElMessage.success('创建成功')
    }
  },
  { onSuccess: () => fetchRoles() }
)

const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: '编码只能包含字母、数字、下划线和连字符，且以字母开头', trigger: 'blur' }
  ]
}

async function handleFormSubmit() {
  if (!formRef.value) return
  try { await formRef.value.validate() } catch { return }
  await handleSubmit()
}

function showCreateDialog() {
  openDialog({ name: '', code: '', description: '', is_default: false } as any)
}

function showEditDialog(role: Role) {
  openDialog({
    name: role.name,
    code: role.code,
    description: role.description || '',
    is_default: role.is_default === 1
  } as any)
}

// 删除
const { confirmDelete, batchDelete } = useConfirmDelete(
  (id) => deleteRole(systemId.value, id),
  { onSuccess: () => fetchRoles(), entityName: '角色' }
)

async function handleBatchDelete() {
  await batchDelete(selectedIds.value, {
    onComplete: () => { selectedIds.value = []; fetchRoles() }
  })
}

function handleSelectionChange(selection: Role[]) {
  selectedIds.value = selection.map(r => r.id)
}

// 授权抽屉
const drawerVisible = ref(false)
const selectedRole = ref<Role | null>(null)

function showAssignDrawer(role: Role) {
  selectedRole.value = role
  drawerVisible.value = true
}

onMounted(() => { fetchRoles() })
```

- [ ] **Step 2: 验证并提交**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
git add frontend/src/views/system/RoleView.vue
git commit -m "refactor(frontend): RoleView 使用 composables 重构"
```

---

### Task 16: 重构 PermissionView 使用 composables

**Files:**
- Modify: `frontend/src/views/system/PermissionView.vue`

- [ ] **Step 1: 重写 PermissionView**

与 Task 14 相同模式，额外处理 `filterMethod` 参数。

```typescript
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getPermissions, createPermission, updatePermission, deletePermission } from '@/api/permission'
import { formatDate } from '@/utils/format'
import { useUrlState } from '@/composables/useUrlState'
import { useTable } from '@/composables/useTable'
import { useFormDialog } from '@/composables/useFormDialog'
import { useConfirmDelete } from '@/composables/useConfirmDelete'

interface Permission {
  id: number
  code: string
  name: string
  method: string
  path: string
  description?: string
  created_at: string
}

const route = useRoute()
const systemId = computed(() => Number(route.params.id))

// URL 状态
const { state: urlState } = useUrlState({
  keyword: '',
  page: 1,
  pageSize: 20,
  method: '',
})

// 表格
const { data: permissions, total, loading, skeleton, page, pageSize, refresh: fetchPermissions, handleSizeChange, handleCurrentChange } = useTable<Permission>(
  () => {
    const params: any = { page: urlState.page, page_size: urlState.pageSize }
    if (urlState.keyword) params.keyword = urlState.keyword
    if (urlState.method) params.method = urlState.method
    return getPermissions(systemId.value, params)
  }
)

// 搜索防抖
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(() => urlState.keyword, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { urlState.page = 1; fetchPermissions() }, 300)
})

watch(() => urlState.method, () => { urlState.page = 1; fetchPermissions() })
watch(page, (p) => { urlState.page = p })
watch(pageSize, (s) => { urlState.pageSize = s })

function getMethodTagType(method: string) {
  const types: Record<string, string> = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger', PATCH: 'info' }
  return types[method] || ''
}

// 表单对话框
const formRef = ref<FormInstance>()
const { visible: dialogVisible, isEditing, formData: form, submitting, open: openDialog, close: closeDialog, submit: handleSubmit, forceClose } = useFormDialog(
  async (data) => {
    if (isEditing.value) {
      await updatePermission(systemId.value, data.id!, {
        name: data.name, method: data.method, path: data.path, description: data.description
      })
      ElMessage.success('更新成功')
    } else {
      await createPermission(systemId.value, {
        code: data.code, name: data.name, method: data.method, path: data.path, description: data.description
      })
      ElMessage.success('创建成功')
    }
  },
  { onSuccess: () => fetchPermissions() }
)

const rules: FormRules = {
  code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  path: [{ required: true, message: '请输入请求路径', trigger: 'blur' }]
}

async function handleFormSubmit() {
  if (!formRef.value) return
  try { await formRef.value.validate() } catch { return }
  await handleSubmit()
}

function showCreateDialog() {
  openDialog({ code: '', name: '', method: 'GET', path: '', description: '' } as any)
}

function showEditDialog(p: Permission) {
  openDialog({ code: p.code, name: p.name, method: p.method, path: p.path, description: p.description || '' } as any)
}

// 删除
const selectedIds = ref<number[]>([])
const { confirmDelete, batchDelete } = useConfirmDelete(
  (id) => deletePermission(systemId.value, id),
  { onSuccess: () => fetchPermissions(), entityName: '权限' }
)

async function handleBatchDelete() {
  await batchDelete(selectedIds.value, {
    onComplete: () => { selectedIds.value = []; fetchPermissions() }
  })
}

function handleSelectionChange(selection: Permission[]) {
  selectedIds.value = selection.map(p => p.id)
}

onMounted(() => { fetchPermissions() })
```

- [ ] **Step 2: 验证并提交**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit 2>&1 | head -20
git add frontend/src/views/system/PermissionView.vue
git commit -m "refactor(frontend): PermissionView 使用 composables 重构"
```

---

### Task 17: 全局错误处理增强

**Files:**
- Modify: `frontend/src/utils/request.ts`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: 在 request.ts 中添加网络错误统一处理**

在 `frontend/src/utils/request.ts` 的错误拦截器中，对网络超时和断网错误使用 `ElNotification`。

在文件顶部 import 中添加：

```typescript
import { ElNotification } from 'element-plus'
```

在错误拦截器的最后部分（约第 95-103 行），修改为：

```typescript
    if (error.response) {
      const { status, data } = error.response
      if (status === 403) {
        return Promise.reject(new Error(data?.message || '无权限访问'))
      }
      return Promise.reject(new Error(data?.message || `请求失败 (${status})`))
    }

    // 网络错误（超时、断网）
    if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
      ElNotification.error({ title: '请求超时', message: '网络连接超时，请检查网络后重试' })
      return Promise.reject(new Error('请求超时'))
    }
    if (!error.response && error.request) {
      ElNotification.error({ title: '网络错误', message: '无法连接到服务器，请检查网络连接' })
      return Promise.reject(new Error('网络连接失败'))
    }

    return Promise.reject(error)
```

- [ ] **Step 2: 在 App.vue 中添加全局错误捕获**

修改 `frontend/src/App.vue`：

```vue
<template>
  <router-view />
</template>

<script setup lang="ts">
import { onErrorCaptured } from 'vue'
import { ElNotification } from 'element-plus'

onErrorCaptured((err, instance, info) => {
  console.error('组件错误:', err, info)
  ElNotification.error({
    title: '页面错误',
    message: err.message || '发生了未知错误',
  })
  return false // 阻止错误向上传播
})
</script>
```

- [ ] **Step 3: 验证并提交**

```bash
git add frontend/src/utils/request.ts frontend/src/App.vue
git commit -m "feat(frontend): 全局错误处理增强 — 网络错误通知与组件错误捕获"
```

---

## 第三阶段：深层优化

### Task 18: 批量操作进度反馈

**Files:**
- Already handled by `useConfirmDelete` composable (Task 13)
- Modify: `frontend/src/views/system/RoleView.vue` (already done in Task 15)
- Modify: `frontend/src/views/system/PermissionView.vue` (already done in Task 16)

- [ ] **Step 1: 验证批量删除进度**

RoleView 和 PermissionView 的 `handleBatchDelete` 已在 Task 15/16 中使用 `useConfirmDelete` 的 `batchDelete` 方法，该方法已内置串行执行和 `ElNotification` 反馈。

验证：在 RoleView 中选中多个角色 → 批量删除 → 确认显示 Notification 而非 ElMessage。

- [ ] **Step 2: 为批量删除按钮添加进度状态**

在 `frontend/src/views/system/RoleView.vue` 的模板中，为批量删除按钮添加进度显示：

```html
<el-button v-if="selectedIds.length > 0" type="danger" @click="handleBatchDelete" :loading="batchDeleting">
  批量删除 ({{ selectedIds.length }})
</el-button>
```

在 script 中添加 `batchDeleting` ref 并在 `handleBatchDelete` 中使用：

```typescript
const batchDeleting = ref(false)

async function handleBatchDelete() {
  batchDeleting.value = true
  try {
    await batchDelete(selectedIds.value, {
      onComplete: () => { selectedIds.value = []; fetchRoles() }
    })
  } finally {
    batchDeleting.value = false
  }
}
```

对 PermissionView 做相同修改。

- [ ] **Step 3: 提交**

```bash
git add frontend/src/views/system/RoleView.vue frontend/src/views/system/PermissionView.vue
git commit -m "feat(frontend): 批量删除按钮进度状态"
```

---

### Task 19: 表格列排序

**Files:**
- Modify: `frontend/src/views/user/UserView.vue`
- Modify: `frontend/src/views/system/RoleView.vue`
- Modify: `frontend/src/views/system/PermissionView.vue`
- Modify: `frontend/src/views/audit/AuditLogView.vue`

- [ ] **Step 1: 为 UserView 添加列排序**

在 `frontend/src/views/user/UserView.vue` 的 `<el-table>` 上添加 `@sort-change` 事件处理。

在 script 中添加排序状态和处理函数：

```typescript
const sortProp = ref('')
const sortOrder = ref('')

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  sortProp.value = prop || ''
  sortOrder.value = order || ''
  // 客户端排序
  if (!prop || !order) {
    fetchUsers() // 恢复默认
    return
  }
  users.value.sort((a: any, b: any) => {
    const va = a[prop]
    const vb = b[prop]
    if (va == null) return 1
    if (vb == null) return -1
    const cmp = typeof va === 'string' ? va.localeCompare(vb) : va - vb
    return order === 'ascending' ? cmp : -cmp
  })
}
```

在 `el-table` 上添加 `@sort-change="handleSortChange"`。

为需要排序的列添加 `sortable="custom"`：

```html
<el-table-column prop="username" label="用户名" min-width="120" sortable="custom" />
<el-table-column prop="created_at" label="创建时间" min-width="180" sortable="custom">
```

- [ ] **Step 2: 为 RoleView 添加列排序**

RoleView 使用客户端过滤，排序也是客户端的。与 UserView 相同模式。

为 `name`、`code`、`created_at` 列添加 `sortable="custom"`。

- [ ] **Step 3: 为 PermissionView 添加列排序**

PermissionView 使用服务端分页，排序需要传给后端。但由于当前后端 API 可能不支持排序参数，先做客户端排序（对当前页数据排序）。

- [ ] **Step 4: 为 AuditLogView 添加列排序**

AuditLogView 的 `created_at` 列添加 `sortable="custom"`，默认降序。

- [ ] **Step 5: 验证并提交**

```bash
git add frontend/src/views/
git commit -m "feat(frontend): 表格列排序支持"
```

---

### Task 20: 智能默认值

**Files:**
- Modify: `frontend/src/views/system/PermissionView.vue`
- Modify: `frontend/src/views/system/RoleView.vue`
- Modify: `frontend/src/views/system/MenuView.vue`
- Modify: `frontend/src/views/system/WebhookView.vue`

- [ ] **Step 1: PermissionView — HTTP 方法默认 GET**

PermissionView 的 `showCreateDialog` 已经默认 `method: 'GET'`（见原代码第 229 行），无需修改。确认即可。

- [ ] **Step 2: RoleView — 状态默认启用**

RoleView 的创建表单没有 status 字段（角色只有 is_default），跳过。

- [ ] **Step 3: MenuView — 排序号自动填入最大值 + 10**

在 `frontend/src/views/system/MenuView.vue` 的 `showCreateDialog` 函数中，计算当前层级的最大排序号：

```typescript
function showCreateDialog(parentId?: number) {
  isEditing.value = false
  editingId.value = null

  // 计算当前层级的最大排序号
  let maxSort = 0
  const siblings = parentId
    ? menus.value.find(m => m.id === parentId)?.children || []
    : menus.value
  for (const item of siblings) {
    if ((item.sort_order || 0) > maxSort) maxSort = item.sort_order || 0
  }

  form.value = {
    parent_id: parentId ? [parentId] : null,
    name: '',
    path: '',
    icon: '',
    sort_order: maxSort + 10
  }
  dialogVisible.value = true
}
```

- [ ] **Step 4: WebhookView — 默认勾选常用事件**

在 `frontend/src/views/system/WebhookView.vue` 的 `showCreateDialog` 中：

```typescript
function showCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.value = { url: '', events: ['role.created', 'role.updated'] }
  dialogVisible.value = true
}
```

- [ ] **Step 5: 验证并提交**

```bash
git add frontend/src/views/system/MenuView.vue frontend/src/views/system/WebhookView.vue
git commit -m "feat(frontend): 智能默认值 — 菜单排序号自动填充、Webhook 默认事件"
```

---

### Task 21: 键盘快捷键

**Files:**
- Create: `frontend/src/composables/useKeyboard.ts`
- Modify: `frontend/src/layouts/DefaultLayout.vue`

- [ ] **Step 1: 创建 useKeyboard.ts**

```typescript
import { onMounted, onUnmounted } from 'vue'

/**
 * 全局键盘快捷键。
 * 仅在非输入框状态下生效。
 */
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
    // Ctrl+Enter: 提交当前表单
    if (e.ctrlKey && e.key === 'Enter') {
      const submitBtn = document.querySelector('.el-dialog:not([style*="display: none"]) .el-button--primary, .el-drawer:not([style*="display: none"]) .el-button--primary') as HTMLButtonElement | null
      if (submitBtn && !submitBtn.disabled) {
        submitBtn.click()
        e.preventDefault()
      }
      return
    }

    if (isInInput()) return

    // /: 聚焦搜索框
    if (e.key === '/') {
      e.preventDefault()
      const searchInput = document.querySelector('.search-bar .el-input__inner, .filter-bar .el-input__inner') as HTMLInputElement | null
      if (searchInput) {
        searchInput.focus()
      }
      return
    }

    // n: 新建
    if (e.key === 'n' && !e.ctrlKey && !e.metaKey && !e.altKey) {
      e.preventDefault()
      handlers.onNew?.()
      return
    }
  }

  onMounted(() => {
    document.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown)
  })
}
```

- [ ] **Step 2: 在 DefaultLayout.vue 中注册全局快捷键**

在 `frontend/src/layouts/DefaultLayout.vue` 的 `<script setup>` 中导入并使用：

```typescript
import { useKeyboard } from '@/composables/useKeyboard'

// 键盘快捷键
useKeyboard({
  onSearch: () => {
    const input = document.querySelector('.search-bar .el-input__inner, .filter-bar .el-input__inner') as HTMLInputElement | null
    input?.focus()
  },
})
```

- [ ] **Step 3: 验证**

在任意列表页面按 `/` → 确认搜索框聚焦。在非输入框状态按 `n` → 无报错（各页面的 `onNew` 需要在各页面单独注册，此处仅做全局基础）。

- [ ] **Step 4: 提交**

```bash
git add frontend/src/composables/useKeyboard.ts frontend/src/layouts/DefaultLayout.vue
git commit -m "feat(frontend): 全局键盘快捷键 composable"
```

---

### Task 22: 页面切换记忆 — keep-alive 与面包屑增强

**Files:**
- Modify: `frontend/src/layouts/DefaultLayout.vue`
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: 在 router 中为路由添加 name**

确保所有列表路由都有 `name` 属性（keep-alive 需要）。检查 `frontend/src/router/index.ts`，所有路由已有 `name`，无需修改。

- [ ] **Step 2: 在 DefaultLayout 中添加 keep-alive**

修改 `frontend/src/layouts/DefaultLayout.vue` 的模板中 `<router-view>` 部分：

```html
<el-main class="layout-main">
  <router-view v-slot="{ Component, route: currentRoute }">
    <keep-alive :include="cachedViews">
      <component :is="Component" :key="currentRoute.fullPath" />
    </keep-alive>
  </router-view>
</el-main>
```

在 script 中添加缓存视图管理：

```typescript
import { ref, computed } from 'vue'

// 缓存的列表页面视图
const cachedViews = ref<string[]>([
  'UserView',
  'RoleView',
  'MenuView',
  'PermissionView',
  'WebhookView',
  'MemberView',
  'AuditLogView',
  'SystemView',
])
```

注意：keep-alive 的 `include` 使用组件的 `name` 选项。需要为每个列表组件添加 `defineOptions({ name: 'UserView' })` 等。

- [ ] **Step 3: 为列表组件添加 name**

在以下每个文件的 `<script setup>` 顶部添加 `defineOptions`：

**UserView.vue:**
```typescript
defineOptions({ name: 'UserView' })
```

**RoleView.vue:**
```typescript
defineOptions({ name: 'RoleView' })
```

**MenuView.vue:**
```typescript
defineOptions({ name: 'MenuView' })
```

**PermissionView.vue:**
```typescript
defineOptions({ name: 'PermissionView' })
```

**WebhookView.vue:**
```typescript
defineOptions({ name: 'WebhookView' })
```

**MemberView.vue:**
```typescript
defineOptions({ name: 'MemberView' })
```

**AuditLogView.vue:**
```typescript
defineOptions({ name: 'AuditLogView' })
```

**SystemView.vue:**
```typescript
defineOptions({ name: 'SystemView' })
```

- [ ] **Step 4: 面包屑增强 — 所有层级可点击**

在 `frontend/src/layouts/DefaultLayout.vue` 的 `breadcrumbs` computed 中，确保每个层级都有 `path`：

当前代码（约第 166-190 行）已经为大部分层级提供了 `path`。需要检查系统名称层级：

```typescript
const breadcrumbs = computed(() => {
  const items: { title: string; path?: string }[] = [{ title: '首页', path: '/dashboard' }]
  const path = route.path

  const sysMatch = path.match(/^\/systems\/(\d+)(?:\/(.+))?$/)
  if (sysMatch) {
    items.push({ title: '系统管理', path: '/systems' })
    const sys = systemStore.systems.find(s => s.id === Number(sysMatch[1]))
    if (sys) {
      // 系统名称也指向系统角色管理页（系统首页）
      items.push({ title: sys.name, path: `/systems/${sysMatch[1]}/roles` })
    }
    if (sysMatch[2] && route.meta.title) {
      items.push({ title: route.meta.title as string })
    }
    return items
  }

  if (path === '/dashboard') return items
  if (route.meta.title) {
    items.push({ title: route.meta.title as string })
  }
  return items
})
```

- [ ] **Step 5: 验证**

从 UserView 点击"详情"进入 UserDetailDrawer → 关闭 Drawer → 确认 UserView 的搜索条件保留。面包屑中"系统管理"可点击返回系统列表，系统名称可点击进入角色管理。

- [ ] **Step 6: 提交**

```bash
git add frontend/src/layouts/DefaultLayout.vue frontend/src/views/
git commit -m "feat(frontend): 页面切换记忆 — keep-alive 缓存与面包屑增强"
```

---

### Task 23: 向导式多步骤操作 — RoleAssignDrawer 增强

**Files:**
- Modify: `frontend/src/views/system/RoleAssignDrawer.vue`
- Modify: `frontend/src/views/system/RoleView.vue`

- [ ] **Step 1: 在 RoleAssignDrawer 中添加向导模式**

在 `frontend/src/views/system/RoleAssignDrawer.vue` 中，添加一个"快速创建"模式，使用 `el-steps` 组件引导用户一步完成角色创建 + 用户/菜单/权限分配。

在 `<script setup>` 中添加向导状态：

```typescript
const wizardMode = ref(false)
const wizardStep = ref(0)
const wizardSubmitting = ref(false)

// 向导数据
const wizardForm = ref({
  name: '',
  code: '',
  description: '',
  is_default: false,
  userIds: [] as number[],
  menuIds: [] as number[],
  permissionIds: [] as number[],
})

function startWizard() {
  wizardMode.value = true
  wizardStep.value = 0
  wizardForm.value = { name: '', code: '', description: '', is_default: false, userIds: [], menuIds: [], permissionIds: [] }
  // 预加载数据
  fetchMenuTree()
  fetchPermissions()
}

function nextWizardStep() {
  if (wizardStep.value < 3) wizardStep.value++
}

function prevWizardStep() {
  if (wizardStep.value > 0) wizardStep.value--
}

async function submitWizard() {
  wizardSubmitting.value = true
  try {
    // Step 1: 创建角色
    const roleData: any = await createRole(props.systemId, {
      name: wizardForm.value.name,
      code: wizardForm.value.code,
      description: wizardForm.value.description,
      is_default: wizardForm.value.is_default ? 1 : 0,
    })
    const newRoleId = roleData.id

    // Step 2: 分配用户
    if (wizardForm.value.userIds.length > 0) {
      await assignUsersApi(props.systemId, newRoleId, wizardForm.value.userIds)
    }

    // Step 3: 分配菜单
    if (wizardForm.value.menuIds.length > 0) {
      await assignMenusApi(props.systemId, newRoleId, wizardForm.value.menuIds)
    }

    // Step 4: 分配权限
    if (wizardForm.value.permissionIds.length > 0) {
      await assignPermissionsApi(props.systemId, newRoleId, wizardForm.value.permissionIds)
    }

    ElNotification.success({ title: '创建成功', message: `角色 "${wizardForm.value.name}" 已创建并完成配置` })
    wizardMode.value = false
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElNotification.error({ title: '创建失败', message: error.message || '部分步骤失败，请检查后重试' })
  } finally {
    wizardSubmitting.value = false
  }
}
```

需要在顶部 import 中添加 `ElNotification` 和 `createRole`：

```typescript
import { ElMessage, ElNotification } from 'element-plus'
import { createRole } from '@/api/role'
```

- [ ] **Step 2: 添加向导模板**

在 `RoleAssignDrawer.vue` 的 `<el-drawer>` 内容顶部添加向导模式切换：

```html
<el-drawer
  v-model="visible"
  :title="wizardMode ? '快速创建角色' : `授权 - ${role?.name}`"
  size="70%"
  @close="handleClose"
>
  <!-- 向导模式 -->
  <template v-if="wizardMode">
    <el-steps :active="wizardStep" finish-status="success" style="margin-bottom: 24px;">
      <el-step title="基本信息" />
      <el-step title="分配成员" />
      <el-step title="分配菜单" />
      <el-step title="分配权限" />
    </el-steps>

    <!-- Step 0: 基本信息 -->
    <div v-show="wizardStep === 0">
      <el-form label-width="100px">
        <el-form-item label="角色名称" required>
          <el-input v-model="wizardForm.name" placeholder="请输入角色名称" autofocus />
        </el-form-item>
        <el-form-item label="角色编码" required>
          <el-input v-model="wizardForm.code" placeholder="请输入角色编码" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="wizardForm.description" type="textarea" :rows="3" placeholder="请输入角色描述" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 1: 分配成员 -->
    <div v-show="wizardStep === 1">
      <p style="color: var(--color-text-secondary); margin-bottom: 12px;">选择要分配给此角色的用户（可跳过）</p>
      <el-select
        v-model="wizardForm.userIds"
        multiple
        filterable
        remote
        :remote-method="searchUsers"
        :loading="searchingUsers"
        placeholder="搜索用户名"
        style="width: 100%"
      >
        <el-option
          v-for="user in availableUsers"
          :key="user.id"
          :label="user.chinese_name ? `${user.chinese_name}(${user.username})` : user.username"
          :value="user.id"
        />
      </el-select>
    </div>

    <!-- Step 2: 分配菜单 -->
    <div v-show="wizardStep === 2">
      <p style="color: var(--color-text-secondary); margin-bottom: 12px;">勾选此角色可访问的菜单（可跳过）</p>
      <el-tree
        ref="wizardMenuTreeRef"
        :data="menuTree"
        :props="{ label: 'name', children: 'children' }"
        show-checkbox
        node-key="id"
        v-loading="loadingMenus"
      />
    </div>

    <!-- Step 3: 分配权限 -->
    <div v-show="wizardStep === 3">
      <p style="color: var(--color-text-secondary); margin-bottom: 12px;">选择此角色的接口权限（可跳过）</p>
      <el-transfer
        v-model="wizardForm.permissionIds"
        :data="transferPermissions"
        :titles="['未分配权限', '已分配权限']"
        filterable
        filter-placeholder="搜索权限编码或名称"
        :props="{ key: 'id', label: 'label' }"
        v-loading="loadingPermissions"
      />
    </div>

    <!-- 向导按钮 -->
    <div style="display: flex; justify-content: space-between; margin-top: 24px;">
      <el-button v-if="wizardStep > 0" @click="prevWizardStep">上一步</el-button>
      <div v-else></div>
      <div>
        <el-button @click="wizardMode = false">取消</el-button>
        <el-button v-if="wizardStep < 3" type="primary" @click="nextWizardStep" :disabled="wizardStep === 0 && (!wizardForm.name || !wizardForm.code)">
          下一步
        </el-button>
        <el-button v-else type="primary" @click="submitWizard" :loading="wizardSubmitting">
          完成创建
        </el-button>
      </div>
    </div>
  </template>

  <!-- 原有的 Tab 模式 -->
  <template v-else>
    <el-tabs v-model="activeTab">
      <!-- 原有内容保持不变 -->
    </el-tabs>
  </template>
</el-drawer>
```

- [ ] **Step 3: 在 RoleView 中添加"快速创建"入口**

在 `frontend/src/views/system/RoleView.vue` 的模板中，添加"快速创建"按钮：

```html
<div class="header-actions">
  <el-button v-if="selectedIds.length > 0" type="danger" @click="handleBatchDelete" :loading="batchDeleting">
    批量删除 ({{ selectedIds.length }})
  </el-button>
  <el-button type="primary" @click="showCreateDialog">
    <el-icon><Plus /></el-icon>
    创建角色
  </el-button>
  <el-button type="success" @click="showWizardDrawer">
    <el-icon><Plus /></el-icon>
    快速创建
  </el-button>
</div>
```

在 script 中添加：

```typescript
function showWizardDrawer() {
  selectedRole.value = { id: 0, name: '', code: '' } // 临时占位
  drawerVisible.value = true
  // RoleAssignDrawer 内部通过 startWizard() 进入向导模式
}
```

在 RoleAssignDrawer 中，当 `role` 的 `id === 0` 时自动进入向导模式：

```typescript
watch(() => props.modelValue, (val) => {
  if (val && props.role) {
    if (props.role.id === 0) {
      startWizard()
    } else {
      loadAllData()
    }
  }
})
```

- [ ] **Step 4: 验证**

点击"快速创建" → 填写角色信息 → 下一步 → 选择用户 → 下一步 → 勾选菜单 → 下一步 → 选择权限 → 完成创建。确认 Notification 显示成功。

- [ ] **Step 5: 提交**

```bash
git add frontend/src/views/system/RoleAssignDrawer.vue frontend/src/views/system/RoleView.vue
git commit -m "feat(frontend): 向导式多步骤角色创建"
```

---

### Task 24: 表格列设置

**Files:**
- Create: `frontend/src/components/ColumnSettings.vue`
- Modify: `frontend/src/views/user/UserView.vue`
- Modify: `frontend/src/views/system/RoleView.vue`
- Modify: `frontend/src/views/system/PermissionView.vue`
- Modify: `frontend/src/views/audit/AuditLogView.vue`

- [ ] **Step 1: 创建 ColumnSettings 通用组件**

```bash
mkdir -p /Users/zqqzqq/05_github/grbac/frontend/src/components
```

创建 `frontend/src/components/ColumnSettings.vue`：

```vue
<template>
  <el-popover placement="bottom-end" :width="200" trigger="click">
    <template #reference>
      <el-button :icon="Setting" circle />
    </template>
    <div class="column-settings">
      <div class="column-settings-header">
        <span>列设置</span>
        <el-button type="primary" link @click="resetColumns">重置</el-button>
      </div>
      <el-checkbox-group v-model="visibleColumns" @change="handleChange">
        <div v-for="col in columns" :key="col.key" class="column-item">
          <el-checkbox :label="col.key">{{ col.label }}</el-checkbox>
        </div>
      </el-checkbox-group>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Setting } from '@element-plus/icons-vue'

interface Column {
  key: string
  label: string
}

const props = defineProps<{
  columns: Column[]
  storageKey: string
}>()

const emit = defineEmits<{
  (e: 'change', visibleKeys: string[]): void
}>()

const defaultKeys = props.columns.map(c => c.key)

function loadFromStorage(): string[] {
  try {
    const stored = localStorage.getItem(`grbac_columns_${props.storageKey}`)
    if (stored) return JSON.parse(stored)
  } catch {}
  return defaultKeys
}

const visibleColumns = ref<string[]>(loadFromStorage())

function handleChange(keys: string[]) {
  localStorage.setItem(`grbac_columns_${props.storageKey}`, JSON.stringify(keys))
  emit('change', keys)
}

function resetColumns() {
  visibleColumns.value = defaultKeys
  localStorage.removeItem(`grbac_columns_${props.storageKey}`)
  emit('change', defaultKeys)
}

// 初始化
emit('change', visibleColumns.value)
</script>

<style scoped>
.column-settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
}
.column-item {
  padding: 4px 0;
}
</style>
```

- [ ] **Step 2: 在 UserView 中集成列设置**

在 `frontend/src/views/user/UserView.vue` 的搜索栏右侧添加列设置按钮：

```html
<div class="search-bar">
  <div class="search-left">
    <el-input ... />
    <el-select ... />
  </div>
  <ColumnSettings
    :columns="allColumns"
    storage-key="user"
    @change="(keys) => visibleColumnKeys = keys"
  />
</div>
```

在 script 中定义列配置：

```typescript
import ColumnSettings from '@/components/ColumnSettings.vue'

const allColumns = [
  { key: 'id', label: 'ID' },
  { key: 'username', label: '用户名' },
  { key: 'chinese_name', label: '中文名' },
  { key: 'email', label: '邮箱' },
  { key: 'phone', label: '手机号' },
  { key: 'status', label: '状态' },
  { key: 'super_admin', label: '超管' },
  { key: 'created_at', label: '创建时间' },
]
const visibleColumnKeys = ref(allColumns.map(c => c.key))
```

在模板中用 `v-if` 控制列显示：

```html
<el-table-column v-if="visibleColumnKeys.includes('id')" prop="id" label="ID" width="80" />
<el-table-column v-if="visibleColumnKeys.includes('username')" prop="username" label="用户名" min-width="120" />
<!-- ... 其他列类似 -->
```

- [ ] **Step 3: 为 RoleView、PermissionView、AuditLogView 集成列设置**

每个页面重复 Step 2 的模式，使用不同的 `storage-key`（`role`、`permission`、`audit`）。

- [ ] **Step 4: 验证**

点击齿轮图标 → 取消勾选某列 → 确认表格隐藏该列 → 刷新页面 → 确认列设置保留。点击"重置" → 确认所有列恢复显示。

- [ ] **Step 5: 提交**

```bash
git add frontend/src/components/ColumnSettings.vue frontend/src/views/
git commit -m "feat(frontend): 表格列设置 — 显隐控制与 localStorage 持久化"
```

---

### Task 25: 最终验证与清理

**Files:**
- All modified files

- [ ] **Step 1: 类型检查**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npx vue-tsc --noEmit
```

确认无类型错误。

- [ ] **Step 2: 构建检查**

```bash
cd /Users/zqqzqq/05_github/grbac/frontend && npm run build
```

确认构建成功。

- [ ] **Step 3: 手动测试清单**

逐项验证所有验收标准：

**第一阶段：**
- [ ] 列表页刷新后筛选条件保留
- [ ] 对话框打开后自动聚焦第一个输入框
- [ ] 表单有修改时关闭对话框有确认提示
- [ ] 所有列表页有骨架屏加载状态
- [ ] 耗时请求显示顶部进度条
- [ ] 批量操作结果用 Notification 展示

**第二阶段：**
- [ ] useUrlState 在所有列表页生效
- [ ] useTable 替代各页面的手动 loading/data/total 管理
- [ ] useFormDialog 替代各页面的手动对话框状态管理
- [ ] useConfirmDelete 替代各页面的手动删除确认逻辑
- [ ] 网络错误全局统一处理

**第三阶段：**
- [ ] 批量删除有实时进度反馈
- [ ] 所有表格支持列排序
- [ ] 表单有智能默认值
- [ ] 全局键盘快捷键生效
- [ ] 面包屑所有层级可点击

- [ ] **Step 4: 提交最终状态**

```bash
git add -A
git commit -m "chore(frontend): 前端交互优化完成 — 最终验证通过"
```
