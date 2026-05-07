package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/crypto"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

// CreateUserRequest holds the payload for creating a new user.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// UpdateUserRequest holds the payload for updating user profile fields.
type UpdateUserRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// UserService provides user CRUD operations.
type UserService struct {
	userRepo *repository.UserRepo
}

// NewUserService creates a new UserService.
func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

// Create registers a new user after validating uniqueness and hashing the password.
func (s *UserService) Create(req *CreateUserRequest) (*model.User, error) {
	// Check if the username already exists.
	existing, _ := s.userRepo.GetByUsername(req.Username)
	if existing != nil {
		return nil, errors.ErrUsernameExists
	}

	// Hash the password.
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("密码加密失败")
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       1, // default active
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return user, nil
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

// List returns a paginated list of users and the total count.
func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return users, total, nil
}

// Update modifies editable profile fields of an existing user.
func (s *UserService) Update(id int64, req *UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return user, nil
}

// Delete removes a user by their ID.
func (s *UserService) Delete(id int64) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if err := s.userRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// UpdateStatus changes the active/disabled status of a user.
func (s *UserService) UpdateStatus(id int64, status int8) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}

	user.Status = status
	if err := s.userRepo.Update(user); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}
