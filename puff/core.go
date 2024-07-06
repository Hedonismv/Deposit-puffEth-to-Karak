package puff

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"log"
	"math/big"
	"puffDep/config"
	"puffDep/formatter"
	"strings"
)

var EthPuffTokenContractAddress = "0xD9A442856C234a39a81a089C06451EBAa4306a72"
var contractABI = `[{"inputs":[{"internalType":"contract IStETH","name":"stETH","type":"address"},{"internalType":"contract IWETH","name":"weth","type":"address"},{"internalType":"contract ILidoWithdrawalQueue","name":"lidoWithdrawalQueue","type":"address"},{"internalType":"contract IStrategy","name":"stETHStrategy","type":"address"},{"internalType":"contract IEigenLayer","name":"eigenStrategyManager","type":"address"},{"internalType":"contract IPufferOracle","name":"oracle","type":"address"},{"internalType":"contract IDelegationManager","name":"delegationManager","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"receiver","type":"address"}],"name":"depositETH","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
var InfoText = color.New(color.FgBlue)

func DepositEth(provider *ethclient.Client, privateKeyECDSA *ecdsa.PrivateKey, amountInEth float64) string {

	ctx := context.Background()

	contractAddress := common.HexToAddress(EthPuffTokenContractAddress)

	fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	//! Parsed ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	//! Get the nonce
	nonce, err := provider.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	//! Gas Price
	gasPrice, err := provider.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	// ! Chain ID
	chainID, err := provider.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	// ! Calldata
	callData, err := parsedABI.Pack("depositETH", fromAddress)
	if err != nil {
		log.Fatalf("Failed to pack function input: %v", err)
	}

	//! Convert ETH amount to Wei
	valueInWei := formatter.ConvertEtherToWei(amountInEth)

	formatter.CheckGasPrice(provider, config.Config{})

	// ! GasLimit
	gasLimit, err := provider.EstimateGas(context.Background(), ethereum.CallMsg{
		To:    &contractAddress,
		Data:  callData,
		Value: valueInWei,
	})
	if err != nil {
		log.Fatalf("Failed to estimate gas: %v", err)
	}

	//! Get wallet balance
	walletBalance, err := provider.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Fatalf("Failed to get wallet balance: %v", err)
	}

	//! Check if the wallet has enough balance
	transactionPrice, hasEnoughBalance := formatter.GetTransactionCost(gasLimit, gasPrice, valueInWei, walletBalance)
	if !hasEnoughBalance {
		log.Fatalf("Insufficient balance. Transaction cost: %v", transactionPrice)
	}

	//!Data for function
	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, chainID)
	if err != nil {
		log.Fatalf("Failed to create keyed transactor %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = valueInWei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	//! Construct the transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasPrice,
		GasFeeCap: gasPrice,
		Gas:       auth.GasLimit,
		To:        &contractAddress,
		Value:     auth.Value,
		Data:      callData,
	})

	//! Sign the transaction
	signedTx, err := auth.Signer(fromAddress, tx)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	//! Send the transaction
	err = provider.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	InfoText.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

	receipt, err := formatter.WaitForTransactionReceipt(provider, signedTx.Hash())
	if err != nil {
		log.Fatalf("Failed to get transaction receipt: %v", err)
	}

	fmt.Printf("Transaction confirmed in block: %d\n", receipt.BlockNumber.Uint64())

	return "https://etherscan.io/tx/" + signedTx.Hash().Hex()
}

func ApprovePuffEth(provider *ethclient.Client, privateKeyECDSA *ecdsa.PrivateKey, amountPuffEth *big.Int, spender string) string {

	ctx := context.Background()

	fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	contractAddress := common.HexToAddress(EthPuffTokenContractAddress)

	//! Get the nonce
	nonce, err := provider.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	//! Gas Price
	gasPrice, err := provider.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	// ! Chain ID
	chainID, err := provider.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	//! Parsed ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// ! Calldata
	callData, err := parsedABI.Pack("approve", common.HexToAddress(spender), amountPuffEth)
	if err != nil {
		log.Fatalf("Failed to pack function input: %v", err)
	}

	// ! GasLimit
	gasLimit, err := provider.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		From: fromAddress,
		Data: callData,
	})
	if err != nil {
		log.Fatalf("Failed to estimate gas: %v", err)
	}

	//!Data for function
	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, chainID)
	if err != nil {
		log.Fatalf("Failed to create keyed transactor %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	//! Construct the transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasPrice,
		GasFeeCap: gasPrice,
		Gas:       auth.GasLimit,
		To:        &contractAddress,
		Value:     auth.Value,
		Data:      callData,
	})

	//! Sign the transaction
	signedTx, err := auth.Signer(fromAddress, tx)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	//! Send the transaction
	err = provider.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	InfoText.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

	receipt, err := formatter.WaitForTransactionReceipt(provider, signedTx.Hash())
	if err != nil {
		log.Fatalf("Failed to get transaction receipt: %v", err)
	}

	fmt.Printf("Transaction confirmed in block: %d\n", receipt.BlockNumber.Uint64())

	return "https://etherscan.io/tx/" + signedTx.Hash().Hex()

}

func GetPuffEthBalance(provider *ethclient.Client, address common.Address) (*big.Int, error) {
	contractAddress := common.HexToAddress(EthPuffTokenContractAddress)
	//! Parsed ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	callData, err := parsedABI.Pack("balanceOf", address)

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: callData,
	}

	result, err := provider.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
	}

	var balance *big.Int
	err = parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		log.Fatalf("Failed to unpack result: %v", err)
	}
	return balance, nil
}
