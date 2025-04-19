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
	cmd := "register_user"
	c.CommonLog(cmd, "register_user")
	if err := c.ReadInput(&req); err != nil {
		return c.Response(fiber.StatusBadRequest, map[string]any{"error": "invalid request"})
	}
	user, err := h.svc.RegisterUser(c, req.Name, req.Email)
	if err != nil {
		return c.Response(http.StatusConflict, map[string]any{"error": err.Error()})
	}
	return c.Response(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c kp.IContext) error {
	cmd := "get_user"

	c.CommonLog(cmd, "get_user_by_id")
	id := c.Param("id")
	book, err := h.svc.GetUserById(c, id)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if book == nil {
		return c.Response(http.StatusNotFound, map[string]any{"error": "book not found"})
	}
	return c.Response(http.StatusOK, book)
}

func (h *UserHandler) GetAllUsers(c kp.IContext) error {
	cmd := "get_all_users"
	c.CommonLog(cmd, "get_all_users")
	c.SummaryLog().AddSuccess("client", cmd, "", "success")
	users, err := h.svc.GetAllUsers(c)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	return c.Response(http.StatusOK, users)
}
