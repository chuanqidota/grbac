# RBAC系统实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建统一的RBAC权限管理系统，支持多系统接入，控制用户菜单和API权限

**Architecture:** 单体Go服务 + Vue3管理后台 + MySQL/Redis + Docker部署

**Tech Stack:** Go 1.21+, Vue 3, Element Plus, Pinia, MySQL 8.0, Redis 7, Docker

---

## Phase 1: 项目初始化与基础设施

### Task 1: Go项目初始化

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/config/config.yaml`

- [ ] **Step 1: 初始化Go模块**

```bash
cd backend && go mod init github.com/jinang/grbac
```

- [ ] **Step 2: 创建配置文件**

```yaml
# backend/config/config.yaml
server:
  port: 8080
  mode: debug

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
  secret: "grbac-jwt-secret-key-2026"
  access_expire: 7200
  refresh_expire: 604800

log:
  level: debug
  format: text
  output: stdout

password:
  min_length: 8
  max_attempts: 5
  lock_duration: 1800
```

- [ ] **Step 3: 创建主入口文件**

```go
// backend/cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jinang/grbac/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Starting RBAC server on port %d...\n", cfg.Server.Port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
```

- [ ] **Step 4: 创建配置加载模块**

```go
// backend/internal/config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	Password PasswordConfig `yaml:"password"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret        string `yaml:"secret"`
	AccessExpire  int    `yaml:"access_expire"`
	RefreshExpire int    `yaml:"refresh_expire"`
}

type LogConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

type PasswordConfig struct {
	MinLength    int `yaml:"min_length"`
	MaxAttempts  int `yaml:"max_attempts"`
	LockDuration int `yaml:"lock_duration"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
```

- [ ] **Step 5: 安装依赖并验证**

```bash
cd backend
go get gopkg.in/yaml.v3
go mod tidy
go build ./cmd/server
```

- [ ] **Step 6: Commit**

```bash
git add backend/
git commit -m "feat: initialize Go project with config module"
```

---

### Task 2: 数据库模型与迁移

**Files:**
- Create: `backend/internal/model/system.go`
- Create: `backend/internal/model/user.go`
- Create: `backend/internal/model/role.go`
- Create: `backend/internal/model/menu.go`
- Create: `backend/internal/model/permission.go`
- Create: `backend/internal/model/webhook.go`
- Create: `backend/internal/model/audit.go`
- Create: `backend/migrations/001_init.sql`

- [ ] **Step 1: 创建系统模型**

```go
// backend/internal/model/system.go
package model

import "time"

type System struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Code        string    `json:"code" gorm:"size:50;uniqueIndex;not null"`
	Secret      string    `json:"-" gorm:"size:128;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Status      int8      `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SystemMember struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	SystemID  int64     `json:"system_id" gorm:"index;not null"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	Role      string    `json:"role" gorm:"size:20;not null"` // admin/member
	CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 2: 创建用户模型**

```go
// backend/internal/model/user.go
package model

import "time"

type User struct {
	ID            int64      `json:"id" gorm:"primaryKey"`
	Username      string     `json:"username" gorm:"size:50;uniqueIndex;not null"`
	PasswordHash  string     `json:"-" gorm:"size:128;not null"`
	Email         string     `json:"email" gorm:"size:100"`
	Phone         string     `json:"phone" gorm:"size:20"`
	IsSuperAdmin  bool       `json:"is_super_admin" gorm:"default:false"`
	Status        int8       `json:"status" gorm:"default:1"`
	LoginAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil   *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type UserRole struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	RoleID    int64     `json:"role_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 3: 创建角色模型**

```go
// backend/internal/model/role.go
package model

import "time"

type Role struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	SystemID    int64     `json:"system_id" gorm:"index;not null"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Code        string    `json:"code" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Version     int       `json:"version" gorm:"default:1"`
	Status      int8      `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoleMenu struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	RoleID    int64     `json:"role_id" gorm:"index;not null"`
	MenuID    int64     `json:"menu_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
}

type RolePermission struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	RoleID       int64     `json:"role_id" gorm:"index;not null"`
	PermissionID int64     `json:"permission_id" gorm:"index;not null"`
	CreatedAt    time.Time `json:"created_at"`
}
```

- [ ] **Step 4: 创建菜单模型**

```go
// backend/internal/model/menu.go
package model

import "time"

type Menu struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	SystemID  int64     `json:"system_id" gorm:"index;not null"`
	ParentID  int64     `json:"parent_id" gorm:"index;default:0"`
	Name      string    `json:"name" gorm:"size:50;not null"`
	Path      string    `json:"path" gorm:"size:200"`
	Icon      string    `json:"icon" gorm:"size:50"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 5: 创建权限模型**

```go
// backend/internal/model/permission.go
package model

import "time"

type Permission struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	SystemID    int64     `json:"system_id" gorm:"index;not null"`
	Code        string    `json:"code" gorm:"size:100;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Method      string    `json:"method" gorm:"size:10;not null"`
	Path        string    `json:"path" gorm:"size:200;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Version     int       `json:"version" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

- [ ] **Step 6: 创建Webhook和审计日志模型**

```go
// backend/internal/model/webhook.go
package model

import "time"

type Webhook struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	SystemID  int64     `json:"system_id" gorm:"index;not null"`
	URL       string    `json:"url" gorm:"size:255;not null"`
	Secret    string    `json:"-" gorm:"size:128;not null"`
	Events    string    `json:"events" gorm:"size:255;not null"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// backend/internal/model/audit.go
package model

import "time"

type AuditLog struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	UserID       int64     `json:"user_id" gorm:"index;not null"`
	Username     string    `json:"username" gorm:"size:50;not null"`
	SystemID     *int64    `json:"system_id" gorm:"index"`
	Action       string    `json:"action" gorm:"size:50;not null"`
	Resource     string    `json:"resource" gorm:"size:50;not null"`
	ResourceID   *int64    `json:"resource_id"`
	ResourceName string    `json:"resource_name" gorm:"size:100"`
	Detail       string    `json:"detail" gorm:"type:text"`
	IP           string    `json:"ip" gorm:"size:50"`
	CreatedAt    time.Time `json:"created_at"`
}
```

- [ ] **Step 7: 创建数据库迁移脚本**

```sql
-- backend/migrations/001_init.sql
CREATE TABLE IF NOT EXISTS systems (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    secret VARCHAR(128) NOT NULL,
    description VARCHAR(500),
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(128) NOT NULL,
    email VARCHAR(100),
    phone VARCHAR(20),
    is_super_admin TINYINT NOT NULL DEFAULT 0,
    status TINYINT NOT NULL DEFAULT 1,
    login_attempts INT NOT NULL DEFAULT 0,
    locked_until DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_user (system_id, user_id)
);

CREATE TABLE IF NOT EXISTS roles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id BIGINT NOT NULL,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(500),
    version INT NOT NULL DEFAULT 1,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_code (system_id, code)
);

CREATE TABLE IF NOT EXISTS menus (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id BIGINT NOT NULL,
    parent_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(50) NOT NULL,
    path VARCHAR(200),
    icon VARCHAR(50),
    sort_order INT NOT NULL DEFAULT 0,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id BIGINT NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(200) NOT NULL,
    description VARCHAR(500),
    version INT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_method_path (system_id, method, path)
);

