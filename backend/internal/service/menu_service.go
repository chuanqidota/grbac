package service

import (
	"github.com/jinang/grbac/internal/model"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/repository"
)

// CreateMenuRequest holds the payload for creating or updating a menu.
type CreateMenuRequest struct {
	ParentID  int64  `json:"parent_id"`
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// MenuTree represents a menu node with its children for tree output.
type MenuTree struct {
	model.Menu
	Children []*MenuTree `json:"children"`
}

// MenuService provides menu CRUD and tree-building operations.
type MenuService struct {
	menuRepo *repository.MenuRepo
}

// NewMenuService creates a new MenuService.
func NewMenuService(menuRepo *repository.MenuRepo) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

// Create registers a new menu within a system.
func (s *MenuService) Create(systemID int64, req *CreateMenuRequest) (*model.Menu, error) {
	menu := &model.Menu{
		SystemID:  systemID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		Status:    1, // default active
	}

	if err := s.menuRepo.Create(menu); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return menu, nil
}

// GetByID retrieves a menu by its ID.
func (s *MenuService) GetByID(id int64) (*model.Menu, error) {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrMenuNotFound
	}
	return menu, nil
}

// GetTree returns the full menu tree for a system.
func (s *MenuService) GetTree(systemID int64) ([]*MenuTree, error) {
	menus, err := s.menuRepo.ListBySystem(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return buildMenuTree(menus, 0), nil
}

// Update modifies an existing menu.
func (s *MenuService) Update(id int64, req *CreateMenuRequest) (*model.Menu, error) {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrMenuNotFound
	}

	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Path = req.Path
	menu.Icon = req.Icon
	menu.SortOrder = req.SortOrder

	if err := s.menuRepo.Update(menu); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return menu, nil
}

// Delete removes a menu after verifying it has no children.
func (s *MenuService) Delete(systemID, id int64) error {
	// 1. Check if the menu exists.
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return errors.ErrMenuNotFound
	}

	// 2. Verify the menu belongs to the specified system.
	if menu.SystemID != systemID {
		return errors.ErrMenuNotFound
	}

	// 3. Check for child menus.
	hasChildren, err := s.menuRepo.HasChildren(systemID, id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if hasChildren {
		return errors.ErrMenuHasChildren
	}

	// 4. Delete the menu.
	if err := s.menuRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	return nil
}

// buildMenuTree recursively constructs a tree from a flat menu list.
// parentID = 0 selects root nodes.
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
