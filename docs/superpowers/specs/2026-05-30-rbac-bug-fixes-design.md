# GRBAC Bug 修复设计文档

日期: 2026-05-30

## 背景

GRBAC 是一个多租户 RBAC 权限管理平台。经过全面代码审查，发现了多个安全、逻辑和数据完整性问题。本文档定义了按模块逐层修复的方案。

## 修复范围

本次修复覆盖四个模块，共 15 项修复：

| 模块 | 修复项数 | 涉及层 |
|------|---------|--------|
| 后端中间件层 | 3 | middleware |
| 后端 Service 层 | 7 | service |
| 后端 Model/DB 层 | 4 | model, repository |
| 前端层 | 1 | utils, stores |

**不包含的项目**（已确认延后）：
- 外部 API 认证增强（保持现状，仅文档说明）
- CORS 策略调整（保持现状）
- Webhook HMAC 签名（暂时跳过）
- Menu.ParentID 语义优化（影响面大，后续优化）

---

## 模块 1：后端中间件层

### 1.1 审计日志异步写入错误处理

**问题**: `middleware/audit.go` 中 `go auditSvc.Record(auditLog)` 丢弃了错误返回值，数据库写入失败时审计日志静默丢失。

**修复**: 在 goroutine 内部包裹错误处理，失败时记录日志。不改为同步写入以避免增加响应延迟。

**改动文件**: `backend/internal/middleware/audit.go`

```go
// 修改前
go auditSvc.Record(auditLog)

// 修改后
go func() {
    if err := auditSvc.Record(auditLog); err != nil {
        log.Printf("[AUDIT] failed to record: user=%d action=%s resource=%s err=%v",
            auditLog.UserID, auditLog.Action, auditLog.Resource, err)
    }
}()
```

### 1.2 Rate limiter 微秒时间戳覆盖

**问题**: `middleware/ratelimit.go` 中使用微秒时间戳作为 Redis sorted set 的 member。同一微秒内的两个请求会使用相同 member，`ZADD` 覆盖而非新增，导致计数不足。

**修复**: member 格式改为 `<timestamp>:<random_suffix>`，用 4 字节随机数确保唯一。

**改动文件**: `backend/internal/middleware/ratelimit.go`

```go
// 修改前
member := fmt.Sprintf("%d", now)

// 修改后
randBytes := make([]byte, 4)
_, _ = crypto_rand.Read(randBytes)
member := fmt.Sprintf("%d:%x", now, randBytes)
```

### 1.3 `extractResource()` 脆弱解析

**问题**: `middleware/audit.go` 中 `extractResource()` 手动解析 URL 路径段，硬编码跳过 "api"、"systems"、"auth" 等前缀。新增路由后会静默失败，记录错误的资源名。

**修复**: 改用 Gin 的 `c.FullPath()` 获取路由模式字符串（如 `/api/systems/:id/roles/:rid`），从模式中提取资源名和参数。同时解析实际路径中的参数值作为 ResourceID。

**改动文件**: `backend/internal/middleware/audit.go`

```go
func extractResource(c *gin.Context) (resource string, resourceID *int64) {
    // 使用路由模式而非实际路径
    pattern := c.FullPath()
    segments := strings.Split(pattern, "/")

    // 从后往前找最后一个命名资源段（非参数段）
    for i := len(segments) - 1; i >= 0; i-- {
        seg := segments[i]
        if strings.HasPrefix(seg, ":") {
            // 参数段，尝试从实际路径获取 ID
            continue
        }
        if seg == "api" || seg == "" {
            continue
        }
        resource = seg
        break
    }

    // 从 URL 参数中提取最后一个 ID
    for _, param := range c.Params {
        if id, err := strconv.ParseInt(param.Value, 10, 64); err == nil {
            resourceID = &id
        }
    }

    return resource, resourceID
}
```

---

## 模块 2：后端 Service 层

### 2.1 Webhook 事件匹配 LIKE 误匹配

**问题**: `repository/webhook/repo.go` 中 `GetByEvent` 使用 `events LIKE '%event%'`，会误匹配子串（如 `role.created` 匹配 `role.created_extra`）。

**修复**: 改用精确匹配逻辑，覆盖事件名在逗号分隔列表中的四种位置：

**改动文件**: `backend/internal/repository/webhook/repo.go`

```go
func (r *repo) GetByEvent(event string) ([]*model.Webhook, error) {
    var webhooks []*model.Webhook
    query := `status = 1 AND (
        events = ? OR
        events LIKE ? OR
        events LIKE ? OR
        events LIKE ?
    )`
    exact := event
    prefix := event + ",%"
    suffix := "%," + event
    middle := "%," + event + ",%"

    err := r.db.Where(query, exact, prefix, suffix, middle).
        Find(&webhooks).Error
    return webhooks, err
}
```

注意：当前实现未按 `system_id` 过滤，修复保持相同行为（查询所有系统的 webhooks）。上述代码中的 `system_id = ?` 条件应去掉，改为仅按 `status` 和 `events` 匹配。

