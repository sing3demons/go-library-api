package postgres

import (
	"fmt"
	"strings"

	"github.com/sing3demons/go-library-api/pkg/entities"
)

func (p *Postgres) GetBookByID(id string) (entities.ProcessData[entities.Book], error) {
	query := "SELECT id, title, author FROM books WHERE id = $1"

	result := entities.ProcessData[entities.Book]{}

	result.Body.Collection = "books"
	result.Body.Table = "books"
	result.Body.Query = map[string]string{"id": id}
	result.RawData = fmt.Sprintf("SELECT id, title, author FROM books WHERE id = %s", id)

	rows, err := p.DB.Query(query, id)
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
func (p *Postgres) GetAllBooks(filter map[string]any) (result entities.ProcessData[[]entities.Book], err error) {
	query := "SELECT id, title, author FROM books"

	var keys []string
	var values []any

	result.Body.Collection = "books"
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

	fmt.Println("query: ", query)

	result.RawData = rawData

	rows, err := p.DB.Query(query, values...)
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
