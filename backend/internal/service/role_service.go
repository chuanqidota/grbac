package service

import (
	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	"grbac/internal/repository"
)

// CreateRoleRequest holds the payload for creating a new role.
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// RoleService provides role CRUD, menu/permission assignment, and user assignment operations.
type RoleService struct {
	roleRepo *repository.RoleRepo
	userRepo *repository.UserRepo
}

// NewRoleService creates a new RoleService.
func NewRoleService(roleRepo *repository.RoleRepo, userRepo *repository.UserRepo) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// Create registers a new role within a system.
func (s *RoleService) Create(systemID int64, req *CreateRoleRequest) (*model.Role, error) {
	// Check if the role code already exists within the system.
	existing, _ := s.roleRepo.GetByCode(systemID, req.Code)
	if existing != nil {
		return nil, errors.ErrRoleCodeExists
	}

	role := &model.Role{
		SystemID:    systemID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      1, // default active
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return role, nil
}

// GetByID retrieves a role by its ID.
func (s *RoleService) GetByID(id int64) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrRoleNotFound
	}
	return role, nil
}

// ListBySystem returns all roles belonging to a system.
func (s *RoleService) ListBySystem(systemID int64) ([]model.Role, error) {
	roles, err := s.roleRepo.ListBySystem(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return roles, nil
}

// Update modifies the name and description of an existing role.
func (s *RoleService) Update(id int64, name, description string) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrRoleNotFound
	}

	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return role, nil
}

// Delete removes a role after checking it has no bound users.
// It also cleans up associated menu and permission mappings.
func (s *RoleService) Delete(id int64) error {
	// 1. Check if the role exists.
	if _, err := s.roleRepo.GetByID(id); err != nil {
		return errors.ErrRoleNotFound
	}

	// 2. Check if any users are bound to this role.
	hasUsers, err := s.roleRepo.HasUsers(id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if hasUsers {
		return errors.ErrInternal.Wrap("角色下存在用户，无法删除")
	}

	// 3. Remove role-menu associations.
	if err := s.roleRepo.RemoveRoleMenus(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	// 4. Remove role-permission associations.
	if err := s.roleRepo.RemoveRolePermissions(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	// 5. Delete the role itself.
	if err := s.roleRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// AssignMenus replaces the menu set of a role (remove old, add new).
func (s *RoleService) AssignMenus(roleID int64, menuIDs []int64) error {
	// Verify the role exists.
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return errors.ErrRoleNotFound
	}

	// Remove existing role-menu associations.
	if err := s.roleRepo.RemoveRoleMenus(roleID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	// Add new associations.
	for _, menuID := range menuIDs {
		rm := &model.RoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		}
		if err := s.roleRepo.AddRoleMenu(rm); err != nil {
			return errors.ErrInternal.Wrap(err.Error())
		}
	}

	return nil
}

// GetRoleMenus returns the menu IDs assigned to a role.
func (s *RoleService) GetRoleMenus(roleID int64) ([]int64, error) {
	menuIDs, err := s.roleRepo.GetRoleMenus(roleID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return menuIDs, nil
}

// AssignPermissions replaces the permission set of a role (remove old, add new).
func (s *RoleService) AssignPermissions(roleID int64, permIDs []int64) error {
	// Verify the role exists.
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return errors.ErrRoleNotFound
	}

	// Remove existing role-permission associations.
	if err := s.roleRepo.RemoveRolePermissions(roleID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	// Add new associations.
	for _, permID := range permIDs {
		rp := &model.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
		if err := s.roleRepo.AddRolePermission(rp); err != nil {
			return errors.ErrInternal.Wrap(err.Error())
		}
	}

	return nil
}

// GetRolePermissions returns the permission IDs assigned to a role.
func (s *RoleService) GetRolePermissions(roleID int64) ([]int64, error) {
	permIDs, err := s.roleRepo.GetRolePermissions(roleID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return permIDs, nil
}

// AssignUsers replaces the user set of a role (remove old, add new).
func (s *RoleService) AssignUsers(roleID int64, userIDs []int64) error {
	// Verify the role exists.
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return errors.ErrRoleNotFound
	}

	// Remove all existing user-role associations for this role.
	if err := s.userRepo.RemoveUsersByRoleID(roleID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	// Add new associations.
	for _, userID := range userIDs {
		ur := &model.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := s.userRepo.AddUserRole(ur); err != nil {
			return errors.ErrInternal.Wrap(err.Error())
		}
	}

	return nil
}

// RemoveUser removes a single user from a role.
func (s *RoleService) RemoveUser(roleID, userID int64) error {
	if err := s.userRepo.RemoveUserRole(userID, roleID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	return nil
}