CREATE TABLE IF NOT EXISTS user_roles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_menus (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id BIGINT NOT NULL,
    menu_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_menu (role_id, menu_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS webhooks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id BIGINT NOT NULL,
    url VARCHAR(255) NOT NULL,
    secret VARCHAR(128) NOT NULL,
    events VARCHAR(255) NOT NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    username VARCHAR(50) NOT NULL,
    system_id BIGINT,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_id BIGINT,
    resource_name VARCHAR(100),
    detail TEXT,
    ip VARCHAR(50),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

- [ ] **Step 8: Commit**

```bash
git add backend/internal/model/ backend/migrations/
git commit -m "feat: add database models and migration scripts"
```

---

### Task 3: 公共工具包

**Files:**
- Create: `backend/internal/pkg/errors/errors.go`
- Create: `backend/internal/pkg/response/response.go`
- Create: `backend/internal/pkg/jwt/jwt.go`
- Create: `backend/internal/pkg/crypto/crypto.go`

- [ ] **Step 1: 创建错误码定义**

```go
// backend/internal/pkg/errors/errors.go
package errors

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	Success             = &AppError{Code: 0, Message: "success"}
	ErrInvalidPassword  = &AppError{Code: 10001, Message: "用户名或密码错误"}
	ErrAccountLocked    = &AppError{Code: 10002, Message: "账号已被锁定"}
	ErrTokenExpired     = &AppError{Code: 10003, Message: "Token已过期"}
	ErrTokenInvalid     = &AppError{Code: 10004, Message: "Token无效"}
	ErrNoPermission     = &AppError{Code: 10005, Message: "无权限访问"}
	ErrSystemNotFound   = &AppError{Code: 20001, Message: "系统不存在"}
	ErrSystemCredential = &AppError{Code: 20002, Message: "系统凭证无效"}
	ErrRoleNotFound     = &AppError{Code: 30001, Message: "角色不存在"}
	ErrRoleCodeExists   = &AppError{Code: 30002, Message: "角色编码已存在"}
	ErrMenuNotFound     = &AppError{Code: 40001, Message: "菜单不存在"}
	ErrMenuHasChildren  = &AppError{Code: 40002, Message: "存在子菜单，无法删除"}
	ErrPermNotFound     = &AppError{Code: 50001, Message: "权限不存在"}
	ErrPermPathExists   = &AppError{Code: 50002, Message: "权限路径已存在"}
	ErrUserNotFound     = &AppError{Code: 60001, Message: "用户不存在"}
	ErrUsernameExists   = &AppError{Code: 60002, Message: "用户名已存在"}
	ErrInternal         = &AppError{Code: 99999, Message: "系统内部错误"}
)

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
```

- [ ] **Step 2: 创建响应封装**

```go
// backend/internal/pkg/response/response.go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func OKPage(c *gin.Context, total int64, list interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			Total: total,
			List:  list,
		},
	})
}

func Error(c *gin.Context, err *errors.AppError) {
	c.JSON(http.StatusOK, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

func ErrorMsg(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
```

- [ ] **Step 3: 创建JWT工具**

```go
// backend/internal/pkg/jwt/jwt.go
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewJWTManager(secret string, accessExpire, refreshExpire int) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		accessExpire:  time.Duration(accessExpire) * time.Second,
		refreshExpire: time.Duration(refreshExpire) * time.Second,
	}
}

func (m *JWTManager) GenerateToken(userID int64, username string, isSuperAdmin bool) (accessToken, refreshToken string, err error) {
	now := time.Now()

	// Access Token
	accessClaims := &Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(m.secret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(m.secret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
```

- [ ] **Step 4: 创建加密工具**

```go
// backend/internal/pkg/crypto/crypto.go
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func VerifyHMAC(message, secret, signature string) bool {
	expected := GenerateHMAC(message, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
```

- [ ] **Step 5: 安装依赖**

```bash
cd backend
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go mod tidy
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/pkg/
git commit -m "feat: add common utility packages (errors, response, jwt, crypto)"
```

---

## Phase 2: 数据库连接与Repository层

### Task 4: 数据库连接

**Files:**
- Create: `backend/internal/database/mysql.go`
- Create: `backend/internal/database/redis.go`

- [ ] **Step 1: 创建MySQL连接**

```go
// backend/internal/database/mysql.go
package database

import (
	"fmt"

	"github.com/jinang/grbac/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}
```

- [ ] **Step 2: 创建Redis连接**

```go
// backend/internal/database/redis.go
package database

import (
	"context"
	"fmt"

	"github.com/jinang/grbac/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
```

- [ ] **Step 3: 安装依赖**

```bash
cd backend
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/redis/go-redis/v9
go mod tidy
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/database/
git commit -m "feat: add database connection modules (MySQL, Redis)"
```

---

### Task 5: Repository层

**Files:**
- Create: `backend/internal/repository/system_repo.go`
- Create: `backend/internal/repository/user_repo.go`
- Create: `backend/internal/repository/role_repo.go`
- Create: `backend/internal/repository/menu_repo.go`
- Create: `backend/internal/repository/permission_repo.go`

- [ ] **Step 1: 创建System Repository**

```go
// backend/internal/repository/system_repo.go
package repository

import (
	"github.com/jinang/grbac/internal/model"
	"gorm.io/gorm"
)

type SystemRepo struct {
	db *gorm.DB
}

func NewSystemRepo(db *gorm.DB) *SystemRepo {
	return &SystemRepo{db: db}
}

func (r *SystemRepo) Create(system *model.System) error {
	return r.db.Create(system).Error
}

func (r *SystemRepo) GetByID(id int64) (*model.System, error) {
	var system model.System
	err := r.db.First(&system, id).Error
	return &system, err
}

func (r *SystemRepo) GetByCode(code string) (*model.System, error) {
	var system model.System
	err := r.db.Where("code = ?", code).First(&system).Error
	return &system, err
}

func (r *SystemRepo) List(page, pageSize int) ([]model.System, int64, error) {
	var systems []model.System
	var total int64

	r.db.Model(&model.System{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&systems).Error
	return systems, total, err
}

func (r *SystemRepo) Update(system *model.System) error {
	return r.db.Save(system).Error
}

func (r *SystemRepo) Delete(id int64) error {
	return r.db.Delete(&model.System{}, id).Error
}

func (r *SystemRepo) AddMember(member *model.SystemMember) error {
	return r.db.Create(member).Error
}

func (r *SystemRepo) RemoveMember(systemID, userID int64) error {
	return r.db.Where("system_id = ? AND user_id = ?", systemID, userID).Delete(&model.SystemMember{}).Error
}

func (r *SystemRepo) GetMembers(systemID int64) ([]model.SystemMember, error) {
	var members []model.SystemMember
	err := r.db.Where("system_id = ?", systemID).Find(&members).Error
	return members, err
}

func (r *SystemRepo) IsMember(systemID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.SystemMember{}).Where("system_id = ? AND user_id = ?", systemID, userID).Count(&count).Error
	return count > 0, err
}

func (r *SystemRepo) IsAdmin(systemID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.SystemMember{}).Where("system_id = ? AND user_id = ? AND role = 'admin'", systemID, userID).Count(&count).Error
	return count > 0, err
}
```

- [ ] **Step 2: 创建User Repository**

```go
// backend/internal/repository/user_repo.go
package repository

import (
	"github.com/jinang/grbac/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepo) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Select("id, username, email, phone, is_super_admin, status, created_at, updated_at").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepo) Delete(id int64) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepo) AddUserRole(userRole *model.UserRole) error {
	return r.db.Create(userRole).Error
}

func (r *UserRepo) RemoveUserRole(userID, roleID int64) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{}).Error
}

func (r *UserRepo) GetUserRoles(userID int64) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := r.db.Where("user_id = ?", userID).Find(&userRoles).Error
	return userRoles, err
}

func (r *UserRepo) GetUsersByRoleID(roleID int64) ([]int64, error) {
	var userIDs []int64
	err := r.db.Model(&model.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &userIDs).Error
	return userIDs, err
}
```

- [ ] **Step 3: 创建Role Repository**

```go
// backend/internal/repository/role_repo.go
package repository

import (
	"github.com/jinang/grbac/internal/model"
	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepo) GetByID(id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *RoleRepo) GetByCode(systemID int64, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("system_id = ? AND code = ?", systemID, code).First(&role).Error
	return &role, err
}

func (r *RoleRepo) ListBySystem(systemID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("system_id = ?", systemID).Order("id ASC").Find(&roles).Error
	return roles, err
}

func (r *RoleRepo) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepo) Delete(id int64) error {
	return r.db.Delete(&model.Role{}, id).Error
}

func (r *RoleRepo) AddRoleMenu(roleMenu *model.RoleMenu) error {
	return r.db.Create(roleMenu).Error
}

func (r *RoleRepo) RemoveRoleMenus(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}

func (r *RoleRepo) GetRoleMenus(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Model(&model.RoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *RoleRepo) AddRolePermission(rolePerm *model.RolePermission) error {
	return r.db.Create(rolePerm).Error
}

func (r *RoleRepo) RemoveRolePermissions(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

func (r *RoleRepo) GetRolePermissions(roleID int64) ([]int64, error) {
	var permIDs []int64
	err := r.db.Model(&model.RolePermission{}).Where("role_id = ?", roleID).Pluck("permission_id", &permIDs).Error
	return permIDs, err
}

func (r *RoleRepo) HasUsers(roleID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	return count > 0, err
}
```

- [ ] **Step 4: 创建Menu Repository**

```go
// backend/internal/repository/menu_repo.go
package repository

import (
	"github.com/jinang/grbac/internal/model"
	"gorm.io/gorm"
)

type MenuRepo struct {
	db *gorm.DB
}

func NewMenuRepo(db *gorm.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

func (r *MenuRepo) Create(menu *model.Menu) error {
	return r.db.Create(menu).Error
}

func (r *MenuRepo) GetByID(id int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.First(&menu, id).Error
	return &menu, err
}

func (r *MenuRepo) ListBySystem(systemID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("system_id = ?", systemID).Order("sort_order ASC").Find(&menus).Error
	return menus, err
}

func (r *MenuRepo) GetByParentID(systemID, parentID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("system_id = ? AND parent_id = ?", systemID, parentID).Order("sort_order ASC").Find(&menus).Error
	return menus, err
}

func (r *MenuRepo) HasChildren(systemID, menuID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Menu{}).Where("system_id = ? AND parent_id = ?", systemID, menuID).Count(&count).Error
	return count > 0, err
}

func (r *MenuRepo) Update(menu *model.Menu) error {
	return r.db.Save(menu).Error
}

func (r *MenuRepo) Delete(id int64) error {
	return r.db.Delete(&model.Menu{}, id).Error
}
```

- [ ] **Step 5: 创建Permission Repository**

```go
// backend/internal/repository/permission_repo.go
package repository

import (
	"github.com/jinang/grbac/internal/model"
	"gorm.io/gorm"
)

type PermissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

func (r *PermissionRepo) Create(perm *model.Permission) error {
	return r.db.Create(perm).Error
}

func (r *PermissionRepo) GetByID(id int64) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.First(&perm, id).Error
	return &perm, err
}

func (r *PermissionRepo) GetByMethodPath(systemID int64, method, path string) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("system_id = ? AND method = ? AND path = ?", systemID, method, path).First(&perm).Error
	return &perm, err
}

func (r *PermissionRepo) ListBySystem(systemID int64, page, pageSize int) ([]model.Permission, int64, error) {
	var perms []model.Permission
	var total int64

	r.db.Model(&model.Permission{}).Where("system_id = ?", systemID).Count(&total)
	err := r.db.Where("system_id = ?", systemID).Offset((page - 1) * pageSize).Limit(pageSize).Order("id ASC").Find(&perms).Error
	return perms, total, err
}

func (r *PermissionRepo) Update(perm *model.Permission) error {
	return r.db.Save(perm).Error
}

func (r *PermissionRepo) Delete(id int64) error {
	return r.db.Delete(&model.Permission{}, id).Error
}

func (r *PermissionRepo) IsUsedByRole(permID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.RolePermission{}).Where("permission_id = ?", permID).Count(&count).Error
	return count > 0, err
}
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/repository/
git commit -m "feat: add repository layer for all entities"
```

---

## Phase 3: Service层与认证

### Task 6: 认证Service

**Files:**
- Create: `backend/internal/service/auth_service.go`

- [ ] **Step 1: 创建认证Service**

```go
// backend/internal/service/auth_service.go
package service

import (
	"context"
	"time"

	"github.com/jinang/grbac/internal/config"
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/crypto"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/repository"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	userRepo   *repository.UserRepo
	jwtManager *jwt.JWTManager
	redis      *redis.Client
	passwordCfg config.PasswordConfig
}

func NewAuthService(userRepo *repository.UserRepo, jwtManager *jwt.JWTManager, redis *redis.Client, passwordCfg config.PasswordConfig) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtManager:  jwtManager,
		redis:       redis,
		passwordCfg: passwordCfg,
	}
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.ErrInvalidPassword
	}

	// 检查账号是否被锁定
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, errors.ErrAccountLocked
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.ErrInvalidPassword
	}

	// 验证密码
	if !crypto.CheckPassword(password, user.PasswordHash) {
		// 增加登录失败次数
		user.LoginAttempts++
		if user.LoginAttempts >= s.passwordCfg.MaxAttempts {
			lockUntil := time.Now().Add(time.Duration(s.passwordCfg.LockDuration) * time.Second)
			user.LockedUntil = &lockUntil
		}
		s.userRepo.Update(user)
		return nil, errors.ErrInvalidPassword
	}

	// 登录成功，重置失败次数
	user.LoginAttempts = 0
	user.LockedUntil = nil
	s.userRepo.Update(user)

	// 生成Token
	accessToken, refreshToken, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.IsSuperAdmin)
	if err != nil {
		return nil, errors.ErrInternal
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    7200,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.Status != 1 {
		return nil, errors.ErrInvalidPassword
	}

	accessToken, newRefreshToken, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.IsSuperAdmin)
	if err != nil {
		return nil, errors.ErrInternal
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    7200,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	// 将Token加入黑名单
	claims, err := s.jwtManager.ParseToken(accessToken)
	if err != nil {
		return nil
	}

	// 计算Token剩余有效期
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		s.redis.Set(ctx, "token:blacklist:"+accessToken, "1", remaining)
	}

	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if !crypto.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.ErrInvalidPassword
	}

	newHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errors.ErrInternal
	}

	user.PasswordHash = newHash
	return s.userRepo.Update(user)
}

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, accessToken string) bool {
	exists, _ := s.redis.Exists(ctx, "token:blacklist:"+accessToken).Result()
	return exists > 0
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/
git commit -m "feat: add authentication service"
```

---

### Task 7: 用户Service

**Files:**
- Create: `backend/internal/service/user_service.go`

- [ ] **Step 1: 创建用户Service**

```go
// backend/internal/service/user_service.go
package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/crypto"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type UpdateUserRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (s *UserService) Create(req *CreateUserRequest) (*model.User, error) {
	// 检查用户名是否存在
	existing, _ := s.userRepo.GetByUsername(req.Username)
	if existing != nil && existing.ID > 0 {
		return nil, errors.ErrUsernameExists
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrInternal
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.ErrInternal
	}

	return user, nil
}

func (s *UserService) GetByID(id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.List(page, pageSize)
}

func (s *UserService) Update(id int64, req *UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	user.Email = req.Email
	user.Phone = req.Phone

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.ErrInternal
	}

	return user, nil
}

func (s *UserService) Delete(id int64) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}
	return s.userRepo.Delete(id)
}

func (s *UserService) UpdateStatus(id int64, status int8) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}
	user.Status = status
	return s.userRepo.Update(user)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/user_service.go
git commit -m "feat: add user service"
```

---

### Task 8: 系统Service

**Files:**
- Create: `backend/internal/service/system_service.go`

- [ ] **Step 1: 创建系统Service**

```go
// backend/internal/service/system_service.go
package service

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

type SystemService struct {
	systemRepo *repository.SystemRepo
	userRepo   *repository.UserRepo
}

func NewSystemService(systemRepo *repository.SystemRepo, userRepo *repository.UserRepo) *SystemService {
	return &SystemService{
		systemRepo: systemRepo,
		userRepo:   userRepo,
	}
}

type CreateSystemRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

func (s *SystemService) Create(req *CreateSystemRequest) (*model.System, error) {
	// 生成系统密钥
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	secret := hex.EncodeToString(secretBytes)

	system := &model.System{
		Name:        req.Name,
		Code:        req.Code,
		Secret:      secret,
		Description: req.Description,
		Status:      1,
	}

	if err := s.systemRepo.Create(system); err != nil {
		return nil, errors.ErrInternal
	}

	return system, nil
}

func (s *SystemService) GetByID(id int64) (*model.System, error) {
	system, err := s.systemRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}
	return system, nil
}

func (s *SystemService) List(page, pageSize int) ([]model.System, int64, error) {
	return s.systemRepo.List(page, pageSize)
}

func (s *SystemService) Update(id int64, name, description string) (*model.System, error) {
	system, err := s.systemRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	system.Name = name
	system.Description = description

	if err := s.systemRepo.Update(system); err != nil {
		return nil, errors.ErrInternal
	}

	return system, nil
}

func (s *SystemService) Delete(id int64) error {
	_, err := s.systemRepo.GetByID(id)
	if err != nil {
		return errors.ErrSystemNotFound
	}
	return s.systemRepo.Delete(id)
}

func (s *SystemService) AddMember(systemID, userID int64, role string) error {
	_, err := s.systemRepo.GetByID(systemID)
	if err != nil {
		return errors.ErrSystemNotFound
	}

	_, err = s.userRepo.GetByID(userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	member := &model.SystemMember{
		SystemID: systemID,
		UserID:   userID,
		Role:     role,
	}

	return s.systemRepo.AddMember(member)
}

func (s *SystemService) RemoveMember(systemID, userID int64) error {
	return s.systemRepo.RemoveMember(systemID, userID)
}

func (s *SystemService) GetMembers(systemID int64) ([]model.SystemMember, error) {
	return s.systemRepo.GetMembers(systemID)
}

func (s *SystemService) IsAdmin(systemID, userID int64) (bool, error) {
	return s.systemRepo.IsAdmin(systemID, userID)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/system_service.go
git commit -m "feat: add system service"
```

---

### Task 9: 角色Service

**Files:**
- Create: `backend/internal/service/role_service.go`

- [ ] **Step 1: 创建角色Service**

```go
// backend/internal/service/role_service.go
package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

type RoleService struct {
	roleRepo *repository.RoleRepo
	userRepo *repository.UserRepo
}

func NewRoleService(roleRepo *repository.RoleRepo, userRepo *repository.UserRepo) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

func (s *RoleService) Create(systemID int64, req *CreateRoleRequest) (*model.Role, error) {
	// 检查角色编码是否存在
	existing, _ := s.roleRepo.GetByCode(systemID, req.Code)
	if existing != nil && existing.ID > 0 {
		return nil, errors.ErrRoleCodeExists
	}

	role := &model.Role{
		SystemID:    systemID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      1,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.ErrInternal
	}

	return role, nil
}

func (s *RoleService) GetByID(id int64) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrRoleNotFound
	}
	return role, nil
}

func (s *RoleService) ListBySystem(systemID int64) ([]model.Role, error) {
	return s.roleRepo.ListBySystem(systemID)
}

func (s *RoleService) Update(id int64, name, description string) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrRoleNotFound
	}

	role.Name = name
	role.Description = description

	if err := s.roleRepo.Update(role); err != nil {
		return nil, errors.ErrInternal
	}

	return role, nil
}

func (s *RoleService) Delete(id int64) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	// 检查是否有用户绑定
	hasUsers, _ := s.roleRepo.HasUsers(role.ID)
	if hasUsers {
		return errors.New(30003, "角色下存在用户，无法删除")
	}

	// 删除角色关联的菜单和权限
	s.roleRepo.RemoveRoleMenus(id)
	s.roleRepo.RemoveRolePermissions(id)

	return s.roleRepo.Delete(id)
}

func (s *RoleService) AssignMenus(roleID int64, menuIDs []int64) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	// 先删除原有菜单
	s.roleRepo.RemoveRoleMenus(roleID)

	// 添加新菜单
	for _, menuID := range menuIDs {
		roleMenu := &model.RoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		}
		s.roleRepo.AddRoleMenu(roleMenu)
	}

	return nil
}

func (s *RoleService) GetRoleMenus(roleID int64) ([]int64, error) {
	return s.roleRepo.GetRoleMenus(roleID)
}

func (s *RoleService) AssignPermissions(roleID int64, permIDs []int64) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	// 先删除原有权限
	s.roleRepo.RemoveRolePermissions(roleID)

	// 添加新权限
	for _, permID := range permIDs {
		rolePerm := &model.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
		s.roleRepo.AddRolePermission(rolePerm)
	}

	return nil
}

func (s *RoleService) GetRolePermissions(roleID int64) ([]int64, error) {
	return s.roleRepo.GetRolePermissions(roleID)
}

func (s *RoleService) AssignUsers(roleID int64, userIDs []int64) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	for _, userID := range userIDs {
		userRole := &model.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		s.userRepo.AddUserRole(userRole)
	}

	return nil
}

func (s *RoleService) RemoveUser(roleID, userID int64) error {
	return s.userRepo.RemoveUserRole(userID, roleID)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/role_service.go
git commit -m "feat: add role service"
```

---

### Task 10: 菜单Service

**Files:**
- Create: `backend/internal/service/menu_service.go`

- [ ] **Step 1: 创建菜单Service**

```go
// backend/internal/service/menu_service.go
package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

type MenuService struct {
	menuRepo *repository.MenuRepo
}

func NewMenuService(menuRepo *repository.MenuRepo) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

type CreateMenuRequest struct {
	ParentID  int64  `json:"parent_id"`
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

type MenuTree struct {
	model.Menu
	Children []*MenuTree `json:"children"`
}

func (s *MenuService) Create(systemID int64, req *CreateMenuRequest) (*model.Menu, error) {
	menu := &model.Menu{
		SystemID:  systemID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		Status:    1,
	}

	if err := s.menuRepo.Create(menu); err != nil {
		return nil, errors.ErrInternal
	}

	return menu, nil
}

func (s *MenuService) GetByID(id int64) (*model.Menu, error) {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrMenuNotFound
	}
	return menu, nil
}

func (s *MenuService) GetTree(systemID int64) ([]*MenuTree, error) {
	menus, err := s.menuRepo.ListBySystem(systemID)
	if err != nil {
		return nil, err
	}

	return buildTree(menus, 0), nil
}

func buildTree(menus []model.Menu, parentID int64) []*MenuTree {
	var tree []*MenuTree
	for _, menu := range menus {
		if menu.ParentID == parentID {
			node := &MenuTree{
				Menu:     menu,
				Children: buildTree(menus, menu.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

func (s *MenuService) Update(id int64, req *CreateMenuRequest) (*model.Menu, error) {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrMenuNotFound
	}

	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Path = req.Path
	menu.Icon = req.Icon
	menu.SortOrder = req.SortOrder

	if err := s.menuRepo.Update(menu); err != nil {
		return nil, errors.ErrInternal
	}

	return menu, nil
}

func (s *MenuService) Delete(systemID, id int64) error {
	_, err := s.menuRepo.GetByID(id)
	if err != nil {
		return errors.ErrMenuNotFound
	}

	// 检查是否有子菜单
	hasChildren, _ := s.menuRepo.HasChildren(systemID, id)
	if hasChildren {
		return errors.ErrMenuHasChildren
	}

	return s.menuRepo.Delete(id)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/menu_service.go
git commit -m "feat: add menu service"
```

---

### Task 11: 权限Service

**Files:**
- Create: `backend/internal/service/permission_service.go`

- [ ] **Step 1: 创建权限Service**

```go
// backend/internal/service/permission_service.go
package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

type PermissionService struct {
	permRepo *repository.PermissionRepo
}

func NewPermissionService(permRepo *repository.PermissionRepo) *PermissionService {
	return &PermissionService{permRepo: permRepo}
}

type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Description string `json:"description"`
}

func (s *PermissionService) Create(systemID int64, req *CreatePermissionRequest) (*model.Permission, error) {
	// 检查method+path是否已存在
	existing, _ := s.permRepo.GetByMethodPath(systemID, req.Method, req.Path)
	if existing != nil && existing.ID > 0 {
		return nil, errors.ErrPermPathExists
	}

	perm := &model.Permission{
		SystemID:    systemID,
		Code:        req.Code,
		Name:        req.Name,
		Method:      req.Method,
		Path:        req.Path,
		Description: req.Description,
		Version:     1,
	}

	if err := s.permRepo.Create(perm); err != nil {
		return nil, errors.ErrInternal
	}

	return perm, nil
}

func (s *PermissionService) GetByID(id int64) (*model.Permission, error) {
	perm, err := s.permRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrPermNotFound
	}
	return perm, nil
}

func (s *PermissionService) ListBySystem(systemID int64, page, pageSize int) ([]model.Permission, int64, error) {
	return s.permRepo.ListBySystem(systemID, page, pageSize)
}

func (s *PermissionService) Update(id int64, req *CreatePermissionRequest) (*model.Permission, error) {
	perm, err := s.permRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrPermNotFound
	}

	perm.Code = req.Code
	perm.Name = req.Name
	perm.Method = req.Method
	perm.Path = req.Path
	perm.Description = req.Description
	perm.Version++

	if err := s.permRepo.Update(perm); err != nil {
		return nil, errors.ErrInternal
	}

	return perm, nil
}

func (s *PermissionService) Delete(id int64) error {
	_, err := s.permRepo.GetByID(id)
	if err != nil {
		return errors.ErrPermNotFound
	}

	// 检查是否被角色使用
	isUsed, _ := s.permRepo.IsUsedByRole(id)
	if isUsed {
		return errors.New(50003, "权限已被角色使用，无法删除")
	}

	return s.permRepo.Delete(id)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/service/permission_service.go
git commit -m "feat: add permission service"
```

---

## Phase 4: HTTP Handler与路由

### Task 12: 认证Handler

**Files:**
- Create: `backend/internal/handler/auth_handler.go`

- [ ] **Step 1: 创建认证Handler**

```go
// backend/internal/handler/auth_handler.go
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	accessToken := c.GetString("access_token")
	h.authService.Logout(c.Request.Context(), accessToken)
	response.OK(c, nil)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/handler/auth_handler.go
git commit -m "feat: add authentication handler"
```

---

### Task 13: 用户Handler

**Files:**
- Create: `backend/internal/handler/user_handler.go`

- [ ] **Step 1: 创建用户Handler**

```go
// backend/internal/handler/user_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	user, err := h.userService.Create(&req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := h.userService.GetByID(id)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.List(page, pageSize)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OKPage(c, total, users)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	user, err := h.userService.Update(id, &req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.userService.Delete(id); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	if err := h.userService.UpdateStatus(id, req.Status); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/handler/user_handler.go
git commit -m "feat: add user handler"
```

---

### Task 14: 系统Handler

**Files:**
- Create: `backend/internal/handler/system_handler.go`

- [ ] **Step 1: 创建系统Handler**

```go
// backend/internal/handler/system_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type SystemHandler struct {
	systemService *service.SystemService
}

func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

func (h *SystemHandler) Create(c *gin.Context) {
	var req service.CreateSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	system, err := h.systemService.Create(&req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, system)
}

func (h *SystemHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	systems, total, err := h.systemService.List(page, pageSize)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OKPage(c, total, systems)
}

func (h *SystemHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	system, err := h.systemService.GetByID(id)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, system)
}

func (h *SystemHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	system, err := h.systemService.Update(id, req.Name, req.Description)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, system)
}

func (h *SystemHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.systemService.Delete(id); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *SystemHandler) AddMember(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		UserID int64  `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	if err := h.systemService.AddMember(systemID, req.UserID, req.Role); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *SystemHandler) RemoveMember(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("uid"), 10, 64)

	if err := h.systemService.RemoveMember(systemID, userID); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *SystemHandler) GetMembers(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	members, err := h.systemService.GetMembers(systemID)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OK(c, members)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/handler/system_handler.go
git commit -m "feat: add system handler"
```

---

### Task 15: 角色、菜单、权限Handler

**Files:**
- Create: `backend/internal/handler/role_handler.go`
- Create: `backend/internal/handler/menu_handler.go`
- Create: `backend/internal/handler/permission_handler.go`

- [ ] **Step 1: 创建角色Handler**

```go
// backend/internal/handler/role_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type RoleHandler struct {
	roleService *service.RoleService
}

