package user

import (
	"grbac/internal/model"
	"grbac/internal/pkg/crypto"
	"grbac/internal/pkg/errors"
	roleRepo "grbac/internal/repository/role"
	systemRepo "grbac/internal/repository/system"
	userRepo "grbac/internal/repository/user"
	"gorm.io/gorm"
)

// CreateRequest holds the payload for creating a new user.
type CreateRequest struct {
	Username    string `json:"username"     binding:"required"`
	ChineseName string `json:"chinese_name"`
	Password    string `json:"password"     binding:"required,min=8"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

// UpdateRequest holds the payload for updating user profile fields.
type UpdateRequest struct {
	ChineseName string `json:"chinese_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

// UserRoleInfo holds role info with system context for a user.
type UserRoleInfo struct {
	SystemID   int64  `json:"system_id"`
	SystemName string `json:"system_name"`
	RoleID     int64  `json:"role_id"`
	RoleName   string `json:"role_name"`
	RoleCode   string `json:"role_code"`
}

// Service provides user CRUD operations.
type Service struct {
	db         *gorm.DB
	userRepo   *userRepo.Repo
	roleRepo   *roleRepo.Repo
	systemRepo *systemRepo.Repo
}

// NewService creates a new Service.
func NewService(db *gorm.DB, userRepo *userRepo.Repo, roleRepo *roleRepo.Repo, systemRepo *systemRepo.Repo) *Service {
	return &Service{db: db, userRepo: userRepo, roleRepo: roleRepo, systemRepo: systemRepo}
}

// Create registers a new user after validating uniqueness and hashing the password.
func (s *Service) Create(req *CreateRequest) (*model.User, error) {
	existing, _ := s.userRepo.GetByUsername(req.Username)
	if existing != nil {
		return nil, errors.ErrUsernameExists
	}

	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("密码加密失败")
	}

	user := &model.User{
		Username:     req.Username,
		ChineseName:  req.ChineseName,
		PasswordHash: hashedPassword,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return user, nil
}

// GetByID retrieves a user by their ID.
func (s *Service) GetByID(id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

// List returns a paginated list of users and the total count.
func (s *Service) List(page, pageSize int, q *userRepo.ListQuery) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize, q)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return users, total, nil
}

// Update modifies editable profile fields of an existing user.
func (s *Service) Update(id int64, req *UpdateRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if req.ChineseName != "" {
		user.ChineseName = req.ChineseName
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

// Delete removes a user and cleans up their role assignments and system memberships.
func (s *Service) Delete(id int64) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.RemoveAllUserRolesTx(tx, id); err != nil {
			return err
		}
		if err := s.userRepo.RemoveAllSystemMembersTx(tx, id); err != nil {
			return err
		}
		return s.userRepo.DeleteTx(tx, id)
	})
}

// UpdateStatus changes the active/disabled status of a user.
func (s *Service) UpdateStatus(id int64, status int8) error {
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

// ResetPassword changes a user's password (admin operation).
func (s *Service) ResetPassword(id int64, newPassword string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}

	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errors.ErrInternal.Wrap("密码加密失败")
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// UpdateSuperAdmin sets or unsets the super-admin flag for a user.
func (s *Service) UpdateSuperAdmin(id int64, isSuperAdmin int8) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// When revoking super admin, ensure at least one remains.
	if isSuperAdmin == 0 && user.IsSuperAdmin == 1 {
		count, err := s.userRepo.CountSuperAdmins()
		if err != nil {
			return errors.ErrInternal.Wrap(err.Error())
		}
		if count <= 1 {
			return errors.ErrInternal.Wrap("系统至少需要保留一个超管")
		}
	}

	user.IsSuperAdmin = isSuperAdmin
	if err := s.userRepo.Update(user); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// GetUserRoles returns all roles assigned to a user across all systems.
func (s *Service) GetUserRoles(userID int64) ([]UserRoleInfo, error) {
	userRoles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	var result []UserRoleInfo
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ur.RoleID)
		if err != nil {
			continue
		}
		sys, err := s.systemRepo.GetByID(role.SystemID)
		if err != nil {
			continue
		}
		result = append(result, UserRoleInfo{
			SystemID:   sys.ID,
			SystemName: sys.Name,
			RoleID:     role.ID,
			RoleName:   role.Name,
			RoleCode:   role.Code,
		})
	}

	if result == nil {
		result = []UserRoleInfo{}
	}
	return result, nil
}
