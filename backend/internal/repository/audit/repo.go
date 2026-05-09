package audit

import (
	"grbac/internal/model"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *Repo) List(page, pageSize int, systemID *int64) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{})
	if systemID != nil {
		query = query.Where("system_id = ?", *systemID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
