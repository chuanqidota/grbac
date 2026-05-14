package menu

import (
	"context"
	"time"

	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	menuRepo "grbac/internal/repository/menu"
)

// DispatchFn is the signature for async webhook event dispatch.
type DispatchFn func(ctx context.Context, event string, payload interface{})

// CreateRequest holds the payload for creating or updating a menu.
type CreateRequest struct {
	ParentID  int64  `json:"parent_id"`
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// Tree represents a menu node with its children for tree output.
type Tree struct {
	model.Menu
	Children []*Tree `json:"children"`
}

// Service provides menu CRUD and tree-building operations.
type Service struct {
	menuRepo   *menuRepo.Repo
	dispatchFn DispatchFn
}

// NewService creates a new Service.
func NewService(menuRepo *menuRepo.Repo, dispatchFn DispatchFn) *Service {
	return &Service{menuRepo: menuRepo, dispatchFn: dispatchFn}
}

// Create registers a new menu within a system.
func (s *Service) Create(systemID int64, req *CreateRequest) (*model.Menu, error) {
	menu := &model.Menu{
		SystemID:  systemID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		Status:    1,
	}

	if err := s.menuRepo.Create(menu); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	if s.dispatchFn != nil {
		s.dispatchFn(context.Background(), "menu.created", map[string]interface{}{
			"event":     "menu.created",
			"timestamp": time.Now(),
			"system_id": systemID,
			"data":      map[string]interface{}{"menu_id": menu.ID, "menu_name": menu.Name},
		})
	}

	return menu, nil
}

// GetByID retrieves a menu by its ID.
func (s *Service) GetByID(id int64) (*model.Menu, error) {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrMenuNotFound
	}
	return menu, nil
}

// GetTree returns the full menu tree for a system.
func (s *Service) GetTree(systemID int64) ([]*Tree, error) {
	menus, err := s.menuRepo.ListBySystem(systemID)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return buildTree(menus, 0), nil
}

// Update modifies an existing menu.
func (s *Service) Update(id int64, req *CreateRequest) (*model.Menu, error) {
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

	if s.dispatchFn != nil {
		s.dispatchFn(context.Background(), "menu.updated", map[string]interface{}{
			"event":     "menu.updated",
			"timestamp": time.Now(),
			"system_id": menu.SystemID,
			"data":      map[string]interface{}{"menu_id": menu.ID, "menu_name": menu.Name},
		})
	}

	return menu, nil
}

// Delete removes a menu after verifying it has no children.
func (s *Service) Delete(systemID, id int64) error {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return errors.ErrMenuNotFound
	}

	if menu.SystemID != systemID {
		return errors.ErrMenuNotFound
	}

	hasChildren, err := s.menuRepo.HasChildren(systemID, id)
	if err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	if hasChildren {
		return errors.ErrMenuHasChildren
	}

	if err := s.menuRepo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}

	if s.dispatchFn != nil {
		s.dispatchFn(context.Background(), "menu.deleted", map[string]interface{}{
			"event":     "menu.deleted",
			"timestamp": time.Now(),
			"system_id": systemID,
			"data":      map[string]interface{}{"menu_id": id, "menu_name": menu.Name},
		})
	}

	return nil
}

func buildTree(menus []model.Menu, parentID int64) []*Tree {
	var trees []*Tree
	for _, m := range menus {
		if m.ParentID == parentID {
			node := &Tree{
				Menu:     m,
				Children: buildTree(menus, m.ID),
			}
			trees = append(trees, node)
		}
	}
	return trees
}
