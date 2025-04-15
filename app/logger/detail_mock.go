package logger

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockDetailLog is a mock implementation of the DetailLog interface.
type MockDetailLog struct {
	mock.Mock
}

// IsRawDataEnabled mocks the IsRawDataEnabled method.
func (m *MockDetailLog) IsRawDataEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// AddInputRequest mocks the AddInputRequest method.
func (m *MockDetailLog) AddInputRequest(node, cmd, invoke string, rawData, data interface{}) {
	m.Called(node, cmd, invoke, rawData, data)
}

// AddInputHttpRequest mocks the AddInputHttpRequest method.
func (m *MockDetailLog) AddInputHttpRequest(node, cmd, invoke string, req *http.Request, rawData bool) {
	m.Called(node, cmd, invoke, req, rawData)
}

// AddOutputRequest mocks the AddOutputRequest method.
func (m *MockDetailLog) AddOutputRequest(node, cmd, invoke string, rawData, data interface{}) {
	m.Called(node, cmd, invoke, rawData, data)
}

// End mocks the End method.
func (m *MockDetailLog) End() {
	m.Called()
}

// AddInputResponse mocks the AddInputResponse method.
func (m *MockDetailLog) AddInputResponse(node, cmd, invoke string, rawData, data interface{}, protocol, protocolMethod string) {
	m.Called(node, cmd, invoke, rawData, data, protocol, protocolMethod)
}

// AddOutputResponse mocks the AddOutputResponse method.
func (m *MockDetailLog) AddOutputResponse(node, cmd, invoke string, rawData, data interface{}) {
	m.Called(node, cmd, invoke, rawData, data)
}

// AutoEnd mocks the AutoEnd method.
func (m *MockDetailLog) AutoEnd() bool {
	args := m.Called()
	return args.Get(0).(bool)
}