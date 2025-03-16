package users

import (
	"context"
	"testing"

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

const (
	mockID    = "f676b9ee-d08d-4758-b016-6c566e8fb573"
	mockName  = "test"
	mockEmail = "test@dev.com"
)

func TestSave(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCol := MockCollection{
			InsertedID: mockID,
			Err:        nil,
		}
		repo := NewMongoUserRepository(&mockCol)

		user := &User{
			Name:  mockName,
			Email: mockEmail,
		}

		err := repo.Save(context.TODO(), user)

		assert.NoError(t, err)
		assert.Equal(t, mockCol.InsertedID, user.ID)
	})

	t.Run("error", func(t *testing.T) {
		mockCol := MockCollection{
			InsertedID: "",
			Err:        mongo.ErrClientDisconnected,
		}
		repo := NewMongoUserRepository(&mockCol)

		user := &User{
			Name:  mockName,
			Email: mockEmail,
		}

		err := repo.Save(context.TODO(), user)

		assert.Error(t, err)
		assert.Equal(t, "", user.ID)
	})
}

func TestGetByID(t *testing.T) {
	user := User{
		ID:    mockID,
		Name:  mockName,
		Email: mockEmail,
	}
	t.Run("success", func(t *testing.T) {
		mockCol := MockCollection{
			Err:  nil,
			User: &user,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetByID(context.TODO(), mockCol.InsertedID)

		assert.NoError(t, err)
		assert.Equal(t, result.ID, mockID)
		assert.Equal(t, result.Name, mockName)
		assert.Equal(t, result.Email, mockEmail)
	})

	t.Run("error", func(t *testing.T) {
		mockCol := MockCollection{
			Err:  mongo.ErrClientDisconnected,
			User: nil,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetByID(context.TODO(), mockCol.InsertedID)

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
		mockCol := MockCollection{
			Err:   nil,
			Users: users,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetALL(context.TODO(), bson.M{
			"name": mockName,
		})

		assert.NoError(t, err)

		assert.Equal(t, len(result), 1)
	})

	t.Run("error", func(t *testing.T) {
		mockCol := MockCollection{
			Err:   mongo.ErrNilCursor,
			Users: nil,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetALL(context.TODO(), bson.M{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error decode", func(t *testing.T) {
		mockCol := MockCollection{
			Err:   mongo.ErrNilCursor,
			Users: users,
			next:  true,
		}
		repo := NewMongoUserRepository(&mockCol)

		result, err := repo.GetALL(context.TODO(), bson.M{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

}
