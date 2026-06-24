package main

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/address/v2"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/v2"
	"github.com/btcsuite/btcd/chaincfg/v2"
)

func main() {
	for _, label := range []string{"Wallet A", "Wallet B"} {
		// Generate a random secp256k1 private key (same curve Bitcoin uses)
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			log.Fatal(err)
		}

		// WIF = Wallet Import Format: a base58-encoded private key that wallets
		// (Electrum, Bitcoin Core, etc.) understand. TestNet3Params marks this
		// key as testnet-only so it can't accidentally be used on mainnet.
		// The third argument (true) means the public key will be compressed (33 bytes).
		wif, err := btcutil.NewWIF(privKey, &chaincfg.TestNet3Params, true)
		if err != nil {
			log.Fatal(err)
		}

		// Hash160 = RIPEMD160(SHA256(pubkey)) — Bitcoin's standard way to
		// shorten a public key into a 20-byte hash that becomes part of the address.
		pubKeyHash := address.Hash160(privKey.PubKey().SerializeCompressed())

		// P2WPKH = Pay-to-Witness-Public-Key-Hash, the modern SegWit address format.
		// On testnet these addresses start with "tb1q".
		addr, err := address.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("=== %s ===\n", label)
		fmt.Printf("Address (public):  %s\n", addr.EncodeAddress())
		fmt.Printf("Private key (WIF): %s\n\n", wif.String())
	}

	fmt.Println("WARNING: store private keys securely. Never share them.")
}
