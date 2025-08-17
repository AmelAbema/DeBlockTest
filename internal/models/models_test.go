package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestIsZeroAddress(t *testing.T) {
	zeroAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
	assert.True(t, IsZeroAddress(zeroAddr))

	nonZeroAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	assert.False(t, IsZeroAddress(nonZeroAddr))

	emptyAddr := common.Address{}
	assert.True(t, IsZeroAddress(emptyAddr))
}

func TestTransactionEvent_ToJSON(t *testing.T) {
	timestamp := time.Now()
	event := &TransactionEvent{
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		BlockHash:       "0xabcdef1234567890",
		UserID:          "user1",
		Source:          "0x1111111111111111111111111111111111111111",
		Destination:     "0x2222222222222222222222222222222222222222",
		Amount:          "1000000000000000000",
		Fees:            "420000000000000",
		GasUsed:         21000,
		GasPrice:        "20000000000",
		Timestamp:       timestamp,
		Status:          1,
		Nonce:           1,
	}

	jsonData, err := event.ToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	assert.Equal(t, event.TransactionHash, jsonMap["transaction_hash"])
	assert.Equal(t, float64(event.BlockNumber), jsonMap["block_number"])
	assert.Equal(t, event.UserID, jsonMap["user_id"])
	assert.Equal(t, event.Amount, jsonMap["amount"])
}

func TestTransactionEvent_FromJSON(t *testing.T) {
	timestamp := time.Now()
	originalEvent := &TransactionEvent{
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		BlockHash:       "0xabcdef1234567890",
		UserID:          "user1",
		Source:          "0x1111111111111111111111111111111111111111",
		Destination:     "0x2222222222222222222222222222222222222222",
		Amount:          "1000000000000000000",
		Fees:            "420000000000000",
		GasUsed:         21000,
		GasPrice:        "20000000000",
		Timestamp:       timestamp,
		Status:          1,
		Nonce:           1,
	}

	jsonData, err := originalEvent.ToJSON()
	assert.NoError(t, err)

	var deserializedEvent TransactionEvent
	err = deserializedEvent.FromJSON(jsonData)
	assert.NoError(t, err)

	assert.Equal(t, originalEvent.TransactionHash, deserializedEvent.TransactionHash)
	assert.Equal(t, originalEvent.BlockNumber, deserializedEvent.BlockNumber)
	assert.Equal(t, originalEvent.BlockHash, deserializedEvent.BlockHash)
	assert.Equal(t, originalEvent.UserID, deserializedEvent.UserID)
	assert.Equal(t, originalEvent.Source, deserializedEvent.Source)
	assert.Equal(t, originalEvent.Destination, deserializedEvent.Destination)
	assert.Equal(t, originalEvent.Amount, deserializedEvent.Amount)
	assert.Equal(t, originalEvent.Fees, deserializedEvent.Fees)
	assert.Equal(t, originalEvent.GasUsed, deserializedEvent.GasUsed)
	assert.Equal(t, originalEvent.GasPrice, deserializedEvent.GasPrice)
	assert.Equal(t, originalEvent.Status, deserializedEvent.Status)
	assert.Equal(t, originalEvent.Nonce, deserializedEvent.Nonce)

	assert.WithinDuration(t, originalEvent.Timestamp, deserializedEvent.Timestamp, time.Second)
}

func TestTransactionEvent_FromJSON_InvalidJSON(t *testing.T) {
	var event TransactionEvent

	err := event.FromJSON([]byte("invalid json"))
	assert.Error(t, err)
}

func TestAddressMatchResult(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	result := &AddressMatchResult{
		IsMatch:       true,
		UserID:        "user1",
		Address:       address,
		IsSource:      true,
		IsDestination: false,
	}

	assert.True(t, result.IsMatch)
	assert.Equal(t, "user1", result.UserID)
	assert.Equal(t, address, result.Address)
	assert.True(t, result.IsSource)
	assert.False(t, result.IsDestination)
}

func TestUserAddress(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	now := time.Now()

	userAddress := &UserAddress{
		ID:        1,
		UserID:    "user1",
		Address:   address,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, uint64(1), userAddress.ID)
	assert.Equal(t, "user1", userAddress.UserID)
	assert.Equal(t, address, userAddress.Address)
	assert.True(t, userAddress.IsActive)
	assert.Equal(t, now, userAddress.CreatedAt)
	assert.Equal(t, now, userAddress.UpdatedAt)
}

