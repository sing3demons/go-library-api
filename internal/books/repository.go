package books

import (
	"fmt"
	"time"

	"github.com/sing3demons/go-library-api/kp"
	"github.com/sing3demons/go-library-api/pkg/entities"
	"github.com/sing3demons/go-library-api/pkg/postgres"
)

type BookRepository interface {
	GetByID(ctx kp.IContext, id string) (*Book, error)
	GetALL(ctx kp.IContext, filter map[string]any) ([]*Book, error)
	Save(ctx kp.IContext, book *Book) error
}

type MongoBookRepository struct {
	Db postgres.DB
}

func NewPostgresBookRepository(db postgres.DB) *MongoBookRepository {
	return &MongoBookRepository{Db: db}
}

const (
	node_postgres = "postgres"
)

func (r *MongoBookRepository) Save(ctx kp.IContext, book *Book) error {
	// var lastInsertId string
	// err := r.Db.QueryRowContext(ctx, "INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id", book.Title, book.Author).Scan(&lastInsertId)
	// if err != nil {
	// 	return err
	// }

	// book.ID = lastInsertId

	result, err := r.Db.CreateBook(entities.Book{
		Title:  book.Title,
		Author: book.Author,
	})
	ctx.DetailLog().AddOutputRequest(node_postgres, "create_book", fmt.Sprintf("pg-%s", time.Nanosecond.String()), result.RawData, result.Body)

	if err != nil {
		return err
	}

	// fmt.Println("raw: ", result.RawData)
	// detailLog.End()
	book.ID = result.Data.ID
	book.Href = r.href(book.ID)
	return nil
}

func (r *MongoBookRepository) GetByID(ctx kp.IContext, id string) (*Book, error) {
	var book Book
	// err := r.Db.QueryRowContext(ctx, "SELECT id, title, author FROM books WHERE id = $1", id).Scan(&book.ID, &book.Title, &book.Author)
	result, err := r.Db.GetBookByID(id)
	if err != nil {
		return nil, err
	}
	book.ID = result.Data.ID
	book.Title = result.Data.Title
	book.Author = result.Data.Author

	book.Href = r.href(book.ID)

	// fmt.Println("RawData: ", result.RawData)
	ctx.DetailLog().AddOutputRequest(node_postgres, "get_book", fmt.Sprintf("pg-%s", time.Nanosecond.String()), result.RawData, result.Body)
	// detailLog.End()

	return &book, nil
}

func (r *MongoBookRepository) href(id string) string {
	return fmt.Sprintf("/books/%s", id)
}

func (r *MongoBookRepository) GetALL(ctx kp.IContext, filter map[string]interface{}) ([]*Book, error) {
	var books []*Book
	// rows, err := r.Db.QueryContext(ctx, "SELECT id, title, author FROM books")
	// if err != nil {
	// 	return nil, err
	// }
	// for rows.Next() {
	// 	var book Book
	// 	err := rows.Scan(&book.ID, &book.Title, &book.Author)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	book.Href = r.href(book.ID)
	// 	books = append(books, &book)
	// }
	result, err := r.Db.GetAllBooks(filter)
	if err != nil {
		return nil, err
	}
	for _, b := range result.Data {
		book := Book{
			ID:     b.ID,
			Title:  b.Title,
			Author: b.Author,
			Href:   r.href(b.ID),
		}
		books = append(books, &book)
	}
	ctx.DetailLog().AddOutputRequest(node_postgres, "get_book", fmt.Sprintf("pg-%s", time.Nanosecond.String()), result.RawData, result.Body)
	// detailLog.End()
	return books, nil
}
