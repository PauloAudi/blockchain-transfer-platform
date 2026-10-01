package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	for _, label := range []string{"Wallet A", "Wallet B"} {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatal(err)
		}

		address := crypto.PubkeyToAddress(key.PublicKey)
		privateKeyBytes := crypto.FromECDSA(key)

		fmt.Printf("=== %s ===\n", label)
		fmt.Printf("Address (public):   %s\n", address.Hex())
		fmt.Printf("Private key:        %x\n\n", privateKeyBytes)
	}

	fmt.Println("WARNING: store private keys securely. Never share them.")
}
