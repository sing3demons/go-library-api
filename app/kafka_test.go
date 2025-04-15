package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	topic = "test-topic"
)

type MockConsumerGroup struct {
	mock.Mock
}

func (m *MockConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return nil
}

func (m *MockConsumerGroup) Close() error {
	return nil
}

func (m *MockConsumerGroup) Errors() <-chan error {
	return make(chan error)
}

func (m *MockConsumerGroup) Pause(partitions map[string][]int32) {
	m.Called(partitions)
}

func (m *MockConsumerGroup) Resume(partitions map[string][]int32) {
	m.Called(partitions)
}

func (m *MockConsumerGroup) PauseAll() {
	m.Called()
}

func (m *MockConsumerGroup) ResumeAll() {
	m.Called()
}

func (m *MockConsumerGroup) Topics() ([]string, error) {
	return []string{topic}, nil
}

func TestKafkaServerStartConsumer(t *testing.T) {
	mockConsumer := &MockConsumerGroup{}

	mockProducer := mocks.NewSyncProducer(t, nil)
	ctx, cancel := context.WithCancel(context.Background())

	server, err := NewKafkaServer(mockProducer, mockConsumer, &KafkaConfig{}, NewMockLogger())
	assert.NoError(t, err)
	assert.NotNil(t, server)

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	err = server.StartConsumer(ctx)
	assert.NoError(t, err)
}

func TestKafkaServerShutdown(t *testing.T) {
	logger := NewMockLogger()

	mockConsumer := &MockConsumerGroup{}

	mockProducer := mocks.NewSyncProducer(t, nil)

	server, err := NewKafkaServer(mockProducer, mockConsumer, &KafkaConfig{}, logger)
	assert.NoError(t, err)
	assert.NotNil(t, server)

	server.Shutdown()
}

func TestKafkaServerConsume(t *testing.T) {
	logger := NewMockLogger()

	mockConsumer := &MockConsumerGroup{}
	mockProducer := mocks.NewSyncProducer(t, nil)

	server, err := NewKafkaServer(mockProducer, mockConsumer, &KafkaConfig{}, logger)
	assert.NoError(t, err)
	assert.NotNil(t, server)

	server.Consume(topic, func(ctx IContext) error { return nil })

	assert.Contains(t, server.topics, topic)
	assert.NotNil(t, server.handlers[topic])
}

// MockConsumerGroupSession is a mock of the sarama.ConsumerGroupSession interface
type MockConsumerGroupSession struct {
	mock.Mock
}

func (m *MockConsumerGroupSession) MarkMessage(message *sarama.ConsumerMessage, metadata string) {
	m.Called(message, metadata)
}

func (m *MockConsumerGroupSession) Claims() map[string][]int32 {
	args := m.Called()
	return args.Get(0).(map[string][]int32)
}

func (m *MockConsumerGroupSession) Commit() {
	m.Called()
}

func (m *MockConsumerGroupSession) GenerationID() int32 {
	args := m.Called()
	return args.Get(0).(int32)
}

func (m *MockConsumerGroupSession) MemberID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	m.Called(topic, partition, offset, metadata)
}

func (m *MockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	m.Called(topic, partition, offset, metadata)
}

func (m *MockConsumerGroupSession) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

// --------------------------------------------

// MockConsumerGroupClaim is a mock of the sarama.ConsumerGroupClaim interface
type MockConsumerGroupClaim struct {
	mock.Mock
}

func (m *MockConsumerGroupClaim) Topic() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConsumerGroupClaim) Partition() int32 {
	args := m.Called()
	return args.Get(0).(int32)
}

func (m *MockConsumerGroupClaim) InitialOffset() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockConsumerGroupClaim) HighWaterMarkOffset() int64 {
	return m.Called().Get(1).(int64)
}

func (m *MockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	args := m.Called()
	return args.Get(0).(chan *sarama.ConsumerMessage)
}

type MockHandler ServiceHandleFunc

func TestConsumeClaim(t *testing.T) {
	// Setup
	mockSession := new(MockConsumerGroupSession)
	mockClaim := new(MockConsumerGroupClaim)
	producer := mocks.NewSyncProducer(t, nil)
	logger := NewMockLogger()

	server := &KafkaServer{
		producer: producer,
		log:      logger,
		handlers: map[string]ServiceHandleFunc{
			"test_topic": func(ctx IContext) error { return nil },
		},
	}

	// Create mock message
	message := &sarama.ConsumerMessage{
		Topic: "test_topic",
		Value: []byte("message value"),
	}

	// Setup mock session expectations
	mockSession.On("MarkMessage", message, "")
	mockSession.On("Claims").Return(map[string][]int32{"test_topic": {0}})
	mockSession.On("Commit")
	mockSession.On("GenerationID").Return(int32(1))                                // Mock GenerationID
	mockSession.On("MemberID").Return("test-member")                               // Mock MemberID
	mockSession.On("MarkOffset", "test_topic", int32(0), int64(1), "").Return(nil) // Mock MarkOffset
	mockSession.On("Context").Return(context.Background())                         // Mock Context

	mockMessageChannel := make(chan *sarama.ConsumerMessage, 1)

	// Setup mock claim expectations
	mockClaim.On("Messages").Return(mockMessageChannel).Once()
	go func() {
		// Simulate sending a message to the channel
		mockMessageChannel <- message
		close(mockMessageChannel)
	}()

	// Run test
	err := server.ConsumeClaim(mockSession, mockClaim)

	// Asserts
	assert.Nil(t, err)

	mockClaim.AssertExpectations(t)
}

func TestConsumeClaimHandlerError(t *testing.T) {
	// Setup
	mockSession := new(MockConsumerGroupSession)
	mockClaim := new(MockConsumerGroupClaim)
	producer := mocks.NewSyncProducer(t, nil)
	logger := NewMockLogger()

	server := &KafkaServer{
		producer: producer,
		log:      logger,
		handlers: map[string]ServiceHandleFunc{
			topic: func(ctx IContext) error {
				return errors.New("handler error") // Simulating an error
			},
		},
	}

	// Create mock message
	message := &sarama.ConsumerMessage{
		Topic: topic,
		Value: []byte("message value"),
	}

	// Mock session method calls
	mockSession.On("MarkMessage", message, "")
	mockSession.On("Claims").Return(map[string][]int32{topic: {0}})
	mockSession.On("Commit")
	mockSession.On("GenerationID").Return(int32(1)) // Mock GenerationID

	mockMessageChannel := make(chan *sarama.ConsumerMessage, 1)

	// Setup mock claim expectations
	mockClaim.On("Messages").Return(mockMessageChannel).Once()
	go func() {
		// Simulate sending a message to the channel
		mockMessageChannel <- message
		close(mockMessageChannel)
	}()

	// Run test
	err := server.ConsumeClaim(mockSession, mockClaim)

	// Asserts
	assert.Nil(t, err)
	// mockSession.AssertExpectations(t)
	mockClaim.AssertExpectations(t)
}

// SendMessage tests
// func TestSendMessage(t *testing.T) {
// 	mockConsumer := &MockConsumerGroup{}

// 	mockProducer := mocks.NewSyncProducer(t, nil)

// 	server, err := NewKafkaServer(mockProducer, mockConsumer, &KafkaConfig{}, NewMockLogger())
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)

// 	// Mock producer method calls
// 	mockProducer.ExpectSendMessageAndSucceed()

// 	// Run test
// 	_, err = server.SendMessage(topic, map[string]string{"message": "hello world"})
// 	// Asserts
// 	assert.Nil(t, err)
// }
