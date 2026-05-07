package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

// CreatePermissionRequest holds the payload for creating a new API permission.
type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Description string `json:"description"`
}

// PermissionService provides permission CRUD operations.
type PermissionService struct {
	permRepo *repository.PermissionRepo
}

// NewPermissionService creates a new PermissionService.
func NewPermissionService(permRepo *repository.PermissionRepo) *PermissionService {
	return &PermissionService{permRepo: permRepo}
}

// Create registers a new API permission within a system.
func (s *PermissionService) Create(systemID int64, req *CreatePermissionRequest) (*model.Permission, error) {
	// Check if method+path already exists within the system.
	existing, _ := s.permRepo.GetByMethodPath(systemID, req.Method, req.Path)
	if existing != nil {
		return nil, errors.ErrPermPathExists
	}

	perm := &model.Permission{
		SystemID:    systemID,
		Code:        req.Code,
		Name:        req.Name,
		Method:      req.Method,
		Path:        req.Path,
		Description: req.Description,
	}

	if err := s.permRepo.Create(perm); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return perm, nil
}

// GetByID retrieves a permission by its ID.
func (s *PermissionService) GetByID(id int64) (*model.Permission, error) {
	perm, err := s.permRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrPermNotFound
	}
	return perm, nil
}

// ListBySystem returns a paginated list of permissions belonging to a system.
func (s *PermissionService) ListBySystem(systemID int64, page, pageSize int) ([]model.Permission, int64, error) {
	perms, total, err := s.permRepo.ListBySystem(systemID, page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return perms, total, nil
}

// Update modifies the fields of an existing permission and increments its version.
func (s *PermissionService) Update(id int64, req *CreatePermissionRequest) (*model.Permission, error) {
	perm, err := s.permRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrPermNotFound
	}

	perm.Code = req.Code
	perm.Name = req.Name
	perm.Method = req.Method
	perm.Path = req.Path
	perm.Description = req.Description
	perm.Version++

	if err := s.permRepo.Update(perm); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return perm, nil
}

// Delete removes a permission after verifying it is not in use by any role.
func (s *PermissionService) Delete(id int64) error {
	// 1. Check if the permission exists.
	if _, err := s.permRepo.GetByID(id); err != nil {
		return errors.ErrPermNotFound
	}

	// 2. Check if the permission is used by any role.
	used, err := s.permRepo.IsUsedByRole(id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if used {
		return errors.ErrPermInUse
	}

	// 3. Delete the permission.
	if err := s.permRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}
