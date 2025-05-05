package users_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sing3demons/go-library-api/internal/users"
	"github.com/sing3demons/go-library-api/pkg/kp"
	"github.com/stretchr/testify/assert"
)

type MockUserRepository struct {
	Err error
}

type UserRepository interface {
	Save(ctx kp.IContext, user *users.User) error
	GetByID(ctx kp.IContext, id string) (*users.User, error)
	GetALL(ctx kp.IContext, filter map[string]any) ([]*users.User, error)
}

func (m *MockUserRepository) Save(ctx kp.IContext, user *users.User) error {
	if m.Err != nil {
		return m.Err
	}
	user.ID = mockID
	user.Href = fmt.Sprintf("/users/%s", user.ID)
	return nil
}

func (m *MockUserRepository) GetByID(ctx kp.IContext, id string) (*users.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return &users.User{
		ID:    mockID,
		Name:  mockName,
		Email: mockEmail,
	}, nil
}

func (m *MockUserRepository) GetALL(ctx kp.IContext, filter map[string]any) ([]*users.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return []*users.User{
		{
			ID:    mockID,
			Name:  mockName,
			Email: mockEmail,
		},
	}, nil
}

const (
	mockID    = "f676b9ee-d08d-4758-b016-6c566e8fb573"
	mockName  = "test"
	mockEmail = "test@dev.com"
)

func TestUserServiceRegisterUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockRepo := MockUserRepository{}

		service := users.NewUserService(&mockRepo)

		user, err := service.RegisterUser(ctx, mockName, mockEmail)

		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, user.Name, mockName)
		assert.Equal(t, user.Email, mockEmail)
	})

	t.Run("error", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockRepo := MockUserRepository{
			Err: errors.New("error"),
		}

		service := users.NewUserService(&mockRepo)

		user, err := service.RegisterUser(ctx, mockName, mockEmail)

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserServiceGetUserById(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockRepo := MockUserRepository{}

		service := users.NewUserService(&mockRepo)

		user, err := service.GetUserById(ctx, mockID)

		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, user.Name, mockName)
		assert.Equal(t, user.Email, mockEmail)
	})

	t.Run("error", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockRepo := MockUserRepository{
			Err: errors.New("error"),
		}

		service := users.NewUserService(&mockRepo)

		user, err := service.GetUserById(ctx, mockID)

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserServiceGetAllUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockRepo := MockUserRepository{}

		service := users.NewUserService(&mockRepo)

		users, err := service.GetAllUsers(ctx)

		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, len(users), 1)
		assert.Equal(t, users[0].ID, mockID)
		assert.Equal(t, users[0].Name, mockName)
		assert.Equal(t, users[0].Email, mockEmail)
	})

	t.Run("error", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockRepo := MockUserRepository{
			Err: errors.New("error"),
		}

		service := users.NewUserService(&mockRepo)

		users, err := service.GetAllUsers(ctx)

		assert.Error(t, err)
		assert.Nil(t, users)
	})
}
