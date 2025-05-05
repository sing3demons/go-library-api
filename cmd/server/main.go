package main

import (
	"github.com/sing3demons/go-library-api/internal/books"
	"github.com/sing3demons/go-library-api/internal/users"
	"github.com/sing3demons/go-library-api/pkg/kp"
	"github.com/sing3demons/go-library-api/pkg/mongo"
	"github.com/sing3demons/go-library-api/pkg/postgres"
)

func main() {
	p, err := postgres.New()
	if err != nil {
		panic(err)
	}
	// p.Db.Exec("INSERT INTO books (title, author) VALUES ($1, $2)", "The Hobbit", "J.R.R. Tolkien")
	// p.Db.Exec("INSERT INTO books (title, author) VALUES ($1, $2)", "The Catcher in the Rye", "J.D. Salinger")
	defer p.Close()

	client := mongo.NewMongo("mongodb://localhost:27017")

	dbname := "my_database"
	dbCollection := "users"
	collection := client.Database(dbname).Collection(dbCollection)

	//
	logger := kp.NewAppLogger()
	server := kp.NewApplication(&kp.Config{
		AppConfig: kp.AppConfig{
			Port:  "8080",
			LogKP: true,
			AppName: "todo",
			Version: "1.0.0",
			TracerHost: "localhost:4318",
		},
		KafkaConfig: kp.KafkaConfig{
			Brokers: []string{"localhost:29092"},
			GroupID: "my-group",
		},
	}, logger)

	// Books module
	bookRepo := books.NewPostgresBookRepository(p)
	bookSvc := books.NewBookService(bookRepo)
	bookHandler := books.NewBookHandler(bookSvc)
	bookHandler.RegisterRoutes(server)

	// Users module
	userRepo := users.NewMongoUserRepository(collection)
	userSvc := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userSvc)
	userHandler.RegisterRoutes(server)

	// log.Fatal(app.Listen(":8080"))
	server.Start()
}
