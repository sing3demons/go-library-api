package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sing3demons/go-library-api/internal/books"
	"github.com/sing3demons/go-library-api/internal/users"
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

	app := fiber.New()

	// Books module
	bookRepo := books.NewPostgresBookRepository(p)
	bookSvc := books.NewBookService(bookRepo)
	bookHandler := books.NewBookHandler(bookSvc)
	bookHandler.RegisterRoutes(app)

	// Users module
	userRepo := users.NewMongoUserRepository(collection)
	userSvc := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userSvc)
	userHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
