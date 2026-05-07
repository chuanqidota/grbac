package model

import "time"

// Permission represents an API permission (method + path) within a system.
type Permission struct {
	ID          int64     `json:"id"          gorm:"column:id;primaryKey;autoIncrement"`
	SystemID    int64     `json:"system_id"   gorm:"column:system_id;not null;uniqueIndex:uk_system_method_path"`
	Code        string    `json:"code"        gorm:"column:code;size:100;not null"`
	Name        string    `json:"name"        gorm:"column:name;size:100;not null"`
	Method      string    `json:"method"      gorm:"column:method;size:10;not null;uniqueIndex:uk_system_method_path"` // GET/POST/PUT/DELETE
	Path        string    `json:"path"        gorm:"column:path;size:200;not null;uniqueIndex:uk_system_method_path"`
	Description string    `json:"description" gorm:"column:description;size:500"`
	Version     int       `json:"version"     gorm:"column:version;not null;default:1"`
	CreatedAt   time.Time `json:"created_at"  gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at"  gorm:"column:updated_at;autoUpdateTime"`
}

func (Permission) TableName() string {
	return "permissions"
}
