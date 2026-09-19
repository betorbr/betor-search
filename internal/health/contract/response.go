package contract

import "time"

type Response struct {
	Status     string    `json:"status"`
	Components []string  `json:"components"`
	Timestamp  time.Time `json:"timestamp"`
}
