package books

import "context"

type BookService interface {
    GetBook(ctx context.Context, id string) (*Book, error)
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