func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) Create(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)

	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	role, err := h.roleService.Create(systemID, &req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, role)
}

func (h *RoleHandler) List(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)

	roles, err := h.roleService.ListBySystem(systemID)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OK(c, roles)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	role, err := h.roleService.Update(id, req.Name, req.Description)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.roleService.Delete(id); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *RoleHandler) AssignMenus(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		MenuIDs []int64 `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	if err := h.roleService.AssignMenus(roleID, req.MenuIDs); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	menuIDs, err := h.roleService.GetRoleMenus(roleID)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OK(c, menuIDs)
}

func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		PermissionIDs []int64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	if err := h.roleService.AssignPermissions(roleID, req.PermissionIDs); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	permIDs, err := h.roleService.GetRolePermissions(roleID)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OK(c, permIDs)
}

func (h *RoleHandler) AssignUsers(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		UserIDs []int64 `json:"user_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	if err := h.roleService.AssignUsers(roleID, req.UserIDs); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}

func (h *RoleHandler) RemoveUser(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("uid"), 10, 64)

	if err := h.roleService.RemoveUser(roleID, userID); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}
```

- [ ] **Step 2: 创建菜单Handler**

```go
// backend/internal/handler/menu_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type MenuHandler struct {
	menuService *service.MenuService
}

func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

