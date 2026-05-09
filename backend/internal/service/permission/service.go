package permission

import (
	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	permRepo "grbac/internal/repository/permission"
)

// CreateRequest holds the payload for creating a new API permission.
type CreateRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Description string `json:"description"`
}

// Service provides permission CRUD operations.
type Service struct {
	permRepo *permRepo.Repo
}

// NewService creates a new Service.
func NewService(permRepo *permRepo.Repo) *Service {
	return &Service{permRepo: permRepo}
}

// Create registers a new API permission within a system.
func (s *Service) Create(systemID int64, req *CreateRequest) (*model.Permission, error) {
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
func (s *Service) GetByID(id int64) (*model.Permission, error) {
	perm, err := s.permRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrPermNotFound
	}
	return perm, nil
}

// ListBySystem returns a paginated list of permissions belonging to a system.
func (s *Service) ListBySystem(systemID int64, page, pageSize int) ([]model.Permission, int64, error) {
	perms, total, err := s.permRepo.ListBySystem(systemID, page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return perms, total, nil
}

// Update modifies the fields of an existing permission and increments its version.
func (s *Service) Update(id int64, req *CreateRequest) (*model.Permission, error) {
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
func (s *Service) Delete(id int64) error {
	if _, err := s.permRepo.GetByID(id); err != nil {
		return errors.ErrPermNotFound
	}

	used, err := s.permRepo.IsUsedByRole(id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if used {
		return errors.ErrPermInUse
	}

	if err := s.permRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}
