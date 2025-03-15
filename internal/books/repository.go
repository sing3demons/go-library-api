package books

import "context"

type BookRepository interface {
    GetByID(ctx context.Context, id string) (*Book, error)
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