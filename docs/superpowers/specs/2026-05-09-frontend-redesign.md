# GRBAC 前端重新设计

## Context

后端 RBAC 系统已完成 30+ API 端点，涵盖用户管理、系统管理、角色/菜单/权限管理、Webhook、审计日志等。当前前端仅有骨架代码（登录 + 空仪表盘），需要根据后端 API 全面重建前端管理界面。

核心目标：让管理员能通过清晰的交互逻辑完成"给角色配权限、给用户配角色"的授权流程。

## 技术栈

- Vue 3 + TypeScript + Vite
- Element Plus + @element-plus/icons-vue
- Pinia (状态管理)
- Vue Router (路由)
- Axios (HTTP)

## 设计决策

1. **系统管理布局**: 左侧系统列表 + 右侧 Tab 切换（角色/菜单/接口权限/Webhook）
2. **角色授权交互**: 点击"授权"→ Drawer 弹出 → 内部 3 个 Tab（分配用户/分配菜单/分配接口权限）
3. **系统上下文**: 顶部全局系统选择器，影响角色/菜单/接口/Webhook 的数据范围
4. **辅助功能**: 审计日志 + 修改密码 + Webhook 管理

---

## 模块总览

| 模块 | 路由 | 权限 | 说明 |
|---|---|---|---|
| 登录 | `/login` | 公开 | 已有，保持不变 |
| 首页 | `/dashboard` | 已登录 | 简要概览 |
| 用户管理 | `/users` | super-admin | CRUD 用户、启停状态 |
| 系统管理 | `/systems` | super-admin | CRUD 系统、管理系统成员 |
| 角色管理 | `/systems/:id/roles` | system-admin (Tab 内) | CRUD 角色 + 授权 Drawer |
| 菜单管理 | `/systems/:id/menus` | system-admin (Tab 内) | 树形菜单 CRUD |
| 接口权限 | `/systems/:id/permissions` | system-admin (Tab 内) | API 权限 CRUD |
| Webhook | `/systems/:id/webhooks` | super-admin (Tab 内) | Webhook 订阅管理 |
| 审计日志 | `/audit-logs` | super-admin | 操作日志查询 |
| 修改密码 | `/profile/password` | 已登录 | 修改当前用户密码 |

---

## 1. 布局结构

### DefaultLayout（改造现有）

```
┌──────────────────────────────────────────────────────┐
│  Header: [折叠] GRBAC    系统: [用户中心 ▼]  [用户 ▼] │
├──────────┬───────────────────────────────────────────┤
│ Sidebar  │  Content (router-view)                    │
│          │                                           │
│ 首页     │                                           │
│ 用户管理  │                                           │
│ 系统管理  │                                           │
│ 审计日志  │                                           │
│ 修改密码  │                                           │
└──────────┴───────────────────────────────────────────┘
```

**Header 改造：**
- 左侧：折叠按钮 + 平台名称 "GRBAC"
- 中间：全局系统选择器 `el-select`（从 `/api/systems` 加载系统列表）
- 右侧：用户下拉菜单（修改密码、退出登录）

**系统选择器逻辑：**
- 存储在 `systemStore`（新建）中：`currentSystemId`, `systems`
- 影响范围：角色/菜单/接口权限/Webhook 页面自动使用当前系统 ID
- 不影响：用户管理、审计日志（super-admin 全局视图）
- 切换系统时，如果当前在系统级页面，自动刷新数据

**Sidebar 改造：**
- 根据用户角色动态显示菜单项
- super-admin 看到：用户管理、系统管理、审计日志
- system-admin 看到：系统管理（进入系统后管理角色/菜单/权限）
- 所有用户看到：首页、修改密码

---

## 2. 用户管理 `/users`

**页面结构：** 标准 CRUD 表格页

**功能：**
- 用户列表：分页表格，显示 username / email / phone / status / created_at
- 创建用户：Dialog 弹窗，字段：username, password, email, phone
- 编辑用户：Dialog 弹窗，字段：email, phone
- 删除用户：确认弹窗后删除
- 启用/禁用：Switch 开关，调用 `PUT /api/users/:id/status`
- 搜索：按用户名搜索（前端筛选或后端支持）

