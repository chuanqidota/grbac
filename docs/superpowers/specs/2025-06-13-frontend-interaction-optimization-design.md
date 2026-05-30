# 前端交互优化设计方案

## 概述

对 grbac 前端进行全面的交互体验优化，采用"快速胜利 + 系统化 + 深层优化"的混合方案，分三阶段交付。

**目标**：解决四大痛点——多步骤操作繁琐、数据查找低效、反馈不够及时、状态不持久。

**技术栈**：Vue 3 + TypeScript + Element Plus + Pinia

**预估工期**：9-14 天，分三批交付。

---

## 第一阶段：快速胜利（1-2 天）

低成本、高收益、全页面覆盖的改进，不涉及架构重构。

### 1.1 URL 状态同步

**现状**：搜索条件、分页、筛选器状态在刷新页面后丢失。

**方案**：将筛选参数同步到 URL query string，页面加载时从 URL 恢复状态。

**涉及参数**：
- 所有页面：`keyword`、`page`、`pageSize`
- PermissionView：`method`（HTTP 方法筛选）
- AuditLogView：`systemId`（系统筛选）

**交互细节**：
- 输入搜索关键字时防抖更新 URL（300ms）
- 切换页码/页大小时立即更新 URL
- 使用 `router.replace` 而非 `router.push`，不产生浏览器历史记录
- 分享 URL 即可复现相同的筛选视图

**涉及页面**：UserView、PermissionView、AuditLogView、RoleView、MenuView

### 1.2 表单体验增强

**现状**：打开对话框后需要手动点击输入框；关闭对话框无确认。

**改进**：
- 对话框打开后自动聚焦第一个输入框（`el-input` 添加 `autofocus` 属性）
- 表单有修改时关闭对话框，弹出 `ElMessageBox.confirm` 提示"确认放弃修改？"
- Enter 键提交表单（Element Plus 的 `el-form` 已支持，确认无冲突）
- Esc 键关闭对话框（Element Plus 默认支持）

**涉及页面**：所有使用 `el-dialog` + `el-form` 的页面（UserView、SystemView、RoleView、MenuView、PermissionView、WebhookView、MemberView、ChangePasswordView）

### 1.3 骨架屏与空状态统一

**现状**：部分页面有骨架屏（SystemView、RoleView、PermissionView），部分没有。

**改进**：
- 为缺失的页面添加 `el-skeleton`：UserView、MemberView、WebhookView、AuditLogView、MenuView、MemberDetailDrawer
- 空数据时展示 `el-empty` 组件，带引导操作按钮（如"暂无角色，点击创建"）
- Drawer 内的加载也统一使用骨架屏

### 1.4 全局加载指示器

**现状**：各页面独立管理 loading 状态，无全局反馈。

**方案**：
- 安装 `nprogress` 包
- 在 Axios 请求拦截器中启动 NProgress，响应拦截器中结束
- 仅对耗时 >300ms 的请求显示（通过 setTimeout 延迟启动）
- 配置为极简样式（无 spinner，仅顶部细线）
- 不替代页面内的 `v-loading`，作为补充

### 1.5 消息反馈优化

**现状**：批量操作结果用 `ElMessage`，容易被忽略。

**改进**：
- 批量操作完成使用 `ElNotification` 替代 `ElMessage`（保留更久，支持手动关闭）
- 成功/失败分别用不同通知类型
- 失败时列出具体失败项名称

---

## 第二阶段：共享基础设施（3-5 天）

构建可复用的 composables，统一各页面的重复模式。

### 2.1 `useUrlState` composable

**职责**：将响应式状态双向绑定到 URL query string。

**接口设计**：
```typescript
function useUrlState<T extends Record<string, any>>(
  defaults: T,
  options?: { debounce?: number }
): {
  state: UnwrapNestedRefs<T>  // reactive 对象，直接解构使用
  reset: () => void
}
```

**核心能力**：
- 类型安全的 query 序列化/反序列化（数字、字符串、布尔）
- 防抖同步（默认 300ms，可配置）
- 与 Vue Router 的 `router.replace` 集成
- 提供 `reset()` 方法清空所有筛选条件
- 浏览器前进/后退时自动恢复状态（监听 `router.afterEach`）

**文件位置**：`frontend/src/composables/useUrlState.ts`

### 2.2 `useTable` composable

**职责**：统一表格的加载、分页、搜索、刷新逻辑。

**接口设计**：
```typescript
function useTable<T>(
  fetchFn: (params: { page: number; pageSize: number; [key: string]: any }) => Promise<{ list: T[]; total: number }>,
  options?: { urlState?: Ref<any> }
): {
  data: Ref<T[]>
  total: Ref<number>
  loading: Ref<boolean>
  skeleton: Ref<boolean>
  pagination: { page: number; pageSize: number }
  search: () => void
  refresh: () => void
  handleSizeChange: (size: number) => void
  handleCurrentChange: (page: number) => void
}
```

