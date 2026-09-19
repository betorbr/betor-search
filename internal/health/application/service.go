package application

import (
	"time"

	"betor-search/internal/health/contract"
)

type CatalogStatus interface {
	Status() string
	LastExecution() time.Time
	LastSuccess() *time.Time
	LastError() string
}

type Component struct {
	Name    string
	Status  string
	Updated time.Time
	LastRun time.Time
	Message string
	Catalog CatalogStatus
}

type Service struct {
	componentName string
	catalog       CatalogStatus
	startedAt     time.Time
}

func NewService() Service {
	return Service{componentName: "betor-search-catalog", startedAt: time.Now().UTC()}
}

func NewServiceWithCatalog(componentName string, catalog CatalogStatus) Service {
	if componentName == "" {
		componentName = "betor-search-catalog"
	}
	return Service{componentName: componentName, catalog: catalog, startedAt: time.Now().UTC()}
}

func (s Service) Response(now time.Time) contract.Response {
	status := "UP"
	components := []contract.Component{}
	if s.catalog != nil {
		componentStatus := s.catalog.Status()
		status = componentStatus
		if componentStatus == "" {
			componentStatus = "DOWN"
		}

		component := contract.Component{
			Name:    s.componentName,
			Status:  componentStatus,
			Updated: now.UTC(),
			LastRun: s.catalog.LastExecution().UTC(),
			Message: s.catalog.LastError(),
		}
		if s.catalog.LastSuccess() != nil {
			component.Updated = s.catalog.LastSuccess().UTC()
		}
		if s.catalog.LastError() == "" && componentStatus == "UP" {
			component.Message = "catalog sync successful"
		}
		components = append(components, component)
		if componentStatus == "DOWN" {
			status = "DOWN"
		} else if componentStatus == "DEGRADED" {
			status = "DEGRADED"
		}
	}
	if len(components) == 0 {
		status = "UP"
	}
	return contract.Response{
		Status:     status,
		Components: components,
		StartedAt:  s.startedAt.UTC(),
		Timestamp:  now.UTC(),
	}
}
