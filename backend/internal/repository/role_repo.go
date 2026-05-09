package repository

import (
	"grbac/internal/model"
	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepo) GetByID(id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) GetByCode(systemID int64, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("system_id = ? AND code = ?", systemID, code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) ListBySystem(systemID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("system_id = ?", systemID).Order("id ASC").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepo) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Role{}).Error
}

func (r *RoleRepo) AddRoleMenu(roleMenu *model.RoleMenu) error {
	return r.db.Create(roleMenu).Error
}

func (r *RoleRepo) RemoveRoleMenus(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}

func (r *RoleRepo) GetRoleMenus(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Model(&model.RoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	if err != nil {
		return nil, err
	}
	return menuIDs, nil
}

func (r *RoleRepo) AddRolePermission(rolePerm *model.RolePermission) error {
	return r.db.Create(rolePerm).Error
}

func (r *RoleRepo) RemoveRolePermissions(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

func (r *RoleRepo) GetRolePermissions(roleID int64) ([]int64, error) {
	var permIDs []int64
	err := r.db.Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permIDs).Error
	if err != nil {
		return nil, err
	}
	return permIDs, nil
}

func (r *RoleRepo) HasUsers(roleID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserRole{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
