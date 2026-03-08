package domain

import "time"

type Event struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}
