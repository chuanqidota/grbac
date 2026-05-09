package audit

import (
	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	auditRepo "grbac/internal/repository/audit"
)

// Service provides audit log operations.
type Service struct {
	repo *auditRepo.Repo
}

// NewService creates a new Service.
func NewService(repo *auditRepo.Repo) *Service {
	return &Service{repo: repo}
}

// Record creates a new audit log entry.
func (s *Service) Record(log *model.AuditLog) error {
	if err := s.repo.Create(log); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	return nil
}

// List returns a paginated list of audit logs.
func (s *Service) List(page, pageSize int, systemID *int64) ([]model.AuditLog, int64, error) {
	logs, total, err := s.repo.List(page, pageSize, systemID)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return logs, total, nil
}
