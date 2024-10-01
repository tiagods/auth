package mysql

import "database/sql"

func NewMysqlDB() *sql.DB {
	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/auth")
	if err != nil {
		panic(err.Error())
	}
	return db
}
