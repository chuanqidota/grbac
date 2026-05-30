package webhook

import (
	"grbac/internal/model"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(webhook *model.Webhook) error {
	return r.db.Create(webhook).Error
}

func (r *Repo) GetByID(id int64) (*model.Webhook, error) {
	var webhook model.Webhook
	err := r.db.Where("id = ?", id).First(&webhook).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *Repo) GetBySystemID(systemID int64) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	err := r.db.Where("system_id = ?", systemID).Find(&webhooks).Error
	if err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (r *Repo) List(page, pageSize int) ([]model.Webhook, int64, error) {
	var webhooks []model.Webhook
	var total int64

	if err := r.db.Model(&model.Webhook{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&webhooks).Error
	if err != nil {
		return nil, 0, err
	}

	return webhooks, total, nil
}

func (r *Repo) Update(webhook *model.Webhook) error {
	return r.db.Save(webhook).Error
}

func (r *Repo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Webhook{}).Error
}

// GetByEvent returns all active webhooks subscribed to the given event type.
func (r *Repo) GetByEvent(event string) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	query := `status = 1 AND (
		events = ? OR
		events LIKE ? OR
		events LIKE ? OR
		events LIKE ?
	)`
	err := r.db.Where(query,
		event,             // exact match: "role.created"
		event+",%",        // prefix: "role.created,menu.updated"
		"%,"+event,        // suffix: "menu.updated,role.created"
		"%,"+event+",%",   // middle: "a,role.created,b"
	).Find(&webhooks).Error
	if err != nil {
		return nil, err
	}
	return webhooks, nil
}
