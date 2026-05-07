package model

import "time"

// User represents a user account in the RBAC system.
type User struct {
	ID            int64      `json:"id"             gorm:"column:id;primaryKey;autoIncrement"`
	Username      string     `json:"username"       gorm:"column:username;size:50;uniqueIndex;not null"`
	PasswordHash  string     `json:"-"              gorm:"column:password_hash;size:128;not null"`
	Email         string     `json:"email"          gorm:"column:email;size:100"`
	Phone         string     `json:"phone"          gorm:"column:phone;size:20"`
	IsSuperAdmin  int8       `json:"is_super_admin" gorm:"column:is_super_admin;not null;default:0"`
	Status        int8       `json:"status"         gorm:"column:status;not null;default:1"`
	LoginAttempts int        `json:"login_attempts" gorm:"column:login_attempts;not null;default:0"`
	LockedUntil   *time.Time `json:"locked_until"   gorm:"column:locked_until"`
	CreatedAt     time.Time  `json:"created_at"     gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at"     gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

// UserRole represents the association between a user and a role.
type UserRole struct {
	ID        int64     `json:"id"         gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id"    gorm:"column:user_id;not null;uniqueIndex:uk_user_role"`
	RoleID    int64     `json:"role_id"    gorm:"column:role_id;not null;uniqueIndex:uk_user_role"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
