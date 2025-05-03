package services

import (
	"encoding/json"
	"errors"
	"fmt"

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

var consumerTopicName = "smstoauth"

func (kafkaConsumer *confluentKafkaConsumer) Start() {
	if kafkaConsumer.isStarted {
		panic("kafka consumer was already started")
	}

	err := kafkaConsumer.kafkaConsumer.Subscribe(consumerTopicName, nil)
	if err != nil {
		panic("while subscribing to topic happened error: " + err.Error())
	}

	kafkaConsumer.isStarted = true

	for {
		message, err := kafkaConsumer.kafkaConsumer.ReadMessage(-1)
		if err != nil {
			kafkaConsumer.logger.Error("while listening for messages happened error: " + err.Error())
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

var singletoneKafkaConsumer = &confluentKafkaConsumer{}

func InitKafkaConsumer(config KafkaConfig, logger *zap.Logger) (KafkaConsumer, error) {
	if singletoneKafkaConsumer.kafkaConsumer == nil {
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": config.Url,
			"group.id":          "sms-to-auth-listener",
			"auto.offset.reset": "earliest"})

		if err != nil {
			return nil, errors.New("while initing kafka consumer happened error: " + err.Error())
		}

		singletoneKafkaConsumer.kafkaConsumer = consumer
		singletoneKafkaConsumer.smsStorage = NewSmsStorage()
		singletoneKafkaConsumer.logger = logger
	}

	return singletoneKafkaConsumer, nil
}
