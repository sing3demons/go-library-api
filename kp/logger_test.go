package kp

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestMockLoggerAllMethods(t *testing.T) {
	mock := NewMockLogger()

	// Should not panic or error
	err := mock.Sync()
	if err != nil {
		t.Errorf("Sync returned error: %v", err)
	}

	mock.Debug("debug")
	mock.Debugf("debug %s", "format")
	mock.Info("info")
	mock.Infof("info %s", "format")
	mock.Warn("warn")
	mock.Warnf("warn %s", "format")
	mock.WarnMsg("warnmsg", errors.New("warn error"))
	mock.Error("error")
	mock.Errorf("error %s", "format")
	mock.Err("err", errors.New("some error"))
	mock.DPanic("dpanic")
	mock.DPanicf("dpanic %s", "format")
	mock.Fatal("fatal")
	mock.Fatalf("fatal %s", "format")
	mock.Printf("print %s", "format")
	mock.WithName("test-logger")
	mock.Println("print line")

	// Session() should return self
	newLogger := mock.Session(mySession)
	if newLogger != mock {
		t.Error("expected Session to return the same logger instance")
	}

	// Optional: check tracking
	if !mock.Called {
		t.Error("expected Session() to be called")
	}
	if mock.SessionID != mySession {
		t.Errorf("expected SessionID to be 'my-session', got '%s'", mock.SessionID)
	}
}

const (
	msgSession = "existing-session"
	mySession  = "my-session"
)

func TestInitSessionGeneratesSessionIfMissing(t *testing.T) {
	ctx := context.Background()
	mock := &MockLogger{}
	newCtx := InitSession(ctx, mock)

	val := newCtx.Value(xSession)
	if val == nil {
		t.Fatal("expected session to be set in context")
	}

	logger := newCtx.Value(key)
	if logger == nil {
		t.Fatal("expected logger to be set in context")
	}

	if mock.SessionID != val {
		t.Errorf("expected session ID '%s', got '%s'", val, mock.SessionID)
	}
}

func TestInitSessionUsesExistingSession(t *testing.T) {
	ctx := context.WithValue(context.Background(), xSession, msgSession)
	mock := &MockLogger{}
	newCtx := InitSession(ctx, mock)

	val := newCtx.Value(xSession)
	if val != msgSession {
		t.Errorf("expected existing session to be reused, got '%v'", val)
	}

	if mock.SessionID != msgSession {
		t.Errorf("mock logger received unexpected session ID: %s", mock.SessionID)
	}
}

func TestLoggerMethodsNoPanic(t *testing.T) {
	logger := NewAppLogger(zap.NewNop())

	testFuncs := []struct {
		name string
		fn   func()
	}{
		{"Debug", func() { logger.Debug("debug") }},
		{"Debugf", func() { logger.Debugf("debug %s", "format") }},
		{"Info", func() { logger.Info("info") }},
		{"Infof", func() { logger.Infof("info %s", "format") }},
		{"Warn", func() { logger.Warn("warn") }},
		{"Warnf", func() { logger.Warnf("warn %s", "format") }},
		{"WarnMsg", func() { logger.WarnMsg("warnmsg", errors.New("warn error")) }},
		{"Error", func() { logger.Error("error") }},
		{"Errorf", func() { logger.Errorf("error %s", "format") }},
		{"Err", func() { logger.Err("err", errors.New("some error")) }},
		{"DPanic", func() { logger.DPanic("dpanic") }},
		{"DPanicf", func() { logger.DPanicf("dpanic %s", "format") }},
		{"Printf", func() { logger.Printf("printf %s", "data") }},
		{"WithName", func() { logger.WithName("logger-name") }},
		{"Println", func() { logger.Println("hello", "world") }},
	}

	for _, test := range testFuncs {
		t.Run(test.name, func(t *testing.T) {
			test.fn() // Just ensure no panic
		})
	}
}

func TestLoggerSessionReturnsNewInstance(t *testing.T) {
	logger := NewAppLogger(zap.NewNop())
	newLogger := logger.Session("mysession")

	if newLogger == nil {
		t.Fatal("expected new logger from Session()")
	}
}
