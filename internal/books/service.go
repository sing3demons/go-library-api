package books

import (
	"context"

	"github.com/sing3demons/go-library-api/kp/logger"
)

type BookService interface {
	GetBook(ctx context.Context, id string) (*Book, error)
	CreateBook(ctx context.Context, book *Book) error
	GetAllBooks(ctx context.Context, filter map[string]any, detailLog logger.DetailLog, summaryLog logger.SummaryLog) ([]*Book, error)
}

type bookService struct {
	repo BookRepository
}

func NewBookService(repo BookRepository) BookService {
	return &bookService{repo: repo}
}

func (s *bookService) GetBook(ctx context.Context, id string) (*Book, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *bookService) CreateBook(ctx context.Context, book *Book) error {
	return s.repo.Save(ctx, book)
}

func (s *bookService) GetAllBooks(ctx context.Context, filter map[string]any, detailLog logger.DetailLog, summaryLog logger.SummaryLog) ([]*Book, error) {
	node := "postgres"
	cmd := "get_books"

	result, err := s.repo.GetALL(ctx, filter, detailLog)
	if err != nil {
		detailLog.AddInputRequest(node, cmd, "", err.Error(), map[string]string{
			"error": err.Error(),
		})
		summaryLog.AddError(node, cmd, "", err.Error())
		return nil, err
	}
	detailLog.AddInputRequest(node, cmd, "", "", result)
	summaryLog.AddSuccess(node, cmd, "20000", "success")

	return result, err
}
