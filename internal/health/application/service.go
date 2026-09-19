package application

import (
	"time"

	"betor-search/internal/health/contract"
)

type Service struct{}

func NewService() Service {
	return Service{}
}

func (Service) Response(now time.Time) contract.Response {
	return contract.Response{
		Status:     "UP",
		Components: []string{},
		Timestamp:  now.UTC(),
	}
}
