package processing

import (
	"DeBlockTest/pkg/storage/postgres"
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ProcessingModule struct {
	db         *postgres.Client
	instanceID string
}

func NewProcessingModule(
	ctx context.Context,
	db *postgres.Client,
	instanceID string,
) (*ProcessingModule, error) {
	return &ProcessingModule{
		db:         db,
		instanceID: instanceID,
	}, nil
}

func (m *ProcessingModule) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	query := `
		SELECT last_processed_block 
		FROM processing_state 
		WHERE instance_id = $1
	`

	var blockNumber uint64
	err := m.db.QueryRow(ctx, query, m.instanceID).Scan(&blockNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// First time running, return 0
			return 0, nil
		}
		return 0, errors.Wrap(err, "failed to get last processed block")
	}

	return blockNumber, nil
}

func (m *ProcessingModule) SetLastProcessedBlock(ctx context.Context, blockNumber uint64) error {
	query := `
		INSERT INTO processing_state (instance_id, last_processed_block, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (instance_id) 
		DO UPDATE SET 
			last_processed_block = EXCLUDED.last_processed_block,
			updated_at = NOW()
	`

	err := m.db.Exec(ctx, query, m.instanceID, blockNumber)
	if err != nil {
		return errors.Wrap(err, "failed to set last processed block")
	}

	return nil
}

func (m *ProcessingModule) UpdateLastProcessedBlock(ctx context.Context, blockNumber uint64) error {
	return m.SetLastProcessedBlock(ctx, blockNumber)
}
