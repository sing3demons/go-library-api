package books

import (
	"context"

	"github.com/sing3demons/go-library-api/kp/logger"
)

type BookService interface {
	GetBook(ctx context.Context, id string, detailLog logger.DetailLog, summaryLog logger.SummaryLog) (*Book, error)
	CreateBook(ctx context.Context, book *Book, detailLog logger.DetailLog, summaryLog logger.SummaryLog) error
	GetAllBooks(ctx context.Context, filter map[string]any, detailLog logger.DetailLog, summaryLog logger.SummaryLog) ([]*Book, error)
}

type bookService struct {
	repo BookRepository
}

func NewBookService(repo BookRepository) BookService {
	return &bookService{repo: repo}
}

func (s *bookService) GetBook(ctx context.Context, id string, detailLog logger.DetailLog, summaryLog logger.SummaryLog) (*Book, error) {
	cmd := "get_book"
	result, err := s.repo.GetByID(ctx, id, detailLog)
	if err != nil {
		detailLog.AddInputRequest(node_postgres, cmd, "", err.Error(), map[string]string{
			"error": err.Error(),
		})
		summaryLog.AddError(node_postgres, cmd, "", err.Error())
		return nil, err
	}
	detailLog.AddInputRequest(node_postgres, cmd, "", "", result)
	summaryLog.AddSuccess(node_postgres, cmd, "20000", "success")
	return result, nil
}

func (s *bookService) CreateBook(ctx context.Context, book *Book, detailLog logger.DetailLog, summaryLog logger.SummaryLog) error {
	err := s.repo.Save(ctx, book, detailLog)
	if err != nil {
		detailLog.AddInputRequest(node_postgres, "create_book", "", err.Error(), map[string]string{
			"error": err.Error(),
		})
		summaryLog.AddError(node_postgres, "create_book", "", err.Error())
		return err
	}
	detailLog.AddInputRequest(node_postgres, "create_book", "", "", book)
	summaryLog.AddSuccess(node_postgres, "create_book", "20000", "success")
	return nil
}

func (s *bookService) GetAllBooks(ctx context.Context, filter map[string]any, detailLog logger.DetailLog, summaryLog logger.SummaryLog) ([]*Book, error) {
	cmd := "get_books"

	result, err := s.repo.GetALL(ctx, filter, detailLog)
	if err != nil {
		detailLog.AddInputRequest(node_postgres, cmd, "", err.Error(), map[string]string{
			"error": err.Error(),
		})
		summaryLog.AddError(node_postgres, cmd, "", err.Error())
		return nil, err
	}
	detailLog.AddInputRequest(node_postgres, cmd, "", "", result)
	summaryLog.AddSuccess(node_postgres, cmd, "20000", "success")

	return result, nil
}
