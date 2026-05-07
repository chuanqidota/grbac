package errors

import "fmt"

// AppError represents a business-level application error.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Wrap returns a new AppError that wraps the original with additional context.
func (e *AppError) Wrap(detail string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: fmt.Sprintf("%s: %s", e.Message, detail),
	}
}

// IsAppError checks whether err is an *AppError.
func IsAppError(err error) (*AppError, bool) {
	if ae, ok := err.(*AppError); ok {
		return ae, true
	}
	return nil, false
}

// ---------- Predefined errors ----------

// Auth errors (10xxx)
var (
	Success             = &AppError{Code: 0, Message: "success"}
	ErrInvalidPassword  = &AppError{Code: 10001, Message: "用户名或密码错误"}
	ErrAccountLocked    = &AppError{Code: 10002, Message: "账号已被锁定"}
	ErrTokenExpired     = &AppError{Code: 10003, Message: "Token已过期"}
	ErrTokenInvalid     = &AppError{Code: 10004, Message: "Token无效"}
	ErrNoPermission     = &AppError{Code: 10005, Message: "无权限访问"}
)

// System errors (2xxx)
var (
	ErrSystemNotFound   = &AppError{Code: 20001, Message: "系统不存在"}
	ErrSystemCredential = &AppError{Code: 20002, Message: "系统凭证无效"}
)

// Role errors (3xxx)
var (
	ErrRoleNotFound   = &AppError{Code: 30001, Message: "角色不存在"}
	ErrRoleCodeExists = &AppError{Code: 30002, Message: "角色编码已存在"}
)

// Menu errors (4xxx)
var (
	ErrMenuNotFound  = &AppError{Code: 40001, Message: "菜单不存在"}
	ErrMenuHasChildren = &AppError{Code: 40002, Message: "存在子菜单，无法删除"}
)

// Permission errors (5xxx)
var (
	ErrPermNotFound   = &AppError{Code: 50001, Message: "权限不存在"}
	ErrPermPathExists = &AppError{Code: 50002, Message: "权限路径已存在"}
)

// User errors (6xxx)
var (
	ErrUserNotFound   = &AppError{Code: 60001, Message: "用户不存在"}
	ErrUsernameExists = &AppError{Code: 60002, Message: "用户名已存在"}
)

// System internal error (99xxx)
var (
	ErrInternal = &AppError{Code: 99999, Message: "系统内部错误"}
)
