package system

import (
	"crypto/rand"
	"math/big"

	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	menuRepo "grbac/internal/repository/menu"
	permRepo "grbac/internal/repository/permission"
	roleRepo "grbac/internal/repository/role"
	systemRepo "grbac/internal/repository/system"
	userRepo "grbac/internal/repository/user"
)

// CreateRequest holds the payload for registering a new system.
type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// Service provides system CRUD and member management operations.
type Service struct {
	systemRepo     *systemRepo.Repo
	userRepo       *userRepo.Repo
	roleRepo       *roleRepo.Repo
	menuRepo       *menuRepo.Repo
	permissionRepo *permRepo.Repo
}

// NewService creates a new Service.
func NewService(
	systemRepo *systemRepo.Repo,
	userRepo *userRepo.Repo,
	roleRepo *roleRepo.Repo,
	menuRepo *menuRepo.Repo,
	permissionRepo *permRepo.Repo,
) *Service {
	return &Service{
		systemRepo:     systemRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		menuRepo:       menuRepo,
		permissionRepo: permissionRepo,
	}
}

// Create registers a new system with an auto-generated code.
func (s *Service) Create(req *CreateRequest) (*model.System, error) {
	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, err
	}

	system := &model.System{
		Name:        req.Name,
		Code:        code,
		Description: req.Description,
		Status:      1,
	}

	if err := s.systemRepo.Create(system); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return system, nil
}

const codeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (s *Service) generateUniqueCode() (string, error) {
	for i := 0; i < 10; i++ {
		b := make([]byte, 8)
		for j := range b {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars))))
			b[j] = codeChars[n.Int64()]
		}
		code := string(b)
		existing, _ := s.systemRepo.GetByCode(code)
		if existing == nil {
			return code, nil
		}
	}
	return "", errors.ErrInternal.Wrap("生成唯一系统编码失败")
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

// ListForUser returns systems visible to the given user.
// Super-admins see all systems; regular users see only systems they belong to.
func (s *Service) ListForUser(userID int64, isSuperAdmin bool, page, pageSize int) ([]model.System, int64, error) {
	if isSuperAdmin {
		return s.List(page, pageSize)
	}

	systems, err := s.systemRepo.ListByUserID(userID)
	if err != nil {
		return nil, 0, errors.ErrInternal.Wrap(err.Error())
	}
	return systems, int64(len(systems)), nil
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
	if role != "admin" && role != "member" {
		return errors.ErrInternal.Wrap("无效的角色值，必须是 admin 或 member")
	}

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

// GetMembers returns all members of a system with user info.
func (s *Service) GetMembers(systemID int64) ([]systemRepo.MemberInfo, error) {
	members, err := s.systemRepo.GetMembersWithUser(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return members, nil
}

// MemberUserInfo holds a user with their roles and permission count in a system.
type MemberUserInfo struct {
	UserID          int64      `json:"user_id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Roles           []model.Role `json:"roles"`
	PermissionCount int        `json:"permission_count"`
}

// GetMemberUsers returns users who have roles in the system, enriched with role and permission info.
func (s *Service) GetMemberUsers(systemID int64) ([]MemberUserInfo, error) {
	users, err := s.systemRepo.GetUsersWithRoles(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	var result []MemberUserInfo
	for _, u := range users {
		roles, err := s.userRepo.GetRolesInSystem(u.UserID, systemID)
		if err != nil {
			continue
		}

		// Count unique permissions across all roles.
		permIDSet := make(map[int64]struct{})
		if len(roles) > 0 {
			roleIDs := make([]int64, len(roles))
			for i, r := range roles {
				roleIDs[i] = r.ID
			}
			rolePerms, err := s.roleRepo.GetRolePermissionsByRoleIDs(roleIDs)
			if err == nil {
				for _, permIDs := range rolePerms {
					for _, pid := range permIDs {
						permIDSet[pid] = struct{}{}
					}
				}
			}
		}

		result = append(result, MemberUserInfo{
			UserID:          u.UserID,
			Username:        u.Username,
			Email:           u.Email,
			Roles:           roles,
			PermissionCount: len(permIDSet),
		})
	}

	if result == nil {
		result = []MemberUserInfo{}
	}
	return result, nil
}

// IsAdmin checks whether a user has the admin role in a system.
func (s *Service) IsAdmin(systemID, userID int64) (bool, error) {
	return s.systemRepo.IsAdmin(systemID, userID)
}

// GetMemberRoles returns the RBAC roles assigned to a user within a system.
func (s *Service) GetMemberRoles(systemID, userID int64) ([]model.Role, error) {
	roles, err := s.userRepo.GetRolesInSystem(userID, systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}
	return roles, nil
}

// MenuTree represents a menu node with children for tree output.
type MenuTree struct {
	model.Menu
	Children []*MenuTree `json:"children"`
}

func buildMenuTree(menus []model.Menu, parentID int64) []*MenuTree {
	var trees []*MenuTree
	for _, m := range menus {
		if m.ParentID == parentID {
			node := &MenuTree{
				Menu:     m,
				Children: buildMenuTree(menus, m.ID),
			}
			trees = append(trees, node)
		}
	}
	return trees
}

// GetMemberMenus returns the effective menu tree for a user within a system.
// Aggregates menus from all the user's RBAC roles (deduplicated).
func (s *Service) GetMemberMenus(systemID, userID int64) ([]*MenuTree, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Super admin gets all system menus.
	if user.IsSuperAdmin == 1 {
		menus, err := s.menuRepo.ListBySystem(systemID)
		if err != nil {
			return nil, errors.ErrInternal.Wrap(err.Error())
		}
		return buildMenuTree(menus, 0), nil
	}

	roles, err := s.userRepo.GetRolesInSystem(userID, systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	if len(roles) == 0 {
		return []*MenuTree{}, nil
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
		return []*MenuTree{}, nil
	}

	ids := make([]int64, 0, len(menuIDSet))
	for id := range menuIDSet {
		ids = append(ids, id)
	}

	menus, err := s.menuRepo.ListByIDs(ids)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return buildMenuTree(menus, 0), nil
}

// GetMemberPermissions returns the effective permissions for a user within a system.
// Aggregates permissions from all the user's RBAC roles (deduplicated).
func (s *Service) GetMemberPermissions(systemID, userID int64) ([]model.Permission, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Super admin has no explicit permissions (they bypass checks).
	if user.IsSuperAdmin == 1 {
		return []model.Permission{}, nil
	}

	roles, err := s.userRepo.GetRolesInSystem(userID, systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	if len(roles) == 0 {
		return []model.Permission{}, nil
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
		return []model.Permission{}, nil
	}

	ids := make([]int64, 0, len(permIDSet))
	for id := range permIDSet {
		ids = append(ids, id)
	}

	perms, err := s.permissionRepo.ListByIDs(ids)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return perms, nil
}
