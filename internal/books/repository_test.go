package books

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/sing3demons/go-library-api/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

type Database interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *FakeRow
}

// FakeRow simulates the behavior of `sql.Row`
type FakeRow struct {
	id  string
	err error
}

// Scan simulates scanning a database row into variables
func (r *FakeRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		ptr, ok := dest[0].(*string)
		if ok {
			*ptr = r.id
		}
	}
	return nil
}

type MockDB struct {
	ExpectedID string
	ShouldFail bool
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) postgres.Row {
	if m.ShouldFail {
		return &MockRow{Err: errors.New("mock database error")}
	}
	return &MockRow{Values: []any{m.ExpectedID}}
}

type MockRow struct {
	Values []any
	Err    error
}

func (r *MockRow) Scan(dest ...any) error {
	if r.Err != nil {
		return r.Err
	}
	if len(dest) > 0 {
		if strPtr, ok := dest[0].(*string); ok {
			*strPtr = r.Values[0].(string)
		}
	}
	return nil
}

// createSQLRows simulates a *sql.Rows object
func (m *MockDB) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockDB) Close() error {
	return nil
}

func TestSave(t *testing.T) {
	t.Run("should save a book success", func(t *testing.T) {
		mockDB := &MockDB{ExpectedID: "123", ShouldFail: false}
		repo := NewPostgresBookRepository(mockDB)

		book := &Book{
			Title:  "Test Book",
			Author: "Test Author",
		}

		err := repo.Save(context.Background(), book)

		assert.NoError(t, err)
		assert.Equal(t, "123", book.ID)

		expectedHref := "/books/123"
		assert.Equal(t, expectedHref, book.Href)
	})

	t.Run("should fail to save a book", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)

		book := &Book{
			Title: "Test Book",
		}

		err := repo.Save(context.Background(), book)

		assert.Error(t, err)
	})

}
