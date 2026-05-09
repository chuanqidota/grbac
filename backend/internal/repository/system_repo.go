package repository

import (
	"grbac/internal/model"
	"gorm.io/gorm"
)

type SystemRepo struct {
	db *gorm.DB
}

func NewSystemRepo(db *gorm.DB) *SystemRepo {
	return &SystemRepo{db: db}
}

func (r *SystemRepo) Create(system *model.System) error {
	return r.db.Create(system).Error
}

func (r *SystemRepo) GetByID(id int64) (*model.System, error) {
	var system model.System
	err := r.db.Where("id = ?", id).First(&system).Error
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (r *SystemRepo) GetByCode(code string) (*model.System, error) {
	var system model.System
	err := r.db.Where("code = ?", code).First(&system).Error
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (r *SystemRepo) List(page, pageSize int) ([]model.System, int64, error) {
	var systems []model.System
	var total int64

	if err := r.db.Model(&model.System{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&systems).Error
	if err != nil {
		return nil, 0, err
	}

	return systems, total, nil
}

func (r *SystemRepo) Update(system *model.System) error {
	return r.db.Save(system).Error
}

func (r *SystemRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.System{}).Error
}

func (r *SystemRepo) AddMember(member *model.SystemMember) error {
	return r.db.Create(member).Error
}

func (r *SystemRepo) RemoveMember(systemID, userID int64) error {
	return r.db.Where("system_id = ? AND user_id = ?", systemID, userID).
		Delete(&model.SystemMember{}).Error
}

func (r *SystemRepo) GetMembers(systemID int64) ([]model.SystemMember, error) {
	var members []model.SystemMember
	err := r.db.Where("system_id = ?", systemID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *SystemRepo) IsMember(systemID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.SystemMember{}).
		Where("system_id = ? AND user_id = ?", systemID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SystemRepo) IsAdmin(systemID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.SystemMember{}).
		Where("system_id = ? AND user_id = ? AND role = ?", systemID, userID, "admin").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
