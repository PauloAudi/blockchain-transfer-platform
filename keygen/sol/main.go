package main

import (
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
)

func main() {
	for _, label := range []string{"Wallet A", "Wallet B"} {
		privateKey, err := solana.NewRandomPrivateKey()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("=== %s ===\n", label)
		fmt.Printf("Address (public):  %s\n", privateKey.PublicKey().String())
		fmt.Printf("Private key:       %s\n\n", privateKey.String())
	}

	fmt.Println("WARNING: store private keys securely. Never share them.")
}
