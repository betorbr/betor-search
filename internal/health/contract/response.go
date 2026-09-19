package contract

import "time"

type Component struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Updated   time.Time `json:"updated_at"`
	LastRun   time.Time `json:"last_run_at"`
	StartedAt time.Time `json:"started_at,omitempty"`
	Message   string    `json:"message,omitempty"`
}

type Response struct {
	Status     string      `json:"status"`
	Components []Component `json:"components"`
	Timestamp  time.Time   `json:"timestamp"`
}
