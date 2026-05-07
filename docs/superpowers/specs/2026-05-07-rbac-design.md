# RBAC系统设计文档

## 1. 系统概述

### 1.1 项目背景
开发一个统一的RBAC（Role-Based Access Control）权限管理系统，支持多个外部系统接入，通过角色控制用户的菜单权限和API接口权限。

### 1.2 设计目标
- 支持5-20个外部系统接入
- 支持1K-10K用户规模
- 提供高性能的权限校验能力
- 系统间权限数据变更实时同步

### 1.3 技术栈
- 后端：Go
- 前端：Vue 3 + Pinia
- 数据库：MySQL
- 缓存：Redis
- 容器化：Docker + docker-compose

---

## 2. 系统架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                          RBAC系统                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  Vue前端     │  │  RBAC API   │  │  MySQL      │            │
│  │  (管理后台)  │──│  (Go服务)   │──│  (数据存储)  │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                          │                                      │
│                     ┌────┴────┐                                │
│                     │  Redis  │                                │
│                     │ (缓存)  │                                │
│                     └─────────┘                                │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│  外部系统A   │       │  外部系统B   │       │  外部系统C   │
│  (RBAC SDK) │       │  (RBAC SDK) │       │  (RBAC SDK) │
└─────────────┘       └─────────────┘       └─────────────┘
```

### 2.2 认证方式
- **用户登录Web端**：JWT Token（有效期2小时，支持刷新）
- **外部系统调用API**：固定凭证（system_code + system_secret）

### 2.3 权限模型
采用扁平RBAC模型：
- 角色不支持继承
- 每个系统独立定义角色、菜单、权限
- 用户可跨系统拥有不同角色

---

## 3. 数据库设计

### 3.1 核心表结构

#### systems（系统表）
```sql
CREATE TABLE systems (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(100) NOT NULL COMMENT '系统名称',
    code            VARCHAR(50) NOT NULL UNIQUE COMMENT '系统编码',
    secret          VARCHAR(128) NOT NULL COMMENT '系统密钥',
    description     VARCHAR(500) COMMENT '系统描述',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### users（用户表）
```sql
CREATE TABLE users (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password_hash   VARCHAR(128) NOT NULL COMMENT '密码哈希',
    email           VARCHAR(100) COMMENT '邮箱',
    phone           VARCHAR(20) COMMENT '手机号',
    is_super_admin  TINYINT NOT NULL DEFAULT 0 COMMENT '是否超级管理员',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    login_attempts  INT NOT NULL DEFAULT 0 COMMENT '登录失败次数',
    locked_until    DATETIME COMMENT '锁定截止时间',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### system_members（系统成员表）
```sql
CREATE TABLE system_members (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT NOT NULL COMMENT '系统ID',
    user_id         BIGINT NOT NULL COMMENT '用户ID',
    role            VARCHAR(20) NOT NULL COMMENT '成员角色：admin/member',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_user (system_id, user_id)
);
```

#### roles（角色表）
```sql
CREATE TABLE roles (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT NOT NULL COMMENT '所属系统ID',
    name            VARCHAR(50) NOT NULL COMMENT '角色名称',
    code            VARCHAR(50) NOT NULL COMMENT '角色编码',
    description     VARCHAR(500) COMMENT '角色描述',
    version         INT NOT NULL DEFAULT 1 COMMENT '版本号（乐观锁）',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_code (system_id, code)
);
```

#### menus（菜单表）
```sql
CREATE TABLE menus (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT NOT NULL COMMENT '所属系统ID',
    parent_id       BIGINT NOT NULL DEFAULT 0 COMMENT '父菜单ID，0表示顶级',
    name            VARCHAR(50) NOT NULL COMMENT '菜单名称',
    path            VARCHAR(200) COMMENT '路由路径',
    icon            VARCHAR(50) COMMENT '菜单图标',
    sort_order      INT NOT NULL DEFAULT 0 COMMENT '排序序号',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### permissions（API权限表）
```sql
CREATE TABLE permissions (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT NOT NULL COMMENT '所属系统ID',
    code            VARCHAR(100) NOT NULL COMMENT '权限标识',
    name            VARCHAR(100) NOT NULL COMMENT '权限名称',
    method          VARCHAR(10) NOT NULL COMMENT 'HTTP方法：GET/POST/PUT/DELETE',
    path            VARCHAR(200) NOT NULL COMMENT 'API路径',
    description     VARCHAR(500) COMMENT '权限描述',
    version         INT NOT NULL DEFAULT 1 COMMENT '版本号',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_method_path (system_id, method, path)
);
```

#### user_roles（用户-角色关联表）
```sql
CREATE TABLE user_roles (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id         BIGINT NOT NULL COMMENT '用户ID',
    role_id         BIGINT NOT NULL COMMENT '角色ID',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id)
);
```

#### role_menus（角色-菜单关联表）
```sql
CREATE TABLE role_menus (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id         BIGINT NOT NULL COMMENT '角色ID',
    menu_id         BIGINT NOT NULL COMMENT '菜单ID',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_menu (role_id, menu_id)
);
```

#### role_permissions（角色-权限关联表）
```sql
CREATE TABLE role_permissions (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id         BIGINT NOT NULL COMMENT '角色ID',
    permission_id   BIGINT NOT NULL COMMENT '权限ID',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_id)
);
```

#### webhooks（Webhook配置表）
```sql
CREATE TABLE webhooks (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT NOT NULL COMMENT '关联系统ID',
    url             VARCHAR(255) NOT NULL COMMENT '回调地址',
    secret          VARCHAR(128) NOT NULL COMMENT '签名密钥',
    events          VARCHAR(255) NOT NULL COMMENT '订阅事件：permission_change,menu_change,role_change',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### audit_logs（审计日志表）
```sql
CREATE TABLE audit_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id         BIGINT NOT NULL COMMENT '操作人ID',
    username        VARCHAR(50) NOT NULL COMMENT '操作人用户名',
    system_id       BIGINT COMMENT '操作的系统ID（可空）',
    action          VARCHAR(50) NOT NULL COMMENT '操作类型：create/update/delete/assign',
    resource        VARCHAR(50) NOT NULL COMMENT '资源类型：user/role/menu/permission/system',
    resource_id     BIGINT COMMENT '资源ID',
    resource_name   VARCHAR(100) COMMENT '资源名称',
    detail          TEXT COMMENT '变更详情JSON',
    ip              VARCHAR(50) COMMENT '操作IP',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 3.2 设计规范
- 所有表使用雪花算法生成ID
- 硬删除，不使用软删除
- 敏感字段（如system_secret）加密存储
- 所有表包含 created_at 和 updated_at 字段

---

## 4. 后端模块设计

### 4.1 模块划分

```
grbac/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/                  # 配置加载
│   ├── middleware/               # 中间件
│   │   ├── auth.go              # 认证中间件
│   │   ├── permission.go        # 权限校验中间件
│   │   ├── cors.go              # CORS中间件
│   │   ├── logger.go            # 日志中间件
│   │   └── recovery.go          # 异常恢复中间件
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据访问层
│   ├── service/                 # 业务逻辑层
│   ├── handler/                 # HTTP处理器
│   ├── router/                  # 路由定义
│   ├── pkg/                     # 公共包
│   │   ├── errors/              # 错误码定义
│   │   ├── response/            # 响应封装
│   │   ├── jwt/                 # JWT工具
│   │   ├── crypto/              # 加密工具
│   │   └── idgen/               # ID生成器
│   └── sdk/                     # 供外部系统集成的SDK
│       ├── client.go            # SDK客户端
│       ├── middleware.go         # 中间件
│       └── cache.go             # 本地缓存
├── config/
│   └── config.yaml              # 配置文件
├── migrations/                  # 数据库迁移脚本
├── Dockerfile
└── go.mod
```

### 4.2 核心接口设计

#### 认证相关
```
POST   /api/auth/login              # 用户登录
POST   /api/auth/refresh            # 刷新Token
POST   /api/auth/logout             # 登出
POST   /api/auth/change-password    # 修改密码
```

#### 用户管理（超级管理员）
```
GET    /api/users                   # 用户列表
POST   /api/users                   # 创建用户
PUT    /api/users/:id               # 更新用户
DELETE /api/users/:id               # 删除用户
PUT    /api/users/:id/status        # 启用/禁用用户
```

#### 系统管理（超级管理员）
```
POST   /api/systems                 # 注册系统
GET    /api/systems                 # 系统列表
PUT    /api/systems/:id             # 更新系统
DELETE /api/systems/:id             # 删除系统
GET    /api/systems/:id/members     # 系统成员列表
POST   /api/systems/:id/members     # 添加系统成员
DELETE /api/systems/:id/members/:uid # 移除系统成员
```

#### 角色管理（系统管理员）
```
GET    /api/systems/:sid/roles      # 角色列表
POST   /api/systems/:sid/roles      # 创建角色
PUT    /api/systems/:sid/roles/:id  # 更新角色
DELETE /api/systems/:sid/roles/:id  # 删除角色
POST   /api/systems/:sid/roles/:id/users    # 分配用户
DELETE /api/systems/:sid/roles/:id/users/:uid # 移除用户
```

#### 菜单管理（系统管理员）
```
GET    /api/systems/:sid/menus          # 菜单树
POST   /api/systems/:sid/menus          # 创建菜单
PUT    /api/systems/:sid/menus/:id      # 更新菜单
DELETE /api/systems/:sid/menus/:id      # 删除菜单
POST   /api/systems/:sid/roles/:id/menus    # 给角色分配菜单
GET    /api/systems/:sid/roles/:id/menus    # 获取角色的菜单
```

#### 权限管理（系统管理员）
```
GET    /api/systems/:sid/permissions          # 权限列表
POST   /api/systems/:sid/permissions          # 创建权限
PUT    /api/systems/:sid/permissions/:id      # 更新权限
DELETE /api/systems/:sid/permissions/:id      # 删除权限
POST   /api/systems/:sid/roles/:id/permissions    # 给角色分配权限
GET    /api/systems/:sid/roles/:id/permissions    # 获取角色的权限
```

#### Webhook管理（系统管理员）
```
GET    /api/systems/:sid/webhooks         # Webhook列表
POST   /api/systems/:sid/webhooks         # 创建Webhook
PUT    /api/systems/:sid/webhooks/:id     # 更新Webhook
DELETE /api/systems/:sid/webhooks/:id     # 删除Webhook
```

#### 审计日志（超级管理员）
```
GET    /api/audit-logs                    # 审计日志列表
GET    /api/audit-logs/export             # 导出日志
```

#### 供外部系统调用的接口
```
POST   /api/external/verify               # 验证Token有效性
GET    /api/external/userinfo             # 获取用户信息和角色
GET    /api/external/menus                # 获取用户在指定系统的菜单
GET    /api/external/permissions          # 获取用户在指定系统的权限
```

#### 健康检查
```
GET    /api/health                        # 健康检查
```

### 4.3 响应格式

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": { ... }
}
```

**分页响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "list": [ ... ]
    }
}
```

**错误响应：**
```json
{
    "code": 10001,
    "message": "用户名或密码错误",
    "data": null
}
```

### 4.4 错误码设计

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 10001 | 用户名或密码错误 |
| 10002 | 账号已被锁定 |
| 10003 | Token已过期 |
| 10004 | Token无效 |
| 10005 | 无权限访问 |
| 20001 | 系统不存在 |
| 20002 | 系统凭证无效 |
| 30001 | 角色不存在 |
| 30002 | 角色编码已存在 |
| 40001 | 菜单不存在 |
| 40002 | 存在子菜单，无法删除 |
| 50001 | 权限不存在 |
| 50002 | 权限路径已存在 |
| 60001 | 用户不存在 |
| 60002 | 用户名已存在 |
| 99999 | 系统内部错误 |

---

## 5. 前端模块设计

### 5.1 技术栈
- Vue 3 + Composition API
- Pinia 状态管理
- Vue Router 动态路由
- Element Plus UI组件库
- Axios HTTP请求

### 5.2 模块划分

```
grbac-web/
├── src/
│   ├── api/                   # API请求封装
│   │   ├── auth.ts
│   │   ├── user.ts
│   │   ├── system.ts
│   │   ├── role.ts
│   │   ├── menu.ts
│   │   ├── permission.ts
│   │   ├── webhook.ts
│   │   └── audit.ts
│   ├── components/            # 公共组件
│   │   ├── PermButton/        # 权限按钮组件
│   │   └── PermMenu/          # 权限菜单组件
│   ├── composables/           # 组合式函数
│   │   ├── useAuth.ts         # 认证相关
│   │   ├── usePermission.ts   # 权限相关
│   │   └── useMenu.ts         # 菜单相关
│   ├── directives/            # 自定义指令
│   │   └── permission.ts      # v-permission指令
│   ├── layouts/               # 布局组件
│   │   ├── DefaultLayout.vue
│   │   └── BlankLayout.vue
│   ├── router/                # 路由配置
│   │   ├── index.ts
│   │   ├── staticRoutes.ts    # 静态路由
│   │   └── dynamicRoutes.ts   # 动态路由生成
│   ├── stores/                # Pinia状态
│   │   ├── auth.ts            # 认证状态
│   │   ├── user.ts            # 用户信息
│   │   ├── menu.ts            # 菜单数据
│   │   └── permission.ts      # 权限数据
│   ├── utils/                 # 工具函数
│   │   ├── request.ts         # Axios封装
│   │   ├── token.ts           # Token管理
│   │   └── storage.ts         # 本地存储
│   ├── views/                 # 页面组件
│   │   ├── login/
│   │   ├── dashboard/
│   │   ├── user/
│   │   ├── system/
│   │   ├── role/
│   │   ├── menu/
│   │   ├── permission/
│   │   ├── webhook/
│   │   └── audit/
│   ├── App.vue
│   └── main.ts
├── Dockerfile
└── package.json
```

### 5.3 权限控制

**路由权限：**
- 登录后从后端获取用户菜单
- 动态生成Vue Router路由
- 无权限的菜单不显示，访问时跳转403

**按钮权限：**
```vue
<!-- 使用v-permission指令 -->
<el-button v-permission="'user:create'" @click="handleCreate">
  新建用户
</el-button>

<!-- 或使用组件 -->
<perm-button code="user:delete" @click="handleDelete">
  删除
</perm-button>
```

### 5.4 Token管理

- Token存储在localStorage
- Axios拦截器自动携带Token
- Token过期自动调用refresh接口
- Refresh失败跳转登录页

---

## 6. 外部系统集成SDK

### 6.1 SDK设计

```go
// 初始化客户端
client := rbac.NewClient(&rbac.Config{
    SystemCode:   "order_system",
    SystemSecret: "xxxxx",
    RBACEndpoint: "http://rbac-server:8080",
})

// 登录获取Token
token, err := client.Login(ctx, "username", "password")

// 获取用户菜单
menus, err := client.GetMenus(ctx, token)

// 获取用户权限
permissions, err := client.GetPermissions(ctx, token)

// 创建权限校验中间件
permMiddleware := client.NewPermissionMiddleware()

// 使用中间件
router.GET("/api/orders", permMiddleware.Check("order:list"), orderHandler)
```

### 6.2 本地缓存策略

SDK内置本地缓存，权限校验无网络开销：

- 首次请求时加载用户菜单和权限到内存
- 缓存有效期30分钟
- 权限变更时通过Webhook通知刷新
- 5分钟轮询兜底检查版本号

### 6.3 Webhook集成

```go
// 注册Webhook处理器
client.RegisterWebhookHandler(func(event rbac.Event) {
    switch event.Type {
    case "permission_change":
        // 刷新权限缓存
        client.RefreshPermissionCache(event.UserID)
    case "menu_change":
        // 刷新菜单缓存
        client.RefreshMenuCache(event.UserID)
    }
})
```

---

## 7. 权限变更通知

### 7.1 方案：Webhook + 轮询兜底

```
RBAC系统 ──────Webhook回调──────▶ 外部系统
    │                               │
    │         失败重试               │
    │    (1s → 5s → 30s)           │
    │                               │
    └──────轮询兜底(每5分钟)────────┘
```

### 7.2 Webhook事件格式

```json
{
    "event": "permission_change",
    "system_code": "order_system",
    "timestamp": 1715059200,
    "data": {
        "action": "update",
        "permission_id": 123,
        "affected_roles": [1, 5, 8],
        "version": 15
    },
    "signature": "sha256=xxxxx"
}
```

### 7.3 签名验证

外部系统收到Webhook后，使用配置的secret验证签名，防止伪造请求。

---

## 8. 安全设计

### 8.1 密码安全
- 使用bcrypt哈希存储密码
- 密码复杂度要求：8位以上，包含大小写字母、数字、特殊字符
- 登录失败5次锁定30分钟

### 8.2 Token安全
- JWT使用HS256签名
- Access Token有效期2小时
- Refresh Token有效期7天
- 支持Token黑名单（登出/改密后失效）

### 8.3 API安全
- 系统凭证（system_secret）加密存储
- 敏感接口限流保护
- CORS配置限制跨域访问
- SQL注入防护（参数化查询）

### 8.4 数据校验规则
- 删除角色：检查是否有用户绑定
- 删除菜单：检查是否有子菜单，级联删除
- 删除权限：检查是否有角色绑定
- 分配角色：检查是否已分配，避免重复

---

## 9. 配置管理

### 9.1 配置文件格式

```yaml
# config.yaml
server:
  port: 8080
  mode: debug  # debug/release

database:
  host: localhost
  port: 3306
  username: root
  password: ""
  dbname: grbac
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret-key"
  access_expire: 7200      # 2小时
  refresh_expire: 604800   # 7天

log:
  level: info              # debug/info/warn/error
  format: json             # json/text
  output: stdout           # stdout/file
  file_path: ./logs/app.log

password:
  min_length: 8
  max_attempts: 5
  lock_duration: 1800      # 30分钟
```

---

## 10. 日志规范

### 10.1 日志格式

```json
{
    "level": "info",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": 123,
    "username": "admin",
    "action": "create_role",
    "resource": "role",
    "resource_id": 456,
    "resource_name": "订单管理员",
    "system_id": 1,
    "ip": "192.168.1.100",
    "method": "POST",
    "path": "/api/systems/1/roles",
    "status": 200,
    "latency": "15ms",
    "message": "created role successfully",
    "timestamp": "2026-05-07T10:00:00Z"
}
```

### 10.2 日志级别
- **debug**：调试信息，开发环境使用
- **info**：正常操作日志
- **warn**：警告信息，如权限校验失败
- **error**：错误信息，需要关注

---

## 11. 部署设计

### 11.1 Docker配置

**后端 Dockerfile：**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./server"]
```

**前端 Dockerfile：**
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### 11.2 docker-compose配置

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: grbac
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  rbac-server:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis

  rbac-web:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "80:80"
    depends_on:
      - rbac-server

volumes:
  mysql_data:
```

---

## 12. 测试策略

### 12.1 单元测试
- 核心业务逻辑：权限校验、角色分配、Token生成
- 目标覆盖率：80%以上

### 12.2 集成测试
- API接口完整流程测试
- 数据库操作测试
- 缓存读写测试

### 12.3 测试命令
```bash
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./internal/service/...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 13. 构建与运行

### 13.1 后端

```bash
# 安装依赖
go mod tidy

# 运行
go run ./cmd/server

# 构建
go build -o server ./cmd/server

# 运行测试
go test ./...
```

### 13.2 前端

```bash
# 安装依赖
npm install

# 开发运行
npm run dev

# 构建
npm run build

# 代码检查
npm run lint
```

### 13.3 Docker部署

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f rbac-server

# 停止服务
docker-compose down
```

---

## 14. 附录

### 14.1 ER图

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   systems   │       │    users    │       │   roles     │
├─────────────┤       ├─────────────┤       ├─────────────┤
│ id          │       │ id          │       │ id          │
│ name        │       │ username    │       │ system_id   │──┐
│ code        │       │ password    │       │ name        │  │
│ secret      │       │ email       │       │ code        │  │
│ status      │       │ is_super    │       │ status      │  │
└──────┬──────┘       └──────┬──────┘       └──────┬──────┘  │
       │                     │                     │         │
       │                     │    ┌────────────────┘         │
       │                     │    │                          │
       ▼                     ▼    ▼                          ▼
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│sys_members  │       │ user_roles  │       │   menus     │
├─────────────┤       ├─────────────┤       ├─────────────┤
│ system_id   │       │ user_id     │       │ id          │
│ user_id     │       │ role_id     │       │ system_id   │
│ role        │       └─────────────┘       │ parent_id   │
└─────────────┘                             │ name        │
                                            │ path        │
       ┌────────────────────────────────────┘
       │
       ▼
┌─────────────┐       ┌─────────────┐
│ role_menus  │       │permissions  │
├─────────────┤       ├─────────────┤
│ role_id     │       │ id          │
│ menu_id     │       │ system_id   │
└─────────────┘       │ method      │
                      │ path        │
┌─────────────┐       │ code        │
│role_perms   │       └──────┬──────┘
├─────────────┤              │
│ role_id     │              │
│ perm_id     │──────────────┘
└─────────────┘
```

### 14.2 时序图

**用户登录流程：**
```
用户          前端           RBAC服务        数据库
 │             │              │              │
 │──登录请求──▶│              │              │
 │             │──POST /login─▶│              │
 │             │              │──查询用户───▶│
 │             │              │◀──用户信息───│
 │             │              │──验证密码    │
 │             │              │──生成JWT     │
 │             │◀──Token──────│              │
 │◀──登录成功──│              │              │
```

**权限校验流程：**
```
用户          外部系统        RBAC SDK       RBAC服务
 │             │              │              │
 │──API请求──▶│              │              │
 │             │──检查权限───▶│              │
 │             │              │──本地缓存    │
 │             │              │──返回结果    │
 │             │◀──允许/拒绝──│              │
 │◀──响应──────│              │              │
```

---

文档版本：v1.0
最后更新：2026-05-07
