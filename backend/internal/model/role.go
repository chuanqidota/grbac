package model

import "time"

// Role represents a role within a system.
type Role struct {
	ID          int64     `json:"id"          gorm:"column:id;primaryKey;autoIncrement"`
	SystemID    int64     `json:"system_id"   gorm:"column:system_id;not null;uniqueIndex:uk_system_code"`
	Name        string    `json:"name"        gorm:"column:name;size:50;not null"`
	Code        string    `json:"code"        gorm:"column:code;size:50;not null;uniqueIndex:uk_system_code"`
	Description string    `json:"description" gorm:"column:description;size:500"`
	Version     int       `json:"version"     gorm:"column:version;not null;default:1"`
	Status      int8      `json:"status"      gorm:"column:status;not null;default:1"`
	CreatedAt   time.Time `json:"created_at"  gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at"  gorm:"column:updated_at;autoUpdateTime"`
}

func (Role) TableName() string {
	return "roles"
}

// RoleMenu represents the association between a role and a menu.
type RoleMenu struct {
	ID        int64     `json:"id"         gorm:"column:id;primaryKey;autoIncrement"`
	RoleID    int64     `json:"role_id"    gorm:"column:role_id;not null;uniqueIndex:uk_role_menu"`
	MenuID    int64     `json:"menu_id"    gorm:"column:menu_id;not null;uniqueIndex:uk_role_menu"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (RoleMenu) TableName() string {
	return "role_menus"
}

// RolePermission represents the association between a role and an API permission.
type RolePermission struct {
	ID           int64     `json:"id"            gorm:"column:id;primaryKey;autoIncrement"`
	RoleID       int64     `json:"role_id"       gorm:"column:role_id;not null;uniqueIndex:uk_role_permission"`
	PermissionID int64     `json:"permission_id" gorm:"column:permission_id;not null;uniqueIndex:uk_role_permission"`
	CreatedAt    time.Time `json:"created_at"    gorm:"column:created_at;autoCreateTime"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
