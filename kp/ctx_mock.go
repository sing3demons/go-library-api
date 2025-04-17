package kp

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/sing3demons/go-library-api/kp/logger"
)

func (m *MockLogger) NewLog(ctx context.Context, initInvoke, scenario string) (logger.DetailLog, logger.SummaryLog) {
	detailLog := &MockDetailLog{
		methodsToCall: make(map[string]bool),
	}

	summaryLog := &MockSummaryLog{
		methodsToCall: make(map[string]bool),
	}

	for mc, c := range detailLog.methodsToCall {
		if c {
			m.methodsToCall[mc] = c
		}
	}

	for mc, c := range summaryLog.methodsToCall {
		if c {
			m.methodsToCall[mc] = c
		}
	}

	return detailLog, summaryLog
}

// Mock implementations for DetailLog and SummaryLog
type MockDetailLog struct {
	methodsToCall map[string]bool
}

func (m *MockDetailLog) IsRawDataEnabled() bool {
	m.methodsToCall["IsRawDataEnabled"] = true
	return false
}

func (m *MockDetailLog) AddInputRequest(node, cmd, invoke string, rawData, data any) {
	m.methodsToCall["AddInputRequest"] = true
}

func (m *MockDetailLog) AddInputHttpRequest(node, cmd, invoke string, req *http.Request, rawData bool) {
	m.methodsToCall["AddInputHttpRequest"] = true
}

func (m *MockDetailLog) AddOutputRequest(node, cmd, invoke string, rawData, data any) {
	m.methodsToCall["AddOutputRequest"] = true
}

func (m *MockDetailLog) End() {
	m.methodsToCall["End"] = true
}

func (m *MockDetailLog) AddInputResponse(node, cmd, invoke string, rawData, data any, protocol, protocolMethod string) {
	m.methodsToCall["AddInputResponse"] = true
}

func (m *MockDetailLog) AddOutputResponse(node, cmd, invoke string, rawData, data any) {
	m.methodsToCall["AddOutputResponse"] = true
}

func (m *MockDetailLog) AutoEnd() bool {
	m.methodsToCall["AutoEnd"] = true
	return true
}

func (m *MockDetailLog) ExpectToCall(methodName string) {
	if m.methodsToCall == nil {
		m.methodsToCall = make(map[string]bool)
	}
	m.methodsToCall[methodName] = false
}

func (m *MockDetailLog) Verify(t *testing.T) {
	for methodName, called := range m.methodsToCall {
		if !called {
			t.Errorf("Expected to call '%s', but it wasn't.", methodName)
		}
	}
}

type MockSummaryLog struct {
	methodsToCall map[string]bool
}

func (m *MockSummaryLog) AddField(fieldName string, fieldValue interface{}) {
	m.methodsToCall["AddField"] = true

}
func (m *MockSummaryLog) AddSuccess(node, cmd, code, desc string) {
	m.methodsToCall["AddSuccess"] = true

}
func (m *MockSummaryLog) AddError(node, cmd, code, desc string) {
	m.methodsToCall["AddError"] = true

}
func (m *MockSummaryLog) IsEnd() bool {
	m.methodsToCall["IsEnd"] = true
	return true

}
func (m *MockSummaryLog) End(resultCode, resultDescription string) error {
	m.methodsToCall["End"] = true
	return nil
}

func (m *MockSummaryLog) ExpectToCall(methodName string) {
	if m.methodsToCall == nil {
		m.methodsToCall = make(map[string]bool)
	}
	m.methodsToCall[methodName] = false
}

func (m *MockSummaryLog) Verify(t *testing.T) {
	for methodName, called := range m.methodsToCall {
		if !called {
			t.Errorf("Expected to call '%s', but it wasn't.", methodName)
		}
	}
}

type MockContext struct {
	Ctx           context.Context
	Headers       map[string]string
	QueryParams   map[string]string
	Params        map[string]string
	Input         any
	Output        any
	LogInstance   ILogger
	methodsToCall map[string]bool

	detailLog  logger.DetailLog
	summaryLog logger.SummaryLog
	// baseCommand string
	// initInvoke  string
}

func NewMockContext() *MockContext {
	return &MockContext{
		Ctx:           context.TODO(),
		Headers:       make(map[string]string),
		QueryParams:   make(map[string]string),
		Params:        make(map[string]string),
		LogInstance:   &MockLogger{},
		methodsToCall: make(map[string]bool),
	}
}

func (m *MockContext) Context() context.Context {
	m.methodsToCall["Context"] = true
	return m.Ctx
}

func (m *MockContext) SetHeader(key, value string) {
	m.methodsToCall["SetHeader"] = true
	m.Headers[key] = value
}

func (m *MockContext) GetHeader(key string) string {
	m.methodsToCall["GetHeader"] = true
	return m.Headers[key]
}

func (m *MockContext) Log() ILogger {
	m.methodsToCall["Log"] = true
	return m.LogInstance
}

func (m *MockContext) Param(name string) string {
	m.methodsToCall["Param"] = true
	return m.Params[name]
}

func (m *MockContext) Query(name string) string {
	m.methodsToCall["Query"] = true
	return m.QueryParams[name]
}

func (m *MockContext) ReadInput(data any) error {
	m.methodsToCall["ReadInput"] = true

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
	m.methodsToCall["Response"] = true
	m.Output = data
	return nil
}

func (m *MockContext) CommonLog(cmd, initInvoke, scenario string) {
	m.methodsToCall["CommonLog"] = true
}

func (m *MockContext) DetailLog() logger.DetailLog {
	if m.methodsToCall == nil {
		m.methodsToCall = make(map[string]bool)
	}

	detailLog := &MockDetailLog{methodsToCall: make(map[string]bool)}
	m.detailLog = detailLog
	m.methodsToCall["DetailLog"] = true
	for mn, c := range detailLog.methodsToCall {
		if c {
			m.methodsToCall[mn] = c
		}
	}
	return m.detailLog
}
func (m *MockContext) SummaryLog() logger.SummaryLog {
	if m.methodsToCall == nil {
		m.methodsToCall = make(map[string]bool)
	}

	summaryLog := &MockSummaryLog{methodsToCall: make(map[string]bool)}
	m.summaryLog = summaryLog

	m.methodsToCall["SummaryLog"] = true

	for mn, c := range summaryLog.methodsToCall {
		if c {
			m.methodsToCall[mn] = c
		}
	}

	return m.summaryLog
}

func (m *MockContext) ExpectToCall(methodName string) {
	if m.methodsToCall == nil {
		m.methodsToCall = make(map[string]bool)
	}
	m.methodsToCall[methodName] = false
}

func (m *MockContext) Verify(t *testing.T) {
	for methodName, called := range m.methodsToCall {
		t.Logf("Verifying method '%s'", methodName)
		if !called {
			t.Errorf("Expected to call '%s', but it wasn't.", methodName)
		}
	}
	m = nil
}
