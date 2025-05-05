package kp

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/sing3demons/go-library-api/pkg/kp/logger"
	"go.opentelemetry.io/otel/sdk/trace"
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
	AppName    string
	Version    string
	Port       string
	Router     Router
	LogKP      bool
	TracerHost string
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
	httpServer    *http.Server
	kafka         *KafkaServer
	router        IRouter
	Log           ILogger
	traceProvider *trace.TracerProvider
}

func NewApplication(config *Config, nLog ILogger) IApplication {
	kafka := &KafkaServer{}

	if len(config.KafkaConfig.Brokers) != 0 {
		producer, err := newProducer(&config.KafkaConfig)
		if err != nil {
			log.Fatalf("Failed to create Kafka consumer: %v", err)
		}

		client, err := newConsumer(&config.KafkaConfig)
		if err != nil {
			log.Fatalf("Failed to create Kafka consumer: %v", err)
		}
		k, err := NewKafkaServer(producer, client, &config.KafkaConfig, nLog)
		if err != nil {
			log.Fatalf("Failed to create Kafka server: %v", err)
		}

		kafka = k
	}

	var traceProvider *trace.TracerProvider
	if config.AppConfig.TracerHost != "" {
		tp, err := startTracing(config.AppConfig.AppName, config.AppConfig.TracerHost)
		if err != nil {
			nLog.Errorf("Failed to start tracing: %v", err)
		} else {
			traceProvider = tp
		}
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
			router = newServer(config, nLog)
		}
	}

	if config.AppConfig.LogKP {
		logger.LoadLogConfig(logger.LogConfig{
			Summary: logger.SummaryLogConfig{
				LogFile:    true,
				LogConsole: true,
			},
			Detail: logger.DetailLogConfig{
				LogFile: true,
			},
		})
	}

	return &Server{
		kafka:         kafka,
		router:        router,
		Log:           nLog,
		traceProvider: traceProvider,
	}
}

func (s *Server) Start() {
	if s.router != nil {
		s.httpServer = s.router.Register()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	if s.httpServer != nil {
		// Start HTTP Server
		go func() {
			s.Log.Println("Starting HTTP server on " + s.httpServer.Addr)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.Log.Printf("HTTP Server Error: %v", err)
				os.Exit(1)
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

	if s.kafka != nil {
		s.kafka.Shutdown()
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.Log.Printf("HTTP Server Shutdown Error: %v", err)
	} else {
		s.Log.Println("HTTP server shutdown complete")
	}

	if s.traceProvider != nil {
		defer func() {
			if err := s.traceProvider.Shutdown(ctx); err != nil {
				s.Log.Println("failed to stop trace provider:", err)
			}
		}()
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
