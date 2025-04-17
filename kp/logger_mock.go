package kp

import (
	"context"
)

type MockLogger struct {
	Called        bool
	SessionID     string
	Calls         []string
	methodsToCall map[string]bool
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		methodsToCall: make(map[string]bool),
	}
}
func (m *MockLogger) Sync() error                          { m.Calls = append(m.Calls, "Sync"); return nil }
func (m *MockLogger) Debug(args ...any)                    { m.Calls = append(m.Calls, "Debug") }
func (m *MockLogger) Debugf(template string, args ...any)  { m.Calls = append(m.Calls, "Debugf") }
func (m *MockLogger) Info(args ...any)                     { m.Calls = append(m.Calls, "Info") }
func (m *MockLogger) Infof(template string, args ...any)   { m.Calls = append(m.Calls, "Infof") }
func (m *MockLogger) Warn(args ...any)                     { m.Calls = append(m.Calls, "Warn") }
func (m *MockLogger) Warnf(template string, args ...any)   { m.Calls = append(m.Calls, "Warnf") }
func (m *MockLogger) WarnMsg(msg string, err error)        { m.Calls = append(m.Calls, "WarnMsg") }
func (m *MockLogger) Error(args ...any)                    { m.Calls = append(m.Calls, "Error") }
func (m *MockLogger) Errorf(template string, args ...any)  { m.Calls = append(m.Calls, "Errorf") }
func (m *MockLogger) Err(msg string, err error)            { m.Calls = append(m.Calls, "Err") }
func (m *MockLogger) DPanic(args ...any)                   { m.Calls = append(m.Calls, "DPanic") }
func (m *MockLogger) DPanicf(template string, args ...any) { m.Calls = append(m.Calls, "DPanicf") }
func (m *MockLogger) Fatal(args ...any)                    { m.Calls = append(m.Calls, "Fatal") }
func (m *MockLogger) Fatalf(template string, args ...any)  { m.Calls = append(m.Calls, "Fatalf") }
func (m *MockLogger) Printf(template string, args ...any)  { m.Calls = append(m.Calls, "Printf") }
func (m *MockLogger) WithName(name string)                 { m.Calls = append(m.Calls, "WithName") }
func (m *MockLogger) Println(v ...any)                     { m.Calls = append(m.Calls, "Println") }
func (m *MockLogger) Session(v string) ILogger {
	m.Called = true
	m.SessionID = v
	m.Calls = append(m.Calls, "Session")
	return m
}

func (m *MockLogger) L(c context.Context) ILogger {
	return &MockLogger{}
}
