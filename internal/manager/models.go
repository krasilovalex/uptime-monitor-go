package manager

import "time"

type Target struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	IntervalSeconds string    `json:"interval_seconds"`
	TimeoutSeconds  int       `json:"timeout_seconds"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

type PingTask struct {
	TargetID       string `json:"target_id"`
	URL            string `json:"url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}
