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

type DB interface {
	// Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
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

func (p *Postgres) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.DB.QueryContext(ctx, query, args...)
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}
