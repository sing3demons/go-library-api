package books

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/sing3demons/go-library-api/kp"
	"github.com/sing3demons/go-library-api/pkg/entities"
	"github.com/sing3demons/go-library-api/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

const (
	mockDatabaseError = "mock database error"
	unsupported       = "unsupported type for Scan"
)

type MockDB struct {
	ExpectedID string
	ShouldFail bool
	book       *Book
	books      []Book
	next       bool
	err        error
}

var book = Book{
	ID:     "123",
	Title:  "Test Book",
	Author: "Test Author",
	Href:   "",
}

func (m *MockDB) CreateBook(book entities.Book) (entities.ProcessData[entities.Book], error) {
	var result entities.ProcessData[entities.Book]

	result.Body.Collection = "books"
	result.Body.Table = "books"
	result.Body.Method = "POST"
	result.Body.Document = book
	result.RawData = fmt.Sprintf("INSERT INTO books (title, author) VALUES (%s, %s)", book.Title, book.Author)

	if m.ShouldFail {
		return result, errors.New(mockDatabaseError)
	}

	if m.err != nil {
		return result, m.err
	}

	book.ID = m.ExpectedID
	if book.Title != "" {
		result.Data.Title = book.Title
	}
	if book.Author != "" {
		result.Data.Author = book.Author
	}

	result.Data.ID = m.ExpectedID

	return result, nil
}

func (m *MockDB) GetAllBooks(filter map[string]any) (result entities.ProcessData[[]entities.Book], err error) {
	result.Body.Collection = "books"
	result.Body.Table = "books"
	result.Body.Query = filter
	result.RawData = "SELECT id, title, author FROM books"
	if m.ShouldFail {
		return result, errors.New(mockDatabaseError)
	}

	if m.err != nil {
		return result, m.err
	}

	if m.book != nil {
		result.Data = append(result.Data, entities.Book{
			ID:     m.book.ID,
			Title:  m.book.Title,
			Author: m.book.Author,
		})
	}

	if m.books != nil {
		for _, book := range m.books {
			result.Data = append(result.Data, entities.Book{
				ID:     book.ID,
				Title:  book.Title,
				Author: book.Author,
			})
		}
	}

	return result, nil
}

func (m *MockDB) GetBookByID(id string) (entities.ProcessData[entities.Book], error) {
	var result entities.ProcessData[entities.Book]

	result.Body.Collection = "books"
	result.Body.Table = "books"
	result.Body.Query = map[string]string{"id": id}
	result.RawData = fmt.Sprintf("SELECT id, title, author FROM books WHERE id = %s", id)

	if m.ShouldFail {
		return result, errors.New(mockDatabaseError)
	}

	if m.book != nil {
		result.Data.ID = m.book.ID
		result.Data.Title = m.book.Title
		result.Data.Author = m.book.Author
	}

	return result, nil
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) postgres.Row {
	if m.ShouldFail {
		return &MockRow{err: errors.New(mockDatabaseError)}
	}

	if m.book != nil {
		return &MockRow{Values: []Book{
			{
				ID:     m.book.ID,
				Title:  m.book.Title,
				Author: m.book.Author,
			},
		},
			query: query,
		}
	}

	return &MockRow{Values: []Book{{ID: m.ExpectedID}}, query: query}
}

type MockRow struct {
	Values []Book
	err    error
	query  string
}

func (r *MockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if len(dest) > 0 {
		if len(r.Values) == 1 {
			columns, _ := extractColumnsFromQuery(r.query)
			if len(columns) == 1 && len(dest) != 0 && columns[0] == "*" {
				v := reflect.ValueOf(r.Values[0])
				destIndex := 0

				for i := 0; i < v.NumField(); i++ {
					if destIndex >= len(dest) {
						break
					}

					field := v.Type().Field(i)
					if field.Name == "Href" {
						continue
					}
					data := v.Field(i).Interface()
					reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(data))
					destIndex++
				}
				r.Values = r.Values[1:]
				return nil
			} else if len(dest) != len(columns) {
				book := r.Values[0]
				switch d := dest[0].(type) {
				case *string:
					*d = book.ID
				default:
					return errors.New(unsupported)
				}
				r.Values = r.Values[1:]
				return nil
			}

			scan(columns, r.Values[0], dest...)
			return nil

		} else if len(r.Values) > 0 {
			book := r.Values[0]
			switch d := dest[0].(type) {
			case *string:
				*d = book.ID
			default:
				return errors.New(unsupported)
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
	query  string
}

func (r *MockRows) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if r.next && len(r.Values) > 0 {
		columns, _ := extractColumnsFromQuery(r.query)
		if len(columns) == 1 && len(dest) != 0 && columns[0] == "*" {
			v := reflect.ValueOf(r.Values[0])
			destIndex := 0

			for i := 0; i < v.NumField(); i++ {
				if destIndex >= len(dest) {
					break
				}

				field := v.Type().Field(i)
				if field.Name == "Href" {
					continue
				}
				data := v.Field(i).Interface()
				reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(data))
				destIndex++
			}
			r.Values = r.Values[1:]
			return nil
		} else if len(dest) != len(columns) {
			book := r.Values[0]
			switch d := dest[0].(type) {
			case *string:
				*d = book.ID
			default:
				return errors.New(unsupported)
			}
			// Remove the processed row
			r.Values = r.Values[1:]
			return nil
		}

		scan(columns, r.Values[0], dest...)
		r.Values = r.Values[1:]
		return nil
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
		return nil, errors.New(mockDatabaseError)
	}

	results := []Book{}

	if m.book != nil {
		results = append(results, *m.book)
	}

	if m.books != nil {
		results = append(results, m.books...)
	}

	return &MockRows{
		Values: results,
		next:   m.next,
		err:    m.err,
		query:  query,
	}, nil
}
func (m *MockDB) Close() error {
	return nil
}