### 2.2 Webhook dead code 清理

**问题**: `service/webhook/event.go` 中定义了 `EventPayload` struct，但实际 dispatch 使用 `map[string]interface{}`。

**修复**: 删除 `EventPayload` struct 及相关未使用代码。

**改动文件**: `backend/internal/service/webhook/event.go`

### 2.3 Webhook context 传播

**问题**: role、permission、menu、system service 中调用 dispatchFn 时传入 `context.Background()`，丢失请求上下文的 tracing 信息。

**修复**: dispatchFn 签名增加 `ctx context.Context` 参数。调用方传入 `c.Request.Context()`。webhook dispatcher 的 goroutine 内使用 `context.WithTimeout(context.Background(), 30*time.Second)` 创建独立超时 context。

**改动文件**:
- `backend/internal/service/webhook/service.go` — 修改 `DispatchFunc` 类型签名
- `backend/internal/service/webhook/dispatcher.go` — `Dispatch` 方法使用独立超时 context
- `backend/internal/service/role/service.go` — 传入 `ctx`
- `backend/internal/service/permission/service.go` — 传入 `ctx`
- `backend/internal/service/menu/service.go` — 传入 `ctx`
- `backend/internal/service/system/service.go` — 传入 `ctx`
- `backend/internal/app/app.go` — 更新 dispatchFn 注入

### 2.4 Permission.Update 唯一性检查

**问题**: `service/permission/service.go` 中 `Update` 方法更新 method/path 时未检查新组合是否已存在。数据库唯一约束错误会报为 `ErrInternal` 而非有意义的业务错误。

**修复**: 如果 method 或 path 发生变化，先查询目标组合是否已存在。

**改动文件**:
- `backend/internal/service/permission/service.go` — 添加唯一性检查逻辑
- `backend/internal/pkg/errors/errors.go` — 新增 `ErrPermPathExists` 错误码 (5005)

```go
func (s *service) Update(ctx context.Context, perm *model.Permission) error {
    existing, err := s.repo.GetByID(perm.ID)
    if err != nil {
        return errors.ErrPermNotFound.Wrap(err.Error())
    }

    // 如果 method 或 path 变化，检查新组合唯一性
    if perm.Method != existing.Method || perm.Path != existing.Path {
        dup, err := s.repo.GetByMethodPath(perm.SystemID, perm.Method, perm.Path)
        if err == nil && dup != nil && dup.ID != perm.ID {
            return errors.ErrPermPathExists
        }
    }

    perm.Version = existing.Version + 1
    return s.repo.Update(perm)
}
```

### 2.5 超管权限返回一致性

**问题**: `service/external/service.go` 中 `GetUserPermissions` 对超管返回空数组，但 `GetUserMenus` 返回全部菜单，`ValidatePermission` 返回 true。行为不一致。

**修复**: `GetUserPermissions` 对超管返回系统全部权限 code 列表，与 `GetUserMenus` 行为一致。

**改动文件**: `backend/internal/service/external/service.go`

```go
func (s *service) GetUserPermissions(ctx context.Context, systemCode string, userID int64) ([]string, error) {
    // ... 获取 system ...

    // 超管返回全部权限
    if user.IsSuperAdmin == 1 {
        perms, err := s.permRepo.ListBySystemID(system.ID)
        if err != nil {
            return nil, err
        }
        codes := make([]string, 0, len(perms))
        for _, p := range perms {
            codes = append(codes, p.Code)
        }
        return codes, nil
    }

    // ... 原有逻辑 ...
}
```

### 2.6 buildTree O(n²) → O(n)

**问题**: `service/menu/service.go` 中 `buildTree` 对每个节点遍历整个列表找子节点，O(n²) 复杂度。

**修复**: 先用 map 建立 parentID → children 索引，再递归构建。

**改动文件**: `backend/internal/service/menu/service.go`

```go
func buildTree(menus []*model.Menu, parentID int64) []*model.Menu {
    // 建立索引
    index := make(map[int64][]*model.Menu)
    for _, m := range menus {
        index[m.ParentID] = append(index[m.ParentID], m)
    }

    var build func(pid int64) []*model.Menu
    build = func(pid int64) []*model.Menu {
        children := index[pid]
        for _, child := range children {
            child.Children = build(child.ID)
        }
        return children
    }

    return build(parentID)
}
```

### 2.7 ListForUser 分页支持

**问题**: `service/system/service.go` 中 `ListForUser` 对非超管用户不分页，`ListByUserIDWithRole` 返回全部系统。

**修复**: 在 repository 层添加 `COUNT` 查询和 `LIMIT/OFFSET`，返回 `(systems, total, error)`。

**改动文件**:
- `backend/internal/repository/system/repo.go` — 添加 `CountByUserID` 方法，修改 `ListByUserIDWithRole` 支持分页参数
- `backend/internal/service/system/service.go` — 传递分页参数

