package users

import "github.com/sing3demons/go-library-api/pkg/kp"

type UserService interface {
	RegisterUser(ctx kp.IContext, name, email string) (*User, error)
	GetUserById(ctx kp.IContext, id string) (*User, error)
	GetAllUsers(ctx kp.IContext) ([]*User, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) RegisterUser(ctx kp.IContext, name, email string) (*User, error) {
	user := &User{Name: name, Email: email}
	if err := s.repo.Save(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserById(ctx kp.IContext, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx kp.IContext) ([]*User, error) {
	return s.repo.GetALL(ctx, nil)
}
