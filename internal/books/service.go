package books

import "context"

type BookService interface {
	GetBook(ctx context.Context, id string) (*Book, error)
	CreateBook(ctx context.Context, book *Book) error
    GetAllBooks(ctx context.Context) ([]*Book, error)
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


func (s *bookService) GetAllBooks(ctx context.Context) ([]*Book, error) {
    return s.repo.GetALL(ctx, nil)
}