func (h *MenuHandler) Create(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)

	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	menu, err := h.menuService.Create(systemID, &req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, menu)
}

func (h *MenuHandler) GetTree(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)

	tree, err := h.menuService.GetTree(systemID)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OK(c, tree)
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	menu, err := h.menuService.Update(id, &req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, menu)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.menuService.Delete(systemID, id); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}
```

- [ ] **Step 3: 创建权限Handler**

```go
// backend/internal/handler/permission_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type PermissionHandler struct {
	permService *service.PermissionService
}

func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

func (h *PermissionHandler) Create(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)

	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	perm, err := h.permService.Create(systemID, &req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, perm)
}

func (h *PermissionHandler) List(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	perms, total, err := h.permService.ListBySystem(systemID, page, pageSize)
	if err != nil {
		response.Error(c, errors.ErrInternal)
		return
	}

	response.OKPage(c, total, perms)
}

func (h *PermissionHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(10000, "参数错误"))
		return
	}

	perm, err := h.permService.Update(id, &req)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, perm)
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.permService.Delete(id); err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, nil)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/
git commit -m "feat: add role, menu, and permission handlers"
```

---

### Task 16: 中间件与路由

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/permission.go`
- Create: `backend/internal/middleware/cors.go`
- Create: `backend/internal/router/router.go`

