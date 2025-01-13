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
	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/otel"
	"strings"
	"time"
)

type (
	DbAdapter struct {
		*sql.DB
		connectionString string
	}

	ResultRows struct {
		rows  *sql.Rows
		close func()
		err   error
	}
	ResultRow struct {
		row *sql.Row
		err error
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
	caller := "database::exec"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	m.Check()
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = m.DB.ExecContext(ctx, query, args...)
	}
	if err != nil {
		otel.SetError(span, err)
		return 0, err
	}
	if returnID {
		id, err := result.LastInsertId()
		if err != nil {
			otel.SetError(span, err)
			return id, err
		}
		if strings.Contains(query, "INSERT") && id == 0 {
			otel.SetError(span, ErrNoRowsAffected)
			return id, ErrNoRowsAffected
		}
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		otel.SetError(span, ErrNoRowsAffected)
		return 0, ErrNoRowsAffected
	}
	return 0, nil
}

func (m *DbAdapter) QueryRows(ctx context.Context, query string, args ...interface{}) ResultRows {
	caller := "database::query_rows"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	m.Check()
	rows, err := m.QueryContext(ctx, query, args...)
	if err != nil {
		otel.SetError(span, err)
		return ResultRows{rows: rows, err: err}
	}
	if rows.Err() != nil {
		otel.SetError(span, rows.Err())
		return ResultRows{rows: rows, err: rows.Err()}
	}
	closeRows := func() {
		err := rows.Close()
		if err != nil {
			logger.Error(ctx, err, err.Error())
		}
	}
	return ResultRows{close: closeRows, rows: rows, err: nil}
}

func (r ResultRows) Result() (rows *sql.Rows, shutdown func(), err error) {
	rows = r.rows
	shutdown = r.close
	err = r.err
	return rows, shutdown, err
}

func (r ResultRows) Close() {
	if r.close != nil && r.rows != nil && r.err == nil {
		r.close()
	}
}

func (m *DbAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) ResultRow {
	caller := "database::query_row"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	m.Check()
	row := m.QueryRowContext(ctx, query, args)
	return ResultRow{row: row, err: row.Err()}
}

func (r ResultRow) Scan(dest ...any) (err error) {
	err = r.err
	if err != nil {
		return err
	}
	err = r.row.Scan(dest...)
	return err
}

func (m *DbAdapter) GetSqlScanner(ctx context.Context) *sqlh.Scanner {
	caller := "database::get_sql_scanner"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	m.Check()
	scn := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"db", "json"},
		},
	}
	return scn
}