**API 调用：**
- `GET /api/users?page=&page_size=` → 列表
- `POST /api/users` → 创建
- `PUT /api/users/:id` → 编辑
- `DELETE /api/users/:id` → 删除
- `PUT /api/users/:id/status` → 启停

---

## 3. 系统管理 `/systems`

**页面结构：** 左侧系统列表 + 右侧 Tab 内容区

### 3.1 系统列表（左侧）

- 系统列表：`el-menu` 或卡片列表，显示系统名 + 描述
- 创建系统：Dialog 弹窗，字段：name, code, description
- 选中系统：点击后右侧显示该系统的 Tab 内容
- 删除系统：长按或右键菜单，确认后级联删除

### 3.2 系统成员（独立 Tab 或 Dialog）

- 系统成员列表：显示 user_id → username + role (admin/member)
- 添加成员：Dialog，选择用户 + 角色(admin/member)
- 移除成员：确认后删除

**API 调用：**
- `GET /api/systems?page=&page_size=` → 列表
- `POST /api/systems` → 创建
- `PUT /api/systems/:id` → 编辑
- `DELETE /api/systems/:id` → 删除
- `GET /api/systems/:id/members` → 成员列表
- `POST /api/systems/:id/members` → 添加成员
- `DELETE /api/systems/:id/members/:uid` → 移除成员

---

## 4. 角色管理（系统管理 Tab 内）

**页面结构：** 角色表格 + 授权 Drawer

### 角色列表

- 表格：name / code / description / 用户数 / 操作
- 创建角色：Dialog，字段：name, code, description
- 编辑角色：Dialog，字段：name, description
- 删除角色：确认后删除（需先移除所有用户绑定）
- **授权按钮**：打开 Drawer

### 授权 Drawer

Drawer 标题："角色授权: {角色名}"

**Tab 1: 分配用户**
- 左侧：已分配用户列表（可移除）
- 右侧：可选用户列表（可添加）
- 使用 `el-transfer` 或双列表格 + 勾选
- 保存调用 `POST /api/systems/:id/roles/:rid/users`（全量替换）

**Tab 2: 分配菜单**
- 树形勾选控件：`el-tree` + `show-checkbox`
- 数据来源：`GET /api/systems/:id/menus`（树形结构）
- 已分配：`GET /api/systems/:id/roles/:rid/menus`（menu_id 数组）
- 保存调用 `POST /api/systems/:id/roles/:rid/menus`（全量替换 menu_ids）

**Tab 3: 分配接口权限**
- 表格勾选：`el-table` + selection 列
- 列：method / path / code / name / description
- 数据来源：`GET /api/systems/:id/permissions`
- 已分配：`GET /api/systems/:id/roles/:rid/permissions`（permission_id 数组）
- 保存调用 `POST /api/systems/:id/roles/:rid/permissions`（全量替换 permission_ids）

**API 调用：**
- `GET /api/systems/:id/roles` → 角色列表
- `POST /api/systems/:id/roles` → 创建
- `PUT /api/systems/:id/roles/:rid` → 编辑
- `DELETE /api/systems/:id/roles/:rid` → 删除
- `GET /api/systems/:id/roles/:rid/menus` → 已分配菜单 ID
- `POST /api/systems/:id/roles/:rid/menus` → 分配菜单
- `GET /api/systems/:id/roles/:rid/permissions` → 已分配权限 ID
- `POST /api/systems/:id/roles/:rid/permissions` → 分配权限
- `POST /api/systems/:id/roles/:rid/users` → 分配用户（全量替换）
- `DELETE /api/systems/:id/roles/:rid/users/:uid` → 移除用户

> **注意：** 后端缺少 `GET /api/systems/:id/roles/:rid/users` 接口（获取角色下的用户列表）。需要新增此接口，或在前端通过 `GET /api/systems/:id/members` 间接获取。

---

## 5. 菜单管理（系统管理 Tab 内）

**页面结构：** 树形表格

**功能：**
- 菜单树：`el-table` + `row-key="id"` + `tree-props` 展示树形结构
- 列：name / path / icon / sort_order / status / 操作
- 创建菜单：Dialog，字段：parent_id(级联选择), name, path, icon, sort_order
- 编辑菜单：Dialog，同上
- 删除菜单：确认后删除（有子菜单时提示先删子菜单）

