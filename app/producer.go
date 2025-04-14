package app

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
)

type RecordMetadata struct {
	TopicName      string `json:"topicName"`
	Partition      int32  `json:"partition"`
	ErrorCode      int    `json:"errorCode"`
	Offset         int64  `json:"offset,omitempty"`
	Timestamp      string `json:"timestamp,omitempty"`
	BaseOffset     string `json:"baseOffset,omitempty"`
	LogAppendTime  string `json:"logAppendTime,omitempty"`
	LogStartOffset string `json:"logStartOffset,omitempty"`
}

func newProducer(option *KafkaConfig) (sarama.SyncProducer, error) {
	if option.producer != nil {
		return option.producer, nil
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_5_0_0

	if option.Username != "" && option.Password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = option.Username
		config.Net.SASL.Password = option.Password
	}

	return sarama.NewSyncProducer(option.Brokers, config)
}

func producer(producer sarama.SyncProducer, topic string, payload any, opts ...OptionProducerMsg) (RecordMetadata, error) {
	timestamp := time.Now()

	data, err := json.Marshal(payload)
	if err != nil {
		return RecordMetadata{}, err
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(data),
		Timestamp: timestamp,
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			if opt.key != "" {
				msg.Key = sarama.StringEncoder(opt.key)
			}

			if len(opt.headers) > 0 {
				for _, header := range opt.headers {
					for key, value := range header {
						msg.Headers = append(msg.Headers, sarama.RecordHeader{
							Key:   []byte(key),
							Value: []byte(value),
						})
					}
				}
			}

			if !opt.Timestamp.IsZero() {
				msg.Timestamp = opt.Timestamp
			}

			if opt.Metadata != nil {
				msg.Metadata = opt.Metadata
			}

			if opt.Offset > 0 {
				msg.Offset = opt.Offset
			}

			if opt.Partition > 0 {
				msg.Partition = opt.Partition
			}
		}
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return RecordMetadata{}, err
	}

	recordMetadata := RecordMetadata{
		TopicName:      topic,
		Partition:      partition,
		Offset:         offset,
		ErrorCode:      0,
		Timestamp:      timestamp.String(),
		BaseOffset:     "",
		LogAppendTime:  "",
		LogStartOffset: "",
	}

	return recordMetadata, nil
}
