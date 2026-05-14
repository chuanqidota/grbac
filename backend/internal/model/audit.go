package model

import "time"

// AuditLog records an administrative action performed by a user.
type AuditLog struct {
	ID           int64     `json:"id"            gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64     `json:"user_id"       gorm:"column:user_id;not null"`
	Username     string    `json:"username"      gorm:"column:username;size:50;not null"`
	SystemID     *int64    `json:"system_id"     gorm:"column:system_id;index"`
	Action       string    `json:"action"        gorm:"column:action;size:50;not null"`       // create/update/delete/assign
	Resource     string    `json:"resource"      gorm:"column:resource;size:50;not null"`     // user/role/menu/permission/system
	ResourceID   *int64    `json:"resource_id"   gorm:"column:resource_id"`
	ResourceName string    `json:"resource_name" gorm:"column:resource_name;size:100"`
	Detail       string    `json:"detail"        gorm:"column:detail;type:text"`               // JSON change details
	IP           string    `json:"ip"            gorm:"column:ip;size:50"`
	CreatedAt    time.Time `json:"created_at"    gorm:"column:created_at;autoCreateTime"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
