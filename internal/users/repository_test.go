package users

import (
	"context"
	"fmt"
	"testing"

	"github.com/sing3demons/go-library-api/kp"
	"github.com/sing3demons/go-library-api/pkg/entities"
	m "github.com/sing3demons/go-library-api/pkg/mongo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MockCollection struct {
	InsertedID string
	User       *User
	Err        error
	Users      []User
	registry   *bson.Registry
	next       bool
}

func (mm *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...options.Lister[options.InsertOneOptions]) (*m.InsertOneResult, error) {
	document.(*User).ID = mm.InsertedID

	mockRetunrn := m.InsertOneResult{
		InsertedID:   mm.InsertedID,
		Acknowledged: true,
	}

	return &mockRetunrn, mm.Err
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) m.SingleResult {
	return mongo.NewSingleResultFromDocument(m.User, m.Err, m.registry)
}

func (mm *MockCollection) UpdateOne(ctx context.Context, filter, update interface{}, opts ...options.Lister[options.UpdateOneOptions]) (*m.UpdateResult, error) {
	mockRetunrn := m.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
		UpsertedCount: 0,
		UpsertedID:    nil,
	}

	return &mockRetunrn, mm.Err
}

func (mm *MockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.DeleteOneOptions]) (*m.DeleteResult, error) {
	mockRetunrn := m.DeleteResult{
		DeletedCount: 1,
		Acknowledged: true,
	}

	return &mockRetunrn, mm.Err
}

func (mm *MockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return 1, mm.Err
}

type mockCursor struct {
	c   m.Cursor
	Err error
}

func (m *mockCursor) Close(ctx context.Context) error {
	return m.c.Close(ctx)
}

func (m *mockCursor) Next(ctx context.Context) bool {
	if m.Err != nil {
		return true
	}
	return m.c.Next(ctx)
}

func (m *mockCursor) Decode(v interface{}) error {
	if m.Err != nil {
		return m.Err
	}
	return m.c.Decode(v)
}

func (m *mockCursor) All(ctx context.Context, v interface{}) error {
	return m.c.All(ctx, v)
}

func (m *MockCollection) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (m.Cursor, error) {
	documents := []interface{}{}
	for _, user := range m.Users {
		documents = append(documents, user)
	}

	cur, err := mongo.NewCursorFromDocuments(documents, m.Err, m.registry)

	mockCursor := &mockCursor{
		c: cur,
	}

	if m.Err != nil && !m.next {
		return nil, m.Err
	}

	if m.next && m.Err != nil {
		mockCursor.Err = m.Err
	}

	return mockCursor, err
}

func (mm *MockCollection) InsertMany(ctx context.Context, documents interface{}, opts ...options.Lister[options.InsertManyOptions]) (*m.InsertManyResult, error) {
	return &m.InsertManyResult{
		InsertedIDs:  []interface{}{mm.InsertedID},
		Acknowledged: true,
	}, mm.Err
}

func (mm *MockCollection) CreateUser(user *entities.User) (entities.ProcessData[entities.User], error) {
	user.ID = mm.InsertedID
	result := entities.ProcessData[entities.User]{}

	result.Body.Method = "InsertOne"
	result.Body.Document = user
	result.Body.Options = nil

	result.Body.Collection = "users"

	result.RawData = fmt.Sprintf("users.insertOne(%s, %s, %s)", user.ID, user.Name, user.Email)
	result.Data.ID = mm.InsertedID

	result.Data.Name = user.Name
	result.Data.Email = user.Email

	return result, mm.Err
}

func (mm *MockCollection) GetUserByID(id string) (entities.ProcessData[entities.User], error) {
	result := entities.ProcessData[entities.User]{}

	result.Body.Method = "FindOne"
	result.Body.Document = nil
	result.Body.Options = nil

	result.Body.Collection = "users"

	result.RawData = fmt.Sprintf("users.findOne({_id: %s})", id)

	var user entities.User
	err := mm.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return result, err
	}

	result.Data = user
	return result, nil
}

func (mm *MockCollection) GetAllUsers(filter map[string]any) (result entities.ProcessData[[]entities.User], err error) {
	result.Body.Method = "Find"
	result.Body.Document = nil
	result.Body.Options = nil

	result.Body.Collection = "users"

	if mm.Err != nil {
		return result, mm.Err
	}

	if mm.next {
		return result, mm.Err
	}

	for _, user := range mm.Users {
		result.Data = append(result.Data, entities.User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	return result, mm.Err
}

const (
	mockID    = "f676b9ee-d08d-4758-b016-6c566e8fb573"
	mockName  = "test"
	mockEmail = "test@dev.com"
)

func TestSave(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := kp.NewMockContext()
		mockCol := MockCollection{
			InsertedID: mockID,
			Err:        nil,
		}
		repo := NewMongoUserRepository(&mockCol)

		user := &User{
			Name:  mockName,
			Email: mockEmail,
		}

		err := repo.Save(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, mockCol.InsertedID, user.ID)
		ctx.Verify(t)
	})

	t.Run("error", func(t *testing.T) {
		ctx := kp.NewMockContext()
		mockCol := MockCollection{
			InsertedID: "",
			Err:        mongo.ErrClientDisconnected,
		}
		repo := NewMongoUserRepository(&mockCol)

		user := &User{
			Name:  mockName,
			Email: mockEmail,
		}

		err := repo.Save(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, "", user.ID)
		ctx.Verify(t)
	})
}

func TestGetByID(t *testing.T) {
	user := User{
		ID:    mockID,
		Name:  mockName,
		Email: mockEmail,
	}
	t.Run("success", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockCol := MockCollection{
			Err:  nil,
			User: &user,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetByID(ctx, mockCol.InsertedID)

		assert.NoError(t, err)
		assert.Equal(t, result.ID, mockID)
		assert.Equal(t, result.Name, mockName)
		assert.Equal(t, result.Email, mockEmail)
	})

	t.Run("error", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockCol := MockCollection{
			Err:  mongo.ErrClientDisconnected,
			User: nil,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetByID(ctx, mockCol.InsertedID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

}

func TestGetAll(t *testing.T) {
	users := []User{{
		ID:    mockID,
		Name:  mockName,
		Email: mockEmail,
	}}

	t.Run("success", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockCol := MockCollection{
			Err:   nil,
			Users: users,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetALL(ctx, bson.M{
			"name": mockName,
		})

		assert.NoError(t, err)

		assert.Equal(t, len(result), 1)
	})

	t.Run("error", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockCol := MockCollection{
			Err:   mongo.ErrNilCursor,
			Users: nil,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetALL(ctx, bson.M{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error decode", func(t *testing.T) {
		ctx := kp.NewMockContext()
		defer ctx.Verify(t)
		mockCol := MockCollection{
			Err:   mongo.ErrNilCursor,
			Users: users,
			next:  true,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetALL(ctx, bson.M{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

}
