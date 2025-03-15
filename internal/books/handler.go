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