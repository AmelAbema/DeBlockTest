package models

import (
	"github.com/ethereum/go-ethereum/common"
	"time"
)

type UserAddress struct {
	ID        uint64         `json:"id" db:"id"`
	UserID    string         `json:"user_id" db:"user_id"`
	Address   common.Address `json:"address" db:"address"`
	IsActive  bool           `json:"is_active" db:"is_active"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

type AddressMatchResult struct {
	IsMatch       bool           `json:"is_match"`
	UserID        string         `json:"user_id"`
	Address       common.Address `json:"address"`
	IsSource      bool           `json:"is_source"`
	IsDestination bool           `json:"is_destination"`
}

type AddressStats struct {
	LoadedAddresses   int       `json:"loaded_addresses"`
	LastUpdated       time.Time `json:"last_updated"`
	AddressesInMemory int       `json:"addresses_in_memory"`
}