func extractColumnsFromQuery(query string) ([]string, error) {
	// Trim spaces and convert the query to lowercase
	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)

	// Validate if the query starts with "select"
	if !strings.HasPrefix(queryLower, "select") {
		if strings.Contains(queryLower, "insert") {
			fromIndex := strings.Index(queryLower, "returning")
			if fromIndex != -1 {
				q := query[fromIndex+len("returning"):]
				q = strings.TrimSpace(q)
				q = strings.ReplaceAll(q, ";", "")

				return []string{q}, nil
			}
		}
		return nil, fmt.Errorf("invalid query: must start with SELECT")
	}

	// Find the position of "FROM" to limit the extraction of columns before it
	fromIndex := strings.Index(queryLower, "from")
	if fromIndex == -1 {
		return nil, fmt.Errorf("invalid query: missing FROM clause")
	}

	// Extract the part of the query between "SELECT" and "FROM"
	columnsPart := query[6:fromIndex]
	columnsPart = strings.TrimSpace(columnsPart)

	// Split the columns by commas
	columns := strings.Split(columnsPart, ",")
	var result []string

	// Clean up each column name by trimming spaces
	for _, column := range columns {
		column = strings.TrimSpace(column)
		result = append(result, column)
	}

	return result, nil
}

func scan(columns []string, data Book, dest ...any) error {
	if len(dest) != len(columns) {
		return errors.New("number of columns does not match number of fields")
	}

	reflectValue := reflect.ValueOf(&data)
	fieldMap := make(map[string]reflect.Value)

	for index := 0; index < reflectValue.Elem().NumField(); index++ {
		field := reflectValue.Elem().Type().Field(index)
		fieldMap[strings.ToUpper(field.Name)] = reflectValue.Elem().Field(index)
	}

	for i, column := range columns {
		normalizedColumn := strings.ToUpper(column)

		if field, found := fieldMap[normalizedColumn]; found {
			reflect.ValueOf(dest[i]).Elem().Set(field)
		}
	}
	return nil
}

func TestSave(t *testing.T) {
	t.Run("should save a book success", func(t *testing.T) {
		mockDB := &MockDB{ExpectedID: "123", ShouldFail: false}
		repo := NewPostgresBookRepository(mockDB)

		ctx := kp.NewMockContext()
		err := repo.Save(ctx, &book)
		defer ctx.Verify(t)

		assert.NoError(t, err)
		assert.Equal(t, "123", book.ID)

		expectedHref := "/books/123"
		assert.Equal(t, expectedHref, book.Href)
	})

	t.Run("should fail to save a book", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)
		ctx := kp.NewMockContext()
		err := repo.Save(ctx, &book)
		ctx.Verify(t)

		assert.Error(t, err)
	})
}

func TestGetByID(t *testing.T) {
	t.Run("should get a book by id", func(t *testing.T) {
		mockDB := &MockDB{
			ShouldFail: false,
			book:       &book,
		}
		repo := NewPostgresBookRepository(mockDB)

		book, err := repo.GetByID(kp.NewMockContext(), "123")

		assert.NoError(t, err)
		assert.Equal(t, "123", book.ID)
	})

	t.Run("should fail to get a book by id", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)

		book, err := repo.GetByID(kp.NewMockContext(), "123")

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
			books:      []Book{{ID: "456", Title: "Test Book 2", Author: "Test Author 2"}},
		}
		repo := NewPostgresBookRepository(mockDB)

		books, err := repo.GetALL(kp.NewMockContext(), nil)

		assert.NoError(t, err)
		assert.NotEmpty(t, books)
	})

	t.Run("should fail to get all books", func(t *testing.T) {
		mockDB := &MockDB{ShouldFail: true}
		repo := NewPostgresBookRepository(mockDB)

		books, err := repo.GetALL(kp.NewMockContext(), nil)

		assert.Error(t, err)
		assert.Nil(t, books)
	})

	t.Run("should error scan", func(t *testing.T) {
		mockDB := &MockDB{
			ShouldFail: false,
			book:       &book,
			err:        errors.New(mockDatabaseError),
		}
		repo := NewPostgresBookRepository(mockDB)

		books, err := repo.GetALL(kp.NewMockContext(), nil)

		assert.Error(t, err)
		assert.Nil(t, books)
	})

}
