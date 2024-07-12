package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"log"
	"math/big"
	"math/rand"
	"os"
	"puffDep/config"
	"puffDep/delayer"
	"puffDep/formatter"
	"puffDep/karak"
	"puffDep/puff"
	"time"
)

func loadConfig() (*config.Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file, %s", err)
	}

	var config config.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("Unable to decode into struct, %v", err)
	}

	return &config, nil
}

func readKeysFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keys []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func getRandomAmount(balance *big.Int, minPercent int, maxPercent int) *big.Int {
	rand.Seed(time.Now().UnixNano())
	percent := rand.Intn(maxPercent-minPercent) + minPercent
	percentFloat := float64(percent) / 100.0
	amountFloat := new(big.Float).Mul(new(big.Float).SetInt(balance), big.NewFloat(percentFloat))
	amount := new(big.Int)
	amountFloat.Int(amount)
	return amount
}

var successText = color.New(color.FgGreen).SprintfFunc()
var greenText = color.New(color.FgGreen)
var warningText = color.New(color.FgYellow)
var errorText = color.New(color.FgRed)
var infoText = color.New(color.FgBlue)

var (
	successLogger *log.Logger
)

func init() {
	successFile, err := os.OpenFile("success.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open success log file: %v", err)
	}

	successLogger = log.New(successFile, "", log.LstdFlags)

	log.SetOutput(os.Stdout)
}
func main() {

	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
	}

	fmt.Printf("App Name: %s\n", config.App.Name)
	fmt.Printf("App Version: %s\n", config.App.Version)
	fmt.Printf("Rpc Provider: %s\n", config.Ethereum.Rpc)
	fmt.Printf("Delays between wallets (Seconds) Min:%d / Max:%d\n", config.Ethereum.Delays.Wallet.Min, config.Ethereum.Delays.Wallet.Max)
	fmt.Printf("Delays between blocks (Seconds) Min:%d / Max:%d\n", config.Ethereum.Delays.Block.Min, config.Ethereum.Delays.Block.Max)
	fmt.Printf("Work Amount Range (Percent) Min:%d / Max:%d\n", config.Ethereum.Workflow.WorkAmountRangePercent.Min, config.Ethereum.Workflow.WorkAmountRangePercent.Max)
	fmt.Printf("Gas Limit (Gwei): %d\n", config.Ethereum.Workflow.GweiLimit)

	client, err := ethclient.Dial(config.Ethereum.Rpc)
	if err != nil {
		log.Printf("Failed to connect to the Ethereum client: %v", err)
	}
	keys, err := readKeysFromFile("keys.txt")
	if err != nil {
		log.Printf("Error reading keys from file: %v", err)
	}

	//! Main Loop
	for _, key := range keys {

		hex256Pk := formatter.PrivateKeyToHex(key)

		privateKeyECDSA, err := crypto.HexToECDSA(hex256Pk)
		if err != nil {
			log.Printf("Failed to parse private key: %v", err)
		}
		fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

		warningText.Printf("Working with address: %s\n", fromAddress.Hex())

		//! Eth Balance
		balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
		if err != nil {
			log.Printf("Failed to get balance: %v", err)
			continue
		}

		//! Generating random amount of Eth for deposit to puffEth
		amount := getRandomAmount(balance, config.Ethereum.Workflow.WorkAmountRangePercent.Min, config.Ethereum.Workflow.WorkAmountRangePercent.Max)

		ethAmount := formatter.ConvertWeiToEther(amount)
		ethBalance := formatter.ConvertWeiToEther(balance)
		warningText.Printf("Randomed value to Deposit:%f / Eth Balance: %f\n", ethAmount, ethBalance)

		//! Main Dep function
		infoText.Printf("Depositing %f ETH to PuffEth\n", ethAmount)
		res := puff.DepositEth(client, privateKeyECDSA, ethAmount, config)
		successLogger.Println(successText("Successful deposit: %s\n", res))
		greenText.Printf("Successful deposit: %s\n", res)

		//! Delay Blocks
		delayer.DelayBlock(config)

		//! Get PuffEth Balance
		puffEthBalance, err := puff.GetPuffEthBalance(client, fromAddress)
		successLogger.Println(successText("puffEth Balance: %f\n", formatter.ConvertWeiToEther(puffEthBalance)))

		//! Approve PuffEth
		infoText.Printf("Approving %f PuffEth\n", formatter.ConvertWeiToEther(puffEthBalance))
		approveResponse := puff.ApprovePuffEth(client, privateKeyECDSA, puffEthBalance, "0x68754d29f2e97B837Cb622ccfF325adAC27E9977")
		successLogger.Println(successText("Successful approve: %s\n", approveResponse))
		greenText.Printf("Successful approve: %s\n", approveResponse)

		//! Delay Blocks
		delayer.DelayBlock(config)

		//! Deposit puffEth to Karak
		infoText.Printf("Depositing %f PuffEth to Karak\n", formatter.ConvertWeiToEther(puffEthBalance))
		karakDepositResponse := karak.DepositToKarak(client, privateKeyECDSA, puffEthBalance, config)
		successLogger.Println(successText("Successful deposit to Karak: %s\n", karakDepositResponse))
		greenText.Printf("Successful deposit to Karak: %s\n", karakDepositResponse)

		//! Delay Wallets
		delayer.DelayWallet(config)
	}
}
