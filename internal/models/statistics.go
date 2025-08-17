package models

import (
	"time"
)

type ProcessingStats struct {
	TotalBlocks        uint64        `json:"total_blocks"`
	TotalTransactions  uint64        `json:"total_transactions"`
	MatchedTxs         uint64        `json:"matched_transactions"`
	SkippedBlocks      uint64        `json:"skipped_blocks"`
	ErrorCount         uint64        `json:"error_count"`
	LastProcessedBlock uint64        `json:"last_processed_block"`
	StartTime          time.Time     `json:"start_time"`
	Uptime             time.Duration `json:"uptime"`
}