- [ ] **Step 1: 创建认证中间件**

```go
// backend/internal/middleware/auth.go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

func AuthMiddleware(jwtManager *jwt.JWTManager, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		accessToken := parts[1]

		// 检查Token是否在黑名单中
		if authService.IsTokenBlacklisted(c.Request.Context(), accessToken) {
			response.Error(c, errors.ErrTokenExpired)
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(accessToken)
		if err != nil {
			response.Error(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_super_admin", claims.IsSuperAdmin)
		c.Set("access_token", accessToken)
		c.Next()
	}
}

func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperAdmin := c.GetBool("is_super_admin")
		if !isSuperAdmin {
			response.Error(c, errors.ErrNoPermission)
			c.Abort()
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: 创建系统管理员中间件**

```go
// backend/internal/middleware/permission.go
package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

func SystemAdminMiddleware(systemService *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		systemID, _ := strconv.ParseInt(c.Param("sid"), 10, 64)

		isAdmin, err := systemService.IsAdmin(systemID, userID)
		if err != nil || !isAdmin {
			response.Error(c, errors.ErrNoPermission)
			c.Abort()
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: 创建CORS中间件**

```go
// backend/internal/middleware/cors.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 4: 创建路由配置**

```go
// backend/internal/router/router.go
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/handler"
	"github.com/jinang/grbac/internal/middleware"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/service"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	systemHandler *handler.SystemHandler,
	roleHandler *handler.RoleHandler,
	menuHandler *handler.MenuHandler,
	permHandler *handler.PermissionHandler,
	jwtManager *jwt.JWTManager,
	authService *service.AuthService,
	systemService *service.SystemService,
) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 认证相关（无需登录）
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// 需要登录的接口
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtManager, authService))
	{
		// 登出和修改密码
		api.POST("/auth/logout", authHandler.Logout)
		api.POST("/auth/change-password", authHandler.ChangePassword)

		// 用户管理（超级管理员）
		users := api.Group("/users")
		users.Use(middleware.SuperAdminMiddleware())
		{
			users.GET("", userHandler.List)
			users.POST("", userHandler.Create)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
			users.PUT("/:id/status", userHandler.UpdateStatus)
		}

		// 系统管理（超级管理员）
		systems := api.Group("/systems")
		systems.Use(middleware.SuperAdminMiddleware())
		{
			systems.GET("", systemHandler.List)
			systems.POST("", systemHandler.Create)
			systems.GET("/:id", systemHandler.GetByID)
			systems.PUT("/:id", systemHandler.Update)
			systems.DELETE("/:id", systemHandler.Delete)
			systems.GET("/:id/members", systemHandler.GetMembers)
			systems.POST("/:id/members", systemHandler.AddMember)
			systems.DELETE("/:id/members/:uid", systemHandler.RemoveMember)
		}

		// 系统管理员接口
		sysAdmin := api.Group("/systems/:sid")
		sysAdmin.Use(middleware.SystemAdminMiddleware(systemService))
		{
			// 角色管理
			sysAdmin.GET("/roles", roleHandler.List)
			sysAdmin.POST("/roles", roleHandler.Create)
			sysAdmin.PUT("/roles/:id", roleHandler.Update)
			sysAdmin.DELETE("/roles/:id", roleHandler.Delete)
			sysAdmin.POST("/roles/:id/menus", roleHandler.AssignMenus)
			sysAdmin.GET("/roles/:id/menus", roleHandler.GetRoleMenus)
			sysAdmin.POST("/roles/:id/permissions", roleHandler.AssignPermissions)
			sysAdmin.GET("/roles/:id/permissions", roleHandler.GetRolePermissions)
			sysAdmin.POST("/roles/:id/users", roleHandler.AssignUsers)
			sysAdmin.DELETE("/roles/:id/users/:uid", roleHandler.RemoveUser)

			// 菜单管理
			sysAdmin.GET("/menus", menuHandler.GetTree)
			sysAdmin.POST("/menus", menuHandler.Create)
			sysAdmin.PUT("/menus/:id", menuHandler.Update)
			sysAdmin.DELETE("/menus/:id", menuHandler.Delete)

			// 权限管理
			sysAdmin.GET("/permissions", permHandler.List)
			sysAdmin.POST("/permissions", permHandler.Create)
			sysAdmin.PUT("/permissions/:id", permHandler.Update)
			sysAdmin.DELETE("/permissions/:id", permHandler.Delete)
		}
	}

	return r
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/middleware/ backend/internal/router/
git commit -m "feat: add middleware and router configuration"
```

---

## Phase 5: 外部API与SDK

### Task 17: 外部系统API

**Files:**
- Create: `backend/internal/handler/external_handler.go`
- Create: `backend/internal/service/external_service.go`

- [ ] **Step 1: 创建外部Service**

```go
// backend/internal/service/external_service.go
package service

