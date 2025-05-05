package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/sing3demons/go-library-api/pkg/entities"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	GetBookByID(ctx context.Context, id string) (entities.ProcessData[entities.Book], error)
	GetAllBooks(ctx context.Context, filter map[string]any) (result entities.ProcessData[[]entities.Book], err error)
	CreateBook(ctx context.Context, book entities.Book) (entities.ProcessData[entities.Book], error)
}

type Postgres struct {
	*sql.DB
	tracer trace.Tracer
}

func New() (DB, error) {
	databaseSource := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", "localhost", 5432, "root", "password", "product_master")

	// otelRegisteredDialect, _ := otelsql.Register("postgres")
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
	return &Postgres{DB: db, tracer: otel.GetTracerProvider().Tracer("gokp-postgres")}, nil
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

func (c *Postgres) addTrace(ctx context.Context, method, table string) (context.Context, trace.Span) {
	if c.tracer != nil {
		contextWithTrace, span := c.tracer.Start(ctx, fmt.Sprintf("postgres-%v", method))

		span.SetAttributes(
			attribute.String("postgres.table", table),
		)

		return contextWithTrace, span
	}

	return ctx, nil
}

func (c *Postgres) sendOperationStats(startTime time.Time, method string, span trace.Span) {
	duration := time.Since(startTime).Microseconds()

	if span != nil {
		defer span.End()
		span.SetAttributes(attribute.Int64(fmt.Sprintf("postgres.%v.duration", method), duration))
	}
}
