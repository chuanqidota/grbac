package service

import (
	"context"

	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/repository"
)

// UserInfo holds the user information returned to external systems.
type UserInfo struct {
	UserID       int64    `json:"user_id"`
	Username     string   `json:"username"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	Roles        []string `json:"roles"`
}

// ExternalService provides APIs consumed by external systems to query
// user identity, menus, and permissions.
type ExternalService struct {
	userRepo       *repository.UserRepo
	systemRepo     *repository.SystemRepo
	roleRepo       *repository.RoleRepo
	menuRepo       *repository.MenuRepo
	permissionRepo *repository.PermissionRepo
	jwtSecret      []byte
	authService    *AuthService
}

// NewExternalService creates a new ExternalService.
func NewExternalService(
	userRepo *repository.UserRepo,
	systemRepo *repository.SystemRepo,
	roleRepo *repository.RoleRepo,
	menuRepo *repository.MenuRepo,
	permissionRepo *repository.PermissionRepo,
	jwtSecret string,
	authService *AuthService,
) *ExternalService {
	return &ExternalService{
		userRepo:       userRepo,
		systemRepo:     systemRepo,
		roleRepo:       roleRepo,
		menuRepo:       menuRepo,
		permissionRepo: permissionRepo,
		jwtSecret:      []byte(jwtSecret),
		authService:    authService,
	}
}

// VerifyToken validates a JWT access token and returns its claims.
// It also checks whether the token has been blacklisted (logged out).
func (s *ExternalService) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	// Check blacklist first.
	if s.authService.IsTokenBlacklisted(ctx, token) {
		return nil, errors.ErrTokenExpired
	}

	claims, err := jwt.ParseToken(token, s.jwtSecret)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	// Only access tokens are accepted.
	if claims.Subject != "access" {
		return nil, errors.ErrTokenInvalid
	}

	return claims, nil
}

// GetUserInfo returns the user's identity and their roles within the
// specified system. Super-admin users get an empty role list since they
// bypass all RBAC checks.
func (s *ExternalService) GetUserInfo(ctx context.Context, userID int64, systemCode string) (*UserInfo, error) {
	// 1. Query user.
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.Status != 1 {
		return nil, errors.ErrAccountDisabled
	}

	info := &UserInfo{
		UserID:       user.ID,
		Username:     user.Username,
		IsSuperAdmin: user.IsSuperAdmin == 1,
		Roles:        []string{},
	}

	// Super-admin does not need role resolution.
	if info.IsSuperAdmin {
		return info, nil
	}

	// 2. Query system.
	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	// 3. Get user's roles within this system.
	roles, err := s.getUserRolesInSystem(userID, system.ID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0, len(roles))
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
	}
	info.Roles = roleNames

	return info, nil
}

// GetUserMenus returns the menus assigned to the user within the
// specified system. Super-admin users receive all menus of the system.
func (s *ExternalService) GetUserMenus(ctx context.Context, userID int64, systemCode string) ([]model.Menu, error) {
	// 1. Query system.
	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	// 2. Check super-admin – return all system menus.
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.IsSuperAdmin == 1 {
		return s.menuRepo.ListBySystem(system.ID)
	}

	// 3. Get user's roles within this system.
	roles, err := s.getUserRolesInSystem(userID, system.ID)
	if err != nil {
		return nil, err
	}

	// 4. Collect all menu IDs from those roles (deduplicated).
	menuIDSet := make(map[int64]struct{})
	for _, role := range roles {
		menuIDs, err := s.roleRepo.GetRoleMenus(role.ID)
		if err != nil {
			return nil, errors.ErrInternal.Wrap(err.Error())
		}
		for _, mid := range menuIDs {
			menuIDSet[mid] = struct{}{}
		}
	}

	if len(menuIDSet) == 0 {
		return []model.Menu{}, nil
	}

	ids := make([]int64, 0, len(menuIDSet))
	for id := range menuIDSet {
		ids = append(ids, id)
	}

	// 5. Query menu details.
	menus, err := s.menuRepo.ListByIDs(ids)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return menus, nil
}

// GetUserPermissions returns the permission codes assigned to the user
// within the specified system. Super-admin users receive an empty list
// since they bypass permission checks.
func (s *ExternalService) GetUserPermissions(ctx context.Context, userID int64, systemCode string) ([]string, error) {
	// 1. Query system.
	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	// 2. Check super-admin – no specific permissions needed.
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.IsSuperAdmin == 1 {
		// Super-admin has all permissions; return empty to signal "skip check".
		return []string{}, nil
	}

	// 3. Get user's roles within this system.
	roles, err := s.getUserRolesInSystem(userID, system.ID)
	if err != nil {
		return nil, err
	}

	// 4. Collect all permission IDs from those roles (deduplicated).
	permIDSet := make(map[int64]struct{})
	for _, role := range roles {
		permIDs, err := s.roleRepo.GetRolePermissions(role.ID)
		if err != nil {
			return nil, errors.ErrInternal.Wrap(err.Error())
		}
		for _, pid := range permIDs {
			permIDSet[pid] = struct{}{}
		}
	}

	if len(permIDSet) == 0 {
		return []string{}, nil
	}

	ids := make([]int64, 0, len(permIDSet))
	for id := range permIDSet {
		ids = append(ids, id)
	}

	// 5. Query permission details and collect codes.
	perms, err := s.permissionRepo.ListByIDs(ids)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	codes := make([]string, 0, len(perms))
	for _, p := range perms {
		codes = append(codes, p.Code)
	}

	return codes, nil
}

// getUserRolesInSystem returns the roles that the user holds within
// the specified system. It fetches all user-role mappings, then filters
// by system ID.
func (s *ExternalService) getUserRolesInSystem(userID, systemID int64) ([]model.Role, error) {
	userRoles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	var roles []model.Role
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ur.RoleID)
		if err != nil {
			// Role may have been deleted; skip.
			continue
		}
		if role.SystemID == systemID {
			roles = append(roles, *role)
		}
	}

	return roles, nil
}
