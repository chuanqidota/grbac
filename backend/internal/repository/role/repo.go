package role

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

func (r *Repo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *Repo) GetByID(id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repo) GetByCode(systemID int64, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("system_id = ? AND code = ?", systemID, code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repo) GetDefaultRoleBySystem(systemID int64) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("system_id = ? AND is_default = 1 AND status = 1", systemID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repo) ListBySystem(systemID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("system_id = ?", systemID).Order("id ASC").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repo) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *Repo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Role{}).Error
}

func (r *Repo) AddRoleMenu(roleMenu *model.RoleMenu) error {
	return r.db.Create(roleMenu).Error
}

func (r *Repo) AddRoleMenuTx(tx *gorm.DB, roleMenu *model.RoleMenu) error {
	return tx.Create(roleMenu).Error
}

func (r *Repo) RemoveRoleMenus(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}

func (r *Repo) RemoveRoleMenusTx(tx *gorm.DB, roleID int64) error {
	return tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}

func (r *Repo) GetRoleMenus(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Model(&model.RoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	if err != nil {
		return nil, err
	}
	return menuIDs, nil
}

func (r *Repo) AddRolePermission(rolePerm *model.RolePermission) error {
	return r.db.Create(rolePerm).Error
}

func (r *Repo) AddRolePermissionTx(tx *gorm.DB, rolePerm *model.RolePermission) error {
	return tx.Create(rolePerm).Error
}

func (r *Repo) RemoveRolePermissions(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

func (r *Repo) RemoveRolePermissionsTx(tx *gorm.DB, roleID int64) error {
	return tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

func (r *Repo) GetRolePermissions(roleID int64) ([]int64, error) {
	var permIDs []int64
	err := r.db.Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permIDs).Error
	if err != nil {
		return nil, err
	}
	return permIDs, nil
}

func (r *Repo) HasUsers(roleID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserRole{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRoleMenusByRoleIDs returns menu IDs grouped by role ID for the given role IDs.
func (r *Repo) GetRoleMenusByRoleIDs(roleIDs []int64) (map[int64][]int64, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var roleMenus []model.RoleMenu
	err := r.db.Where("role_id IN ?", roleIDs).Find(&roleMenus).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]int64)
	for _, rm := range roleMenus {
		result[rm.RoleID] = append(result[rm.RoleID], rm.MenuID)
	}
	return result, nil
}

// GetRolePermissionsByRoleIDs returns permission IDs grouped by role ID for the given role IDs.
func (r *Repo) GetRolePermissionsByRoleIDs(roleIDs []int64) (map[int64][]int64, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var rolePerms []model.RolePermission
	err := r.db.Where("role_id IN ?", roleIDs).Find(&rolePerms).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]int64)
	for _, rp := range rolePerms {
		result[rp.RoleID] = append(result[rp.RoleID], rp.PermissionID)
	}
	return result, nil
}
