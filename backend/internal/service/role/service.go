package role

import (
	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	roleRepo "grbac/internal/repository/role"
	userRepo "grbac/internal/repository/user"
	"gorm.io/gorm"
)

// CreateRequest holds the payload for creating a new role.
type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// Service provides role CRUD, menu/permission assignment, and user assignment operations.
type Service struct {
	db       *gorm.DB
	roleRepo *roleRepo.Repo
	userRepo *userRepo.Repo
}

// NewService creates a new Service.
func NewService(db *gorm.DB, roleRepo *roleRepo.Repo, userRepo *userRepo.Repo) *Service {
	return &Service{
		db:       db,
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// Create registers a new role within a system.
func (s *Service) Create(systemID int64, req *CreateRequest) (*model.Role, error) {
	existing, _ := s.roleRepo.GetByCode(systemID, req.Code)
	if existing != nil {
		return nil, errors.ErrRoleCodeExists
	}

	role := &model.Role{
		SystemID:    systemID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      1,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
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
func (s *Service) Update(id int64, name, description string) (*model.Role, error) {
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

// Delete removes a role and its associations within a transaction.
func (s *Service) Delete(id int64) error {
	if _, err := s.roleRepo.GetByID(id); err != nil {
		return errors.ErrRoleNotFound
	}

	hasUsers, err := s.roleRepo.HasUsers(id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if hasUsers {
		return errors.ErrInternal.Wrap("角色下存在用户，无法删除")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.roleRepo.RemoveRoleMenusTx(tx, id); err != nil {
			return err
		}
		if err := s.roleRepo.RemoveRolePermissionsTx(tx, id); err != nil {
			return err
		}
		return s.roleRepo.Delete(id)
	})
}

// AssignMenus replaces the menu set of a role within a transaction.
func (s *Service) AssignMenus(roleID int64, menuIDs []int64) error {
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return errors.ErrRoleNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
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
func (s *Service) AssignPermissions(roleID int64, permIDs []int64) error {
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return errors.ErrRoleNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
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
func (s *Service) AssignUsers(roleID int64, userIDs []int64) error {
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		return errors.ErrRoleNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
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
}

// RemoveUser removes a single user from a role.
func (s *Service) RemoveUser(roleID, userID int64) error {
	if err := s.userRepo.RemoveUserRole(userID, roleID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
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
