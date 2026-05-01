package worker

import "time"

type PingTask struct {
	TargetID       string `json:"target_id"`
	URL            string `json:"url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type PingResult struct {
	TargetID     string    `json:"target_id"`
	IsUp         bool      `json:"is_up"`
	StatusCode   int       `json:"status_code,omitempty"`
	LatencyMs    int       `json:"latency_ms"`
	ErrorMessage string    `json:"erorr_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}