import (
	"context"

	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/repository"
	"github.com/redis/go-redis/v9"
)

type ExternalService struct {
	systemRepo *repository.SystemRepo
	userRepo   *repository.UserRepo
	roleRepo   *repository.RoleRepo
	menuRepo   *repository.MenuRepo
	permRepo   *repository.PermissionRepo
	jwtManager *jwt.JWTManager
	redis      *redis.Client
}

func NewExternalService(
	systemRepo *repository.SystemRepo,
	userRepo *repository.UserRepo,
	roleRepo *repository.RoleRepo,
	menuRepo *repository.MenuRepo,
	permRepo *repository.PermissionRepo,
	jwtManager *jwt.JWTManager,
	redis *redis.Client,
) *ExternalService {
	return &ExternalService{
		systemRepo: systemRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		menuRepo:   menuRepo,
		permRepo:   permRepo,
		jwtManager: jwtManager,
		redis:      redis,
	}
}

type UserInfo struct {
	UserID       int64    `json:"user_id"`
	Username     string   `json:"username"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	Roles        []string `json:"roles"`
}

func (s *ExternalService) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ParseToken(token)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}
	return claims, nil
}

func (s *ExternalService) GetUserInfo(ctx context.Context, userID int64, systemCode string) (*UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	// 获取用户在该系统的角色
	userRoles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errors.ErrInternal
	}

	var roleNames []string
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ur.RoleID)
		if err == nil && role.SystemID == system.ID {
			roleNames = append(roleNames, role.Code)
		}
	}

	return &UserInfo{
		UserID:       user.ID,
		Username:     user.Username,
		IsSuperAdmin: user.IsSuperAdmin,
		Roles:        roleNames,
	}, nil
}

func (s *ExternalService) GetUserMenus(ctx context.Context, userID int64, systemCode string) ([]model.Menu, error) {
	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	// 获取用户所有角色
	userRoles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errors.ErrInternal
	}

	// 收集所有菜单ID
	menuIDSet := make(map[int64]bool)
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ur.RoleID)
		if err != nil || role.SystemID != system.ID {
			continue
		}

		menuIDs, _ := s.roleRepo.GetRoleMenus(role.ID)
		for _, id := range menuIDs {
			menuIDSet[id] = true
		}
	}

	// 查询菜单详情
	var menus []model.Menu
	for menuID := range menuIDSet {
		menu, err := s.menuRepo.GetByID(menuID)
		if err == nil && menu.Status == 1 {
			menus = append(menus, *menu)
		}
	}

	return menus, nil
}

func (s *ExternalService) GetUserPermissions(ctx context.Context, userID int64, systemCode string) ([]string, error) {
	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	// 获取用户所有角色
	userRoles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errors.ErrInternal
	}

	// 收集所有权限码
	permCodeSet := make(map[string]bool)
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ur.RoleID)
		if err != nil || role.SystemID != system.ID {
			continue
		}

		permIDs, _ := s.roleRepo.GetRolePermissions(role.ID)
		for _, id := range permIDs {
			perm, err := s.permRepo.GetByID(id)
			if err == nil {
				permCodeSet[perm.Code] = true
			}
		}
	}

	var codes []string
	for code := range permCodeSet {
		codes = append(codes, code)
	}

	return codes, nil
}
```

- [ ] **Step 2: 创建外部Handler**

```go
// backend/internal/handler/external_handler.go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

type ExternalHandler struct {
	externalService *service.ExternalService
}

func NewExternalHandler(externalService *service.ExternalService) *ExternalHandler {
	return &ExternalHandler{externalService: externalService}
}

func (h *ExternalHandler) Verify(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Error(c, errors.ErrTokenInvalid)
		return
	}

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, claims)
}

func (h *ExternalHandler) GetUserInfo(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	systemCode := c.GetHeader("X-System-Code")

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	userInfo, err := h.externalService.GetUserInfo(c.Request.Context(), claims.UserID, systemCode)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, userInfo)
}

func (h *ExternalHandler) GetMenus(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	systemCode := c.GetHeader("X-System-Code")

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	menus, err := h.externalService.GetUserMenus(c.Request.Context(), claims.UserID, systemCode)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, menus)
}

func (h *ExternalHandler) GetPermissions(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	systemCode := c.GetHeader("X-System-Code")

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	perms, err := h.externalService.GetUserPermissions(c.Request.Context(), claims.UserID, systemCode)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	response.OK(c, perms)
}

func (h *ExternalHandler) ValidatePermission(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	systemCode := c.GetHeader("X-System-Code")
	requiredPerm := c.Query("permission")

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	perms, err := h.externalService.GetUserPermissions(c.Request.Context(), claims.UserID, systemCode)
	if err != nil {
		response.Error(c, err.(*errors.AppError))
		return
	}

	hasPermission := false
	for _, p := range perms {
		if p == requiredPerm {
			hasPermission = true
			break
		}
	}

	response.OK(c, gin.H{"has_permission": hasPermission})
}
```

- [ ] **Step 3: 更新路由添加外部API**

```go
// 在 router.go 中添加外部API路由

// 外部系统API（使用系统凭证认证）
external := r.Group("/api/external")
external.Use(middleware.ExternalAuthMiddleware(systemService))
{
    external.POST("/verify", externalHandler.Verify)
    external.GET("/userinfo", externalHandler.GetUserInfo)
    external.GET("/menus", externalHandler.GetMenus)
    external.GET("/permissions", externalHandler.GetPermissions)
    external.GET("/validate", externalHandler.ValidatePermission)
}
```

- [ ] **Step 4: 创建外部认证中间件**

```go
// backend/internal/middleware/external.go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

func ExternalAuthMiddleware(systemService *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		systemCode := c.GetHeader("X-System-Code")
		systemSecret := c.GetHeader("X-System-Secret")

		if systemCode == "" || systemSecret == "" {
			response.Error(c, errors.ErrSystemCredential)
			c.Abort()
			return
		}

		system, err := systemService.GetByCode(systemCode)
		if err != nil || system.Secret != systemSecret {
			response.Error(c, errors.ErrSystemCredential)
			c.Abort()
			return
		}

		c.Set("system_id", system.ID)
		c.Next()
	}
}

func (s *SystemService) GetByCode(code string) (*model.System, error) {
	return s.systemRepo.GetByCode(code)
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/external_handler.go backend/internal/service/external_service.go backend/internal/middleware/external.go
git commit -m "feat: add external system API endpoints"
```

---

## Phase 6: 主程序与Docker

### Task 18: 更新主程序

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 更新主程序**

```go
// backend/cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jinang/grbac/internal/config"
	"github.com/jinang/grbac/internal/database"
	"github.com/jinang/grbac/internal/handler"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/repository"
	"github.com/jinang/grbac/internal/router"
	"github.com/jinang/grbac/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redis, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	// 初始化Repository
	systemRepo := repository.NewSystemRepo(db)
	userRepo := repository.NewUserRepo(db)
	roleRepo := repository.NewRoleRepo(db)
	menuRepo := repository.NewMenuRepo(db)
	permRepo := repository.NewPermissionRepo(db)

	// 初始化工具
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpire, cfg.JWT.RefreshExpire)

	// 初始化Service
	authService := service.NewAuthService(userRepo, jwtManager, redis, cfg.Password)
	userService := service.NewUserService(userRepo)
	systemService := service.NewSystemService(systemRepo, userRepo)
	roleService := service.NewRoleService(roleRepo, userRepo)
	menuService := service.NewMenuService(menuRepo)
	permService := service.NewPermissionService(permRepo)
	externalService := service.NewExternalService(systemRepo, userRepo, roleRepo, menuRepo, permRepo, jwtManager, redis)

	// 初始化Handler
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	systemHandler := handler.NewSystemHandler(systemService)
	roleHandler := handler.NewRoleHandler(roleService)
	menuHandler := handler.NewMenuHandler(menuService)
	permHandler := handler.NewPermissionHandler(permService)
	externalHandler := handler.NewExternalHandler(externalService)

	// 设置路由
	r := router.SetupRouter(
		authHandler, userHandler, systemHandler,
		roleHandler, menuHandler, permHandler, externalHandler,
		jwtManager, authService, systemService,
	)

	// 启动服务
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Starting RBAC server on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "feat: update main program with dependency injection"
```

---

### Task 19: Docker配置

**Files:**
- Create: `backend/Dockerfile`
- Create: `frontend/Dockerfile`
- Create: `docker-compose.yaml`

- [ ] **Step 1: 创建后端Dockerfile**

```dockerfile
# backend/Dockerfile
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

- [ ] **Step 2: 创建前端Dockerfile**

```dockerfile
# frontend/Dockerfile
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
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Step 3: 创建docker-compose.yaml**

```yaml
# docker-compose.yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: grbac-mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: grbac
      MYSQL_CHARACTER_SET_SERVER: utf8mb4
      MYSQL_COLLATION_SERVER: utf8mb4_unicode_ci
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./backend/migrations:/docker-entrypoint-initdb.d
    networks:
      - grbac-network

  redis:
    image: redis:7-alpine
    container_name: grbac-redis
    ports:
      - "6379:6379"
    networks:
      - grbac-network

  rbac-server:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: grbac-server
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USERNAME=root
      - DB_PASSWORD=rootpassword
      - DB_NAME=grbac
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    networks:
      - grbac-network

  rbac-web:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: grbac-web
    ports:
      - "80:80"
    depends_on:
      - rbac-server
    networks:
      - grbac-network

volumes:
  mysql_data:

networks:
  grbac-network:
    driver: bridge
```

- [ ] **Step 4: Commit**

```bash
git add backend/Dockerfile frontend/Dockerfile docker-compose.yaml
git commit -m "feat: add Docker and docker-compose configuration"
```

---

## Phase 7: 前端项目

### Task 20: Vue项目初始化

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`

- [ ] **Step 1: 初始化Vue项目**

```bash
cd frontend && npm create vite@latest . -- --template vue-ts
npm install
npm install vue-router@4 pinia element-plus @element-plus/icons-vue axios
npm install -D @types/node
```

- [ ] **Step 2: 配置vite.config.ts**

```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 3: 创建main.ts**

```typescript
// frontend/src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'

const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
```

- [ ] **Step 4: Commit**

```bash
git add frontend/
git commit -m "feat: initialize Vue 3 project with Element Plus"
```

---

### Task 21: 前端工具与Store

**Files:**
- Create: `frontend/src/utils/request.ts`
- Create: `frontend/src/utils/token.ts`
- Create: `frontend/src/stores/auth.ts`
- Create: `frontend/src/stores/user.ts`
- Create: `frontend/src/stores/menu.ts`

- [ ] **Step 1: 创建Axios封装**

```typescript
// frontend/src/utils/request.ts
import axios from 'axios'
import { getToken, removeToken } from './token'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code === 0) {
      return data
    }
    if (code === 10003 || code === 10004) {
      removeToken()
      router.push('/login')
    }
    return Promise.reject(new Error(message))
  },
  (error) => {
    return Promise.reject(error)
  }
)

export default request
```

- [ ] **Step 2: 创建Token工具**

```typescript
// frontend/src/utils/token.ts
const TOKEN_KEY = 'grbac_access_token'
const REFRESH_TOKEN_KEY = 'grbac_refresh_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_TOKEN_KEY, token)
}
```

- [ ] **Step 3: 创建Auth Store**

```typescript
// frontend/src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'
import { setToken, setRefreshToken, removeToken, getToken } from '@/utils/token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())

  async function login(username: string, password: string) {
    const data: any = await request.post('/auth/login', { username, password })
    token.value = data.access_token
    setToken(data.access_token)
    setRefreshToken(data.refresh_token)
    return data
  }

  async function logout() {
    try {
      await request.post('/auth/logout')
    } finally {
      token.value = null
      removeToken()
    }
  }

  async function refreshToken() {
    const refreshToken = localStorage.getItem('grbac_refresh_token')
    if (!refreshToken) throw new Error('No refresh token')
    
    const data: any = await request.post('/auth/refresh', { refresh_token: refreshToken })
    token.value = data.access_token
    setToken(data.access_token)
    setRefreshToken(data.refresh_token)
    return data
  }

  return { token, login, logout, refreshToken }
})
```

- [ ] **Step 4: 创建User Store**

```typescript
// frontend/src/stores/user.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

interface UserInfo {
  id: number
  username: string
  email: string
  is_super_admin: boolean
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)

  async function fetchUserInfo() {
    const data: any = await request.get('/users/me')
    userInfo.value = data
    return data
  }

  function isSuperAdmin() {
    return userInfo.value?.is_super_admin ?? false
  }

  return { userInfo, fetchUserInfo, isSuperAdmin }
})
```

- [ ] **Step 5: 创建Menu Store**

```typescript
// frontend/src/stores/menu.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

interface MenuItem {
  id: number
  name: string
  path: string
  icon: string
  children?: MenuItem[]
}

export const useMenuStore = defineStore('menu', () => {
  const menus = ref<MenuItem[]>([])
  const currentSystemId = ref<number | null>(null)

  async function fetchMenus(systemId: number) {
    const data: any = await request.get(`/systems/${systemId}/menus`)
    menus.value = data
    currentSystemId.value = systemId
  }

  return { menus, currentSystemId, fetchMenus }
})
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/utils/ frontend/src/stores/
git commit -m "feat: add frontend utilities and Pinia stores"
```

---

### Task 22: 前端路由与布局

**Files:**
- Create: `frontend/src/router/index.ts`
- Create: `frontend/src/router/staticRoutes.ts`
- Create: `frontend/src/layouts/DefaultLayout.vue`
- Create: `frontend/src/views/login/LoginView.vue`

- [ ] **Step 1: 创建路由配置**

```typescript
// frontend/src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { staticRoutes } from './staticRoutes'
import { getToken } from '@/utils/token'

const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes,
})

router.beforeEach((to, from, next) => {
  const token = getToken()
  
  if (to.path === '/login') {
    next()
    return
  }
  
  if (!token) {
    next('/login')
    return
  }
  
  next()
})

export default router
```

- [ ] **Step 2: 创建静态路由**

```typescript
// frontend/src/router/staticRoutes.ts
import type { RouteRecordRaw } from 'vue-router'

export const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { title: '首页', icon: 'HomeFilled' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/ForbiddenView.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/NotFoundView.vue'),
  },
]
```

- [ ] **Step 3: 创建默认布局**

```vue
<!-- frontend/src/layouts/DefaultLayout.vue -->
<template>
  <el-container class="layout-container">
    <el-aside width="200px" class="aside">
      <div class="logo">RBAC管理系统</div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>
        <el-menu-item index="/system">
          <el-icon><Setting /></el-icon>
          <span>系统管理</span>
        </el-menu-item>
        <el-menu-item index="/user">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">RBAC管理系统</div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="el-dropdown-link">
              {{ userStore.userInfo?.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()

async function handleCommand(command: string) {
  if (command === 'logout') {
    await authStore.logout()
    router.push('/login')
  }
}

// 获取用户信息
userStore.fetchUserInfo()
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.aside {
  background-color: #304156;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #eee;
}

.header-right {
  cursor: pointer;
}

.main {
  background-color: #f5f7fa;
}
</style>
```

- [ ] **Step 4: 创建登录页面**

```vue
<!-- frontend/src/views/login/LoginView.vue -->
<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2 class="title">RBAC管理系统</h2>
      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleLogin" style="width: 100%">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error: any) {
    ElMessage.error(error.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
}

.title {
  text-align: center;
  margin-bottom: 30px;
  color: #303133;
}
</style>
```

- [ ] **Step 5: 创建占位页面**

```vue
<!-- frontend/src/views/dashboard/DashboardView.vue -->
<template>
  <div>
    <h2>欢迎使用RBAC管理系统</h2>
  </div>
</template>

<!-- frontend/src/views/error/ForbiddenView.vue -->
<template>
  <div class="error-page">
    <h1>403</h1>
    <p>抱歉，您没有权限访问此页面</p>
    <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
  </div>
</template>

<!-- frontend/src/views/error/NotFoundView.vue -->
<template>
  <div class="error-page">
    <h1>404</h1>
    <p>页面不存在</p>
    <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
  </div>
</template>
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/router/ frontend/src/layouts/ frontend/src/views/
git commit -m "feat: add router, layout, and login page"
```

---

## 完成

实现计划已完成。按照以上任务顺序执行，即可完成RBAC系统的后端和前端基础框架搭建。

**执行方式：**

1. **Subagent-Driven (推荐)** - 每个任务使用独立子代理执行，任务间进行审查
2. **Inline Execution** - 在当前会话中按顺序执行所有任务

选择哪种执行方式？
