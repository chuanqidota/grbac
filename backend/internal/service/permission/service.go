package permission

import (
	"context"
	"time"

	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	permRepo "grbac/internal/repository/permission"
)

// DispatchFn is the signature for async webhook event dispatch.
type DispatchFn func(ctx context.Context, event string, payload interface{})

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
	permRepo   *permRepo.Repo
	dispatchFn DispatchFn
}

// NewService creates a new Service.
func NewService(permRepo *permRepo.Repo, dispatchFn DispatchFn) *Service {
	return &Service{permRepo: permRepo, dispatchFn: dispatchFn}
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

	if s.dispatchFn != nil {
		s.dispatchFn(context.Background(), "permission.created", map[string]interface{}{
			"event":     "permission.created",
			"timestamp": time.Now(),
			"system_id": systemID,
			"data":      map[string]interface{}{"permission_id": perm.ID, "permission_code": perm.Code, "path": perm.Path, "method": perm.Method},
		})
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
func (s *Service) ListBySystem(systemID int64, page, pageSize int, method, keyword string) ([]model.Permission, int64, error) {
	perms, total, err := s.permRepo.ListBySystem(systemID, page, pageSize, method, keyword)
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

	if s.dispatchFn != nil {
		s.dispatchFn(context.Background(), "permission.updated", map[string]interface{}{
			"event":     "permission.updated",
			"timestamp": time.Now(),
			"system_id": perm.SystemID,
			"data":      map[string]interface{}{"permission_id": perm.ID, "permission_code": perm.Code, "path": perm.Path, "method": perm.Method},
		})
	}

	return perm, nil
}

// Delete removes a permission after verifying it is not in use by any role.
func (s *Service) Delete(id int64) error {
	perm, err := s.permRepo.GetByID(id)
	if err != nil {
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

	if s.dispatchFn != nil {
		s.dispatchFn(context.Background(), "permission.deleted", map[string]interface{}{
			"event":     "permission.deleted",
			"timestamp": time.Now(),
			"system_id": perm.SystemID,
			"data":      map[string]interface{}{"permission_id": id, "permission_code": perm.Code, "path": perm.Path, "method": perm.Method},
		})
	}

	return nil
}
