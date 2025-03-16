package books

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/sing3demons/go-library-api/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

type MockDB struct {
	ExpectedID string
	ShouldFail bool
	book       *Book
	next       bool
	err        error
	query      string
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) postgres.Row {
	if m.ShouldFail {
		return &MockRow{err: errors.New("mock database error")}
	}

	if m.book != nil {
		return &MockRow{Values: []Book{
			{
				ID:     m.ExpectedID,
				Title:  m.book.Title,
				Author: m.book.Author,
			},
		}}
	}

	return &MockRow{Values: []Book{{ID: m.ExpectedID}}}
}

type MockRow struct {
	Values []Book
	err    error
}

func (r *MockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if len(dest) > 0 {
		if len(r.Values) > 0 {
			book := r.Values[0]
			switch d := dest[0].(type) {
			case *string:
				*d = book.ID
			default:
				return errors.New("unsupported type for Scan")
			}
			r.Values = r.Values[1:]
		}
	}

	return nil
}

func (r MockRow) Err() error {
	return r.err
}

// createSQLRows simulates a *sql.Rows object
func (m *MockDB) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}

type MockRows struct {
	Values []Book
	err    error
	next   bool
}

func (r *MockRows) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	// Handle the scanning logic
	if len(dest) > 0 {
		// Check if we have at least one row to scan
		if len(r.Values) > 0 {
			// Extract the first value from the Values slice
			book := r.Values[0]
			// Assign fields of `book` to the dest slice
			switch d := dest[0].(type) {
			case *string:
				*d = book.ID
			default:
				return errors.New("unsupported type for Scan")
			}
			// Remove the processed row
			r.Values = r.Values[1:]
		}
	}
	return nil
}

func (r *MockRows) Next() bool {
	return len(r.Values) != 0
}

func (r *MockRows) Close() error {
	return nil
}

func (r *MockRows) Err() error {
	return r.err
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...any) (postgres.Rows, error) {
	if m.ShouldFail {
		return nil, errors.New("mock database error")
	}

	return &MockRows{
		Values: []Book{
			{
				ID:     m.ExpectedID,
				Title:  m.book.Title,
				Author: m.book.Author,
			},
		},
		next: m.next,
		err:  m.err,
	}, nil
}
func (m *MockDB) Close() error {
	return nil
}

var book = Book{
	ID:    "123",
	Title:  "Test Book",
	Author: "Test Author",
}

func TestSave(t *testing.T) {
	t.Run("should save a book success", func(t *testing.T) {
		mockDB := &MockDB{ExpectedID: "123", ShouldFail: false}
		repo := NewPostgresBookRepository(mockDB)

		err := repo.Save(context.Background(), &book)

		assert.NoError(t, err)
		assert.Equal(t, "123", book.ID)

		expectedHref := "/books/123"
		assert.Equal(t, expectedHref, book.Href)
	})

	t.Run("should fail to save a book", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)

		err := repo.Save(context.Background(), &book)

		assert.Error(t, err)
	})
}

func TestGetByID(t *testing.T) {
	t.Run("should get a book by id", func(t *testing.T) {

		mockDB := &MockDB{
			ExpectedID: "123",
			ShouldFail: false,
			book:       &book,
		}
		repo := NewPostgresBookRepository(mockDB)

		book, err := repo.GetByID(context.Background(), "123")

		assert.NoError(t, err)
		assert.Equal(t, "123", book.ID)
	})

	t.Run("should fail to get a book by id", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)

		book, err := repo.GetByID(context.Background(), "123")

		assert.Error(t, err)
		assert.Nil(t, book)
	})
}

func TestGetALL(t *testing.T) {
	t.Run("should get all books", func(t *testing.T) {
		mockDB := &MockDB{
			ExpectedID: "123",
			ShouldFail: false,
			book:       &book,
			next:       true,
		}
		repo := NewPostgresBookRepository(mockDB)

		books, err := repo.GetALL(context.Background(), nil)

		assert.NoError(t, err)
		assert.Len(t, books, 1)
	})

	t.Run("should fail to get all books", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)

		books, err := repo.GetALL(context.Background(), nil)

		assert.Error(t, err)
		assert.Nil(t, books)
	})
	t.Run("should error scan", func(t *testing.T) {
		mockDB := &MockDB{
			ShouldFail: false,
			book:       &book,
			err:        errors.New("mock database error"),
		}
		repo := NewPostgresBookRepository(mockDB)

		books, err := repo.GetALL(context.Background(), nil)

		assert.Error(t, err)
		assert.Nil(t, books)
	})

}
