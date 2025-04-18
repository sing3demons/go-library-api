package kp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	handlerCalledErr = "handler was not called"
)

type MockRouter struct {
	mock.Mock
}

func (m *MockRouter) Get(path string, handler func(http.ResponseWriter, *http.Request), middlewares ...func(http.Handler) http.Handler) {
	m.Called(path, handler, middlewares)
}

func (m *MockRouter) Post(path string, handler func(http.ResponseWriter, *http.Request), middlewares ...func(http.Handler) http.Handler) {
	m.Called(path, handler, middlewares)
}

func (m *MockRouter) Use(middlewares ...func(http.Handler) http.Handler) {
	m.Called(middlewares)
}

func (m *MockRouter) Register() *http.Server {
	args := m.Called()
	return args.Get(0).(*http.Server)
}
func TestNewApplication(t *testing.T) {
	mockRouter := new(MockRouter)

	logger := NewMockLogger()

	mockRouter.On("Register").Return(&http.Server{})

	config := &Config{
		AppConfig: AppConfig{
			Port: "3000",
		},
	}

	app := NewApplication(config, logger)

	assert.NotNil(t, app, "Application should not be nil")
}

func TestApplicationGet(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port: "3000",
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false
	var name string

	app.Get("/test", func(ctx IContext) error {
		handlerCalled = true
		name = ctx.Query("name")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test?name=xx", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, "xx", name)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApplicationPost(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port: "3002",
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false
	var data map[string]any

	app.Post("/test", func(ctx IContext) error {
		handlerCalled = true
		if err := ctx.ReadInput(&data); err != nil {
			assert.Fail(t, err.Error())
		}

		// ctx.Response(http.StatusCreated, data)
		return nil
	})

	body := []byte(`{"name":"xx"}`)
	req := httptest.NewRequest(http.MethodPost, "/test?name=xx", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, "xx", data["name"])
	// assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestApplicationPut(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Gin,
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false
	var id string
	var data map[string]interface{}

	app.Put("/test/:id", func(ctx IContext) error {
		handlerCalled = true
		id = ctx.Param("id")

		if err := ctx.ReadInput(&data); err != nil {
			assert.Fail(t, err.Error())
		}

		return nil
	})

	body := []byte(`{"name":"xx"}`)
	req := httptest.NewRequest(http.MethodPut, "/test/1", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, "1", id)
	assert.Equal(t, "xx", data["name"])
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApplicationDelete(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Echo,
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false
	var id string

	app.Delete("/test/:id", func(ctx IContext) error {
		handlerCalled = true
		id = ctx.Param("id")
		return nil
	})

	req := httptest.NewRequest(http.MethodDelete, "/test/1", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, "1", id)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApplicationPatch(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Gin,
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false
	var id string
	var data map[string]interface{}

	app.Patch("/x/:id", func(ctx IContext) error {
		handlerCalled = true
		id = ctx.Param("id")

		if err := ctx.ReadInput(&data); err != nil {
			assert.Fail(t, err.Error())
		}

		return nil
	})

	body := []byte(`{"name":"xx"}`)
	req := httptest.NewRequest(http.MethodPatch, "/x/1", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, "1", id)
	assert.Equal(t, "xx", data["name"])
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApplicationUse(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Gin,
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false

	app.Use(func(next HandleFunc) HandleFunc {
		return func(ctx IContext) error {
			handlerCalled = true
			return next(ctx)
		}
	})

	app.Get("/test", func(ctx IContext) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApplicationServeHTTP(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Gin,
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false

	app.Get("/test", func(ctx IContext) error {
		handlerCalled = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, handlerCalledErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApplicationServeHTTPWithNoRoute(t *testing.T) {
	logger := NewMockLogger()

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Gin,
		},
	}

	app := NewApplication(config, logger)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestApplicationConsume(t *testing.T) {
	logger := NewMockLogger()
	mockConsumer := &MockConsumerGroup{}

	config := &Config{
		AppConfig: AppConfig{
			Port:   "3000",
			Router: Gin,
		},
		KafkaConfig: KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			consumer: mockConsumer,
			producer: mocks.NewSyncProducer(t, nil),
		},
	}

	app := NewApplication(config, logger)

	handlerCalled := false

	app.Consume("test", func(ctx IContext) error {
		handlerCalled = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.False(t, handlerCalled, handlerCalledErr)
}

func TestServerStart(t *testing.T) {

	mockLogger := NewMockLogger()
	mockConsumer := &MockConsumerGroup{}

	server := NewApplication(&Config{
		AppConfig: AppConfig{
			Port: "8080",
		},
		KafkaConfig: KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			producer: mocks.NewSyncProducer(t, nil),
			consumer: mockConsumer,
		},
	}, mockLogger)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		time.Sleep(1 * time.Second)
		signalChan <- os.Interrupt
	}()

	go func() {
		server.Start()
		cancel() // Ensure cleanup after shutdown
	}()

	// Wait for shutdown
	<-ctx.Done()

	t.Log("Server shut down successfully")
}
