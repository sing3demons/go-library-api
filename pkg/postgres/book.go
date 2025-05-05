package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sing3demons/go-library-api/pkg/entities"
)

func (p *Postgres) GetBookByID(ctx context.Context, id string) (entities.ProcessData[entities.Book], error) {
	query := "SELECT id, title, author FROM books WHERE id = $1"

	result := entities.ProcessData[entities.Book]{}

	result.Body.Collection = "books"
	result.Body.Table = "books"
	result.Body.Query = map[string]string{"id": id}
	result.RawData = fmt.Sprintf("SELECT id, title, author FROM books WHERE id = %s", id)

	result.Body.Method = "findOne"
	ctx, span := p.addTrace(ctx, result.Body.Method, result.Body.Table)
	defer p.sendOperationStats(time.Now(), result.Body.Method, span)

	rows, err := p.DB.QueryContext(ctx, query, id)
	if err != nil {
		return result, err
	}

	defer rows.Close()

	for rows.Next() {
		var b entities.Book
		err := rows.Scan(&b.ID, &b.Title, &b.Author)
		if err != nil {
			return result, err
		}
		result.Data = b
	}
	return result, nil
}

// GetAllBooks returns all books from the database
func (p *Postgres) GetAllBooks(ctx context.Context, filter map[string]any) (result entities.ProcessData[[]entities.Book], err error) {
	query := "SELECT id, title, author FROM books"

	var keys []string
	var values []any

	result.Body.Table = "books"

	index := 1
	for k, v := range filter {
		keys = append(keys, fmt.Sprintf("%s = $%d", k, index))
		index++
		values = append(values, v)
	}

	if len(keys) > 0 {
		query += " WHERE " + keys[0]
		for i := 1; i < len(keys); i++ {
			query += " AND " + keys[i]
		}
	}

	var rawData string = query

	if len(values) > 0 {
		for i := 1; i <= len(values); i++ {
			rawData = strings.Replace(rawData, fmt.Sprintf("$%d", i), fmt.Sprintf("%v", values[i-1]), 1)
		}
	}

	result.RawData = rawData

	result.Body.Method = "find"
	ctx, span := p.addTrace(ctx, result.Body.Method, result.Body.Table)
	defer p.sendOperationStats(time.Now(), result.Body.Method, span)

	rows, err := p.DB.QueryContext(ctx, query, values...)
	if err != nil {
		return result, err
	}

	defer rows.Close()

	var books []entities.Book
	for rows.Next() {
		var b entities.Book
		err := rows.Scan(&b.ID, &b.Title, &b.Author)
		if err != nil {
			return result, err
		}
		books = append(books, b)
	}
	result.Data = books
	return result, nil
}

func (p *Postgres) CreateBook(ctx context.Context, book entities.Book) (entities.ProcessData[entities.Book], error) {
	query := "INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id"

	var result entities.ProcessData[entities.Book]
	result.RawData = query
	result.Body.Table = "books"

	result.Body.Method = "insert"
	ctx, span := p.addTrace(ctx, result.Body.Method, result.Body.Table)
	defer p.sendOperationStats(time.Now(), result.Body.Method, span)
	result.RawData = strings.Replace(result.RawData, "$1", book.Title, 1)
	result.RawData = strings.Replace(result.RawData, "$2", book.Author, 1)

	err := p.DB.QueryRowContext(ctx, query, book.Title, book.Author).Scan(&book.ID)
	if err != nil {
		return result, err
	}

	result.Data = book
	return result, nil
}
