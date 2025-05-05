package books

import "github.com/sing3demons/go-library-api/pkg/kp"

type BookService interface {
	GetBook(ctx kp.IContext, id string) (*Book, error)
	CreateBook(ctx kp.IContext, book *Book) error
	GetAllBooks(ctx kp.IContext, filter map[string]any) ([]*Book, error)
}

type bookService struct {
	repo BookRepository
}

func NewBookService(repo BookRepository) BookService {
	return &bookService{repo: repo}
}

func (s *bookService) GetBook(ctx kp.IContext, id string) (*Book, error) {
	cmd := "get_book"
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		ctx.SummaryLog().AddError(node_postgres, cmd, "", err.Error())
		return nil, err
	}

	ctx.SummaryLog().AddSuccess(node_postgres, cmd, "20000", "success")
	return result, nil
}

func (s *bookService) CreateBook(ctx kp.IContext, book *Book) error {
	cmd := "create_book"
	err := s.repo.Save(ctx, book)
	if err != nil {
		ctx.SummaryLog().AddError(node_postgres, cmd, "", err.Error())
		return err
	}
	ctx.SummaryLog().AddSuccess(node_postgres, cmd, "20000", "success")
	return nil
}

func (s *bookService) GetAllBooks(ctx kp.IContext, filter map[string]any) ([]*Book, error) {
	cmd := "get_books"

	result, err := s.repo.GetALL(ctx, filter)
	if err != nil {
		ctx.SummaryLog().AddError(node_postgres, cmd, "", err.Error())
		return nil, err
	}
	ctx.SummaryLog().AddSuccess(node_postgres, cmd, "20000", "success")

	return result, nil
}
