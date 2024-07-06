package formatter

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"puffDep/config"
	"time"
)

func ConvertWeiToGwei(wei *big.Int) *big.Float {
	gwei := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e9))
	return gwei
}

func ConvertEtherToWei(ethAmount float64) *big.Int {
	value := new(big.Float).Mul(big.NewFloat(ethAmount), big.NewFloat(1e18))
	weiValue := new(big.Int)
	value.Int(weiValue)
	return weiValue
}

// ConvertWeiToEther преобразует значение wei в ETH
func ConvertWeiToEther(weiAmount *big.Int) float64 {
	value := new(big.Float).SetInt(weiAmount)
	ethValue := new(big.Float).Quo(value, big.NewFloat(1e18))
	ethAmount, _ := ethValue.Float64()
	return ethAmount
}
func GetTransactionCost(gasLimit uint64, gasPrice *big.Int, amountToSend *big.Int, balance *big.Int) (*big.Int, bool) {
	gasCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	totalCost := new(big.Int).Add(amountToSend, gasCost)
	hasEnoughBalance := balance.Cmp(totalCost) >= 0
	return totalCost, hasEnoughBalance
}

func CalculateSlippage(valueInWei *big.Int) *big.Int {
	onePercent := new(big.Int).Div(new(big.Int).Mul(valueInWei, big.NewInt(1)), big.NewInt(100))
	minValue := new(big.Int).Sub(valueInWei, onePercent)
	return minValue
}

func CheckGasPrice(client *ethclient.Client, cfg config.Config) {
	limit := big.NewInt(int64(cfg.Ethereum.Workflow.GweiLimit))
	for {
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatalf("Failed to get gas price: %v", err)
		}

		gasPriceGwei := new(big.Int).Div(gasPrice, big.NewInt(1e9))
		fmt.Printf("Current gas price: %s Gwei\n", gasPriceGwei.String())

		if gasPriceGwei.Cmp(limit) <= 0 {
			fmt.Println("Gas price is within the limit, proceeding...")
			break
		}

		fmt.Println("Gas price is too high, waiting...")
		time.Sleep(30 * time.Second)
	}
}
