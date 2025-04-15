package users

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sing3demons/go-library-api/kp"
)

type UserHandler struct {
	svc UserService
}

func NewUserHandler(svc UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(r kp.IApplication) {
	r.Post("/users/register", h.RegisterUser)
	r.Get("/users/:id", h.GetUser)
	r.Get("/users", h.GetAllUsers)
}

func (h *UserHandler) RegisterUser(c kp.IContext) error {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.ReadInput(&req); err != nil {
		return c.Response(fiber.StatusBadRequest, map[string]any{"error": "invalid request"})
	}
	user, err := h.svc.RegisterUser(c.Context(), req.Name, req.Email)
	if err != nil {
		return c.Response(http.StatusConflict, map[string]any{"error": err.Error()})
	}
	return c.Response(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c kp.IContext) error {
	id := c.Param("id")
	book, err := h.svc.GetUserById(c.Context(), id)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if book == nil {
		return c.Response(http.StatusNotFound, map[string]any{"error": "book not found"})
	}
	return c.Response(http.StatusOK, book)
}

func (h *UserHandler) GetAllUsers(c kp.IContext) error {
	users, err := h.svc.GetAllUsers(c.Context())
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	return c.Response(http.StatusOK, users)
}
