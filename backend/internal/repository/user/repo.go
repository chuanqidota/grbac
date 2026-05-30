package user

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

func (r *Repo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *Repo) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListQuery holds optional filters for listing users.
type ListQuery struct {
	IsSuperAdmin *int // nil = all, 0 = non-super-admin, 1 = super-admin
}

func (r *Repo) List(page, pageSize int, q *ListQuery) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	tx := r.db.Model(&model.User{})
	if q != nil && q.IsSuperAdmin != nil {
		tx = tx.Where("is_super_admin = ?", *q.IsSuperAdmin)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := tx.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *Repo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *Repo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

func (r *Repo) AddUserRole(userRole *model.UserRole) error {
	return r.db.Create(userRole).Error
}

func (r *Repo) AddUserRoleTx(tx *gorm.DB, userRole *model.UserRole) error {
	return tx.Create(userRole).Error
}

func (r *Repo) RemoveUserRole(userID, roleID int64) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error
}

func (r *Repo) GetUserRoles(userID int64) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := r.db.Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}
	return userRoles, nil
}

func (r *Repo) GetUsersByRoleID(roleID int64) ([]int64, error) {
	var userIDs []int64
	err := r.db.Model(&model.UserRole{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *Repo) RemoveUsersByRoleID(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.UserRole{}).Error
}

func (r *Repo) RemoveUsersByRoleIDTx(tx *gorm.DB, roleID int64) error {
	return tx.Where("role_id = ?", roleID).Delete(&model.UserRole{}).Error
}

// GetRolesInSystem returns all roles assigned to the user within the specified system,
// including the system's default role (is_default=1) which applies to all users.
func (r *Repo) GetRolesInSystem(userID, systemID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Model(&model.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.system_id = ?", userID, systemID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	// Include the default role for this system.
	var defaultRole model.Role
	if err := r.db.Where("system_id = ? AND is_default = 1 AND status = 1", systemID).First(&defaultRole).Error; err == nil {
		exists := make(map[int64]bool, len(roles))
		for _, role := range roles {
			exists[role.ID] = true
		}
		if !exists[defaultRole.ID] {
			roles = append(roles, defaultRole)
		}
	}

	return roles, nil
}

// RemoveAllUserRoles removes all role associations for a user.
func (r *Repo) RemoveAllUserRoles(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error
}

// RemoveAllUserRolesTx removes all role associations for a user within a transaction.
func (r *Repo) RemoveAllUserRolesTx(tx *gorm.DB, userID int64) error {
	return tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error
}

// RemoveAllSystemMembers removes all system memberships for a user.
func (r *Repo) RemoveAllSystemMembers(userID int64) error {
	var member model.SystemMember
	return r.db.Where("user_id = ?", userID).Delete(&member).Error
}

// RemoveAllSystemMembersTx removes all system memberships for a user within a transaction.
func (r *Repo) RemoveAllSystemMembersTx(tx *gorm.DB, userID int64) error {
	return tx.Where("user_id = ?", userID).Delete(&model.SystemMember{}).Error
}

// DeleteTx removes a user by ID within a transaction.
func (r *Repo) DeleteTx(tx *gorm.DB, id int64) error {
	return tx.Where("id = ?", id).Delete(&model.User{}).Error
}

// GetByRoleID returns all user objects assigned to the given role.
func (r *Repo) GetByRoleID(roleID int64) ([]model.User, error) {
	var users []model.User
	err := r.db.Model(&model.User{}).
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error
	return users, err
}

// CountSuperAdmins returns the number of users with super-admin flag set.
func (r *Repo) CountSuperAdmins() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("is_super_admin = 1").Count(&count).Error
	return count, err
}
