package models

import (
	"github.com/ethereum/go-ethereum/common"
)

func IsZeroAddress(addr common.Address) bool {
	return addr == common.HexToAddress("0x0000000000000000000000000000000000000000")
}

type ErrorType string

const (
	ErrorTypeConnection ErrorType = "CONNECTION"
	ErrorTypeProcessing ErrorType = "PROCESSING"
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeNotFound   ErrorType = "NOT_FOUND"
	ErrorTypeTimeout    ErrorType = "TIMEOUT"
)

type ProcessingError struct {
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	Code        string    `json:"code"`
	BlockNumber uint64    `json:"block_number,omitempty"`
	TxHash      string    `json:"tx_hash,omitempty"`
	Retryable   bool      `json:"retryable"`
}

func (e ProcessingError) Error() string {
	return e.Message
}
