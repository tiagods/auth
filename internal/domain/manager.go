package domain

import (
	service "github.com/tiagods/auth/internal/domain/services"
)

type (
	Manager struct {
		Token service.TokenService
	}
)

func NewManager(services ...any) Manager {
	manager := Manager{}

	for _, s := range services {
		switch s.(type) {
		case service.TokenService:
			manager.Token = s
		default:
		}
	}
	return manager
}
