package books

import (
	"context"
	"fmt"

	"github.com/sing3demons/go-library-api/pkg/postgres"
)

type BookRepository interface {
	GetByID(ctx context.Context, id string) (*Book, error)
	GetALL(ctx context.Context, filter map[string]interface{}) ([]*Book, error)
	Save(ctx context.Context, book *Book) error
}

type MongoBookRepository struct {
	Db *postgres.Postgres
}

func NewMongoBookRepository(db *postgres.Postgres) *MongoBookRepository {
	return &MongoBookRepository{Db: db}
}

func (r *MongoBookRepository) Save(ctx context.Context, book *Book) error {
	var lastInsertId string
	err := r.Db.Db.QueryRowContext(ctx, "INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id", book.Title, book.Author).Scan(&lastInsertId)
	if err != nil {
		return err
	}

	book.ID = lastInsertId
	book.Href = r.href(book.ID)
	return nil
}

func (r *MongoBookRepository) GetByID(ctx context.Context, id string) (*Book, error) {
	var book Book
	err := r.Db.Db.QueryRowContext(ctx, "SELECT id, title, author FROM books WHERE id = $1", id).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		return nil, err
	}
	book.Href = r.href(book.ID)
	return &book, nil
}

func (r *MongoBookRepository) href(id string) string {
	return fmt.Sprintf("/books/%s", id)
}

func (r *MongoBookRepository) GetALL(ctx context.Context, filter map[string]interface{}) ([]*Book, error) {
	var books []*Book
	rows, err := r.Db.Db.QueryContext(ctx, "SELECT id, title, author FROM books")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author)
		if err != nil {
			return nil, err
		}
		book.Href = r.href(book.ID)
		books = append(books, &book)
	}
	return books, nil
}

type InMemoryBookRepository struct {
	books map[string]*Book
}

func NewInMemoryBookRepository() *InMemoryBookRepository {
	return &InMemoryBookRepository{
		books: map[string]*Book{
			"1": {ID: "1", Title: "The Go Programming Language", Author: "Donovan"},
		},
	}
}

func (r *InMemoryBookRepository) GetByID(ctx context.Context, id string) (*Book, error) {
	book, exists := r.books[id]
	if !exists {
		return nil, nil // Simulate "not found"
	}
	return book, nil
}
