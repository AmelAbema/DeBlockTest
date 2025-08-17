package models

import (
	"time"
)

type ProcessedBlock struct {
	BlockNumber  uint64    `json:"block_number"`
	BlockHash    string    `json:"block_hash"`
	TxCount      int       `json:"tx_count"`
	MatchedTxs   int       `json:"matched_txs"`
	ProcessedAt  time.Time `json:"processed_at"`
	ProcessingMs int64     `json:"processing_ms"`
}

type ProcessingState struct {
	InstanceID         string    `json:"instance_id" db:"instance_id"`
	LastProcessedBlock uint64    `json:"last_processed_block" db:"last_processed_block"`
	StatsData          []byte    `json:"stats_data" db:"stats_data"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}
