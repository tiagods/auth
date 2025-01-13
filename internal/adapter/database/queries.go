package database

const (
	// Table Users
	InsertUser                    = `INSERT INTO Users (Username, Password) VALUES (?,?)`
	FindUserByUserNameAndPassword = `SELECT u.ID as id, u.Username as username, u.Password as password FROM Users u WHERE u.Username=? AND u.Password=?`
	ListUsers                     = `SELECT u.ID as id, u.Username as username, u.Password as password FROM Users u LIMIT ? OFFSET ?`
	// Table RefreshTokens
	FindRefreshToken = `
		SELECT r.ID as id, r.User_ID as user_id, r.CreatedAt as created_at, r.ExpiresAt as expires_at
		FROM RefreshTokens r
		WHERE ExpiresAt < ?
	`
	FindRefreshTokenAddUserID = `AND r.User_ID = ?`

	FindRefreshTokenAddID = `AND r.ID = ?`

	FindRefreshTokenByUser = `
		SELECT r.ID as id, r.User_ID as user_id, r.CreatedAt as created_at, r.ExpiresAt as expires_at
		FROM RefreshTokens r
		WHERE r.User_ID = ?
	`

	DeleteRefreshToken = `DELETE FROM RefreshTokens WHERE ID = ?`

	InsertRefreshToken = `INSERT INTO RefreshTokens(ID,User_ID,CreatedAt,ExpiresAt) VALUES (?,?,?,?)`
)
