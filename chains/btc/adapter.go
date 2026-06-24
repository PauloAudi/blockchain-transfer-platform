package btc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/btcsuite/btcd/address/v2"
	"github.com/btcsuite/btcd/btcutil/v2"
	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/txscript/v2"
	"github.com/btcsuite/btcd/wire/v2"

	"tcc-eth/core"
)

type Adapter struct {
	params      *chaincfg.Params // network rules: mainnet vs testnet (address format, magic bytes, etc.)
	apiURL      string
	explorerURL string
}

func NewAdapter(apiURL, explorerURL string, testnet bool) *Adapter {
	params := &chaincfg.MainNetParams
	if testnet {
		params = &chaincfg.TestNet3Params
	}
	return &Adapter{params: params, apiURL: apiURL, explorerURL: explorerURL}
}

// utxo represents a single Unspent Transaction Output.
// Bitcoin doesn't have "account balances" — your balance is the sum of all UTXOs
// locked to your address. To send BTC you pick UTXOs as inputs and create new outputs.
type utxo struct {
	TxID  string `json:"txid"`
	Vout  uint32 `json:"vout"`  // index of this output inside that transaction
	Value int64  `json:"value"` // amount in satoshis (1 BTC = 100,000,000 sat)
}

// prevOutFetcher implements txscript.PrevOutputFetcher.
// The BIP143 signing algorithm (used by SegWit) needs the value and script of each
// input's *previous* output. This map provides that lookup during signing.
type prevOutFetcher map[wire.OutPoint]*wire.TxOut

func (f prevOutFetcher) FetchPrevOutput(op wire.OutPoint) *wire.TxOut {
	if out, ok := f[op]; ok {
		return out
	}
	return &wire.TxOut{}
}

// getUTXOs queries the mempool.space API for all UTXOs belonging to addr.
func (a *Adapter) getUTXOs(addr string) ([]utxo, error) {
	resp, err := http.Get(fmt.Sprintf("%s/address/%s/utxo", a.apiURL, addr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", body)
	}
	var utxos []utxo
	return utxos, json.NewDecoder(resp.Body).Decode(&utxos)
}

// broadcast sends the raw serialized transaction (hex) to the network.
// Returns the txid on success.
func (a *Adapter) broadcast(txHex string) (string, error) {
	resp, err := http.Post(fmt.Sprintf("%s/tx", a.apiURL), "text/plain", bytes.NewBufferString(txHex))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("broadcast failed: %s", body)
	}
	return string(body), nil
}

func (a *Adapter) Transfer(req core.TransferRequest) (core.TransferResult, error) {
	// WIF = Wallet Import Format, the standard encoding for Bitcoin private keys.
	wif, err := btcutil.DecodeWIF(req.PrivateKey)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid private key (expected WIF format): %w", err)
	}

	// Derive the sender's P2WPKH (SegWit) address from the public key.
	// Hash160 = RIPEMD160(SHA256(pubkey)) — shrinks 33-byte pubkey to 20 bytes.
	pubKeyHash := address.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	senderAddr, err := address.NewAddressWitnessPubKeyHash(pubKeyHash, a.params)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to derive sender address: %w", err)
	}

	toAddr, err := address.DecodeAddress(req.To, a.params)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid destination address: %w", err)
	}

	satoshis, err := btcToSatoshis(req.Amount)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("invalid amount: %w", err)
	}

	// Fetch UTXOs so we know which "coins" we can spend.
	utxos, err := a.getUTXOs(senderAddr.EncodeAddress())
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to get UTXOs: %w", err)
	}
	if len(utxos) == 0 {
		return core.TransferResult{}, fmt.Errorf("no UTXOs — fund address %s via testnet faucet first", senderAddr.EncodeAddress())
	}

	// Fee estimate: 10 sat/vbyte × ~141 vbytes (typical 1-input, 2-output P2WPKH tx).
	const fee = int64(10 * 141)

	// Coin selection: pick just enough UTXOs to cover amount + fee.
	var selected []utxo
	var totalIn int64
	for _, u := range utxos {
		selected = append(selected, u)
		totalIn += u.Value
		if totalIn >= satoshis+fee {
			break
		}
	}
	if totalIn < satoshis+fee {
		return core.TransferResult{}, fmt.Errorf("insufficient funds: have %d sat, need %d sat", totalIn, satoshis+fee)
	}

	tx := wire.NewMsgTx(wire.TxVersion)

	// fetcher stores each UTXO's previous output so BIP143 signing can read it.
	fetcher := make(prevOutFetcher)

	// P2WPKH locking script for the sender: OP_0 <20-byte-pubkey-hash>
	senderScript, err := txscript.PayToAddrScript(senderAddr)
	if err != nil {
		return core.TransferResult{}, err
	}

	// Add one input per selected UTXO (empty scriptSig — SegWit uses the witness field).
	for _, u := range selected {
		hash, err := chainhash.NewHashFromStr(u.TxID)
		if err != nil {
			return core.TransferResult{}, fmt.Errorf("invalid utxo txid: %w", err)
		}
		op := wire.NewOutPoint(hash, u.Vout)
		tx.AddTxIn(wire.NewTxIn(op, nil, nil))
		fetcher[*op] = &wire.TxOut{Value: u.Value, PkScript: senderScript}
	}

	// Output 1: payment to the recipient.
	toScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to build output script: %w", err)
	}
	tx.AddTxOut(&wire.TxOut{Value: satoshis, PkScript: toScript})

	// Output 2: change back to sender (anything below 546 sat is "dust" and would be rejected).
	if change := totalIn - satoshis - fee; change > 546 {
		tx.AddTxOut(&wire.TxOut{Value: change, PkScript: senderScript})
	}

	// Sign each input with the SegWit (BIP143) algorithm.
	// sigHashes is a cache of the pre-computed sighash components; computing it once
	// here avoids redundant hashing when there are multiple inputs.
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)
	for i, u := range selected {
		hash, _ := chainhash.NewHashFromStr(u.TxID)
		op := wire.NewOutPoint(hash, u.Vout)
		// WitnessSignature produces the [sig, pubkey] witness stack required by P2WPKH.
		witness, err := txscript.WitnessSignature(tx, sigHashes, i,
			fetcher[*op].Value, senderScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return core.TransferResult{}, fmt.Errorf("failed to sign input %d: %w", i, err)
		}
		tx.TxIn[i].Witness = witness
	}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return core.TransferResult{}, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	txID, err := a.broadcast(hex.EncodeToString(buf.Bytes()))
	if err != nil {
		return core.TransferResult{}, err
	}

	return core.TransferResult{
		TxHash: txID,
		URL:    fmt.Sprintf("%s/%s", a.explorerURL, txID),
	}, nil
}

// btcToSatoshis converts a human-readable BTC amount (e.g. "0.001") to satoshis.
// 1 BTC = 100,000,000 satoshis (sat), the smallest indivisible unit on Bitcoin.
func btcToSatoshis(amount string) (int64, error) {
	f, _, err := big.ParseFloat(amount, 10, 256, big.ToNearestEven)
	if err != nil {
		return 0, err
	}
	f.Mul(f, new(big.Float).SetFloat64(1e8))
	result, _ := f.Int(nil)
	return result.Int64(), nil
}
