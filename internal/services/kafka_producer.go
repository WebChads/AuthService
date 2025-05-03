package services

import (
	"encoding/json"
	"errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer interface {
	SendPhoneNumber(phoneNumber string) error
}

type confluentKafkaProducer struct {
	kafkaProducer *kafka.Producer
}

type phoneNumberRequestDto struct {
	PhoneNumber string `json:"phone_number"`
}

var topicName string = "authtosms"
var singletoneKafkaProducer *confluentKafkaProducer = &confluentKafkaProducer{}

func (kafkaProducer *confluentKafkaProducer) SendPhoneNumber(phoneNumber string) error {
	dto := phoneNumberRequestDto{PhoneNumber: phoneNumber}
	encodedMessage, err := json.Marshal(dto)

	if err != nil {
		return errors.New("while encoding phone number in dto happened error: " + err.Error())
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
		Value:          encodedMessage,
	}

	err = kafkaProducer.kafkaProducer.Produce(message, nil)
	if err != nil {
		return errors.New("while producing message in kafka happened error: " + err.Error())
	}

	return nil
}

func NewKafkaProducer(config KafkaConfig) (KafkaProducer, error) {
	if singletoneKafkaProducer.kafkaProducer == nil {
		producer, err := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": config.Url,
		})

		if err != nil {
			return nil, err
		}

		singletoneKafkaProducer.kafkaProducer = producer
	}

	return singletoneKafkaProducer, nil
}
