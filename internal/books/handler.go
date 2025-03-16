package books

import (
	"github.com/gofiber/fiber/v2"
)

type BookHandler struct {
	svc BookService
}

func NewBookHandler(svc BookService) *BookHandler {
	return &BookHandler{svc: svc}
}

func (h *BookHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/books/:id", h.GetBook)
	app.Post("/books", h.CreateBook)
	app.Get("/books", h.GetAllBooks)
}

func (h *BookHandler) GetBook(c *fiber.Ctx) error {
	id := c.Params("id")
	book, err := h.svc.GetBook(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if book == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "book not found"})
	}
	return c.JSON(book)
}

func (h *BookHandler) CreateBook(c *fiber.Ctx) error {
	var req Book
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	err := h.svc.CreateBook(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "book created"})
}

func (h *BookHandler) GetAllBooks(c *fiber.Ctx) error {
	books, err := h.svc.GetAllBooks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(books)
}
