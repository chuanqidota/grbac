package permission

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

func (r *Repo) Create(perm *model.Permission) error {
	return r.db.Create(perm).Error
}

func (r *Repo) GetByID(id int64) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *Repo) GetByMethodPath(systemID int64, method, path string) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("system_id = ? AND method = ? AND path = ?", systemID, method, path).
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *Repo) ListBySystem(systemID int64, page, pageSize int, method, keyword string) ([]model.Permission, int64, error) {
	var perms []model.Permission
	var total int64

	query := r.db.Model(&model.Permission{}).Where("system_id = ?", systemID)
	if method != "" {
		query = query.Where("method = ?", method)
	}
	if keyword != "" {
		query = query.Where("(code LIKE ? OR name LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

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

func (r *Repo) ListByIDs(ids []int64) ([]model.Permission, error) {
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

// ListAllBySystem returns all permissions for a system without pagination.
func (r *Repo) ListAllBySystem(systemID int64) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Where("system_id = ?", systemID).Order("id DESC").Find(&perms).Error
	return perms, err
}

func (r *Repo) Update(perm *model.Permission) error {
	return r.db.Save(perm).Error
}

func (r *Repo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Permission{}).Error
}

func (r *Repo) IsUsedByRole(permID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.RolePermission{}).
		Where("permission_id = ?", permID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasUserPermission checks directly in the database whether the user has a permission
// matching the given method and path within the specified system.
// This includes permissions from the user's explicitly assigned roles AND the system's default role.
func (r *Repo) HasUserPermission(userID, systemID int64, method, path string) (bool, error) {
	var count int64
	err := r.db.Model(&model.RolePermission{}).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.system_id = ? AND permissions.method = ? AND permissions.path = ?",
			userID, systemID, method, path).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Also check the system's default role.
	err = r.db.Model(&model.RolePermission{}).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.system_id = ? AND roles.is_default = 1 AND roles.status = 1 AND permissions.method = ? AND permissions.path = ?",
			systemID, method, path).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
