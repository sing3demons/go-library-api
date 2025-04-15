package books

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/sing3demons/go-library-api/kp"
)

type BookHandler struct {
	svc BookService
}

func NewBookHandler(svc BookService) *BookHandler {
	return &BookHandler{svc: svc}
}

func (h *BookHandler) RegisterRoutes(r kp.IApplication) {
	r.Get("/books/:id", h.GetBook)
	r.Post("/books", h.CreateBook)
	r.Get("/books", h.GetAllBooks)
}

func (h *BookHandler) GetBook(c kp.IContext) error {
	node := "client"
	cmd := "get_book"
	detailLog, summaryLog := c.Log().NewLog("session:"+uuid.New().String(), "", "book")
	id := c.Param("id")

	detailLog.AddInputRequest(node, cmd, "", "", map[string]any{"id": id})
	summaryLog.AddSuccess(node, cmd, "", "success")

	book, err := h.svc.GetBook(c.Context(), id, detailLog, summaryLog)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if book == nil {
		return c.Response(http.StatusNotFound, map[string]any{"error": "book not found"})
	}
	detailLog.AddOutputRequest(node, cmd, "", "", book)
	detailLog.End()
	summaryLog.End("200", "")
	return c.Response(http.StatusOK, book)
}

func (h *BookHandler) CreateBook(c kp.IContext) error {
	node := "client"
	cmd := "create_book"
	detailLog, summaryLog := c.Log().NewLog("session:"+uuid.New().String(), "", "book")
	detailLog.AddInputRequest(node, cmd, "", "", nil)
	summaryLog.AddSuccess(node, cmd, "", "success")
	defer detailLog.End()

	var req Book
	if err := c.ReadInput(&req); err != nil {
		return c.Response(http.StatusBadRequest, map[string]any{"error": "invalid request"})
	}
	err := h.svc.CreateBook(c.Context(), &req, detailLog, summaryLog)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	detailLog.AddOutputRequest(node, cmd, "", "", req)
	summaryLog.End("200", "")
	return c.Response(http.StatusCreated, map[string]any{"message": "book created", "id": req.ID})
}

func (h *BookHandler) GetAllBooks(c kp.IContext) error {
	node := "client"
	cmd := "get_books"

	detailLog, summaryLog := c.Log().NewLog("session:"+uuid.New().String(), "", "book")
	filter := map[string]any{}

	if c.Query("id") != "" {
		filter["id"] = c.Query("id")
	}

	if c.Query("title") != "" {
		filter["title"] = c.Query("title")
	}

	detailLog.AddInputRequest(node, cmd, "", "", filter)
	summaryLog.AddSuccess(node, cmd, "", "success")

	books, err := h.svc.GetAllBooks(c.Context(), filter, detailLog, summaryLog)
	if err != nil {
		msg := map[string]string{
			"error": err.Error(),
		}
		detailLog.AddOutputRequest(node, cmd, "", "", msg)
		summaryLog.AddError(node, cmd, "", "")

		return c.Response(http.StatusInternalServerError, msg)
	}
	detailLog.AddOutputRequest(node, cmd, "", "", books)

	detailLog.End()
	summaryLog.End("200", "")
	return c.Response(http.StatusOK, books)
}
