package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"tcc-eth/core"
)

type Adapter struct {
	rpcURL      string
	chainID     *big.Int
	explorerURL string
}

func NewAdapter(rpcURL string, chainID int64, explorerURL string) *Adapter {
	return &Adapter{
		rpcURL:      rpcURL,
		chainID:     big.NewInt(chainID),
		explorerURL: explorerURL,
	}
}

func (a *Adapter) Transfer(req core.TransferRequest) (core.TransferResult, error) {
	client, err := ethclient.Dial(a.rpcURL)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(req.PrivateKey)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid private key: %w", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	value, err := etherToWei(req.Amount)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid amount: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	toAddress := common.HexToAddress(req.To)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      21000,
		GasPrice: gasPrice,
	})

	signer := types.NewEIP155Signer(a.chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	if err := client.SendTransaction(context.Background(), signedTx); err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	hash := signedTx.Hash().Hex()
	return core.TransferResult{
		TxHash: hash,
		URL:    fmt.Sprintf("%s/tx/%s", a.explorerURL, hash),
	}, nil
}

// etherToWei converts a human-readable ETH amount (e.g., "0.001") to wei.
func etherToWei(amount string) (*big.Int, error) {
	f, _, err := big.ParseFloat(amount, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, err
	}
	weiPerEther := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	f.Mul(f, weiPerEther)
	result, _ := f.Int(nil)
	return result, nil
}
