package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"grbac/internal/pkg/crypto"
	webhookRepo "grbac/internal/repository/webhook"
)

// retryDelays defines the backoff schedule: 1s, 5s, 30s.
var retryDelays = []time.Duration{
	time.Second,
	5 * time.Second,
	30 * time.Second,
}

// Dispatcher sends webhook events to subscribed endpoints.
type Dispatcher struct {
	repo   *webhookRepo.Repo
	client *http.Client
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(repo *webhookRepo.Repo) *Dispatcher {
	return &Dispatcher{
		repo: repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Dispatch finds all webhooks subscribed to the given event and sends the payload.
func (d *Dispatcher) Dispatch(ctx context.Context, event string, payload interface{}) {
	webhooks, err := d.repo.GetByEvent(event)
	if err != nil {
		log.Printf("webhook dispatch: failed to get webhooks for event %s: %v", event, err)
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhook dispatch: failed to marshal payload: %v", err)
		return
	}

	for _, wh := range webhooks {
		events := strings.Split(wh.Events, ",")
		matched := false
		for _, e := range events {
			if strings.TrimSpace(e) == event {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		go d.sendWithRetry(wh.URL, wh.Secret, event, body)
	}
}

func (d *Dispatcher) sendWithRetry(url, secret, event string, body []byte) {
	signature := crypto.GenerateHMAC(body, []byte(secret))

	for attempt, delay := range retryDelays {
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("webhook send: invalid URL %s: %v", url, err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Event", event)
		req.Header.Set("X-Webhook-Signature", signature)

		resp, err := d.client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		if attempt < len(retryDelays)-1 {
			time.Sleep(delay)
		}
	}

	log.Printf("webhook send: all retries exhausted for %s (event: %s)", url, event)
}
