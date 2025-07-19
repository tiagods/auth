package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"github.com/nofeaturesonlybugs/set"
	"github.com/nofeaturesonlybugs/sqlh"
	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/tracer"
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

func NewDB(user, password, host, schema string) (*DbAdapter, error) {
	datasource := fmt.Sprintf("%s:%s@tcp(%s)/%s?loc=UTC&parseTime=true", user, password, host, schema)
	db, err := sql.Open("mysql", datasource)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar conexão: %w", err)
	}
	adapter := &DbAdapter{
		DB:               db,
		connectionString: datasource,
	}
	return adapter, nil
}

func (m *DbAdapter) Close() {
	if m.DB != nil {
		err := m.DB.Close()
		if err != nil {
			log.Error(err)
		}
	}
}

func (m *DbAdapter) Check(ctx context.Context) error {
	if m.DB == nil {
		err := m.Connect(ctx)
		if err != nil {
			return err
		}
	}
	if err := m.DB.Ping(); err != nil {
		return err
	}
	return nil
}

func (m *DbAdapter) StartPeriodicHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.HealthCheck(ctx); err != nil {
					logger.Error(ctx, err, "falha no health check periódico")
				}
			}
		}
	}()
}

func (m *DbAdapter) Connect(ctx context.Context) error {
	if m.DB != nil {
		return nil
	}

	db, err := sql.Open("mysql", m.connectionString)
	if err != nil {
		logger.Error(ctx, err, "falha ao abrir conexão com o banco")
		return err
	}

	// Adicionar timeout de contexto para teste de conexão
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error(ctx, err, "falha ao pingar o banco")
		return err
	}

	m.DB = db
	return nil
}

func (m *DbAdapter) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx, err, "falha ao iniciar transação")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (m *DbAdapter) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.DB.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("falha no health check: %w", err)
	}
	stats := m.DB.Stats()
	logger.Debug(ctx, fmt.Sprintf("conexões ociosas: %d", stats.Idle))
	logger.Debug(ctx, fmt.Sprintf("conexões em uso: %d", stats.InUse))
	logger.Debug(ctx, fmt.Sprintf("conexões ativas: %d", stats.OpenConnections))
	return nil
}

func (m *DbAdapter) Exec(ctx context.Context, returnID bool, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	caller := "database::exec"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindDB)
	defer span.End()

	if err := m.Check(ctx); err != nil {
		return 0, err
	}
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = m.DB.ExecContext(ctx, query, args...)
	}
	if err != nil {
		tracer.SetError(span, err)
		logger.Error(ctx, err, fmt.Sprintf("falha ao executar query: %s", query))
		return 0, err
	}
	var id int64
	if returnID {
		id, err = result.LastInsertId()
		if err != nil {
			tracer.SetError(span, err)
			logger.Error(ctx, err, "falha ao carregar ID do registro inserido")
			return id, err
		}
		if strings.Contains(query, "INSERT") && id == 0 {
			tracer.SetError(span, ErrNoRowsAffected)
			logger.Error(ctx, ErrNoRowsAffected, "nenhuma linha afetada na inserção")
			return id, ErrNoRowsAffected
		}
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		tracer.SetError(span, ErrNoRowsAffected)
		logger.Error(ctx, ErrNoRowsAffected, "nenhuma linha afetada na execução")
		return 0, ErrNoRowsAffected
	}
	return id, nil
}

func (m *DbAdapter) QueryRows(ctx context.Context, query string, args ...interface{}) ResultRows {
	caller := "database::query_rows"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindDB)
	defer span.End()

	if err := m.Check(ctx); err != nil {
		return ResultRows{err: err}
	}
	rows, err := m.QueryContext(ctx, query, args...)
	if err != nil {
		tracer.SetError(span, err)
		return ResultRows{rows: rows, err: err}
	}
	if rows.Err() != nil {
		tracer.SetError(span, rows.Err())
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
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindDB)
	defer span.End()

	if err := m.Check(ctx); err != nil {
		return ResultRow{err: err}
	}
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
	_, span := tracer.Start(ctx, caller, tracer.SpanKindDB)
	defer span.End()

	if err := m.Check(ctx); err != nil {
		return nil
	}
	scn := &sqlh.Scanner{
		Mapper: &set.Mapper{
			Tags: []string{"db", "json"},
		},
	}
	return scn
}
