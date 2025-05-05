package books

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sing3demons/go-library-api/pkg/entities"
	"github.com/sing3demons/go-library-api/pkg/kp"
	"github.com/sing3demons/go-library-api/pkg/postgres"
	"go.opentelemetry.io/otel"
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
	cmd := "create_book"
	c, span := otel.GetTracerProvider().Tracer("gokp").Start(ctx.Context(), fmt.Sprintf("%s-%s", node_postgres, cmd))
	defer span.End()
	// var lastInsertId string
	// err := r.Db.QueryRowContext(ctx, "INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id", book.Title, book.Author).Scan(&lastInsertId)
	// if err != nil {
	// 	return err
	// }

	// book.ID = lastInsertId
	invoke := uuid.NewString()
	result, err := r.Db.CreateBook(c, entities.Book{
		Title:  book.Title,
		Author: book.Author,
	})
	ctx.DetailLog().AddOutputRequest(node_postgres, cmd, invoke, result.RawData, result.Body, node_postgres, "")

	if err != nil {
		ctx.DetailLog().AddInputResponse(node_postgres, cmd, invoke, err.Error(), map[string]string{
			"error": err.Error(),
		})
		return err
	}

	// fmt.Println("raw: ", result.RawData)
	// detailLog.End()
	book.ID = result.Data.ID
	book.Href = r.href(book.ID)

	ctx.DetailLog().AddInputResponse(node_postgres, cmd, invoke, book, book)

	return nil
}

func (r *MongoBookRepository) GetByID(ctx kp.IContext, id string) (*Book, error) {
	cmd := "get_book"
	c, span := otel.GetTracerProvider().Tracer("gokp").Start(ctx.Context(), fmt.Sprintf("%s-%s", node_postgres, cmd))
	defer span.End()

	invoke := uuid.NewString()
	var book Book
	// err := r.Db.QueryRowContext(ctx, "SELECT id, title, author FROM books WHERE id = $1", id).Scan(&book.ID, &book.Title, &book.Author)
	result, err := r.Db.GetBookByID(c, id)
	ctx.DetailLog().AddOutputRequest(node_postgres, cmd, invoke, result.RawData, result.Body, node_postgres, "")

	if err != nil {
		ctx.DetailLog().AddInputResponse(node_postgres, cmd, "", err.Error(), map[string]string{
			"error": err.Error(),
		})
		return nil, err
	}
	book.ID = result.Data.ID
	book.Title = result.Data.Title
	book.Author = result.Data.Author

	book.Href = r.href(book.ID)

	// fmt.Println("RawData: ", result.RawData)
	// detailLog.End()
	ctx.DetailLog().AddInputResponse(node_postgres, cmd, "", "", book)

	return &book, nil
}

func (r *MongoBookRepository) href(id string) string {
	return fmt.Sprintf("/books/%s", id)
}

func (r *MongoBookRepository) GetALL(ctx kp.IContext, filter map[string]interface{}) ([]*Book, error) {
	cmd := "get_books"
	c, span := otel.GetTracerProvider().Tracer("gokp").Start(ctx.Context(), fmt.Sprintf("%s-%s", node_postgres, cmd))
	defer span.End()
	invoke := uuid.NewString()
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
	result, err := r.Db.GetAllBooks(c, filter)
	ctx.DetailLog().AddOutputRequest(node_postgres, "get_book", invoke, result.RawData, result.Body, node_postgres, "")

	if err != nil {
		ctx.DetailLog().AddInputResponse(node_postgres, cmd, invoke, err.Error(), map[string]string{
			"error": err.Error(),
		})
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
	ctx.DetailLog().AddInputResponse(node_postgres, cmd, invoke, "", result)

	// detailLog.End()
	return books, nil
}
