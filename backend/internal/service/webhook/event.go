package webhook

import "time"

// EventPayload is the standard envelope for all webhook events.
type EventPayload struct {
	Event      string      `json:"event"`
	Timestamp  time.Time   `json:"timestamp"`
	SystemCode string      `json:"system_code"`
	Data       interface{} `json:"data"`
}
