package system

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

func (r *Repo) Create(system *model.System) error {
	return r.db.Create(system).Error
}

func (r *Repo) GetByID(id int64) (*model.System, error) {
	var system model.System
	err := r.db.Where("id = ?", id).First(&system).Error
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (r *Repo) GetByCode(code string) (*model.System, error) {
	var system model.System
	err := r.db.Where("code = ?", code).First(&system).Error
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (r *Repo) List(page, pageSize int) ([]model.System, int64, error) {
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

func (r *Repo) Update(system *model.System) error {
	return r.db.Save(system).Error
}

func (r *Repo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.System{}).Error
}

func (r *Repo) AddMember(member *model.SystemMember) error {
	return r.db.Create(member).Error
}

func (r *Repo) RemoveMember(systemID, userID int64) error {
	return r.db.Where("system_id = ? AND user_id = ?", systemID, userID).
		Delete(&model.SystemMember{}).Error
}

func (r *Repo) GetMembers(systemID int64) ([]model.SystemMember, error) {
	var members []model.SystemMember
	err := r.db.Where("system_id = ?", systemID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// MemberInfo is a SystemMember enriched with user fields.
type MemberInfo struct {
	ID          int64  `json:"id"`
	SystemID    int64  `json:"system_id"`
	UserID      int64  `json:"user_id"`
	Role        string `json:"role"`
	Username    string `json:"username"`
	ChineseName string `json:"chinese_name"`
	Email       string `json:"email"`
	CreatedAt   string `json:"created_at"`
}

// GetMembersWithUser returns system members joined with user info.
func (r *Repo) GetMembersWithUser(systemID int64) ([]MemberInfo, error) {
	var results []MemberInfo
	err := r.db.Table("system_members sm").
		Select("sm.id, sm.system_id, sm.user_id, sm.role, u.username, u.chinese_name, u.email, sm.created_at").
		Joins("LEFT JOIN users u ON u.id = sm.user_id").
		Where("sm.system_id = ?", systemID).
		Order("sm.id DESC").
		Scan(&results).Error
	return results, err
}

func (r *Repo) IsMember(systemID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.SystemMember{}).
		Where("system_id = ? AND user_id = ?", systemID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repo) IsAdmin(systemID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.SystemMember{}).
		Where("system_id = ? AND user_id = ? AND role = ?", systemID, userID, "admin").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByUserID returns all systems where the user is a member.
func (r *Repo) ListByUserID(userID int64) ([]model.System, error) {
	var systems []model.System
	err := r.db.Model(&model.System{}).
		Joins("JOIN system_members ON system_members.system_id = systems.id").
		Where("system_members.user_id = ?", userID).
		Order("systems.id DESC").
		Find(&systems).Error
	return systems, err
}

// SystemWithRole extends System with the current user's role in that system.
type SystemWithRole struct {
	model.System
	CurrentUserRole string `json:"current_user_role"`
}

// ListByUserIDWithRole returns systems where the user is a member, enriched with their role.
func (r *Repo) ListByUserIDWithRole(userID int64, page, pageSize int) ([]SystemWithRole, int64, error) {
	var total int64
	base := r.db.Table("systems s").
		Joins("JOIN system_members sm ON sm.system_id = s.id").
		Where("sm.user_id = ?", userID)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []SystemWithRole
	offset := (page - 1) * pageSize
	err := base.Select("s.*, sm.role AS current_user_role").
		Order("s.id DESC").
		Offset(offset).Limit(pageSize).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// RoleMemberInfo is a user enriched with their roles in a system.
type RoleMemberInfo struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	ChineseName string `json:"chinese_name"`
	Email       string `json:"email"`
}

// GetUsersWithRoles returns distinct users who have at least one role in the system.
func (r *Repo) GetUsersWithRoles(systemID int64) ([]RoleMemberInfo, error) {
	var results []RoleMemberInfo
	err := r.db.Table("users u").
		Select("DISTINCT u.id AS user_id, u.username, u.chinese_name, u.email").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("r.system_id = ?", systemID).
		Order("u.id DESC").
		Scan(&results).Error
	return results, err
}

// DeleteCascade removes a system and all associated data within a transaction.
func (r *Repo) DeleteCascade(systemID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get all role IDs for this system.
		var roleIDs []int64
		if err := tx.Model(&model.Role{}).Where("system_id = ?", systemID).Pluck("id", &roleIDs).Error; err != nil {
			return err
		}

		if len(roleIDs) > 0 {
			// Delete role_permissions for these roles.
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&model.RolePermission{}).Error; err != nil {
				return err
			}
			// Delete role_menus for these roles.
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&model.RoleMenu{}).Error; err != nil {
				return err
			}
			// Delete user_roles for these roles.
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&model.UserRole{}).Error; err != nil {
				return err
			}
			// Delete the roles themselves.
			if err := tx.Where("id IN ?", roleIDs).Delete(&model.Role{}).Error; err != nil {
				return err
			}
		}

		// Delete system_members.
		if err := tx.Where("system_id = ?", systemID).Delete(&model.SystemMember{}).Error; err != nil {
			return err
		}
		// Delete menus for this system.
		if err := tx.Where("system_id = ?", systemID).Delete(&model.Menu{}).Error; err != nil {
			return err
		}
		// Delete permissions for this system.
		if err := tx.Where("system_id = ?", systemID).Delete(&model.Permission{}).Error; err != nil {
			return err
		}
		// Delete the system itself.
		return tx.Where("id = ?", systemID).Delete(&model.System{}).Error
	})
}
