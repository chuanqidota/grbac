package repository

import (
	"grbac/internal/model"
	"gorm.io/gorm"
)

type PermissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

func (r *PermissionRepo) Create(perm *model.Permission) error {
	return r.db.Create(perm).Error
}

func (r *PermissionRepo) GetByID(id int64) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *PermissionRepo) GetByMethodPath(systemID int64, method, path string) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("system_id = ? AND method = ? AND path = ?", systemID, method, path).
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *PermissionRepo) ListBySystem(systemID int64, page, pageSize int) ([]model.Permission, int64, error) {
	var perms []model.Permission
	var total int64

	query := r.db.Model(&model.Permission{}).Where("system_id = ?", systemID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&perms).Error
	if err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

// ListByIDs returns permissions matching the given IDs.
func (r *PermissionRepo) ListByIDs(ids []int64) ([]model.Permission, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var perms []model.Permission
	err := r.db.Where("id IN ?", ids).Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *PermissionRepo) Update(perm *model.Permission) error {
	return r.db.Save(perm).Error
}

func (r *PermissionRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Permission{}).Error
}

func (r *PermissionRepo) IsUsedByRole(permID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.RolePermission{}).
		Where("permission_id = ?", permID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
