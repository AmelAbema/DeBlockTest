package monitoring

import (
	"DeBlockTest/internal/models"
	"DeBlockTest/pkg/addresses"
	"DeBlockTest/pkg/processing"
	"DeBlockTest/pkg/transport"
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

type MonitoringModule struct {
	transport  *transport.TransportModule
	addresses  *addresses.AddressModule
	processing *processing.ProcessingModule
	instanceID string
}

func NewMonitoringModule(
	transport *transport.TransportModule,
	addresses *addresses.AddressModule,
	processing *processing.ProcessingModule,
	instanceID string,
) *MonitoringModule {
	return &MonitoringModule{
		transport:  transport,
		addresses:  addresses,
		processing: processing,
		instanceID: instanceID,
	}
}

func (m *MonitoringModule) StartMonitoring(ctx context.Context) error {
	tel.Global().Info("starting blockchain monitoring", tel.String("instance_id", m.instanceID))

	startBlock, err := m.getStartingBlock(ctx)
	if err != nil {
		return err
	}

	if err := m.processHistoricalBlocks(ctx, startBlock); err != nil {
		return err
	}

	return m.startRealTimeMonitoring(ctx)
}

func (m *MonitoringModule) getStartingBlock(ctx context.Context) (uint64, error) {
	lastBlock, err := m.processing.GetLastProcessedBlock(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get last processed block")
	}

	currentBlock, err := m.transport.GetEthereumClient().GetLatestBlockNumber(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current block number")
	}

	tel.Global().Info("monitoring status",
		tel.Uint64("last_processed", lastBlock),
		tel.Uint64("current_block", currentBlock),
		tel.Int("monitored_addresses", m.addresses.GetAddressCount()))

	if lastBlock == 0 {
		return currentBlock, nil
	}
	return lastBlock + 1, nil
}

func (m *MonitoringModule) processHistoricalBlocks(ctx context.Context, startBlock uint64) error {
	currentBlock, err := m.transport.GetEthereumClient().GetLatestBlockNumber(ctx)
	if err != nil {
		return err
	}

	for blockNum := startBlock; blockNum <= currentBlock; blockNum++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.processBlock(ctx, blockNum); err != nil {
				tel.Global().Error("block processing failed",
					tel.Error(err), tel.Uint64("block", blockNum))
				continue
			}
		}
	}
	return nil
}

func (m *MonitoringModule) startRealTimeMonitoring(ctx context.Context) error {
	headerChan, err := m.transport.GetEthereumClient().SubscribeNewHead(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to new blocks")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case header := <-headerChan:
			if header == nil {
				continue
			}

			blockNumber := header.Number.Uint64()
			if err := m.processBlock(ctx, blockNumber); err != nil {
				tel.Global().Error("real-time block processing failed",
					tel.Error(err), tel.Uint64("block", blockNumber))
			}
		}
	}
}

func (m *MonitoringModule) processBlock(ctx context.Context, blockNumber uint64) error {
	ethClient := m.transport.GetEthereumClient()

	block, err := ethClient.GetBlockByNumber(ctx, blockNumber)
	if err != nil {
		return errors.Wrap(err, "failed to get block")
	}

	tel.Global().Debug("processing block",
		tel.Uint64("block_number", blockNumber),
		tel.Int("transaction_count", len(block.Transactions())))

	for _, tx := range block.Transactions() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.processTransaction(ctx, tx, block); err != nil {
				tel.Global().Error("failed to process transaction",
					tel.Error(err),
					tel.String("tx_hash", tx.Hash().Hex()))
			}
		}
	}

	if err := m.processing.UpdateLastProcessedBlock(ctx, blockNumber); err != nil {
		return errors.Wrap(err, "failed to update last processed block")
	}

	return nil
}

func (m *MonitoringModule) processTransaction(ctx context.Context, tx *types.Transaction, block *types.Block) error {
	from, to, err := m.extractTransactionAddresses(tx)
	if err != nil {
		return err
	}

	matches, err := m.addresses.CheckTransactionAddresses(ctx, from, to)
	if err != nil {
		return errors.Wrap(err, "address check failed")
	}
	if len(matches) == 0 {
		return nil // No monitored addresses involved
	}

	receipt, err := m.getTransactionReceipt(ctx, tx)
	if err != nil {
		return err // Skip if can't get receipt
	}

	return m.publishTransactionEvents(ctx, tx, block, receipt, matches, from, to)
}

func (m *MonitoringModule) extractTransactionAddresses(tx *types.Transaction) (from, to common.Address, err error) {
	if tx.To() != nil {
		to = *tx.To()
	}

	signer := types.LatestSignerForChainID(m.transport.GetEthereumClient().GetChainID())
	from, err = types.Sender(signer, tx)
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "failed to get sender")
	}
	return from, to, nil
}

func (m *MonitoringModule) getTransactionReceipt(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := m.transport.GetEthereumClient().GetTransactionReceipt(ctx, tx.Hash())
	if err != nil {
		tel.Global().Warn("receipt unavailable, skipping transaction",
			tel.Error(err), tel.String("tx_hash", tx.Hash().Hex()))
		return nil, err
	}
	return receipt, nil
}

func (m *MonitoringModule) publishTransactionEvents(ctx context.Context, tx *types.Transaction, block *types.Block, receipt *types.Receipt, matches []*models.AddressMatchResult, from, to common.Address) error {
	for _, match := range matches {
		event := &models.TransactionEvent{
			TransactionHash: tx.Hash().Hex(),
			BlockNumber:     block.Number().Uint64(),
			BlockHash:       block.Hash().Hex(),
			UserID:          match.UserID,
			Source:          from.Hex(),
			Destination:     to.Hex(),
			Amount:          m.extractTransactionAmount(tx).String(),
			Fees:            m.calculateTransactionFees(tx, receipt).String(),
			GasUsed:         receipt.GasUsed,
			GasPrice:        tx.GasPrice().String(),
			Timestamp:       time.Unix(int64(block.Time()), 0),
			Status:          receipt.Status,
			Nonce:           tx.Nonce(),
		}

		if err := m.transport.PublishTransaction(ctx, event); err != nil {
			tel.Global().Error("event publish failed",
				tel.Error(err), tel.String("tx_hash", tx.Hash().Hex()), tel.String("user_id", match.UserID))
			continue
		}

		tel.Global().Info("transaction processed",
			tel.String("tx_hash", tx.Hash().Hex()),
			tel.String("user_id", match.UserID),
			tel.String("amount", event.Amount))
	}
	return nil
}

func (m *MonitoringModule) extractTransactionAmount(tx *types.Transaction) *big.Int {
	if len(tx.Data()) == 0 {
		return tx.Value()
	}

	if len(tx.Data()) >= 68 && len(tx.Data()[:4]) == 4 {
		methodSig := common.Bytes2Hex(tx.Data()[:4])
		if methodSig == "a9059cbb" {
			amountBytes := tx.Data()[36:68]
			amount := new(big.Int).SetBytes(amountBytes)

			tel.Global().Debug("extracted ERC-20 transfer amount",
				tel.String("tx_hash", tx.Hash().Hex()),
				tel.String("amount", amount.String()),
				tel.String("method_sig", methodSig))

			return amount
		}
	}

	return tx.Value()
}

func (m *MonitoringModule) calculateTransactionFees(tx *types.Transaction, receipt *types.Receipt) *big.Int {
	gasUsed := big.NewInt(int64(receipt.GasUsed))
	gasPrice := tx.GasPrice()

	fees := new(big.Int)
	fees.Mul(gasUsed, gasPrice)

	return fees
}
