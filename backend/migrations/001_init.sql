-- 001_init.sql
-- Initial schema for grbac RBAC system

CREATE TABLE IF NOT EXISTS systems (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(100) NOT NULL COMMENT '系统名称',
    code            VARCHAR(50)  NOT NULL UNIQUE COMMENT '系统编码',
    secret          VARCHAR(128) NOT NULL COMMENT '系统密钥',
    description     VARCHAR(500) COMMENT '系统描述',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统表';

CREATE TABLE IF NOT EXISTS users (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(50)  NOT NULL UNIQUE COMMENT '用户名',
    password_hash   VARCHAR(128) NOT NULL COMMENT '密码哈希',
    email           VARCHAR(100) COMMENT '邮箱',
    phone           VARCHAR(20)  COMMENT '手机号',
    is_super_admin  TINYINT      NOT NULL DEFAULT 0 COMMENT '是否超级管理员',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    login_attempts  INT          NOT NULL DEFAULT 0 COMMENT '登录失败次数',
    locked_until    DATETIME     COMMENT '锁定截止时间',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS system_members (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT      NOT NULL COMMENT '系统ID',
    user_id         BIGINT      NOT NULL COMMENT '用户ID',
    role            VARCHAR(20) NOT NULL COMMENT '成员角色：admin/member',
    created_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_user (system_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统成员表';

CREATE TABLE IF NOT EXISTS roles (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT      NOT NULL COMMENT '所属系统ID',
    name            VARCHAR(50) NOT NULL COMMENT '角色名称',
    code            VARCHAR(50) NOT NULL COMMENT '角色编码',
    description     VARCHAR(500) COMMENT '角色描述',
    version         INT         NOT NULL DEFAULT 1 COMMENT '版本号（乐观锁）',
    status          TINYINT     NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_code (system_id, code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS menus (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT       NOT NULL COMMENT '所属系统ID',
    parent_id       BIGINT       NOT NULL DEFAULT 0 COMMENT '父菜单ID，0表示顶级',
    name            VARCHAR(50)  NOT NULL COMMENT '菜单名称',
    path            VARCHAR(200) COMMENT '路由路径',
    icon            VARCHAR(50)  COMMENT '菜单图标',
    sort_order      INT          NOT NULL DEFAULT 0 COMMENT '排序序号',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

CREATE TABLE IF NOT EXISTS permissions (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT       NOT NULL COMMENT '所属系统ID',
    code            VARCHAR(100) NOT NULL COMMENT '权限标识',
    name            VARCHAR(100) NOT NULL COMMENT '权限名称',
    method          VARCHAR(10)  NOT NULL COMMENT 'HTTP方法：GET/POST/PUT/DELETE',
    path            VARCHAR(200) NOT NULL COMMENT 'API路径',
    description     VARCHAR(500) COMMENT '权限描述',
    version         INT          NOT NULL DEFAULT 1 COMMENT '版本号',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_system_method_path (system_id, method, path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API权限表';

CREATE TABLE IF NOT EXISTS user_roles (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id         BIGINT   NOT NULL COMMENT '用户ID',
    role_id         BIGINT   NOT NULL COMMENT '角色ID',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-角色关联表';

CREATE TABLE IF NOT EXISTS role_menus (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id         BIGINT   NOT NULL COMMENT '角色ID',
    menu_id         BIGINT   NOT NULL COMMENT '菜单ID',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_menu (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-菜单关联表';

CREATE TABLE IF NOT EXISTS role_permissions (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id         BIGINT   NOT NULL COMMENT '角色ID',
    permission_id   BIGINT   NOT NULL COMMENT '权限ID',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-权限关联表';

CREATE TABLE IF NOT EXISTS webhooks (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    system_id       BIGINT       NOT NULL COMMENT '关联系统ID',
    url             VARCHAR(255) NOT NULL COMMENT '回调地址',
    secret          VARCHAR(128) NOT NULL COMMENT '签名密钥',
    events          VARCHAR(255) NOT NULL COMMENT '订阅事件：permission_change,menu_change,role_change',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Webhook配置表';

-- ============================================================
-- Seed data: initial super admin
-- Username: admin  Password: admin123
-- ============================================================
INSERT INTO users (username, password_hash, email, is_super_admin, status)
VALUES ('admin', '$2a$10$TIX.8yiuDKqeZPPUzBOWo.pszQ7Ybz9ZN6XrdaXfBQxcdCb7gFUrO', 'admin@grbac.local', 1, 1);

CREATE TABLE IF NOT EXISTS audit_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id         BIGINT       NOT NULL COMMENT '操作人ID',
    username        VARCHAR(50)  NOT NULL COMMENT '操作人用户名',
    system_id       BIGINT       COMMENT '操作的系统ID（可空）',
    action          VARCHAR(50)  NOT NULL COMMENT '操作类型：create/update/delete/assign',
    resource        VARCHAR(50)  NOT NULL COMMENT '资源类型：user/role/menu/permission/system',
    resource_id     BIGINT       COMMENT '资源ID',
    resource_name   VARCHAR(100) COMMENT '资源名称',
    detail          TEXT         COMMENT '变更详情JSON',
    ip              VARCHAR(50)  COMMENT '操作IP',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';