**核心能力**：
- 封装 `loading`、`skeleton`、`data`、`total` 四个 ref
- 封装分页切换逻辑（当前每个页面重复实现）
- 提供 `refresh()` 方法（当前每个页面各自写 `fetchXxx()`）
- 错误统一处理（`ElMessage.error`），调用方可覆盖
- 骨架屏在首次加载时显示，后续刷新用 `v-loading`

**文件位置**：`frontend/src/composables/useTable.ts`

### 2.3 `useFormDialog` composable

**职责**：统一创建/编辑对话框的状态管理。

**接口设计**：
```typescript
function useFormDialog<T extends Record<string, any>>(
  submitFn: (data: T) => Promise<any>,
  options?: { onSuccess?: () => void }
): {
  visible: Ref<boolean>
  isEditing: Ref<boolean>
  formData: UnwrapNestedRefs<T>
  formChanged: Ref<boolean>
  submitting: Ref<boolean>
  open: (row?: T) => void
  close: () => void
  submit: () => Promise<void>
}
```

**核心能力**：
- `open(row?)` 传入行数据为编辑模式，不传为创建模式
- 自动深拷贝行数据，取消时不污染原数据
- `submitting` ref 自动管理
- `formChanged` ref 追踪表单是否被修改（dirty 检测）
- `close()` 有修改时触发确认对话框

**文件位置**：`frontend/src/composables/useFormDialog.ts`

### 2.4 `useConfirmDelete` composable

**职责**：统一删除操作的确认、loading、成功回调。

**接口设计**：
```typescript
function useConfirmDelete(
  deleteFn: (id: number) => Promise<any>,
  options?: { onSuccess?: () => void }
): {
  confirmDelete: (row: { id: number; name?: string }) => Promise<void>
  batchDelete: (ids: number[], options?: {
    onProgress?: (done: number, total: number) => void
    onComplete?: (success: number, failed: number) => void
  }) => Promise<void>
}
```

**核心能力**：
- 统一的 `ElMessageBox.confirm` 调用
- 单个删除：确认 → loading → 成功提示 → 刷新
- 批量删除：确认 → 串行执行（带进度回调） → 汇总结果 → 通知
- 批量删除失败时显示"已完成 N/M，失败 X 个"

**文件位置**：`frontend/src/composables/useConfirmDelete.ts`

### 2.5 全局错误处理增强

**现状**：每个 view 的 catch 块都写 `ElMessage.error(error.message || '...')`。

**改进**：
- 在 `request.ts` 响应拦截器中统一处理网络错误（超时、断网、500）
- 网络错误显示 `ElNotification` 带重试提示
- 业务错误（code !== 0）仍由拦截器 reject，view 层可做差异化处理
- 在 `App.vue` 添加 `onErrorCaptured` 捕获未处理的组件错误

**文件位置**：修改 `frontend/src/utils/request.ts`、`frontend/src/App.vue`

---

## 第三阶段：深层优化（5-7 天）

解决结构性交互问题，需要较大改动。

### 3.1 向导式多步骤操作

**痛点**：创建角色后需要分别去三个 Tab 分配用户、菜单、权限，来回切换。

**方案**：RoleAssignDrawer 增强为可选的向导模式。

**流程**：
```
Step 1: 填写角色名称、描述、状态
Step 2: 选择成员（远程搜索多选，可跳过）
Step 3: 勾选菜单权限（树形复选框，可跳过）
Step 4: 勾选 API 权限（穿梭框，可跳过）
→ 完成
```

**交互细节**：
- 使用 `el-steps` 组件展示进度
- 每步可跳过，最后统一提交
- 提交时按顺序调用：创建角色 → 分配用户 → 分配菜单 → 分配权限
- 任一步失败显示具体错误，已完成步骤不回滚

**涉及页面**：RoleView（RoleAssignDrawer）

### 3.2 批量操作进度反馈

**痛点**：批量删除用 `Promise.all` 并行执行，无进度反馈。

**方案**：改为串行执行 + 实时进度。

**交互**：
- 点击"批量删除 (N)" → 确认对话框显示"将删除 N 个 XXX"
- 确认后，按钮变为进度状态：`删除中 3/10...`
- 串行执行，每完成一个更新进度
- 完成后显示 `ElNotification`：成功 N 个，失败 M 个

**涉及页面**：RoleView、PermissionView（通过 `useConfirmDelete` 的 `batchDelete` 方法）

### 3.3 表格列排序与列设置

**痛点**：列表没有排序功能，找数据靠翻页。

