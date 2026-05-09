package webhook

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

// ValidateEvents checks that all event strings are valid.
func ValidateEvents(events string) bool {
	validEvents := map[string]bool{
		"permission_change": true,
		"menu_change":       true,
		"role_change":       true,
	}
	for _, e := range strings.Split(events, ",") {
		if !validEvents[strings.TrimSpace(e)] {
			return false
		}
	}
	return true
}
