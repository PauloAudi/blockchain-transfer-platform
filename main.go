package main

import (
	"fmt"
	"log"
	"os"

	"tcc-eth/chains/eth"
	"tcc-eth/core"
)

func main() {
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY environment variable not set")
	}
	toAddress := os.Getenv("TO_ADDRESS")
	if toAddress == "" {
		log.Fatal("TO_ADDRESS environment variable not set")
	}

	// Swapping the adapter here is all it takes to change the target blockchain.
	var adapter core.BlockchainAdapter = eth.NewAdapter(
		"https://ethereum-sepolia-rpc.publicnode.com",
		11155111, // Sepolia chain ID
		"https://sepolia.etherscan.io",
	)

	result, err := adapter.Transfer(core.TransferRequest{
		PrivateKey: privateKey,
		To:         toAddress,
		Amount:     "0.001",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction sent successfully!")
	fmt.Println("Hash:", result.TxHash)
	fmt.Println("Track at:", result.URL)
}
