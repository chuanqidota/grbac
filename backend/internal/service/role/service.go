package role

import (
	"context"
	"time"

	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	roleRepo "grbac/internal/repository/role"
	userRepo "grbac/internal/repository/user"
	"gorm.io/gorm"
)

// DispatchFn is the signature for async webhook event dispatch.
type DispatchFn func(ctx context.Context, event string, payload interface{})

// CreateRequest holds the payload for creating a new role.
type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	IsDefault   int8   `json:"is_default"`
}

// Service provides role CRUD, menu/permission assignment, and user assignment operations.
type Service struct {
	db         *gorm.DB
	roleRepo   *roleRepo.Repo
	userRepo   *userRepo.Repo
	dispatchFn DispatchFn
}

// NewService creates a new Service.
func NewService(db *gorm.DB, roleRepo *roleRepo.Repo, userRepo *userRepo.Repo, dispatchFn DispatchFn) *Service {
	return &Service{
		db:         db,
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		dispatchFn: dispatchFn,
	}
}

// Create registers a new role within a system.
func (s *Service) Create(ctx context.Context, systemID int64, req *CreateRequest) (*model.Role, error) {
	existing, _ := s.roleRepo.GetByCode(systemID, req.Code)
	if existing != nil {
		return nil, errors.ErrRoleCodeExists
	}

	if req.IsDefault == 1 {
		defaultRole, _ := s.roleRepo.GetDefaultRoleBySystem(systemID)
		if defaultRole != nil {
			return nil, errors.ErrInternal.Wrap("该系统已存在默认角色")
		}
	}

	role := &model.Role{
		SystemID:    systemID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		Status:      1,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.created", map[string]interface{}{
			"event":     "role.created",
			"timestamp": time.Now(),
			"system_id": systemID,
			"data":      map[string]interface{}{"role_id": role.ID, "role_name": role.Name, "role_code": role.Code},
		})
	}

	return role, nil
}

// GetByID retrieves a role by its ID.
func (s *Service) GetByID(id int64) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrRoleNotFound
	}
	return role, nil
}

// ListBySystem returns all roles belonging to a system.
func (s *Service) ListBySystem(systemID int64) ([]model.Role, error) {
	roles, err := s.roleRepo.ListBySystem(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return roles, nil
}

// Update modifies the name and description of an existing role.
func (s *Service) Update(ctx context.Context, id int64, name, description string) (*model.Role, error) {
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

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.updated", map[string]interface{}{
			"event":     "role.updated",
			"timestamp": time.Now(),
			"system_id": role.SystemID,
			"data":      map[string]interface{}{"role_id": role.ID, "role_name": role.Name},
		})
	}

	return role, nil
}

// Delete removes a role and its associations within a transaction.
func (s *Service) Delete(ctx context.Context, id int64) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	if role.IsDefault == 1 {
		return errors.ErrInternal.Wrap("默认角色不可删除")
	}

	hasUsers, err := s.roleRepo.HasUsers(id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if hasUsers {
		return errors.ErrInternal.Wrap("角色下存在用户，无法删除")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.roleRepo.RemoveRoleMenusTx(tx, id); err != nil {
			return err
		}
		if err := s.roleRepo.RemoveRolePermissionsTx(tx, id); err != nil {
			return err
		}
		return s.roleRepo.Delete(id)
	})
	if err != nil {
		return err
	}

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.deleted", map[string]interface{}{
			"event":     "role.deleted",
			"timestamp": time.Now(),
			"system_id": role.SystemID,
			"data":      map[string]interface{}{"role_id": id, "role_name": role.Name},
		})
	}

	return nil
}

// AssignMenus replaces the menu set of a role within a transaction.
func (s *Service) AssignMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.roleRepo.RemoveRoleMenusTx(tx, roleID); err != nil {
			return err
		}
		for _, menuID := range menuIDs {
			rm := &model.RoleMenu{
				RoleID: roleID,
				MenuID: menuID,
			}
			if err := s.roleRepo.AddRoleMenuTx(tx, rm); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.menus_assigned", map[string]interface{}{
			"event":     "role.menus_assigned",
			"timestamp": time.Now(),
			"system_id": role.SystemID,
			"data":      map[string]interface{}{"role_id": roleID, "menu_ids": menuIDs},
		})
	}

	return nil
}

// GetRoleMenus returns the menu IDs assigned to a role.
func (s *Service) GetRoleMenus(roleID int64) ([]int64, error) {
	menuIDs, err := s.roleRepo.GetRoleMenus(roleID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return menuIDs, nil
}

// AssignPermissions replaces the permission set of a role within a transaction.
func (s *Service) AssignPermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.roleRepo.RemoveRolePermissionsTx(tx, roleID); err != nil {
			return err
		}
		for _, permID := range permIDs {
			rp := &model.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := s.roleRepo.AddRolePermissionTx(tx, rp); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.permissions_assigned", map[string]interface{}{
			"event":     "role.permissions_assigned",
			"timestamp": time.Now(),
			"system_id": role.SystemID,
			"data":      map[string]interface{}{"role_id": roleID, "permission_ids": permIDs},
		})
	}

	return nil
}

// GetRolePermissions returns the permission IDs assigned to a role.
func (s *Service) GetRolePermissions(roleID int64) ([]int64, error) {
	permIDs, err := s.roleRepo.GetRolePermissions(roleID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return permIDs, nil
}

// AssignUsers replaces the user set of a role within a transaction.
func (s *Service) AssignUsers(ctx context.Context, roleID int64, userIDs []int64) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.RemoveUsersByRoleIDTx(tx, roleID); err != nil {
			return err
		}
		for _, userID := range userIDs {
			ur := &model.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
			if err := s.userRepo.AddUserRoleTx(tx, ur); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.users_assigned", map[string]interface{}{
			"event":     "role.users_assigned",
			"timestamp": time.Now(),
			"system_id": role.SystemID,
			"data":      map[string]interface{}{"role_id": roleID, "user_ids": userIDs},
		})
	}

	return nil
}

// RemoveUser removes a single user from a role.
func (s *Service) RemoveUser(ctx context.Context, roleID, userID int64) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}

	if err := s.userRepo.RemoveUserRole(userID, roleID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	if s.dispatchFn != nil {
		s.dispatchFn(ctx, "role.user_removed", map[string]interface{}{
			"event":     "role.user_removed",
			"timestamp": time.Now(),
			"system_id": role.SystemID,
			"data":      map[string]interface{}{"role_id": roleID, "user_id": userID},
		})
	}

	return nil
}

// GetRoleUsers returns all users assigned to a role.
func (s *Service) GetRoleUsers(roleID int64) ([]model.User, error) {
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return nil, errors.ErrRoleNotFound
	}
	users, err := s.userRepo.GetByRoleID(roleID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return users, nil
}