func TestAddressStats(t *testing.T) {
	now := time.Now()

	stats := &AddressStats{
		LoadedAddresses:   500000,
		LastUpdated:       now,
		AddressesInMemory: 499950,
	}

	assert.Equal(t, 500000, stats.LoadedAddresses)
	assert.Equal(t, now, stats.LastUpdated)
	assert.Equal(t, 499950, stats.AddressesInMemory)
}

func TestProcessedTransactionLog(t *testing.T) {
	now := time.Now()

	log := &ProcessedTransactionLog{
		ID:                 1,
		TransactionHash:    "0x1234567890abcdef",
		BlockNumber:        18000000,
		UserID:             "user1",
		SourceAddress:      "0x1111111111111111111111111111111111111111",
		DestinationAddress: "0x2222222222222222222222222222222222222222",
		Amount:             "1000000000000000000",
		Fees:               "420000000000000",
		GasUsed:            21000,
		GasPrice:           "20000000000",
		ProcessedAt:        now,
		KafkaPublished:     true,
	}

	assert.Equal(t, uint64(1), log.ID)
	assert.Equal(t, "0x1234567890abcdef", log.TransactionHash)
	assert.Equal(t, uint64(18000000), log.BlockNumber)
	assert.Equal(t, "user1", log.UserID)
	assert.True(t, log.KafkaPublished)
}

func TestFailedTransaction(t *testing.T) {
	now := time.Now()

	failed := &FailedTransaction{
		ID:              1,
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		ErrorMessage:    "RPC timeout",
		RetryCount:      2,
		MaxRetries:      3,
		CreatedAt:       now,
		LastRetryAt:     &now,
		Resolved:        false,
	}

	assert.Equal(t, uint64(1), failed.ID)
	assert.Equal(t, "0x1234567890abcdef", failed.TransactionHash)
	assert.Equal(t, "RPC timeout", failed.ErrorMessage)
	assert.Equal(t, 2, failed.RetryCount)
	assert.Equal(t, 3, failed.MaxRetries)
	assert.False(t, failed.Resolved)
	assert.NotNil(t, failed.LastRetryAt)
}

func TestProcessingError(t *testing.T) {
	err := ProcessingError{
		Type:        ErrorTypeConnection,
		Message:     "Connection failed",
		Code:        "CONN_001",
		BlockNumber: 18000000,
		TxHash:      "0x1234567890abcdef",
		Retryable:   true,
	}

	assert.Equal(t, ErrorTypeConnection, err.Type)
	assert.Equal(t, "Connection failed", err.Message)
	assert.Equal(t, "CONN_001", err.Code)
	assert.Equal(t, uint64(18000000), err.BlockNumber)
	assert.Equal(t, "0x1234567890abcdef", err.TxHash)
	assert.True(t, err.Retryable)

	assert.Equal(t, "Connection failed", err.Error())
}

func TestErrorTypes(t *testing.T) {
	assert.Equal(t, ErrorType("CONNECTION"), ErrorTypeConnection)
	assert.Equal(t, ErrorType("PROCESSING"), ErrorTypeProcessing)
	assert.Equal(t, ErrorType("VALIDATION"), ErrorTypeValidation)
	assert.Equal(t, ErrorType("NOT_FOUND"), ErrorTypeNotFound)
	assert.Equal(t, ErrorType("TIMEOUT"), ErrorTypeTimeout)
}

func BenchmarkTransactionEvent_ToJSON(b *testing.B) {
	event := &TransactionEvent{
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		BlockHash:       "0xabcdef1234567890",
		UserID:          "user1",
		Source:          "0x1111111111111111111111111111111111111111",
		Destination:     "0x2222222222222222222222222222222222222222",
		Amount:          "1000000000000000000",
		Fees:            "420000000000000",
		GasUsed:         21000,
		GasPrice:        "20000000000",
		Timestamp:       time.Now(),
		Status:          1,
		Nonce:           1,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := event.ToJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionEvent_FromJSON(b *testing.B) {
	event := &TransactionEvent{
		TransactionHash: "0x1234567890abcdef",
		BlockNumber:     18000000,
		UserID:          "user1",
		Amount:          "1000000000000000000",
		Timestamp:       time.Now(),
	}

	jsonData, _ := event.ToJSON()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var newEvent TransactionEvent
		err := newEvent.FromJSON(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
