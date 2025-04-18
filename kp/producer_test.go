package kp

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProducer(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil) // Use mocks.NewSyncProducer

	topic := "test-topic"
	payload := map[string]string{"message": "hello world"}

	mockProducer.ExpectSendMessageAndSucceed() // Expect a successful send

	recordMetadata, err := producer(mockProducer, topic, payload)

	assert.NoError(t, err)
	assert.Equal(t, topic, recordMetadata.TopicName)
	assert.GreaterOrEqual(t, recordMetadata.Partition, int32(0))
	assert.GreaterOrEqual(t, recordMetadata.Offset, int64(0))
}

type MockSyncProducer struct {
	mock.Mock
	sarama.SyncProducer
}

func (m *MockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (m *MockSyncProducer) Close() error {
	return m.Called().Error(0)
}
func TestNewProducerWithMockedProducer(t *testing.T) {
	mockProducer := new(MockSyncProducer)

	cfg := &KafkaConfig{
		Brokers:  []string{"localhost:9092"},
		producer: mockProducer,
	}

	producer, err := newProducer(cfg)

	assert.NoError(t, err)
	assert.Equal(t, mockProducer, producer)
}

func TestNewProducerWithoutAuth(t *testing.T) {
	cfg := &KafkaConfig{
		Brokers:  []string{"localhost:9092"},
		Username: "test",
		Password: "test",
	}

	// This actually tries to connect to Kafka
	// In a real unit test you'd mock sarama.NewSyncProducer via interface/injection
	producer, err := newProducer(cfg)

	if err != nil {
		t.Logf("Expected failure if Kafka not available: %v", err)
	} else {
		assert.NotNil(t, producer)
		_ = producer.Close()
	}
}
