package models

import (
	"encoding/json"
	"time"
)

type TransactionEvent struct {
	TransactionHash string    `json:"transaction_hash"`
	BlockNumber     uint64    `json:"block_number"`
	BlockHash       string    `json:"block_hash"`
	UserID          string    `json:"user_id"`
	Source          string    `json:"source"`
	Destination     string    `json:"destination"`
	Amount          string    `json:"amount"`
	Fees            string    `json:"fees"`
	GasUsed         uint64    `json:"gas_used"`
	GasPrice        string    `json:"gas_price"`
	Timestamp       time.Time `json:"timestamp"`
	Status          uint64    `json:"status"`
	Nonce           uint64    `json:"nonce"`
}

type ProcessedTransactionLog struct {
	ID                 uint64    `json:"id" db:"id"`
	TransactionHash    string    `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber        uint64    `json:"block_number" db:"block_number"`
	UserID             string    `json:"user_id" db:"user_id"`
	SourceAddress      string    `json:"source_address" db:"source_address"`
	DestinationAddress string    `json:"destination_address" db:"destination_address"`
	Amount             string    `json:"amount" db:"amount"`
	Fees               string    `json:"fees" db:"fees"`
	GasUsed            uint64    `json:"gas_used" db:"gas_used"`
	GasPrice           string    `json:"gas_price" db:"gas_price"`
	ProcessedAt        time.Time `json:"processed_at" db:"processed_at"`
	KafkaPublished     bool      `json:"kafka_published" db:"kafka_published"`
}

type FailedTransaction struct {
	ID              uint64     `json:"id" db:"id"`
	TransactionHash string     `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber     uint64     `json:"block_number" db:"block_number"`
	ErrorMessage    string     `json:"error_message" db:"error_message"`
	RetryCount      int        `json:"retry_count" db:"retry_count"`
	MaxRetries      int        `json:"max_retries" db:"max_retries"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	LastRetryAt     *time.Time `json:"last_retry_at" db:"last_retry_at"`
	Resolved        bool       `json:"resolved" db:"resolved"`
}

func (te *TransactionEvent) ToJSON() ([]byte, error) {
	return json.Marshal(te)
}

func (te *TransactionEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, te)
}
