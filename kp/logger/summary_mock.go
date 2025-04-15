package logger

import (
	"github.com/stretchr/testify/mock"
)

// MockSummaryLog is a mock implementation of the SummaryLog interface
type MockSummaryLog struct {
	mock.Mock
}

// AddField mocks the AddField method
func (m *MockSummaryLog) AddField(fieldName string, fieldValue interface{}) {
	m.Called(fieldName, fieldValue)
}

// AddSuccess mocks the AddSuccess method
func (m *MockSummaryLog) AddSuccess(node, cmd, code, desc string) {
	m.Called(node, cmd, code, desc)
}

// AddError mocks the AddError method
func (m *MockSummaryLog) AddError(node, cmd, code, desc string) {
	m.Called(node, cmd, code, desc)
}

// IsEnd mocks the IsEnd method
func (m *MockSummaryLog) IsEnd() bool {
	args := m.Called()
	return args.Bool(0)
}

// End mocks the End method
func (m *MockSummaryLog) End(resultCode, resultDescription string) error {
	args := m.Called(resultCode, resultDescription)
	return args.Error(0)
}
