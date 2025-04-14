package app

import (
	"context"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type KafkaServer struct {
	client   sarama.ConsumerGroup
	producer sarama.SyncProducer
	options  *KafkaConfig
	mutex    sync.Mutex
	handlers map[string]ServiceHandleFunc
	topics   []string
	log      ILogger
}

func NewKafkaServer(producer sarama.SyncProducer, client sarama.ConsumerGroup, options *KafkaConfig, log ILogger) (*KafkaServer, error) {
	return &KafkaServer{
		client:   client,
		producer: producer,
		options:  options,
		handlers: make(map[string]ServiceHandleFunc),
		log:      log,
	}, nil
}

func (s *KafkaServer) StartConsumer(ctx context.Context) error {
	if len(s.topics) == 0 {
		return nil
	}

	for {
		select {
		case <-ctx.Done():

			s.log.Println("Stopping Kafka consumer...")
			return nil
		default:
			if err := s.client.Consume(ctx, s.topics, s); err != nil {
				s.log.Printf("Error consuming messages: %v", err)
				time.Sleep(time.Second) // Avoid tight loop on failure
				return err
			}
		}
	}

}
func (s *KafkaServer) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.client == nil {
		return
	}

	s.log.Println("Closing Kafka consumer...")
	if err := s.client.Close(); err != nil {
		s.log.Printf("Error closing Kafka consumer: %v", err)
	}

	s.log.Println("Closing Kafka producer...")
	if err := s.producer.Close(); err != nil {
		s.log.Printf("Error closing Kafka producer: %v", err)
	}
}

func (s *KafkaServer) SendMessage(topic string, payload any, opts ...OptionProducerMsg) (RecordMetadata, error) {
	return producer(s.producer, topic, payload, opts...)
}

func (s *KafkaServer) Consume(topic string, handler ServiceHandleFunc) {
	s.topics = append(s.topics, topic)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handlers[topic] = handler
}

func (s *KafkaServer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (s *KafkaServer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (s *KafkaServer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		ctx := NewConsumerContext(message.Topic, string(message.Value), s.producer, s.log)

		handler, exists := s.handlers[message.Topic]
		if !exists {
			s.log.Printf("No handler for topic: %s", message.Topic)
			continue
		}

		if err := handler(ctx); err != nil {
			s.log.Printf("Handler error: %v", err)
		}

		session.MarkMessage(message, "")
	}
	return nil
}
