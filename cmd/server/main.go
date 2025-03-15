package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sing3demons/go-library-api/internal/users"
	"github.com/sing3demons/go-library-api/pkg/mongo"
)

func main() {

	// client, err := mongo.Connect(
	// 	options.Client().ApplyURI("mongodb://localhost:27017/daov2"),
	// )

	// if err != nil {
	// 	panic(err)
	// }

	// defer client.Disconnect(context.Background())

	client := mongo.NewMongo("mongodb://localhost:27017")

	dbname := "my_database"
	dbCollection := "users"
	collection := client.Database(dbname).Collection(dbCollection)

	app := fiber.New()

	// // Books module
	// bookRepo := books.NewInMemoryBookRepository()
	// bookSvc := books.NewBookService(bookRepo)
	// bookHandler := books.NewBookHandler(bookSvc)
	// bookHandler.RegisterRoutes(app)

	// Users module
	userRepo := users.NewMongoUserRepository(collection)
	userSvc := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userSvc)
	userHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
