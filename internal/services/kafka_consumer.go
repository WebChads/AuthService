package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type KafkaConsumer interface {
	Start()
	GetSmsCode(phoneNumber string) (string, bool)
}

type confluentKafkaConsumer struct {
	logger *zap.Logger

	kafkaConsumer *kafka.Consumer
	smsStorage    SmsStorage

	isStarted bool
}

type SmsCodeMessage struct {
	PhoneNumber string `json:"phone_number"`
	SmsCode     string `json:"sms_code"`
}

var consumerTopicName = "sms-to-auth"
var compiledPhoneNumberRegex *regexp.Regexp

func (kafkaConsumer *confluentKafkaConsumer) Start() {
	if kafkaConsumer.isStarted {
		panic("kafka consumer was already started")
	}

	err := kafkaConsumer.ensureTopicExists()
	if err != nil {
		kafkaConsumer.logger.Error("failed to ensure topic exists", zap.Error(err))
	}

	backoff := 5 * time.Second
	maxBackoff := 405 * time.Second

	for {
		err := kafkaConsumer.kafkaConsumer.Subscribe(consumerTopicName, nil)
		if err == nil {
			kafkaConsumer.logger.Info("successfully subscribed to Kafka topic", zap.String("topic", consumerTopicName))
			break
		}

		kafkaConsumer.logger.Warn("failed to subscribe to Kafka topic, retrying...",
			zap.String("topic", consumerTopicName),
			zap.Error(err),
			zap.Duration("retry_in", backoff))

		time.Sleep(backoff)
		if backoff < maxBackoff {
			backoff *= 3
		}
	}

	kafkaConsumer.isStarted = true

	for {
		message, err := kafkaConsumer.kafkaConsumer.ReadMessage(-1)

		if err != nil {
			kafkaConsumer.logger.Error("while listening for messages happened error", zap.Error(err))
			time.Sleep(time.Second * 5)
			continue
		}

		var codeMessage SmsCodeMessage
		err = json.Unmarshal(message.Value, &codeMessage)
		if err != nil || codeMessage.PhoneNumber == "" || codeMessage.SmsCode == "" {
			if err != nil {
				kafkaConsumer.logger.Error("while unmarshalling message in listener happened error: " + err.Error())
			} else {
				kafkaConsumer.logger.Error("while unmarshalling message in listener happened error: fields of dto remained empty somehow")
			}

			continue
		}

		kafkaConsumer.logger.Info(fmt.Sprintf("received sms code message for %s: %s", codeMessage.PhoneNumber, codeMessage.SmsCode))

		isPhoneNumberCorrect := compiledPhoneNumberRegex.Match([]byte(codeMessage.PhoneNumber))
		if !isPhoneNumberCorrect {
			kafkaConsumer.logger.Error(fmt.Errorf("user sent invalid phone number: %s", codeMessage.PhoneNumber).Error())
			continue
		}

		if len(codeMessage.SmsCode) != 4 {
			kafkaConsumer.logger.Error(fmt.Errorf("wrong format of sms code (must be 4 digits): %s", codeMessage.SmsCode).Error())
			continue
		}

		kafkaConsumer.smsStorage.Set(codeMessage.PhoneNumber, codeMessage.SmsCode)
	}
}

func (kafkaConsumer *confluentKafkaConsumer) GetSmsCode(phoneNumber string) (string, bool) {
	code, exists := kafkaConsumer.smsStorage.Get(phoneNumber)
	if !exists {
		return "", false
	}

	return code, true
}

func (kafkaConsumer *confluentKafkaConsumer) ensureTopicExists() error {
	adminClient, err := kafka.NewAdminClientFromConsumer(kafkaConsumer.kafkaConsumer)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer adminClient.Close()

	metadata, err := adminClient.GetMetadata(&consumerTopicName, false, 5000)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	if _, exists := metadata.Topics[consumerTopicName]; exists {
		kafkaConsumer.logger.Info("topic already exists", zap.String("topic", consumerTopicName))
		return nil
	}

	// Создаем топик
	topicSpec := kafka.TopicSpecification{
		Topic:             consumerTopicName,
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

	kafkaConsumer.logger.Info("topic successfully created", zap.String("topic", consumerTopicName))
	return nil
}

var singletoneKafkaConsumer = &confluentKafkaConsumer{}

func InitKafkaConsumer(config KafkaConfig, logger *zap.Logger) (KafkaConsumer, error) {
	if singletoneKafkaConsumer.kafkaConsumer == nil {
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": config.Url,
			"group.id":          "sms-to-auth-listener",
			"auto.offset.reset": "earliest",

			"socket.timeout.ms":       10000,
			"session.timeout.ms":      6000,
			"heartbeat.interval.ms":   2000,
			"max.poll.interval.ms":    300000,
			"enable.auto.commit":      true,
			"auto.commit.interval.ms": 5000,
		})

		if err != nil {
			return nil, errors.New("while initing kafka consumer happened error: " + err.Error())
		}

		singletoneKafkaConsumer.kafkaConsumer = consumer
		singletoneKafkaConsumer.smsStorage = NewSmsStorage()
		singletoneKafkaConsumer.logger = logger

		compiledPhoneNumberRegex, _ = regexp.Compile(`^(8|\+7)(\s|\(|-)?(\d{3})(\s|\)|-)?(\d{3})(\s|-)?(\d{2})(\s|-)?(\d{2})$`)
	}

	return singletoneKafkaConsumer, nil
}
