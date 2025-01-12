package model

import "github.com/tiagods/auth/internal/domain/entity"

type (
	User struct {
		ID       int64  `db:"id"`
		Username string `db:"Username"`
		Password string `db:"password"`
	}
)

func (u User) ToEntity() *entity.User {
	return &entity.User{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
	}
}
