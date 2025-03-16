package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

type DB interface {
	// Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	Close() error
}

type Postgres struct {
	*sql.DB
}

func New() (DB, error) {
	databaseSource := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", "localhost", 5432, "root", "password", "product_master")
	db, err := sql.Open("postgres", databaseSource)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// return &Postgres{Db: &sqlDBWrapper{DB: db}}, nil
	return &Postgres{DB: db}, nil
}

// func (p *Postgres) Exec(query string, args ...interface{}) (sql.Result, error) {
// 	return p.Db.Exec(query, args...)
// }

func (p *Postgres) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return p.DB.QueryRowContext(ctx, query, args...)
}

func (p *Postgres) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	// return p.DB.QueryContext(ctx, query, args...)
	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	// return NewSqlRowsWrapper(rows), nil
	return rows, nil
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}

func NewSqlRowsWrapper(rows *sql.Rows) Rows {
	return &sqlRowsWrapper{rows: rows}
}

type sqlRowsWrapper struct {
	rows *sql.Rows
}

func (r *sqlRowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *sqlRowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *sqlRowsWrapper) Close() error {
	return r.rows.Close()
}

func (r *sqlRowsWrapper) Err() error {
	return r.rows.Err()
}
