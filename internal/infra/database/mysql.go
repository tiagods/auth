package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh"
	"strings"
	"time"
)

type (
	DbAdapter struct {
		*sql.DB
		connectionString string
	}
)

var ErrNoRowsAffected = errors.New("no rows affected")

func NewDB(user, password, host, schema string) *DbAdapter {
	datasource := fmt.Sprintf("%s:%s@tcp(%s)/%s?loc=UTC&parseTime=true", user, password, host, schema)
	db := newDBWithConnectionString(datasource)
	return db
}

func newDBWithConnectionString(connectionString string) *DbAdapter {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return &DbAdapter{DB: db}
}

func (m *DbAdapter) Close() {
	if m.DB != nil {
		err := m.DB.Close()
		if err != nil {
			log.Error(err)
		}
	}
}

func (m *DbAdapter) Check() {
	if m.DB == nil {
		m.DB = newDBWithConnectionString(m.connectionString).DB
	}
	if err := m.DB.Ping(); err != nil {
		panic(err.Error())
	}
}

func (m *DbAdapter) Exec(ctx context.Context, returnID bool, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	m.Check()
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = m.DB.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return 0, err
	}
	if returnID {
		id, err := result.LastInsertId()
		if err != nil {
			return id, err
		}
		if strings.Contains(query, "INSERT") && id == 0 {
			return id, ErrNoRowsAffected
		}
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, ErrNoRowsAffected
	}
	return 0, nil
}

func (m *DbAdapter) QueryRows(ctx context.Context, query string, slice []interface{}, scanner *sqlh.Scanner, args ...interface{}) error {
	m.Check()
	if scanner == nil {
		scanner = &sqlh.Scanner{
			Mapper: &set.Mapper{
				Tags: []string{"DB", "json"},
			},
		}
	}
	err := scanner.Select(m.DB, &slice, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (m *DbAdapter) QueryRow(ctx context.Context, query string, dest interface{}, args ...interface{}) error {
	m.Check()
	scn := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"DB", "json"},
		},
	}

	err := scn.Select(m.DB, dest, query, args...)
	if err != nil {
		return err
	}
	return nil
}
