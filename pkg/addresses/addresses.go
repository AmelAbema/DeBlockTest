package addresses

import (
	"DeBlockTest/internal/models"
	"DeBlockTest/pkg/storage/postgres"
	"DeBlockTest/pkg/storage/redis"
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

const (
	addressCacheKeyPrefix = "addr:"
	addressCacheTTL       = 24 * time.Hour
)

type AddressModule struct {
	db    *postgres.Client
	cache *redis.Client

	addressMap map[common.Address]string
	mu         sync.RWMutex
}

func NewAddressModule(
	ctx context.Context,
	db *postgres.Client,
	cache *redis.Client,
) (*AddressModule, error) {
	mod := &AddressModule{
		db:         db,
		cache:      cache,
		addressMap: make(map[common.Address]string),
	}

	if err := mod.LoadAddresses(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to load addresses during initialization")
	}

	return mod, nil
}

func (m *AddressModule) LoadAddresses(ctx context.Context) error {
	tel.Global().Info("loading monitored addresses from database")

	addresses, err := m.loadAddressesFromDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load addresses from database")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.addressMap = make(map[common.Address]string, len(addresses))

	for _, addr := range addresses {
		m.addressMap[addr.Address] = addr.UserID

		if err := m.setAddressCache(ctx, addr.Address.Hex(), addr.UserID); err != nil {
			tel.Global().Error("failed to cache address",
				tel.Error(err),
				tel.String("address", addr.Address.Hex()))
		}
	}

	tel.Global().Info("addresses loaded successfully",
		tel.Int("count", len(addresses)))

	return nil
}

func (m *AddressModule) loadAddressesFromDB(ctx context.Context) ([]*models.UserAddress, error) {
	query := `
		SELECT id, user_id, address, is_active, created_at, updated_at
		FROM monitored_addresses 
		WHERE is_active = true
		ORDER BY user_id
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query addresses")
	}
	defer rows.Close()

	var addresses []*models.UserAddress
	for rows.Next() {
		var addr models.UserAddress
		var addressStr string

		if err := rows.Scan(&addr.ID, &addr.UserID, &addressStr, &addr.IsActive, &addr.CreatedAt, &addr.UpdatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan address row")
		}

		addr.Address = common.HexToAddress(addressStr)
		addresses = append(addresses, &addr)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating address rows")
	}

	return addresses, nil
}

func (m *AddressModule) setAddressCache(ctx context.Context, address string, userID string) error {
	key := addressCacheKeyPrefix + address
	return m.cache.SetString(ctx, key, userID, addressCacheTTL)
}

func (m *AddressModule) getAddressCache(ctx context.Context, address string) (string, error) {
	key := addressCacheKeyPrefix + address
	userID, err := m.cache.GetString(ctx, key)
	if err != nil {
		if errors.Is(err, redis.ErrNotFound) {
			return "", nil
		}
		return "", errors.Wrap(err, "failed to get address from cache")
	}
	return userID, nil
}

func (m *AddressModule) IsMonitoredAddress(ctx context.Context, address common.Address) (*models.AddressMatchResult, error) {
	m.mu.RLock()
	userID, exists := m.addressMap[address]
	m.mu.RUnlock()

	if exists {
		return &models.AddressMatchResult{
			IsMatch: true,
			UserID:  userID,
			Address: address,
		}, nil
	}

	cachedUserID, err := m.getAddressCache(ctx, address.Hex())
	if err != nil {
		tel.Global().Error("failed to check address cache",
			tel.Error(err),
			tel.String("address", address.Hex()))
	}

	if cachedUserID != "" {
		m.mu.Lock()
		m.addressMap[address] = cachedUserID
		m.mu.Unlock()

		return &models.AddressMatchResult{
			IsMatch: true,
			UserID:  cachedUserID,
			Address: address,
		}, nil
	}

	return &models.AddressMatchResult{
		IsMatch: false,
		Address: address,
	}, nil
}

func (m *AddressModule) CheckTransactionAddresses(ctx context.Context, from, to common.Address) ([]*models.AddressMatchResult, error) {
	var results []*models.AddressMatchResult

	if !models.IsZeroAddress(from) {
		fromResult, err := m.IsMonitoredAddress(ctx, from)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check source address")
		}
		if fromResult.IsMatch {
			fromResult.IsSource = true
			results = append(results, fromResult)
		}
	}

	if !models.IsZeroAddress(to) {
		toResult, err := m.IsMonitoredAddress(ctx, to)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check destination address")
		}
		if toResult.IsMatch {
			toResult.IsDestination = true
			results = append(results, toResult)
		}
	}

	return results, nil
}

func (m *AddressModule) GetAddressCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.addressMap)
}

func (m *AddressModule) ReloadAddresses(ctx context.Context) error {
	tel.Global().Info("reloading addresses from database")
	return m.LoadAddresses(ctx)
}
