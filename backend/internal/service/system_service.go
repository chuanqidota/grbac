package service

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

// CreateSystemRequest holds the payload for registering a new system.
type CreateSystemRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// SystemService provides system CRUD and member management operations.
type SystemService struct {
	systemRepo *repository.SystemRepo
	userRepo   *repository.UserRepo
}

// NewSystemService creates a new SystemService.
func NewSystemService(systemRepo *repository.SystemRepo, userRepo *repository.UserRepo) *SystemService {
	return &SystemService{
		systemRepo: systemRepo,
		userRepo:   userRepo,
	}
}

// Create registers a new system with an auto-generated secret key.
func (s *SystemService) Create(req *CreateSystemRequest) (*model.System, error) {
	// Check if the code already exists.
	existing, _ := s.systemRepo.GetByCode(req.Code)
	if existing != nil {
		return nil, errors.ErrInternal.Wrap("系统编码已存在")
	}

	// Generate a 32-byte random secret (hex-encoded to 64 characters).
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, errors.ErrInternal.Wrap("生成系统密钥失败")
	}
	secret := hex.EncodeToString(secretBytes)

	system := &model.System{
		Name:        req.Name,
		Code:        req.Code,
		Secret:      secret,
		Description: req.Description,
		Status:      1, // default active
	}

	if err := s.systemRepo.Create(system); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return system, nil
}

// GetByID retrieves a system by its ID.
func (s *SystemService) GetByID(id int64) (*model.System, error) {
	system, err := s.systemRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}
	return system, nil
}

// GetByCode retrieves a system by its unique code.
func (s *SystemService) GetByCode(code string) (*model.System, error) {
	system, err := s.systemRepo.GetByCode(code)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}
	return system, nil
}

// List returns a paginated list of systems and the total count.
func (s *SystemService) List(page, pageSize int) ([]model.System, int64, error) {
	systems, total, err := s.systemRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return systems, total, nil
}

// Update modifies the name and description of an existing system.
func (s *SystemService) Update(id int64, name, description string) (*model.System, error) {
	system, err := s.systemRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	if name != "" {
		system.Name = name
	}
	if description != "" {
		system.Description = description
	}

	if err := s.systemRepo.Update(system); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return system, nil
}

// Delete removes a system by its ID.
func (s *SystemService) Delete(id int64) error {
	_, err := s.systemRepo.GetByID(id)
	if err != nil {
		return errors.ErrSystemNotFound
	}

	if err := s.systemRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// AddMember adds a user to a system with the specified role (admin / member).
func (s *SystemService) AddMember(systemID, userID int64, role string) error {
	// Verify the system exists.
	if _, err := s.systemRepo.GetByID(systemID); err != nil {
		return errors.ErrSystemNotFound
	}

	// Verify the user exists.
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return errors.ErrUserNotFound
	}

	member := &model.SystemMember{
		SystemID: systemID,
		UserID:   userID,
		Role:     role,
	}

	if err := s.systemRepo.AddMember(member); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// RemoveMember removes a user from a system.
func (s *SystemService) RemoveMember(systemID, userID int64) error {
	if err := s.systemRepo.RemoveMember(systemID, userID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	return nil
}

// GetMembers returns all members of a system.
func (s *SystemService) GetMembers(systemID int64) ([]model.SystemMember, error) {
	members, err := s.systemRepo.GetMembers(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return members, nil
}

// IsAdmin checks whether a user has the admin role in a system.
func (s *SystemService) IsAdmin(systemID, userID int64) (bool, error) {
	return s.systemRepo.IsAdmin(systemID, userID)
}
