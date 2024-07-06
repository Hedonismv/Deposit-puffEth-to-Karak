package delayer

import (
	"github.com/fatih/color"
	"math/rand"
	"puffDep/config"
	"time"
)

var warningText = color.New(color.FgYellow)

func DelayBlock(config *config.Config) {
	blockDelay := rand.Intn(config.Ethereum.Delays.Block.Max-config.Ethereum.Delays.Block.Min) + config.Ethereum.Delays.Block.Min
	warningText.Printf("[Block] Waiting for %d seconds\n", blockDelay)
	time.Sleep(time.Duration(blockDelay) * time.Second)
}

func DelayWallet(config *config.Config) {
	walletDelay := rand.Intn(config.Ethereum.Delays.Wallet.Max-config.Ethereum.Delays.Wallet.Min) + config.Ethereum.Delays.Wallet.Min
	warningText.Printf("[Wallet] Waiting for %d seconds\n", walletDelay)
	time.Sleep(time.Duration(walletDelay) * time.Second)
}
