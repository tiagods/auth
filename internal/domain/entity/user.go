package entity

import (
	"fmt"
	"time"
)

type (
	UserCredential struct {
		ID       int64
		Username string
		Password string
	}

	User struct {
		ID       int64
		Username string
	}

	Token struct {
		ID     string
		UserID int64
	}

	RefreshToken struct {
		ID        string
		UserID    int64
		ExpiresAt time.Time
	}
)

func (r RefreshToken) GetKey() string {
	return fmt.Sprintf("refreshtoken::id::%d", r.UserID)
}

func (r Token) GetKey() string {
	return fmt.Sprintf("token::id::%d", r.UserID)
}
