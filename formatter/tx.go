package formatter

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

// WaitForTransactionReceipt waits for the transaction to be mined and confirmed
func WaitForTransactionReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	ctx := context.Background()
	for {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == ethereum.NotFound {
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			return nil, err
		}
		return receipt, nil
	}
}
