package webhook

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"

	"grbac/internal/model"
	"grbac/internal/pkg/errors"
	webhookRepo "grbac/internal/repository/webhook"
)

// CreateRequest holds the payload for registering a webhook.
type CreateRequest struct {
	URL    string `json:"url" binding:"required"`
	Events string `json:"events" binding:"required"` // comma-separated event types
}

// UpdateRequest holds the payload for updating a webhook.
type UpdateRequest struct {
	URL    string `json:"url"`
	Events string `json:"events"`
	Status *int8  `json:"status"`
}

// Service provides webhook CRUD and event dispatch operations.
type Service struct {
	repo       *webhookRepo.Repo
	dispatcher *Dispatcher
}

// NewService creates a new Service.
func NewService(repo *webhookRepo.Repo) *Service {
	svc := &Service{
		repo:       repo,
		dispatcher: NewDispatcher(repo),
	}
	return svc
}

// Create registers a new webhook with an auto-generated secret.
func (s *Service) Create(systemID int64, req *CreateRequest) (*model.Webhook, error) {
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, errors.ErrInternal.Wrap("生成webhook密钥失败")
	}
	secret := hex.EncodeToString(secretBytes)

	webhook := &model.Webhook{
		SystemID: systemID,
		URL:      req.URL,
		Secret:   secret,
		Events:   req.Events,
		Status:   1,
	}

	if err := s.repo.Create(webhook); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return webhook, nil
}

// GetByID retrieves a webhook by its ID.
func (s *Service) GetByID(id int64) (*model.Webhook, error) {
	webhook, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("webhook不存在")
	}
	return webhook, nil
}

// GetBySystemID returns all webhooks for a system.
func (s *Service) GetBySystemID(systemID int64) ([]model.Webhook, error) {
	return s.repo.GetBySystemID(systemID)
}

// List returns a paginated list of webhooks.
func (s *Service) List(page, pageSize int) ([]model.Webhook, int64, error) {
	return s.repo.List(page, pageSize)
}

// Update modifies a webhook's URL, events, or status.
func (s *Service) Update(id int64, req *UpdateRequest) (*model.Webhook, error) {
	webhook, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.ErrInternal.Wrap("webhook不存在")
	}

	if req.URL != "" {
		webhook.URL = req.URL
	}
	if req.Events != "" {
		webhook.Events = req.Events
	}
	if req.Status != nil {
		webhook.Status = *req.Status
	}

	if err := s.repo.Update(webhook); err != nil {
		return nil, errors.ErrInternal.Wrap(err.Error())
	}

	return webhook, nil
}

// Delete removes a webhook.
func (s *Service) Delete(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return errors.ErrInternal.Wrap(err.Error())
	}
	return nil
}

// DispatchEvent sends an event to all matching webhooks asynchronously.
func (s *Service) DispatchEvent(ctx context.Context, event string, payload interface{}) {
	go s.dispatcher.Dispatch(ctx, event, payload)
}

// validEvents is the set of all recognized webhook event types.
var validEvents = map[string]bool{
	"role.created":              true,
	"role.updated":              true,
	"role.deleted":              true,
	"role.menus_assigned":       true,
	"role.permissions_assigned": true,
	"role.users_assigned":       true,
	"role.user_removed":         true,
	"menu.created":              true,
	"menu.updated":              true,
	"menu.deleted":              true,
	"permission.created":        true,
	"permission.updated":        true,
	"permission.deleted":        true,
	"system.created":            true,
	"system.updated":            true,
	"system.deleted":            true,
	"system.member_added":       true,
	"system.member_removed":     true,
}

// ParseEvents parses an events string that may be a JSON array ("[\"a\",\"b\"]")
// or comma-separated ("a,b") and returns a normalized comma-separated string.
func ParseEvents(raw string) (string, []string) {
	raw = strings.TrimSpace(raw)
	var list []string
	// Try JSON array first.
	if strings.HasPrefix(raw, "[") {
		_ = json.Unmarshal([]byte(raw), &list)
	}
	// Fall back to comma-separated.
	if list == nil {
		for _, e := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(e); t != "" {
				list = append(list, t)
			}
		}
	}
	return strings.Join(list, ","), list
}

// ValidateEvents checks that all event strings are valid.
func ValidateEvents(events string) bool {
	_, list := ParseEvents(events)
	for _, e := range list {
		if !validEvents[e] {
			return false
		}
	}
	return len(list) > 0
}
