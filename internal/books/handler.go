package books

import (
	"net/http"

	"github.com/sing3demons/go-library-api/pkg/kp"
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

	c.CommonLog(cmd, "book")

	// logger := c.Log()
	// detailLog, summaryLog := logger.NewLog(c.Context(), "", "book")
	id := c.Param("id")

	c.SummaryLog().AddSuccess(node, cmd, "", "success")

	book, err := h.svc.GetBook(c, id)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if book == nil {
		return c.Response(http.StatusNotFound, map[string]any{"error": "book not found"})
	}
	c.SummaryLog().End("200", "")
	return c.Response(http.StatusOK, book)
}

func (h *BookHandler) CreateBook(c kp.IContext) error {
	node := "client"
	cmd := "create_book"
	// detailLog, summaryLog := c.Log().NewLog(c.Context(), "", "book")
	// detailLog.AddInputRequest(node, cmd, "", "", nil)
	c.CommonLog(cmd, "book")

	var req Book
	if err := c.ReadInput(&req); err != nil {
		c.SummaryLog().AddError(node, cmd, "", err.Error())
		return c.Response(http.StatusBadRequest, map[string]any{"error": "invalid request"})
	}
	c.SummaryLog().AddSuccess(node, cmd, "", "success")
	err := h.svc.CreateBook(c, &req)
	if err != nil {
		return c.Response(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	result, err := kp.RequestHttp(kp.RequestAttributes{
		Method: http.MethodGet,
		URL:    "http://localhost:8080/books/{id}",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Params: map[string]string{
			"id": req.ID,
		},
		Invoke:     "x-request-id",
		Service:    "http_book",
		Command:    "get_book",
		RetryCount: 3,
		Timeout:    5,
	}, c.DetailLog(), c.SummaryLog())

	if err != nil {
		c.SummaryLog().AddField("ErrorCause", err.Error())
		return c.Response(http.StatusCreated, map[string]any{
			"message": "book created",
			"id":      req.ID,
		})
	}

	c.SummaryLog().End("200", "")
	return c.Response(http.StatusCreated, map[string]any{
		"message": "book created",
		"data":    result,
	})
}

func (h *BookHandler) GetAllBooks(c kp.IContext) error {
	node := "client"
	cmd := "get_books"

	// detailLog, summaryLog := c.Log().NewLog(c.Context(), "", "book")
	c.CommonLog(cmd, "book")

	filter := map[string]any{}

	if c.Query("id") != "" {
		filter["id"] = c.Query("id")
	}

	if c.Query("title") != "" {
		filter["title"] = c.Query("title")
	}

	c.SummaryLog().AddSuccess(node, cmd, "", "success")

	books, err := h.svc.GetAllBooks(c, filter)
	if err != nil {
		msg := map[string]string{
			"error": err.Error(),
		}
		c.SummaryLog().AddError(node, cmd, "", "")
		return c.Response(http.StatusInternalServerError, msg)
	}

	c.SummaryLog().End("200", "")
	return c.Response(http.StatusOK, books)
}
