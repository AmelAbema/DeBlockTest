package transport

import (
	"DeBlockTest/internal/config"
	"DeBlockTest/internal/models"
	"context"

	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

type TransportModule struct {
	kafkaProducer  *KafkaProducer
	ethereumClient *EthereumClient
}

func NewTransportModule(ctx context.Context, kafkaConfig *config.KafkaConfig, ethereumConfig *config.EthereumConfig) (*TransportModule, error) {
	kafkaProducer, err := NewKafkaProducer(kafkaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Kafka producer")
	}

	ethereumClient, err := NewEthereumClient(ethereumConfig)
	if err != nil {
		kafkaProducer.Close()
		return nil, errors.Wrap(err, "failed to create Ethereum client")
	}

	tel.Global().Info("transport module initialized",
		tel.Strings("kafka_brokers", kafkaConfig.Brokers),
		tel.String("kafka_topic", kafkaConfig.Topic),
		tel.String("ethereum_rpc", ethereumConfig.RPCURL))

	return &TransportModule{
		kafkaProducer:  kafkaProducer,
		ethereumClient: ethereumClient,
	}, nil
}

func (t *TransportModule) Close() error {
	tel.Global().Info("closing transport module")

	var errs []error

	if err := t.kafkaProducer.Close(); err != nil {
		errs = append(errs, errors.Wrap(err, "failed to close Kafka producer"))
	}

	t.ethereumClient.Close()

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (t *TransportModule) GetKafkaProducer() *KafkaProducer {
	return t.kafkaProducer
}

func (t *TransportModule) GetEthereumClient() *EthereumClient {
	return t.ethereumClient
}

func (t *TransportModule) PublishTransaction(ctx context.Context, event *models.TransactionEvent) error {
	return t.kafkaProducer.PublishTransaction(ctx, event)
}
