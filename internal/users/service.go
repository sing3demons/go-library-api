package users

import (
	"context"
)

type UserService interface {
	RegisterUser(ctx context.Context, name, email string) (*User, error)
	GetUserById(ctx context.Context, id string) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) RegisterUser(ctx context.Context, name, email string) (*User, error) {
	user := &User{Name: name, Email: email}
	if err := s.repo.Save(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserById(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*User, error) {
	return s.repo.GetALL(ctx, nil)
}

// func generateID() string {
// 	b := make([]byte, 4)
// 	rand.Read(b)
// 	return fmt.Sprintf("%x", b)
// }
