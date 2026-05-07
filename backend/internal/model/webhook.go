package model

import "time"

// Webhook represents a webhook subscription for a system.
type Webhook struct {
	ID        int64     `json:"id"         gorm:"column:id;primaryKey;autoIncrement"`
	SystemID  int64     `json:"system_id"  gorm:"column:system_id;not null"`
	URL       string    `json:"url"        gorm:"column:url;size:255;not null"`
	Secret    string    `json:"-"          gorm:"column:secret;size:128;not null"`
	Events    string    `json:"events"     gorm:"column:events;size:255;not null"` // comma-separated: permission_change,menu_change,role_change
	Status    int8      `json:"status"     gorm:"column:status;not null;default:1"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Webhook) TableName() string {
	return "webhooks"
}
