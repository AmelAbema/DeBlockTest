package monitoring

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestExtractTransactionAmount_ETHTransfer(t *testing.T) {
	module := &MonitoringModule{}

	tx := createETHTransferTransaction(t, big.NewInt(1000000000000000000))

	amount := module.extractTransactionAmount(tx)
	assert.Equal(t, big.NewInt(1000000000000000000), amount)
}

func TestExtractTransactionAmount_ERC20Transfer(t *testing.T) {
	module := &MonitoringModule{}

	transferAmount := big.NewInt(1000000)
	tx := createERC20TransferTransaction(t, transferAmount)

	amount := module.extractTransactionAmount(tx)
	assert.Equal(t, transferAmount, amount)
}

func TestCalculateTransactionFees(t *testing.T) {
	module := &MonitoringModule{}

	tx := createTestTransaction(t)
	receipt := &types.Receipt{
		GasUsed: 21000,
	}

	fees := module.calculateTransactionFees(tx, receipt)

	expectedFees := big.NewInt(420000000000000)
	assert.Equal(t, expectedFees, fees)
}

func createTestTransaction(t *testing.T) *types.Transaction {
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		To:       &to,
		Value:    big.NewInt(1000000000000000000),
		Gas:      21000,
		GasPrice: big.NewInt(20000000000),
		Data:     []byte{},
	})

	return tx
}

func createETHTransferTransaction(t *testing.T, amount *big.Int) *types.Transaction {
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		To:       &to,
		Value:    amount,
		Gas:      21000,
		GasPrice: big.NewInt(20000000000),
		Data:     []byte{},
	})

	return tx
}

func createERC20TransferTransaction(t *testing.T, amount *big.Int) *types.Transaction {
	to := common.HexToAddress("0xA0b86a33E6441E2B7c66C52C4C8F8f7E7b5c1234")

	data := make([]byte, 68)

	methodSig := []byte{0xa9, 0x05, 0x9c, 0xbb}
	copy(data[0:4], methodSig)

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
	copy(data[4:36], common.LeftPadBytes(recipient.Bytes(), 32))

	copy(data[36:68], common.LeftPadBytes(amount.Bytes(), 32))

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      60000,
		GasPrice: big.NewInt(20000000000),
		Data:     data,
	})

	return tx
}
