package main

import (
	"fmt"
	"log"
	"os"

	"tcc-eth/chains/btc"
	"tcc-eth/chains/eth"
	"tcc-eth/chains/sol"
	"tcc-eth/core"
)

func main() {
	chain := os.Getenv("CHAIN")
	if chain == "" {
		log.Fatal("CHAIN environment variable not set (use 'eth' or 'sol')")
	}
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY environment variable not set")
	}
	toAddress := os.Getenv("TO_ADDRESS")
	if toAddress == "" {
		log.Fatal("TO_ADDRESS environment variable not set")
	}

	var adapter core.BlockchainAdapter

	switch chain {
	case "eth":
		adapter = eth.NewAdapter(
			"https://ethereum-sepolia-rpc.publicnode.com",
			11155111, // Sepolia chain ID
			"https://sepolia.etherscan.io",
		)
	case "sol":
		adapter = sol.NewAdapter(
			"https://api.devnet.solana.com",
			"https://explorer.solana.com",
		)
	case "btc":
		adapter = btc.NewAdapter(
			"https://mempool.space/testnet/api",
			"https://mempool.space/testnet/tx",
			true, // testnet
		)
	default:
		log.Fatalf("unsupported chain: %q (use 'eth' or 'sol')", chain)
	}

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
