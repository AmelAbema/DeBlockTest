package transport

import (
	"DeBlockTest/internal/config"
	"DeBlockTest/internal/models"
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(cfg *config.KafkaConfig) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Kafka producer")
	}

	tel.Global().Info("Kafka producer initialized",
		tel.Strings("brokers", cfg.Brokers),
		tel.String("topic", cfg.Topic))

	return &KafkaProducer{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

func (k *KafkaProducer) PublishTransaction(ctx context.Context, event *models.TransactionEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal transaction event")
	}

	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Key:   sarama.StringEncoder(event.TransactionHash),
		Value: sarama.ByteEncoder(eventData),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		tel.Global().Error("failed to publish transaction event",
			tel.Error(err),
			tel.String("transaction_hash", event.TransactionHash))
		return errors.Wrap(err, "failed to send message to Kafka")
	}

	tel.Global().Debug("transaction event published successfully",
		tel.String("transaction_hash", event.TransactionHash),
		tel.Int32("partition", partition),
		tel.Int64("offset", offset))

	return nil
}

func (k *KafkaProducer) Close() error {
	if k.producer != nil {
		return k.producer.Close()
	}
	return nil
}
