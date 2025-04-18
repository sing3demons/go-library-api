package kp

import (
	"context"
	"testing"
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
func (m *MockLogger) Sync() error {
	m.Calls = append(m.Calls, "Sync")
	m.methodsToCall["Sync"] = true
	return nil
}
func (m *MockLogger) Debug(args ...any) {
	m.Calls = append(m.Calls, "Debug")
	m.methodsToCall["Debug"] = true
}
func (m *MockLogger) Debugf(template string, args ...any) {
	m.Calls = append(m.Calls, "Debugf")
	m.methodsToCall["Debugf"] = true
}
func (m *MockLogger) Info(args ...any) {
	m.Calls = append(m.Calls, "Info")
	m.methodsToCall["Info"] = true
}
func (m *MockLogger) Infof(template string, args ...any) {
	m.Calls = append(m.Calls, "Infof")
	m.methodsToCall["Infof"] = true
}
func (m *MockLogger) Warn(args ...any) {
	m.Calls = append(m.Calls, "Warn")
	m.methodsToCall["Warn"] = true
}
func (m *MockLogger) Warnf(template string, args ...any) {
	m.Calls = append(m.Calls, "Warnf")
	m.methodsToCall["Warnf"] = true
}
func (m *MockLogger) WarnMsg(msg string, err error) {
	m.Calls = append(m.Calls, "WarnMsg")
	m.methodsToCall["WarnMsg"] = true
}
func (m *MockLogger) Error(args ...any) {
	m.Calls = append(m.Calls, "Error")
	m.methodsToCall["Error"] = true
}
func (m *MockLogger) Errorf(template string, args ...any) {
	m.Calls = append(m.Calls, "Errorf")
	m.methodsToCall["Errorf"] = true
}
func (m *MockLogger) Err(msg string, err error) {
	m.Calls = append(m.Calls, "Err")
	m.methodsToCall["Err"] = true
}
func (m *MockLogger) DPanic(args ...any) {
	m.Calls = append(m.Calls, "DPanic")
	m.methodsToCall["DPanic"] = true
}
func (m *MockLogger) DPanicf(template string, args ...any) {
	m.Calls = append(m.Calls, "DPanicf")
	m.methodsToCall["DPanicf"] = true
}
func (m *MockLogger) Fatal(args ...any) {
	m.Calls = append(m.Calls, "Fatal")
	m.methodsToCall["Fatal"] = true
}
func (m *MockLogger) Fatalf(template string, args ...any) {
	m.Calls = append(m.Calls, "Fatalf")
	m.methodsToCall["Fatalf"] = true
}
func (m *MockLogger) Printf(template string, args ...any) {
	m.Calls = append(m.Calls, "Printf")
	m.methodsToCall["Printf"] = true
}
func (m *MockLogger) WithName(name string) {
	m.Calls = append(m.Calls, "WithName")
	m.methodsToCall["WithName"] = true
}
func (m *MockLogger) Println(v ...any) {
	m.Calls = append(m.Calls, "Println")
	m.methodsToCall["Println"] = true
}
func (m *MockLogger) Session(v string) ILogger {
	m.Called = true
	m.SessionID = v
	m.Calls = append(m.Calls, "Session")
	m.methodsToCall["Session"] = true
	return m
}

func (m *MockLogger) L(c context.Context) ILogger {
	return m
}

func (m *MockLogger) Verify(t *testing.T) {
	for methodName, called := range m.methodsToCall {
		if !called {
			printError(t, methodName)
		}
	}
}
