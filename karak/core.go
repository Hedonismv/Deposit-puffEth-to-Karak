package karak

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

var InfoText = color.New(color.FgBlue)
var karakVaultContract = "0x54e44DbB92dBA848ACe27F44c0CB4268981eF1CC"

var karakVaultAddress = "0x68754d29f2e97B837Cb622ccfF325adAC27E9977"

var karakABI = `[{"inputs":[{"internalType":"contract IVault","name":"vault","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"minSharesOut","type":"uint256"}],"name":"deposit","outputs":[{"internalType":"uint256","name":"shares","type":"uint256"}],"stateMutability":"nonpayable","type":"function"}]`

func DepositToKarak(provider *ethclient.Client, privateKeyECDSA *ecdsa.PrivateKey, amountPuffEth *big.Int) string {
	ctx := context.Background()

	contractAddress := common.HexToAddress(karakVaultContract)

	fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	//! Parsed ABI
	parsedABI, err := abi.JSON(strings.NewReader(karakABI))
	if err != nil {
		log.Printf("Failed to parse contract ABI: %v", err)
	}

	//! Get the nonce
	nonce, err := provider.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Printf("Failed to get nonce: %v", err)
	}

	//! Gas Price
	gasPrice, err := provider.SuggestGasPrice(context.Background())
	if err != nil {
		log.Printf("Failed to get gas price: %v", err)
	}

	// ! Chain ID
	chainID, err := provider.ChainID(context.Background())
	if err != nil {
		log.Printf("Failed to get chain ID: %v", err)
	}

	minShareOut := formatter.CalculateSlippage(amountPuffEth) // ! 1% slippage

	// ! Calldata
	callData, err := parsedABI.Pack("deposit", common.HexToAddress(karakVaultAddress), amountPuffEth, minShareOut)
	if err != nil {
		log.Printf("Failed to pack function input: %v", err)
	}

	formatter.CheckGasPrice(provider, config.Config{})

	// ! GasLimit
	gasLimit, err := provider.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &contractAddress,
		Data: callData,
	})
	if err != nil {
		log.Printf("Failed to estimate gas: %v", err)
	}

	//! Get wallet balance
	walletBalance, err := provider.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Printf("Failed to get wallet balance: %v", err)
	}

	//! Check if the wallet has enough balance
	transactionPrice, hasEnoughBalance := formatter.GetTransactionCost(gasLimit, gasPrice, big.NewInt(0), walletBalance)
	if !hasEnoughBalance {
		log.Printf("Insufficient balance. Transaction cost: %v", transactionPrice)
	}

	//!Data for function
	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, chainID)
	if err != nil {
		log.Printf("Failed to create keyed transactor %v", err)
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
		log.Printf("Failed to sign transaction: %v", err)
	}

	//! Send the transaction
	err = provider.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Printf("Failed to send transaction: %v", err)
	}

	InfoText.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

	receipt, err := formatter.WaitForTransactionReceipt(provider, signedTx.Hash())
	if err != nil {
		log.Printf("Failed to get transaction receipt: %v", err)
	}

	fmt.Printf("Transaction confirmed in block: %d\n", receipt.BlockNumber.Uint64())

	return "https://etherscan.io/tx/" + signedTx.Hash().Hex()

}
