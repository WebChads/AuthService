package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type KafkaProducer interface {
	SendPhoneNumber(phoneNumber string) error
}

type confluentKafkaProducer struct {
	kafkaProducer *kafka.Producer
	logger        *zap.Logger
}

type phoneNumberRequestDto struct {
	PhoneNumber string `json:"phone_number"`
}

var producerTopicName = "auth-to-sms"
var singletoneKafkaProducer *confluentKafkaProducer = &confluentKafkaProducer{}

func (kafkaProducer *confluentKafkaProducer) SendPhoneNumber(phoneNumber string) error {
	err := kafkaProducer.ensureTopicExists()
	if err != nil {
		kafkaProducer.logger.Error("failed to ensure topic exists, proceeding anyway...", zap.Error(err))
		return err
	}

	dto := phoneNumberRequestDto{PhoneNumber: phoneNumber}
	encodedMessage, err := json.Marshal(dto)

	if err != nil {
		return errors.New("while encoding phone number in dto happened error: " + err.Error())
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &producerTopicName, Partition: kafka.PartitionAny},
		Value:          encodedMessage,
	}

	err = kafkaProducer.kafkaProducer.Produce(message, nil)
	if err != nil {
		return errors.New("while producing message in kafka happened error: " + err.Error())
	}

	return nil
}

func (kafkaProducer *confluentKafkaProducer) ensureTopicExists() error {
	adminClient, err := kafka.NewAdminClientFromProducer(kafkaProducer.kafkaProducer)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer adminClient.Close()

	metadata, err := adminClient.GetMetadata(&producerTopicName, false, 5000)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	if _, exists := metadata.Topics[producerTopicName]; exists {
		kafkaProducer.logger.Info("topic already exists", zap.String("topic", producerTopicName))
		return nil
	}

	topicSpec := kafka.TopicSpecification{
		Topic:             producerTopicName,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	results, err := adminClient.CreateTopics(
		context.Background(),
		[]kafka.TopicSpecification{topicSpec},
		kafka.SetAdminOperationTimeout(10000),
		kafka.SetAdminRequestTimeout(10000),
	)

	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to create topic %s: %s", result.Topic, result.Error.String())
		}
	}

	kafkaProducer.logger.Info("topic successfully created", zap.String("topic", producerTopicName))
	return nil
}

func NewKafkaProducer(config KafkaConfig, logger *zap.Logger) (KafkaProducer, error) {
	if singletoneKafkaProducer.kafkaProducer == nil {
		producer, err := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":      config.Url,
			"socket.timeout.ms":      10000,
			"message.timeout.ms":     30000,
			"request.timeout.ms":     5000,
			"retries":                5,
			"retry.backoff.ms":       1000,
			"enable.idempotence":     true,
			"queue.buffering.max.ms": 100,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to init kafka producer: %w", err)
		}

		singletoneKafkaProducer.kafkaProducer = producer
		singletoneKafkaProducer.logger = logger

		// Запускаем горутину для обработки delivery reports (иначе могут быть memory leaks)
		go func() {
			for e := range producer.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						logger.Error("delivery failed",
							zap.String("topic", *ev.TopicPartition.Topic),
							zap.Error(ev.TopicPartition.Error))
					} else {
						logger.Debug("delivered message",
							zap.String("topic", *ev.TopicPartition.Topic),
							zap.Int32("partition", ev.TopicPartition.Partition),
						)
					}
				}
			}
		}()
	}

	return singletoneKafkaProducer, nil
}
