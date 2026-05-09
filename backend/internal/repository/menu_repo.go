package repository

import (
	"grbac/internal/model"
	"gorm.io/gorm"
)

type MenuRepo struct {
	db *gorm.DB
}

func NewMenuRepo(db *gorm.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

func (r *MenuRepo) Create(menu *model.Menu) error {
	return r.db.Create(menu).Error
}

func (r *MenuRepo) GetByID(id int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.Where("id = ?", id).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *MenuRepo) ListBySystem(systemID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("system_id = ?", systemID).Order("sort_order ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *MenuRepo) GetByParentID(systemID, parentID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("system_id = ? AND parent_id = ?", systemID, parentID).
		Order("sort_order ASC, id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *MenuRepo) HasChildren(systemID, menuID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Menu{}).
		Where("system_id = ? AND parent_id = ?", systemID, menuID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByIDs returns menus matching the given IDs.
func (r *MenuRepo) ListByIDs(ids []int64) ([]model.Menu, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var menus []model.Menu
	err := r.db.Where("id IN ?", ids).Order("sort_order ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *MenuRepo) Update(menu *model.Menu) error {
	return r.db.Save(menu).Error
}

func (r *MenuRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Menu{}).Error
}
