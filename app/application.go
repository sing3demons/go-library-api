package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

type IApplication interface {
	Get(path string, handler HandleFunc, middlewares ...Middleware)
	Post(path string, handler HandleFunc, middlewares ...Middleware)
	Put(path string, handler HandleFunc, middlewares ...Middleware)
	Delete(path string, handler HandleFunc, middlewares ...Middleware)
	Patch(path string, handler HandleFunc, middlewares ...Middleware)

	Use(middlewares ...Middleware)
	Start()
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	Consume(topic string, handler ServiceHandleFunc)
	SendMessage(topic string, payload any, opts ...OptionProducerMsg) (RecordMetadata, error)
}

type IRouter interface {
	Get(path string, handler HandleFunc, middlewares ...Middleware)
	Post(path string, handler HandleFunc, middlewares ...Middleware)
	Put(path string, handler HandleFunc, middlewares ...Middleware)
	Delete(path string, handler HandleFunc, middlewares ...Middleware)
	Patch(path string, handler HandleFunc, middlewares ...Middleware)
	Use(middlewares ...Middleware)
	Register() *http.Server
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type AppConfig struct {
	Port   string
	Router Router
}

type KafkaConfig struct {
	Brokers  []string
	GroupID  string
	Username string
	Password string

	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
}

type KafkaProducerOptions struct {
	ReturnSuccesses bool
	ReturnErrors    bool
}

type Config struct {
	AppConfig   AppConfig
	KafkaConfig KafkaConfig
}

// enum Router {gin, mux}
type Router int

const (
	None Router = iota
	Gin
	Mux
	Echo
	Fiber
)

type Server struct {
	httpServer *http.Server
	kafka      *KafkaServer
	router     IRouter
	Log        ILogger
}

func NewApplication(config *Config, logger ILogger) IApplication {
	kafka := &KafkaServer{}

	if len(config.KafkaConfig.Brokers) != 0 {
		producer, err := newProducer(&config.KafkaConfig)
		if err != nil {
			logger.Fatalf("Failed to create Kafka consumer: %v", err)

		}

		client, err := newConsumer(&config.KafkaConfig)
		if err != nil {
			logger.Fatalf("Failed to create Kafka consumer: %v", err)
		}
		k, err := NewKafkaServer(producer, client, &config.KafkaConfig, logger)
		if err != nil {
			logger.Fatalf("Failed to create Kafka server: %v", err)
		}

		kafka = k
	}

	var router IRouter

	if config.AppConfig.Port != "" {
		if kafka != nil {
			config.KafkaConfig.producer = kafka.producer
		}
		switch config.AppConfig.Router {
		// case Gin:
		// 	router = newGinServer(config, logger)
		// case Echo:
		// 	router = newEchoServer(config, logger)
		default:
			router = newServer(config, logger)
		}
	}

	return &Server{
		kafka:  kafka,
		router: router,
		Log:    logger,
	}
}

func (s *Server) Start() {
	if s.router != nil {
		s.httpServer = s.router.Register()
	}
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	if s.httpServer != nil {
		// Start HTTP Server
		go func() {
			s.Log.Println("Starting HTTP server on " + s.httpServer.Addr)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.Log.Fatalf("HTTP Server Error: %v", err)
			}
		}()
	}

	if s.kafka != nil {
		// Start Kafka Consumer
		go func() {
			s.Log.Println("Starting Kafka consumer...")
			if err := s.kafka.StartConsumer(ctx); err != nil {
				s.Log.Printf("Kafka consumer error: %v", err)
			}
		}()
	}

	// Wait for termination signal
	<-signalChan
	s.Log.Println("Shutdown signal received")

	// Gracefully shutdown Kafka
	cancel()

	if s.kafka != nil {
		s.kafka.Shutdown()
	}

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.Log.Printf("HTTP Server Shutdown Error: %v", err)
	} else {
		s.Log.Println("HTTP server shutdown complete")
	}

	s.Log.Println("Application exited cleanly")
}

func (s *Server) Consume(topic string, handler ServiceHandleFunc) {
	s.kafka.Consume(topic, handler)
}

func (s *Server) SendMessage(topic string, payload any, opts ...OptionProducerMsg) (RecordMetadata, error) {
	return producer(s.kafka.producer, topic, payload, opts...)
}

func (s *Server) Get(path string, handler HandleFunc, middlewares ...Middleware) {
	s.router.Get(path, handler, middlewares...)
}

func (s *Server) Post(path string, handler HandleFunc, middlewares ...Middleware) {
	s.router.Post(path, handler, middlewares...)
}

func (s *Server) Put(path string, handler HandleFunc, middlewares ...Middleware) {
	s.router.Put(path, handler, middlewares...)
}

func (s *Server) Delete(path string, handler HandleFunc, middlewares ...Middleware) {
	s.router.Delete(path, handler, middlewares...)
}

func (s *Server) Patch(path string, handler HandleFunc, middlewares ...Middleware) {
	s.router.Patch(path, handler, middlewares...)
}

func (s *Server) Use(middlewares ...Middleware) {
	s.router.Use(middlewares...)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
