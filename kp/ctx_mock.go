package kp

import (
	"context"
	"errors"
	"net/http"

	"github.com/sing3demons/go-library-api/kp/logger"
)

func (m *MockLogger) NewLog(ctx context.Context, initInvoke, scenario string) (logger.DetailLog, logger.SummaryLog) {
	return &MockDetailLog{}, &logger.MockSummaryLog{}
}

// Mock implementations for DetailLog and SummaryLog
type MockDetailLog struct{}

func (m *MockDetailLog) IsRawDataEnabled() bool {
	return false
}

func (m *MockDetailLog) AddInputRequest(node, cmd, invoke string, rawData, data any) {}

func (m *MockDetailLog) AddInputHttpRequest(node, cmd, invoke string, req *http.Request, rawData bool) {
}

func (m *MockDetailLog) AddOutputRequest(node, cmd, invoke string, rawData, data any) {}

func (m *MockDetailLog) End() {}

func (m *MockDetailLog) AddInputResponse(node, cmd, invoke string, rawData, data any, protocol, protocolMethod string) {
}

func (m *MockDetailLog) AddOutputResponse(node, cmd, invoke string, rawData, data any) {}

func (m *MockDetailLog) AutoEnd() bool {
	return true
}

type MockSummaryLog struct{}

type MockContext struct {
	Ctx         context.Context
	Headers     map[string]string
	QueryParams map[string]string
	Params      map[string]string
	Input       any
	Output      any
	LogInstance ILogger
}

func NewMockContext() *MockContext {
	return &MockContext{
		Ctx:         context.Background(),
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Params:      make(map[string]string),
		LogInstance: &MockLogger{},
	}
}

func (m *MockContext) Context() context.Context {
	return m.Ctx
}

func (m *MockContext) SetHeader(key, value string) {
	m.Headers[key] = value
}

func (m *MockContext) GetHeader(key string) string {
	return m.Headers[key]
}

func (m *MockContext) Log() ILogger {
	return m.LogInstance
}

func (m *MockContext) Param(name string) string {
	return m.Params[name]
}

func (m *MockContext) Query(name string) string {
	return m.QueryParams[name]
}

func (m *MockContext) ReadInput(data any) error {
	if m.Input == nil {
		return errors.New("no input set")
	}
	// Simulate decoding input
	switch v := data.(type) {
	case *string:
		if s, ok := m.Input.(string); ok {
			*v = s
			return nil
		}
	}
	return errors.New("type mismatch")
}

func (m *MockContext) Response(code int, data any) error {
	m.Output = data
	return nil
}

func (m *MockContext) CommonLog(cmd, initInvoke, scenario string) {}
func (m *MockContext) DetailLog() logger.DetailLog {
	return &MockDetailLog{}
}
func (m *MockContext) SummaryLog() logger.SummaryLog {
	return &logger.MockSummaryLog{}
}
