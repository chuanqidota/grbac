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

func (r *Repo) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
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

// GetRolesInSystem returns all roles assigned to the user within the specified system.
// Uses a single JOIN query instead of loading all user roles then filtering.
func (r *Repo) GetRolesInSystem(userID, systemID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Model(&model.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.system_id = ?", userID, systemID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// RemoveAllUserRoles removes all role associations for a user.
func (r *Repo) RemoveAllUserRoles(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error
}

// RemoveAllSystemMembers removes all system memberships for a user.
func (r *Repo) RemoveAllSystemMembers(userID int64) error {
	var member model.SystemMember
	return r.db.Where("user_id = ?", userID).Delete(&member).Error
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
