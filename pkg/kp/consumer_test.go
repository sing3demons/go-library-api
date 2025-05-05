package kp

import (
	"testing"

	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewConsumerContext(t *testing.T) {
	logger := NewMockLogger()
	mockProducer := mocks.NewSyncProducer(t, nil)
	topic := "test-topic"
	body := `{"message": "hello world"}`

	ctx := NewConsumerContext(topic, body, mockProducer, logger)

	assert.NotNil(t, ctx)
	assert.Equal(t, topic, ctx.(*kafkaContext).topic)
	assert.Equal(t, body, ctx.(*kafkaContext).body)
	assert.Equal(t, mockProducer, ctx.(*kafkaContext).producer)
}

func TestReadInput(t *testing.T) {
	logger := NewMockLogger()

	mockProducer := mocks.NewSyncProducer(t, nil)
	topic := "test-topic"
	body := `{"message": "hello world"}`

	ctx := NewConsumerContext(topic, body, mockProducer, logger)

	type testStruct struct {
		Message string `json:"message"`
	}

	var input testStruct
	err := ctx.ReadInput(&input)

	assert.NoError(t, err)
	assert.Equal(t, "hello world", input.Message)
}
