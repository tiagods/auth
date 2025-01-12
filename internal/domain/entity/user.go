package entity

import "fmt"

type (
	User struct {
		ID       int64
		Username string
		Password string
	}

	Token struct {
		UserID int64
		Token  string
	}

	RefreshToken struct {
		ID     string
		UserID int64
	}
)

func (r RefreshToken) GetKey() string {
	return fmt.Sprintf("refreshtoken::id::%d", r.UserID)
}

func (r Token) GetKey() string {
	return fmt.Sprintf("token::id::%s", r.UserID)
}
