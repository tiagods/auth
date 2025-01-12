package database

const (
	// Table Users
	InsertUser                    = `INSERT INTO Users (Username, Password) VALUES (?,?)`
	FindUserByUserNameAndPassword = `SELECT ID, Username, Password FROM Users WHERE username=? AND password=?`

	// Table RefreshTokens
	FindRefreshToken = `
		SELECT ID, User_ID, CreatedAt, ExpiresAt
		FROM RefreshTokens
		WHERE ExpiresAt < ?
	`
	FindRefreshTokenAddUserID = `AND User_ID = ?`

	FindRefreshTokenAddID = `AND ID = ?`

	FindRefreshTokenByUser = `
		SELECT ID, User_ID, CreatedAt, ExpiresAt
		FROM RefreshTokens
		WHERE User_ID = ?
	`

	DeleteRefreshToken = `DELETE FROM RefreshTokens WHERE ID = ?`

	InsertRefreshToken = `INSERT INTO RefreshTokens(ID,User_ID,CreatedAt,ExpiresAt) VALUES (?,?,?,?)`
)
