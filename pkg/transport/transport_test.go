package transport

import (
	"DeBlockTest/internal/models"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSyncProducer struct {
	mock.Mock
}

func (m *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (m *mockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func (m *mockSyncProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockSyncProducer) AbortTxn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	args := m.Called(msg, groupId, metadata)
	return args.Error(0)
}

func (m *mockSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	args := m.Called(offsets, groupId)
	return args.Error(0)
}

func (m *mockSyncProducer) BeginTxn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockSyncProducer) CommitTxn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	args := m.Called()
	return args.Get(0).(sarama.ProducerTxnStatusFlag)
}

func (m *mockSyncProducer) IsTransactional() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestKafkaProducer_PublishTransaction_Success(t *testing.T) {
	ctx := context.Background()

	mockProducer := &mockSyncProducer{}

	event := &models.TransactionEvent{
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		UserID:          "user1",
		Amount:          "1000000000000000000",
		Timestamp:       time.Now(),
	}

	mockProducer.On("SendMessage", mock.MatchedBy(func(msg *sarama.ProducerMessage) bool {
		if msg.Topic != "test-topic" {
			return false
		}

		key, _ := msg.Key.Encode()
		if string(key) != event.TransactionHash {
			return false
		}

		value, _ := msg.Value.Encode()
		var receivedEvent models.TransactionEvent
		if err := json.Unmarshal(value, &receivedEvent); err != nil {
			return false
		}

		return receivedEvent.TransactionHash == event.TransactionHash &&
			receivedEvent.UserID == event.UserID &&
			receivedEvent.Amount == event.Amount
	})).Return(int32(0), int64(123), nil)

	kafkaProducer := &KafkaProducer{
		producer: mockProducer,
		topic:    "test-topic",
	}

	err := kafkaProducer.PublishTransaction(ctx, event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestKafkaProducer_Close(t *testing.T) {
	mockProducer := &mockSyncProducer{}
	mockProducer.On("Close").Return(nil)

	kafkaProducer := &KafkaProducer{
		producer: mockProducer,
		topic:    "test-topic",
	}

	err := kafkaProducer.Close()

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestTransactionEvent_JSONSerialization(t *testing.T) {
	timestamp := time.Now()
	event := &models.TransactionEvent{
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		UserID:          "user1",
		Source:          "0x1111111111111111111111111111111111111111",
		Destination:     "0x2222222222222222222222222222222222222222",
		Amount:          "1000000000000000000",
		Fees:            "420000000000000",
		Timestamp:       timestamp,
		Status:          1,
		Nonce:           1,
	}

	jsonData, err := event.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var deserializedEvent models.TransactionEvent
	err = deserializedEvent.FromJSON(jsonData)
	assert.NoError(t, err)

	assert.Equal(t, event.TransactionHash, deserializedEvent.TransactionHash)
	assert.Equal(t, event.UserID, deserializedEvent.UserID)
	assert.Equal(t, event.Amount, deserializedEvent.Amount)
}