**API 调用：**
- `GET /api/systems/:id/menus` → 树形菜单
- `POST /api/systems/:id/menus` → 创建
- `PUT /api/systems/:id/menus/:mid` → 编辑
- `DELETE /api/systems/:id/menus/:mid` → 删除

---

## 6. 接口权限管理（系统管理 Tab 内）

**页面结构：** 分页表格

**功能：**
- 权限列表：分页表格，列：code / name / method / path / description / 操作
- 创建权限：Dialog，字段：code, name, method(GET/POST/PUT/DELETE), path, description
- 编辑权限：Dialog，同上
- 删除权限：确认后删除（被角色使用时提示）
- 搜索：按 code / name / path 搜索

**API 调用：**
- `GET /api/systems/:id/permissions?page=&page_size=` → 列表
- `POST /api/systems/:id/permissions` → 创建
- `PUT /api/systems/:id/permissions/:pid` → 编辑
- `DELETE /api/systems/:id/permissions/:pid` → 删除

---

## 7. Webhook 管理（系统管理 Tab 内）

**页面结构：** 表格

**功能：**
- Webhook 列表：列：url / events / status / created_at / 操作
- 创建 Webhook：Dialog，字段：url, events(多选 checkbox: permission_change / menu_change / role_change)
- 编辑 Webhook：Dialog，同上 + status 开关
- 删除 Webhook：确认后删除

**API 调用：**
- `GET /api/systems/:id/webhooks` → 列表
- `POST /api/systems/:id/webhooks` → 创建
- `PUT /api/systems/:id/webhooks/:wid` → 编辑
- `DELETE /api/systems/:id/webhooks/:wid` → 删除

---

## 8. 审计日志 `/audit-logs`

**页面结构：** 分页表格 + 筛选

**功能：**
- 日志列表：分页表格，列：username / action / resource / resource_name / ip / created_at
- 筛选：按系统筛选（el-select），按操作类型筛选
- 详情：点击行展开或 Dialog 显示 JSON detail

**API 调用：**
- `GET /api/audit-logs?page=&page_size=&system_id=` → 列表

---

## 9. 修改密码

**页面结构：** 简单表单

**功能：**
- 表单：旧密码、新密码、确认新密码
- 前端校验：新密码最少 8 位，两次输入一致
- 提交调用 `POST /api/auth/change-password`

---

## 10. 首页 Dashboard

**改造为有意义的概览：**
- 欢迎信息 + 当前用户角色
- 如果是 super-admin：显示系统总数、用户总数、最近审计日志
- 如果是 system-admin：显示当前系统的角色数、菜单数、权限数

---

## Store 设计

| Store | 文件 | 职责 |
|---|---|---|
| authStore | `stores/auth.ts` | 登录/登出/刷新 token（已有，需修复 refreshToken bug） |
| userStore | `stores/user.ts` | 当前用户信息（已有） |
| systemStore | `stores/system.ts` | **新建** — 系统列表、当前选中系统、系统切换 |
| menuStore | `stores/menu.ts` | 当前系统的菜单树（已有但未使用，需改造） |

---

## API Service 层

新增 API 模块（`api/` 目录）：

| 文件 | 职责 |
|---|---|
| `api/auth.ts` | login / logout / refresh / changePassword |
| `api/user.ts` | CRUD users + status |
| `api/system.ts` | CRUD systems + members |
| `api/role.ts` | CRUD roles + assign menus/permissions/users |
| `api/menu.ts` | CRUD menus (tree) |
| `api/permission.ts` | CRUD permissions |
| `api/webhook.ts` | CRUD webhooks |
| `api/audit.ts` | list audit logs |

---

## 清理

删除前端脚手架残留文件：
- `components/HelloWorld.vue`
- `components/TheWelcome.vue`
- `components/WelcomeItem.vue`
- `components/icons/` 整个目录
- `assets/logo.svg`
- 更新 `index.html` 标题为 "GRBAC"

---

## 验证

1. `npm run build` 编译通过，无 TS 错误
2. `npm run dev` 启动开发服务器
3. 登录流程正常（admin / admin123）
4. 系统选择器加载并切换
5. 用户管理 CRUD 正常
6. 系统管理 + 角色/菜单/权限 Tab 切换正常
7. 角色授权 Drawer：分配用户/菜单/接口权限正常
8. 审计日志查询正常
9. 修改密码正常
