package mongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type SingleResult interface {
	Decode(any) error
}

type Cursor interface {
	Close(context.Context) error
	Next(context.Context) bool
	Decode(any) error
	All(context.Context, any) error
}

type InsertOneResult struct {
	InsertedID   any
	Acknowledged bool
}

type Collection interface {
	FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResult
	InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*InsertOneResult, error)
	// InsertMany(context.Context, []interface{}) ([]interface{}, error)
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*DeleteResult, error)
	Find(context.Context, any, ...options.Lister[options.FindOptions]) (Cursor, error)
	CountDocuments(context.Context, any, ...options.Lister[options.CountOptions]) (int64, error)
	UpdateOne(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*UpdateResult, error)
}

type Database interface {
	Collection(string) Collection
	// Client() Client
}

type DeleteResult struct {
	DeletedCount int64
	Acknowledged bool
}

type Client interface {
	Database(string) Database
	Disconnect(context.Context) error
	// StartSession() (mongo.Session, error)
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    any
}
type mongoClient struct {
	cl *mongo.Client
}
type mongoDatabase struct {
	db *mongo.Database
}
type mongoCollection struct {
	coll *mongo.Collection
}

type mongoSingleResult struct {
	sr *mongo.SingleResult
}

type mongoCursor struct {
	mc *mongo.Cursor
}

// type mongoSession struct {
// 	mongo.Session
// }

func NewMongo(uri string) Client {
	if uri == "" {
		log.Fatal("uri is empty")
	}
	cl, err := mongo.Connect(options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	return &mongoClient{
		cl: cl,
	}
}

func (m *mongoClient) Disconnect(ctx context.Context) error {
	return m.cl.Disconnect(ctx)
}

func (m *mongoClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return m.cl.Ping(ctx, rp)
}

func (m *mongoClient) Database(name string) Database {
	return &mongoDatabase{
		db: m.cl.Database(name),
	}
}

func (m *mongoDatabase) Collection(name string) Collection {
	return &mongoCollection{
		coll: m.db.Collection(name),
	}
}
func (m *mongoCollection) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return m.coll.CountDocuments(ctx, filter, opts...)
}
func (m *mongoCollection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*InsertOneResult, error) {
	r, err := m.coll.InsertOne(ctx, document, opts...)
	return &InsertOneResult{
		InsertedID:   r.InsertedID,
		Acknowledged: r.Acknowledged,
	}, err
}

func newDeleteResult(r *mongo.DeleteResult) *DeleteResult {
	return &DeleteResult{
		DeletedCount: r.DeletedCount,
		Acknowledged: r.Acknowledged,
	}
}

func (m *mongoCollection) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*DeleteResult, error) {
	r, err := m.coll.DeleteOne(ctx, filter, opts...)
	return newDeleteResult(r), err
}

func (m *mongoCollection) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (Cursor, error) {
	cursor, err := m.coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &mongoCursor{mc: cursor}, nil
}

func (m *mongoCursor) Close(ctx context.Context) error {
	return m.mc.Close(ctx)
}

func (m *mongoCursor) Next(ctx context.Context) bool {
	return m.mc.Next(ctx)
}

func (m *mongoCursor) Decode(v interface{}) error {
	return m.mc.Decode(v)
}

func (m *mongoCursor) All(ctx context.Context, v any) error {
	return m.mc.All(ctx, v)
}

func (m *mongoSingleResult) Decode(v any) error {
	return m.sr.Decode(v)
}

func (m *mongoCollection) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResult {
	return &mongoSingleResult{
		sr: m.coll.FindOne(ctx, filter, opts...),
	}
}

func (m *mongoCollection) UpdateOne(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*UpdateResult, error) {
	r, err := m.coll.UpdateOne(ctx, filter, update, opts...)
	return &UpdateResult{
		MatchedCount:  r.MatchedCount,
		ModifiedCount: r.ModifiedCount,
		UpsertedCount: r.UpsertedCount,
		UpsertedID:    r.UpsertedID,
	}, err
}
