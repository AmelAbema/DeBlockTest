package transport

import (
	"DeBlockTest/internal/config"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

type EthereumClient struct {
	client  *ethclient.Client
	chainID *big.Int
	config  *config.EthereumConfig
}

func NewEthereumClient(cfg *config.EthereumConfig) (*EthereumClient, error) {
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Ethereum RPC")
	}

	chainID := big.NewInt(cfg.ChainID)

	tel.Global().Info("Ethereum client initialized",
		tel.String("rpc_url", cfg.RPCURL),
		tel.Int64("chain_id", cfg.ChainID))

	return &EthereumClient{
		client:  client,
		chainID: chainID,
		config:  cfg,
	}, nil
}

func (e *EthereumClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := e.client.BlockNumber(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get latest block number")
	}
	return blockNumber, nil
}

func (e *EthereumClient) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	block, err := e.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block by number")
	}
	return block, nil
}

func (e *EthereumClient) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	receipt, err := e.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction receipt")
	}
	return receipt, nil
}

func (e *EthereumClient) SubscribeNewHead(ctx context.Context) (<-chan *types.Header, error) {
	headerChan := make(chan *types.Header)

	sub, err := e.client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return nil, errors.Wrap(err, "failed to subscribe to new heads")
	}

	go func() {
		defer close(headerChan)
		<-sub.Err()
	}()

	return headerChan, nil
}

func (e *EthereumClient) Close() {
	if e.client != nil {
		e.client.Close()
	}
}

func (e *EthereumClient) GetChainID() *big.Int {
	return e.chainID
}

func (e *EthereumClient) GetConfig() *config.EthereumConfig {
	return e.config
}
