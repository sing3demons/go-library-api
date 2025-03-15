package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	m "github.com/sing3demons/go-library-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mongoUserRepository struct {
	col m.Collection
}

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetALL(ctx context.Context, filter map[string]interface{}) ([]*User, error)
}

func NewMongoUserRepository(col m.Collection) UserRepository {
	return &mongoUserRepository{col: col}
}

func (r *mongoUserRepository) Save(ctx context.Context, user *User) error {
	id := uuid.NewString()
	user.ID = id

	_, err := r.col.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.Href = r.href(user.ID)
	return nil
}

func (r *mongoUserRepository) href(id string) string {
	return fmt.Sprintf("/users/%s", id)
}

func (r *mongoUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var user *User
	decode := r.col.FindOne(ctx, bson.M{"_id": id})

	err := decode.Decode(&user)
	if err != nil {
		return nil, err
	}
	user.Href = r.href(user.ID)
	return user, nil
}

func (r *mongoUserRepository) GetALL(ctx context.Context, filter map[string]interface{}) ([]*User, error) {
	var users []*User
	filters := bson.M{}
	for k, v := range filter {
		filters[k] = v
	}
	cursor, err := r.col.Find(ctx, filters)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		user.Href = r.href(user.ID)
		users = append(users, &user)
	}

	return users, nil
}

// InMemoryUserRepository is a simple in-memory database
// type InMemoryUserRepository struct {
// 	users []*User
// }

// func NewInMemoryUserRepository() *InMemoryUserRepository {
// 	return &InMemoryUserRepository{
// 		users: make([]*User, 0),
// 	}
// }

// func (r *InMemoryUserRepository) Save(ctx context.Context, user *User) error {
// 	// check if user already exists
// 	for _, u := range r.users {
// 		if u.Email == user.Email {
// 			return errors.New("user with this email already exists")
// 		}
// 	}

// 	r.users = append(r.users, user)
// 	return nil
// }

// func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
// 	var user *User
// 	for _, u := range r.users {
// 		if u.ID == id {
// 			user = u
// 			break
// 		}
// 	}

// 	if user == nil {
// 		return nil, nil
// 	}
// 	return user, nil
// }

// func (r *InMemoryUserRepository) GetALL(ctx context.Context, filter map[string]interface{}) ([]*User, error) {

// 	if filterName, ok := filter["name"]; filterName != nil && ok {
// 		for _, u := range r.users {
// 			if u.Name == filterName {
// 				return []*User{u}, nil
// 			}
// 		}
// 	} else {
// 		return r.users, nil
// 	}
// 	return r.users, nil
// }
