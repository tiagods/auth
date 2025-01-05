package service

import "reflect"

type (
	Manager struct {
		Token TokenService
	}
)

func NewManager(services ...any) Manager {
	manager := Manager{}

	for _, s := range services {
		name := reflect.TypeOf(s).String()
		switch name {
		case "*service.TokenService":
			manager.Token = s.(*tokenService)
		}
	}
	return manager
}
