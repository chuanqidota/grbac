package system

import (
	"crypto/rand"
	"encoding/hex"

	"grbac/internal/model"
	"grbac/internal/pkg/crypto"
	"grbac/internal/pkg/errors"
	systemRepo "grbac/internal/repository/system"
	userRepo "grbac/internal/repository/user"
)

// CreateRequest holds the payload for registering a new system.
type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// Service provides system CRUD and member management operations.
type Service struct {
	systemRepo    *systemRepo.Repo
	userRepo      *userRepo.Repo
	encryptionKey []byte
}

// NewService creates a new Service.
func NewService(systemRepo *systemRepo.Repo, userRepo *userRepo.Repo, encryptionKey string) *Service {
	return &Service{
		systemRepo:    systemRepo,
		userRepo:      userRepo,
		encryptionKey: []byte(encryptionKey),
	}
}

// Create registers a new system with an auto-generated secret key.
// The secret is encrypted before storage. The plaintext secret is returned
// in the response (shown once to the admin).
func (s *Service) Create(req *CreateRequest) (*model.System, error) {
	existing, _ := s.systemRepo.GetByCode(req.Code)
	if existing != nil {
		return nil, errors.ErrInternal.Wrap("系统编码已存在")
	}

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, errors.ErrInternal.Wrap("生成系统密钥失败")
	}
	plaintextSecret := hex.EncodeToString(secretBytes)

	encryptedSecret, err := crypto.Encrypt([]byte(plaintextSecret), s.encryptionKey)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("加密系统密钥失败")
	}

	system := &model.System{
		Name:        req.Name,
		Code:        req.Code,
		Secret:      encryptedSecret,
		Description: req.Description,
		Status:      1,
	}

	if err := s.systemRepo.Create(system); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	// Return plaintext secret in response (shown once).
	system.Secret = plaintextSecret
	return system, nil
}

// ValidateSecret checks whether the given secret matches the stored (encrypted) secret.
func (s *Service) ValidateSecret(code, secret string) (*model.System, error) {
	system, err := s.systemRepo.GetByCode(code)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	decrypted, err := crypto.Decrypt(system.Secret, s.encryptionKey)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("解密系统密钥失败")
	}

	if string(decrypted) != secret {
		return nil, errors.ErrSystemCredential
	}

	return system, nil
}

// GetByID retrieves a system by its ID.
func (s *Service) GetByID(id int64) (*model.System, error) {
	system, err := s.systemRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}
	return system, nil
}

// GetByCode retrieves a system by its unique code.
func (s *Service) GetByCode(code string) (*model.System, error) {
	system, err := s.systemRepo.GetByCode(code)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}
	return system, nil
}

// List returns a paginated list of systems and the total count.
func (s *Service) List(page, pageSize int) ([]model.System, int64, error) {
	systems, total, err := s.systemRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return systems, total, nil
}

// Update modifies the name and description of an existing system.
func (s *Service) Update(id int64, name, description string) (*model.System, error) {
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

// Delete removes a system and all associated data (roles, menus, permissions, members).
func (s *Service) Delete(id int64) error {
	_, err := s.systemRepo.GetByID(id)
	if err != nil {
		return errors.ErrSystemNotFound
	}

	if err := s.systemRepo.DeleteCascade(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// AddMember adds a user to a system with the specified role (admin / member).
func (s *Service) AddMember(systemID, userID int64, role string) error {
	if _, err := s.systemRepo.GetByID(systemID); err != nil {
		return errors.ErrSystemNotFound
	}

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
func (s *Service) RemoveMember(systemID, userID int64) error {
	if err := s.systemRepo.RemoveMember(systemID, userID); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	return nil
}

// GetMembers returns all members of a system.
func (s *Service) GetMembers(systemID int64) ([]model.SystemMember, error) {
	members, err := s.systemRepo.GetMembers(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return members, nil
}

// IsAdmin checks whether a user has the admin role in a system.
func (s *Service) IsAdmin(systemID, userID int64) (bool, error) {
	return s.systemRepo.IsAdmin(systemID, userID)
}
