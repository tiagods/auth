package entity

type (
	User struct {
		ID       string
		Username string
	}

	Token struct {
		UserID string
		Token  string
	}

	RefreshToken struct {
		RefreshToken string
		UserID       string
	}
)

func (r RefreshToken) GetKey() string {
	return "refreshtoken::id" + r.UserID
}

func (r Token) GetKey() string {
	return "token::id::" + r.UserID
}