---

## 模块 3：后端 Model/DB 层

### 3.1 Permission.Code 唯一约束

**问题**: `Permission.Code` 无唯一索引，同系统内可能出现重复 code。

**修复**: 在 model GORM tag 中添加 `uniqueIndex:uk_system_code`。

**改动文件**: `backend/internal/model/permission.go`

```go
Code string `gorm:"type:varchar(100);not null;uniqueIndex:uk_system_code" json:"code"`
```

注意：这与现有的 `uk_system_method_path` 唯一索引共存。`uk_system_code` 确保 code 在系统内唯一，`uk_system_method_path` 确保 method+path 组合唯一。

### 3.2 SystemMember.Role 枚举约束

**问题**: `SystemMember.Role` 是自由字符串，可写入任意值。

**修复**: service 层已有验证逻辑（仅允许 "admin"/"member"），DB 层暂不加 CHECK 约束（GORM AutoMigrate 不支持）。添加 model 注释说明合法值。

**改动文件**: `backend/internal/model/system.go`

```go
// Role is the system-level membership role.
// Valid values: "admin" (can manage system config), "member" (regular member).
Role string `gorm:"type:varchar(20);not null" json:"role"`
```

### 3.3 AuditLog.UserID 索引

**问题**: `AuditLog.UserID` 无索引，按用户查询审计日志会全表扫描。

**修复**: 添加 GORM index tag。

**改动文件**: `backend/internal/model/audit.go`

```go
UserID int64 `gorm:"not null;index" json:"user_id"`
```

### 3.4 错误码新增

新增 `ErrPermPathExists` 错误码用于 Permission.Update 唯一性检查。

**改动文件**: `backend/internal/pkg/errors/errors.go`

```go
ErrPermPathExists = &AppError{Code: 5005, Message: "permission method+path already exists"}
```

---

## 模块 4：前端层

### 4.1 Token 自动刷新

**问题**: Axios 响应拦截器遇到 401 直接清 token 跳登录页，`authStore.refreshToken()` 存在但未接入。

**修复方案**:

1. 在 `request.ts` 中引入 `isRefreshing` 标志和 `pendingRequests` 队列
2. 收到 401 时：
   - 如果正在刷新中，将当前请求加入队列等待
   - 如果未在刷新中，启动刷新流程
   - 刷新成功：用新 token 重试原请求 + 执行队列中的请求
   - 刷新失败：清 token 跳登录页
3. 需要从 `request.ts` 访问 Pinia store，通过在 main.ts 中注入或直接导入

**改动文件**: `frontend/src/utils/request.ts`, `frontend/src/stores/auth.ts`

```typescript
// request.ts 核心逻辑
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

// 响应拦截器中
if (error.response?.status === 401) {
    const config = error.config

    if (isRefreshing) {
        return new Promise((resolve) => {
            pendingRequests.push((token: string) => {
                config.headers.Authorization = `Bearer ${token}`
                resolve(axios(config))
            })
        })
    }

    isRefreshing = true
    try {
        const newToken = await authStore.refreshToken()
        // 重试当前请求
        config.headers.Authorization = `Bearer ${newToken}`
        // 执行队列中的请求
        pendingRequests.forEach(cb => cb(newToken))
        pendingRequests = []
        return axios(config)
    } catch {
        authStore.logout()
        router.push('/login')
        return Promise.reject(error)
    } finally {
        isRefreshing = false
    }
}
```

---

## 数据库迁移影响

以下改动需要数据库 schema 变更，通过 GORM AutoMigrate 自动处理：

| 改动 | 操作 | 风险 |
|------|------|------|
| Permission.Code 添加唯一索引 | ADD UNIQUE INDEX | 如果已有重复 code 会失败，需先清理 |
| AuditLog.UserID 添加索引 | ADD INDEX | 无风险，Online DDL |

**部署步骤建议**：
1. 先检查 permissions 表中是否有重复 code：`SELECT system_id, code, COUNT(*) FROM permissions GROUP BY system_id, code HAVING COUNT(*) > 1`
2. 如有重复，手动清理后再部署
3. AutoMigrate 会自动添加索引

---

## 测试策略

每个模块修复后需要验证：

1. **中间件层**：
   - 审计日志写入失败时能看到日志输出
   - 高并发下 Rate limiter 计数准确
   - 新增路由后审计日志正确记录资源名

2. **Service 层**：
   - Webhook 事件匹配不误匹配子串
   - Permission.Update 重复 method+path 返回友好错误
   - 超管 GetUserPermissions 返回全部权限
   - 大量菜单时 buildTree 性能正常
   - 系统列表分页正常

3. **Model/DB 层**：
   - Permission.Code 重复创建被拒绝
   - AuditLog 按 UserID 查询使用索引

4. **前端层**：
   - Access token 过期后自动刷新，用户无感知
   - Refresh token 也过期时正确跳转登录页
   - 并发请求只触发一次刷新