**方案**：
- 为所有表格添加 `sortable="custom"` + `@sort-change` 事件
- 排序状态同步到 URL（通过 useUrlState）
- 添加"列设置"按钮（齿轮图标），使用 `el-popover` + `el-checkbox-group` 控制列显隐
- 列设置保存到 `localStorage`，按页面 key 存储

**默认排序策略**：
- 用户列表：创建时间倒序
- 角色列表：名称字母序
- 权限列表：HTTP 方法分组
- 审计日志：时间倒序

**涉及页面**：UserView、RoleView、PermissionView、AuditLogView、WebhookView、MemberView

### 3.4 智能默认值

**痛点**：创建表单时重复输入相同内容。

**方案**：
- 权限创建：HTTP 方法默认 `GET`，路径输入框根据已有同前缀权限自动建议
- 角色创建：状态默认 `启用`
- 菜单创建：排序号自动填入"当前层级最大值 + 10"（留间隔便于插入）
- Webhook 创建：事件类型默认勾选 `role.created, role.updated`

**涉及页面**：PermissionView、RoleView、MenuView、WebhookView

### 3.5 键盘快捷键

**痛点**：频繁使用鼠标效率低。

**方案**：全局快捷键，仅在非输入框时生效。

| 快捷键 | 功能 | 生效条件 |
|--------|------|----------|
| `/` | 聚焦当前页面搜索框 | 非输入框状态 |
| `n` | 打开新建对话框 | 非输入框状态，当前页面有且仅有一个新建入口 |
| `Esc` | 关闭当前最上层 Drawer/Dialog | 有打开的 Drawer/Dialog |
| `Ctrl+Enter` | 提交当前表单 | 有打开的表单 Dialog/Drawer |

**实现**：
- 新建 `frontend/src/composables/useKeyboard.ts`
- 在 `DefaultLayout.vue` 注册全局 `keydown` 监听
- 通过 `document.activeElement` 判断是否在输入框中
- 各页面声明自己支持的快捷键

**涉及页面**：全局 + 各页面声明

### 3.6 页面切换记忆

**痛点**：从列表进入详情后返回，筛选条件和页码丢失。

**方案**：
- URL 状态同步（useUrlState）天然解决刷新后状态恢复
- 使用 Vue Router 的 `keep-alive` 缓存已访问的列表页面组件
- 面包屑增强：所有层级都可点击返回（当前只有最后一级可点击）

**涉及页面**：DefaultLayout（keep-alive + 面包屑）、所有列表页

---

## 新增文件清单

```
frontend/src/composables/
  useUrlState.ts          # URL 状态同步
  useTable.ts             # 表格数据管理
  useFormDialog.ts        # 表单对话框管理
  useConfirmDelete.ts     # 删除确认管理
  useKeyboard.ts          # 键盘快捷键
```

## 修改文件清单

```
frontend/src/utils/request.ts       # NProgress + 全局错误处理
frontend/src/App.vue                # onErrorCaptured
frontend/src/layouts/DefaultLayout.vue  # keep-alive + 面包屑增强 + 快捷键
frontend/src/views/user/UserView.vue
frontend/src/views/system/SystemView.vue
frontend/src/views/system/RoleView.vue
frontend/src/views/system/MenuView.vue
frontend/src/views/system/PermissionView.vue
frontend/src/views/system/WebhookView.vue
frontend/src/views/system/MemberView.vue
frontend/src/views/system/MemberDetailDrawer.vue
frontend/src/views/system/RoleAssignDrawer.vue
frontend/src/views/audit/AuditLogView.vue
```

## 依赖变更

```json
{
  "dependencies": {
    "nprogress": "^0.2.0"
  },
  "devDependencies": {
    "@types/nprogress": "^0.2.3"
  }
}
```

## 验收标准

### 第一阶段
- [ ] 列表页刷新后筛选条件保留
- [ ] 对话框打开后自动聚焦第一个输入框
- [ ] 表单有修改时关闭对话框有确认提示
- [ ] 所有列表页有骨架屏加载状态
- [ ] 耗时请求显示顶部进度条
- [ ] 批量操作结果用 Notification 展示

### 第二阶段
- [ ] useUrlState 在所有列表页生效
- [ ] useTable 替代各页面的手动 loading/data/total 管理
- [ ] useFormDialog 替代各页面的手动对话框状态管理
- [ ] useConfirmDelete 替代各页面的手动删除确认逻辑
- [ ] 网络错误全局统一处理

### 第三阶段
- [ ] 创建角色可通过向导一步完成用户/菜单/权限分配
- [ ] 批量删除有实时进度反馈
- [ ] 所有表格支持列排序
- [ ] 列设置可保存到 localStorage
- [ ] 表单有智能默认值
- [ ] 全局键盘快捷键生效
- [ ] 面包屑所有层级可点击
