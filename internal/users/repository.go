package users

import (
	"context"
	"fmt"

	"github.com/sing3demons/go-library-api/pkg/entities"
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

	// _, err := r.col.InsertOne(ctx, user)
	// if err != nil {
	// 	return err
	// }

	body := &entities.User{
		Name:  user.Name,
		Email: user.Email,
	}
	result, err := r.col.CreateUser(body)
	if err != nil {
		return err
	}

	user.ID = result.Data.ID

	user.Href = r.href(user.ID)
	return nil
}

func (r *mongoUserRepository) href(id string) string {
	return fmt.Sprintf("/users/%s", id)
}

func (r *mongoUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	result, err := r.col.GetUserByID(id)
	// decode := r.col.FindOne(ctx, bson.M{"_id": id})

	// err := decode.Decode(&user)
	if err != nil {
		return nil, err
	}
	user.ID = result.Data.ID
	user.Name = result.Data.Name
	user.Email = result.Data.Email
	user.Href = r.href(user.ID)
	return &user, nil
}

func (r *mongoUserRepository) GetALL(ctx context.Context, filter map[string]interface{}) ([]*User, error) {
	var users []*User
	filters := bson.M{}
	for k, v := range filter {
		filters[k] = v
	}
	// cursor, err := r.col.Find(ctx, filters)
	// if err != nil {
	// 	return nil, err
	// }
	// defer cursor.Close(ctx)

	// for cursor.Next(ctx) {
	// 	var user User
	// 	err := cursor.Decode(&user)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	user.Href = r.href(user.ID)
	// 	users = append(users, &user)
	// }

	result, err := r.col.GetAllUsers(filter)
	if err != nil {
		return nil, err
	}
	for _, u := range result.Data {
		user := User{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Href:  r.href(u.ID),
		}
		users = append(users, &user)
	}

	return users, nil
}
