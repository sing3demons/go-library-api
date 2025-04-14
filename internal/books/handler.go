package books

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sing3demons/go-library-api/app"
)

type BookHandler struct {
	svc BookService
}

func NewBookHandler(svc BookService) *BookHandler {
	return &BookHandler{svc: svc}
}

func (h *BookHandler) RegisterRoutes(r app.IApplication) {
	r.Get("/books/:id", h.GetBook)
	r.Post("/books", h.CreateBook)
	r.Get("/books", h.GetAllBooks)
}

func (h *BookHandler) GetBook(c app.IContext) error {
	id := c.Param("id")
	book, err := h.svc.GetBook(c.Context(), id)
	if err != nil {
		return c.Response(fiber.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if book == nil {
		return c.Response(http.StatusNotFound, map[string]any{"error": "book not found"})
	}
	return c.Response(http.StatusOK, book)
}

func (h *BookHandler) CreateBook(c app.IContext) error {
	var req Book
	if err := c.ReadInput(&req); err != nil {
		return c.Response(fiber.StatusBadRequest, map[string]any{"error": "invalid request"})
	}
	err := h.svc.CreateBook(c.Context(), &req)
	if err != nil {
		return c.Response(fiber.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	return c.Response(fiber.StatusCreated, map[string]any{"message": "book created", "id": req.ID})
}

func (h *BookHandler) GetAllBooks(c app.IContext) error {
	filter := map[string]any{}

	if c.Query("id") != "" {
		filter["id"] = c.Query("id")
	}

	if c.Query("title") != "" {
		filter["title"] = c.Query("title")
	}

	books, err := h.svc.GetAllBooks(c.Context(), filter)
	if err != nil {
		return c.Response(fiber.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	return c.Response(http.StatusOK, books)
}
