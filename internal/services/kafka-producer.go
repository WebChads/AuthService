package services

type KafkaProducer interface {
	SendPhoneNumber(phoneNumber string) error
}

type confluentKafkaProducer struct {
}

func (kafkaProducer *confluentKafkaProducer) SendPhoneNumber(phoneNumber string) error {
	return nil
}

func InitKafkaProducer(config KafkaConfig) (KafkaProducer, error) {
	return &confluentKafkaProducer{}, nil
}
