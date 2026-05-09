package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/jwt"
	menuRepo "grbac/internal/repository/menu"
	permRepo "grbac/internal/repository/permission"
	roleRepo "grbac/internal/repository/role"
	systemRepo "grbac/internal/repository/system"
	userRepo "grbac/internal/repository/user"
	authService "grbac/internal/service/auth"
)

const (
	cacheTTL       = 5 * time.Minute
	cachePrefix    = "grbac:ext:"
	permCacheFmt   = "user:perms:%s:%d"
	menuCacheFmt   = "user:menus:%s:%d"
	infoCacheFmt   = "user:info:%s:%d"
)

// UserInfo holds the user information returned to external systems.
type UserInfo struct {
	UserID       int64    `json:"user_id"`
	Username     string   `json:"username"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	Roles        []string `json:"roles"`
}

// Service provides APIs consumed by external systems to query
// user identity, menus, and permissions.
type Service struct {
	userRepo       *userRepo.Repo
	systemRepo     *systemRepo.Repo
	roleRepo       *roleRepo.Repo
	menuRepo       *menuRepo.Repo
	permissionRepo *permRepo.Repo
	jwtSecret      []byte
	authSvc        *authService.Service
	rdb            *redis.Client
}

// NewService creates a new Service.
func NewService(
	userRepo *userRepo.Repo,
	systemRepo *systemRepo.Repo,
	roleRepo *roleRepo.Repo,
	menuRepo *menuRepo.Repo,
	permissionRepo *permRepo.Repo,
	jwtSecret string,
	authSvc *authService.Service,
	rdb *redis.Client,
) *Service {
	return &Service{
		userRepo:       userRepo,
		systemRepo:     systemRepo,
		roleRepo:       roleRepo,
		menuRepo:       menuRepo,
		permissionRepo: permissionRepo,
		jwtSecret:      []byte(jwtSecret),
		authSvc:        authSvc,
		rdb:            rdb,
	}
}

// VerifyToken validates a JWT access token and returns its claims.
func (s *Service) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	if s.authSvc.IsTokenBlacklisted(ctx, token) {
		return nil, errors.ErrTokenExpired
	}

	claims, err := jwt.ParseToken(token, s.jwtSecret)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims.Subject != "access" {
		return nil, errors.ErrTokenInvalid
	}

	return claims, nil
}

// GetUserInfo returns the user's identity and their roles within the specified system.
func (s *Service) GetUserInfo(ctx context.Context, userID int64, systemCode string) (*UserInfo, error) {
	cacheKey := fmt.Sprintf(cachePrefix+infoCacheFmt, systemCode, userID)

	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var info UserInfo
		if json.Unmarshal(cached, &info) == nil {
			return &info, nil
		}
	}

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

	if info.IsSuperAdmin {
		s.cacheResult(ctx, cacheKey, info)
		return info, nil
	}

	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	roles, err := s.userRepo.GetRolesInSystem(userID, system.ID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	roleNames := make([]string, 0, len(roles))
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
	}
	info.Roles = roleNames

	s.cacheResult(ctx, cacheKey, info)
	return info, nil
}

// GetUserMenus returns the menus assigned to the user within the specified system.
func (s *Service) GetUserMenus(ctx context.Context, userID int64, systemCode string) ([]model.Menu, error) {
	cacheKey := fmt.Sprintf(cachePrefix+menuCacheFmt, systemCode, userID)

	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var menus []model.Menu
		if json.Unmarshal(cached, &menus) == nil {
			return menus, nil
		}
	}

	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.IsSuperAdmin == 1 {
		menus, err := s.menuRepo.ListBySystem(system.ID)
		if err != nil {
			return nil, errors.ErrInternal.Wrap(err.Error())
		}
		s.cacheResult(ctx, cacheKey, menus)
		return menus, nil
	}

	roles, err := s.userRepo.GetRolesInSystem(userID, system.ID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	if len(roles) == 0 {
		empty := []model.Menu{}
		s.cacheResult(ctx, cacheKey, empty)
		return empty, nil
	}

	roleIDs := make([]int64, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	roleMenus, err := s.roleRepo.GetRoleMenusByRoleIDs(roleIDs)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	menuIDSet := make(map[int64]struct{})
	for _, menuIDs := range roleMenus {
		for _, mid := range menuIDs {
			menuIDSet[mid] = struct{}{}
		}
	}

	if len(menuIDSet) == 0 {
		empty := []model.Menu{}
		s.cacheResult(ctx, cacheKey, empty)
		return empty, nil
	}

	ids := make([]int64, 0, len(menuIDSet))
	for id := range menuIDSet {
		ids = append(ids, id)
	}

	menus, err := s.menuRepo.ListByIDs(ids)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	s.cacheResult(ctx, cacheKey, menus)
	return menus, nil
}

// GetUserPermissions returns the permission codes assigned to the user within the specified system.
func (s *Service) GetUserPermissions(ctx context.Context, userID int64, systemCode string) ([]string, error) {
	cacheKey := fmt.Sprintf(cachePrefix+permCacheFmt, systemCode, userID)

	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var codes []string
		if json.Unmarshal(cached, &codes) == nil {
			return codes, nil
		}
	}

	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return nil, errors.ErrSystemNotFound
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.IsSuperAdmin == 1 {
		empty := []string{}
		s.cacheResult(ctx, cacheKey, empty)
		return empty, nil
	}

	roles, err := s.userRepo.GetRolesInSystem(userID, system.ID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	if len(roles) == 0 {
		empty := []string{}
		s.cacheResult(ctx, cacheKey, empty)
		return empty, nil
	}

	roleIDs := make([]int64, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	rolePerms, err := s.roleRepo.GetRolePermissionsByRoleIDs(roleIDs)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	permIDSet := make(map[int64]struct{})
	for _, permIDs := range rolePerms {
		for _, pid := range permIDs {
			permIDSet[pid] = struct{}{}
		}
	}

	if len(permIDSet) == 0 {
		empty := []string{}
		s.cacheResult(ctx, cacheKey, empty)
		return empty, nil
	}

	ids := make([]int64, 0, len(permIDSet))
	for id := range permIDSet {
		ids = append(ids, id)
	}

	perms, err := s.permissionRepo.ListByIDs(ids)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	codes := make([]string, 0, len(perms))
	for _, p := range perms {
		codes = append(codes, p.Code)
	}

	s.cacheResult(ctx, cacheKey, codes)
	return codes, nil
}

// ValidatePermission checks whether the user holds a specific permission (method + path).
// Uses a direct DB JOIN query instead of loading all permission codes.
func (s *Service) ValidatePermission(ctx context.Context, userID int64, systemCode, method, path string) (bool, error) {
	system, err := s.systemRepo.GetByCode(systemCode)
	if err != nil {
		return false, errors.ErrSystemNotFound
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, errors.ErrUserNotFound
	}

	if user.IsSuperAdmin == 1 {
		return true, nil
	}

	return s.permissionRepo.HasUserPermission(userID, system.ID, method, path)
}

func (s *Service) cacheResult(ctx context.Context, key string, value interface{}) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	s.rdb.Set(ctx, key, data, cacheTTL)
}
