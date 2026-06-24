package sol

import (
	"context"
	"fmt"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"

	"tcc-eth/core"
)

type Adapter struct {
	rpcURL      string
	explorerURL string
}

func NewAdapter(rpcURL, explorerURL string) *Adapter {
	return &Adapter{rpcURL: rpcURL, explorerURL: explorerURL}
}

func (a *Adapter) Transfer(req core.TransferRequest) (core.TransferResult, error) {
	client := rpc.New(a.rpcURL)

	privateKey, err := solana.PrivateKeyFromBase58(req.PrivateKey)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid private key: %w", err)
	}

	toAddr, err := solana.PublicKeyFromBase58(req.To)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid destination address: %w", err)
	}

	lamports, err := solToLamports(req.Amount)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid amount: %w", err)
	}

	recent, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to get blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(lamports, privateKey.PublicKey(), toAddr).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(privateKey.PublicKey()),
	)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to build transaction: %w", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(privateKey.PublicKey()) {
			return &privateKey
		}
		return nil
	})
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := client.SendTransactionWithOpts(context.Background(), tx,
		rpc.TransactionOpts{SkipPreflight: false},
	)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return core.TransferResult{
		TxHash: sig.String(),
		URL:    fmt.Sprintf("%s/tx/%s?cluster=devnet", a.explorerURL, sig.String()),
	}, nil
}

// solToLamports converts a human-readable SOL amount (e.g., "0.001") to lamports.
func solToLamports(amount string) (uint64, error) {
	f, _, err := big.ParseFloat(amount, 10, 256, big.ToNearestEven)
	if err != nil {
		return 0, err
	}
	lamportsPerSol := new(big.Float).SetFloat64(1e9)
	f.Mul(f, lamportsPerSol)
	result, _ := f.Int(nil)
	return result.Uint64(), nil
}
