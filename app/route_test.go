package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGinApplicationGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	app.Get("/test", func(ctx IContext) error {
		handlerCalled = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, printErr(http.StatusOK, rec.Code))
	assert.True(t, handlerCalled, handlerCalledErr)
}

func printErr(s, code int) string {
	return fmt.Sprintf("expected status code %d but got %d", s, code)
}

func TestGinApplicationPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	app.Post("/test", func(ctx IContext) error {
		handlerCalled = true
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, printErr(http.StatusOK, rec.Code))
	assert.True(t, handlerCalled, handlerCalledErr)
}

func TestGinApplicationPut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	app.Put("/test", func(ctx IContext) error {
		handlerCalled = true
		return nil
	})

	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, printErr(http.StatusOK, rec.Code))
	assert.True(t, handlerCalled, handlerCalledErr)
}

func TestGinApplicationDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	app.Delete("/test", func(ctx IContext) error {
		handlerCalled = true
		return nil
	})

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, printErr(http.StatusOK, rec.Code))
	assert.True(t, handlerCalled, handlerCalledErr)
}

func TestGinApplicationPatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	var id string
	var data map[string]interface{}
	app.Patch("/test/:id", func(ctx IContext) error {
		handlerCalled = true
		id = ctx.Param("id")

		if err := ctx.ReadInput(&data); err != nil {
			assert.Fail(t, err.Error())
		}

		return nil
	})

	body := `{"name":"test"}`

	req := httptest.NewRequest(http.MethodPatch, "/test/123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, printErr(http.StatusOK, rec.Code))
	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, "123", id, "expected id to be 123")
	assert.Equal(t, "test", data["name"])
}

func TestGinApplicationUse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	handlerCalled2 := false

	app.Use(func(next HandleFunc) HandleFunc {
		return func(ctx IContext) error {
			handlerCalled = true
			return next(ctx)
		}
	})
	app.Get("/test", func(ctx IContext) error {
		handlerCalled2 = true
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.True(t, handlerCalled2, handlerCalledErr)
}

func TestGinApplicationRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	srv := app.Register()

	assert.NotNil(t, srv)
}

// test/:id
func TestGinApplicationParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := NewMockLogger()

	cfg := &Config{
		AppConfig: AppConfig{
			Port: "8888",
		},
	}
	app := newServer(cfg, log).(*httpApplication)

	handlerCalled := false
	app.Get("/test/:id", func(ctx IContext) error {
		handlerCalled = true
		id := ctx.Param("id")
		assert.Equal(t, "123", id)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, printErr(http.StatusOK, rec.Code))
	assert.True(t, handlerCalled, handlerCalledErr)
}
