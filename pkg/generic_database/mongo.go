package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDbCollection[T any] struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewMongoDbWithCollection[T any](collection *mongo.Collection, timeout time.Duration) *MongoDbCollection[T] {
	return &MongoDbCollection[T]{
		collection: collection,
		timeout:    timeout,
	}
}

// GenerateContext generates a new context with the configured timeout
func (m *MongoDbCollection[T]) GenerateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), m.timeout)
}

func (m *MongoDbCollection[T]) Add(record T) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	r, err := m.collection.InsertOne(ctx, record)
	if err != nil {
		fmt.Println("error inserting record: ", err)
	}

	fmt.Println("record inserted", r.InsertedID)
	return err
}

func (m *MongoDbCollection[T]) Update(id string, data T) error {
	ctx, cancel := m.GenerateContext()
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": data}

	_, err := m.collection.UpdateOne(ctx, filter, update)
	return err
}

func (m *MongoDbCollection[T]) Get(q QueryFunc[T]) ([]T, error) {
	return q()
}

func (m *MongoDbCollection[T]) Find(filter map[string]interface{}) ([]T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	cursor, err := m.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	fmt.Println("results: ", results)

	for _, result := range results {
		res, _ := bson.MarshalExtJSON(result, false, false)
		fmt.Println(string(res))
	}

	return results, nil
}

func (m *MongoDbCollection[T]) Delete(id string) error {
	ctx, cancel := m.GenerateContext()
	defer cancel()

	filter := bson.M{"_id": id}
	_, err := m.collection.DeleteOne(ctx, filter)
	return err
}

// func main() {
// 	type User struct {
// 		Username string `bson:"username"`
// 		Active   bool   `bson:"active"`
// 	}
// 	users := []User{
// 		{Username: "Adam", Active: false},
// 		{Username: "Eve", Active: true},
// 		{Username: "Bob", Active: true},
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
// 	defer cancel()

// 	client, err := mongo.Connect(
// 		ctx,
// 		options.Client().ApplyURI("mongodb://localhost:27017"),
// 	)

// 	if err != nil {
// 		panic(err)
// 	}

// 	defer client.Disconnect(context.Background())

// 	dbname := "my_database"
// 	dbcollection := "my_collection"
// 	collection := client.Database(dbname).Collection(dbcollection)
// 	// delete after test
// 	defer collection.Drop(context.Background())

// 	db := database.NewMongoDbWithCollection[User](
// 		collection,
// 		5*time.Second,
// 	)

// 	for _, u := range users {
// 		err := database.AddToDAO(db, u)
// 		if err != nil {
// 			log.Printf("error inserting %s : %s", u.Username, err)
// 		}
// 	}

// 	result, err := database.QueryDAOWith(db, func() ([]User, error) {

// 		ctx, cancel := db.GenerateContext()
// 		defer cancel()

// 		cursor, err := collection.Find(ctx, bson.M{"active": true})
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer cursor.Close(ctx)

// 		var results []User
// 		if err := cursor.All(ctx, &results); err != nil {
// 			return nil, err
// 		}

// 		return results, nil
// 	})

// 	if err != nil {
// 		log.Fatalf("error finding data: %s", err)
// 	}

// 	log.Printf("data returned %+v", result)

// }
