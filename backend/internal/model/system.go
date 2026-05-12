package model

import "time"

// System represents an external system registered in the RBAC platform.
type System struct {
	ID          int64     `json:"id"          gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `json:"name"        gorm:"column:name;size:100;not null"`
	Code        string    `json:"code"        gorm:"column:code;size:50;uniqueIndex;not null"`
	Secret      string    `json:"-"           gorm:"column:secret;size:512;not null"`
	Description string    `json:"description" gorm:"column:description;size:500"`
	Status      int8      `json:"status"      gorm:"column:status;not null;default:1"`
	CreatedAt   time.Time `json:"created_at"  gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at"  gorm:"column:updated_at;autoUpdateTime"`
}

func (System) TableName() string {
	return "systems"
}

// SystemMember represents the membership relationship between a user and a system.
type SystemMember struct {
	ID        int64     `json:"id"         gorm:"column:id;primaryKey;autoIncrement"`
	SystemID  int64     `json:"system_id"  gorm:"column:system_id;not null;uniqueIndex:uk_system_user"`
	UserID    int64     `json:"user_id"    gorm:"column:user_id;not null;uniqueIndex:uk_system_user"`
	Role      string    `json:"role"       gorm:"column:role;size:20;not null"` // admin / member
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (SystemMember) TableName() string {
	return "system_members"
}
