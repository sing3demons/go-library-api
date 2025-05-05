package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sing3demons/go-library-api/pkg/entities"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore interface {
	GetUserByID(id string) (entities.ProcessData[entities.User], error)
	GetAllUsers(filter map[string]any) (result entities.ProcessData[[]entities.User], err error)
	CreateUser(ctx context.Context, user *entities.User) (entities.ProcessData[entities.User], error)
}

type mongoStore struct {
	mongo *mongo.Collection
}

// func NewMongoStore(mongo *mongo.Collection) MongoStore {
// 	return &mongoStore{mongo: mongo}
// }

func (m *mongoCollection) CreateUser(ctx context.Context, user *entities.User) (entities.ProcessData[entities.User], error) {

	user.ID = uuid.NewString()
	result := entities.ProcessData[entities.User]{}

	result.Body.Method = "InsertOne"
	result.Body.Document = user
	result.Body.Options = nil
	result.Body.Collection = "users"
	ctx, span := m.addTrace(ctx, result.Body.Method, result.Body.Collection)
	defer m.sendOperationStats(time.Now(), result.Body.Method, span)

	jsonDocumentBytes, _ := json.Marshal(user)
	jsonDocument := strings.ReplaceAll(string(jsonDocumentBytes), `"`, "'")
	result.RawData = fmt.Sprintf("%s.%s(%s)", result.Body.Collection, result.Body.Method, jsonDocument)
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// defer cancel()

	_, err := m.coll.InsertOne(ctx, user)
	if err != nil {
		return result, err
	}

	result.Data = *user
	return result, nil
}

func (m *mongoCollection) GetUserByID(ctx context.Context, id string) (entities.ProcessData[entities.User], error) {
	result := entities.ProcessData[entities.User]{}

	result.Body.Method = "FindOne"
	result.Body.Document = nil
	result.Body.Options = nil

	result.Body.Collection = "users"

	ctx, span := m.addTrace(ctx, result.Body.Method, result.Body.Collection)
	defer m.sendOperationStats(time.Now(), result.Body.Method, span)

	result.RawData = fmt.Sprintf("users.findOne({_id: %s})", id)

	var user entities.User
	err := m.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return result, err
	}

	result.Data = user
	return result, nil
}

func (m *mongoCollection) GetAllUsers(ctx context.Context, filter map[string]any) (result entities.ProcessData[[]entities.User], err error) {
	result.Body.Method = "Find"
	result.Body.Document = nil
	result.Body.Options = nil

	result.Body.Collection = "users"
	ctx, span := m.addTrace(ctx, result.Body.Method, result.Body.Collection)
	defer m.sendOperationStats(time.Now(), result.Body.Method, span)
	opt := &options.FindOptions{}

	result.RawData = buildMongoRawData("users", bson.D{}, opt)

	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return result, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user entities.User
		err := cursor.Decode(&user)
		if err != nil {
			return result, err
		}

		result.Data = append(result.Data, user)
	}

	return result, nil
}

func buildMongoRawData(name string, filter bson.D, opts *options.FindOptions) string {
	rawData := fmt.Sprintf("%s.find(%s, {projection,sort})", name, ConvertDToJSON(filter))
	if opts.Sort != nil {
		rawData = strings.Replace(rawData, "sort", fmt.Sprintf("sort:%s", ConvertDToJSON(opts.Sort.(bson.D))), 1)
	} else {
		rawData = strings.Replace(rawData, "sort", "", 1)
	}
	if opts.Projection != nil {
		rawData = strings.Replace(rawData, "projection", fmt.Sprintf("projection:%s", ConvertDToJSON(opts.Projection.(bson.D))), 1)
	} else {
		rawData = strings.Replace(rawData, "projection,", "", 1)
	}

	if opts.Projection == nil && opts.Sort == nil {
		rawData = strings.Replace(rawData, "{", "", 1)
		rawData = strings.Replace(rawData, "}", "", 1)
		rawData = strings.Replace(rawData, ", ", "", 1)
	}

	return rawData
}
func ConvertDToJSON(d bson.D) string {
	// Marshal primitive.D to JSON-compatible byte slice
	data := convertDToMap(d)
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error converting map to JSON:", err)
		return ""
	}
	return string(jsonData)
}

func convertDToMap(d bson.D) map[string]interface{} {
	result := make(map[string]interface{})
	for _, elem := range d {
		result[elem.Key] = elem.Value
	}
	return result
}
