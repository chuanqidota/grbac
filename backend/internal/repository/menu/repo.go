package menu

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

func (r *Repo) Create(menu *model.Menu) error {
	return r.db.Create(menu).Error
}

func (r *Repo) GetByID(id int64) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.Where("id = ?", id).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *Repo) ListBySystem(systemID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("system_id = ?", systemID).Order("sort_order ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *Repo) GetByParentID(systemID, parentID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("system_id = ? AND parent_id = ?", systemID, parentID).
		Order("sort_order ASC, id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *Repo) HasChildren(systemID, menuID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Menu{}).
		Where("system_id = ? AND parent_id = ?", systemID, menuID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repo) ListByIDs(ids []int64) ([]model.Menu, error) {
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

func (r *Repo) Update(menu *model.Menu) error {
	return r.db.Save(menu).Error
}

func (r *Repo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Menu{}).Error
}
