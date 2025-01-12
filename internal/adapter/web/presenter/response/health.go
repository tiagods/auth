package response

import "github.com/tiagods/auth/internal/domain/entity"

type Health struct {
	Status  bool            `json:"status"`
	Service map[string]bool `json:"service"`
}

func FromEntity(health entity.Health) Health {
	return Health{Status: health.Status, Service: health.Service}
}
