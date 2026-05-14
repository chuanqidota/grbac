package model

import "time"

// Menu represents a menu entry within a system. Menus form a tree via ParentID.
type Menu struct {
	ID        int64     `json:"id"         gorm:"column:id;primaryKey;autoIncrement"`
	SystemID  int64     `json:"system_id"  gorm:"column:system_id;not null;index:idx_menus_system_parent,priority:1"`
	ParentID  int64     `json:"parent_id"  gorm:"column:parent_id;not null;default:0;index:idx_menus_system_parent,priority:2"`
	Name      string    `json:"name"       gorm:"column:name;size:50;not null"`
	Path      string    `json:"path"       gorm:"column:path;size:200"`
	Icon      string    `json:"icon"       gorm:"column:icon;size:50"`
	SortOrder int       `json:"sort_order" gorm:"column:sort_order;not null;default:0"`
	Status    int8      `json:"status"     gorm:"column:status;not null;default:1"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Menu) TableName() string {
	return "menus"
}
