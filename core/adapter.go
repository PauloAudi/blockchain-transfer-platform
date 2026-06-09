package core

// TransferRequest holds the data required for any transfer,
// regardless of the target blockchain.
type TransferRequest struct {
	PrivateKey string // hex-encoded private key, without 0x prefix
	To         string // destination address
	Amount     string // amount in the network's main unit (e.g., "0.001" for 0.001 ETH)
}

// TransferResult is the standardized response of a transfer operation.
type TransferResult struct {
	TxHash string
	URL    string // link to the network's block explorer
}

// BlockchainAdapter is the interface every blockchain must implement.
// Adding a new network means creating a new adapter that satisfies this interface.
type BlockchainAdapter interface {
	Transfer(req TransferRequest) (TransferResult, error)
}
