# Multi-Chain Transfer Platform

A modular and reusable platform for transferring digital assets on multiple blockchain networks (Ethereum, Solana, Bitcoin), built as part of an MBA Software Engineering thesis.

## Architecture

The platform uses the **Adapter Pattern** to abstract each blockchain behind a common interface. Adding a new network only requires creating a new adapter — no changes to the core orchestration logic.

```
core/adapter.go          → BlockchainAdapter interface
chains/eth/adapter.go    → Ethereum implementation
chains/sol/adapter.go    → Solana implementation
chains/btc/adapter.go    → Bitcoin implementation
main.go                  → Orchestrator
keygen/eth/main.go       → Ethereum wallet key pair generator
keygen/sol/main.go       → Solana wallet key pair generator
keygen/btc/main.go       → Bitcoin wallet key pair generator
```

## Requirements

- [Go 1.25+](https://go.dev/dl/)
- A funded wallet on the target network's testnet

## Setup

### 1. Generate wallets

```bash
go run keygen/eth/main.go   # Ethereum (Sepolia)
go run keygen/sol/main.go   # Solana (Devnet)
go run keygen/btc/main.go   # Bitcoin (Testnet3)
```

Each command prints two wallet addresses and their private keys. Save the output — you will need it in the next step.

### 2. Create the environment file

Create a `.env` file in the project root:

```
CHAIN=eth              # eth | sol | btc
PRIVATE_KEY=<private key of the sender wallet>
TO_ADDRESS=<address of the recipient wallet>
```

This file is listed in `.gitignore` and will never be committed.

> Private key format depends on the chain: hex without `0x` for Ethereum, base58 for Solana, WIF for Bitcoin (as printed by the respective keygen).

### 3. Fund the sender wallet

Get testnet funds for the sender wallet address:

- Ethereum (Sepolia): [Sepolia faucet](https://cloud.google.com/application/web3/faucet/ethereum/sepolia)
- Solana (Devnet): `solana airdrop` via the Solana CLI, or any public Devnet faucet
- Bitcoin (Testnet3): any active testnet3 faucet

### 4. Run the transfer

```bash
export $(cat .env) && go run main.go
```

The program sends a fixed amount (`0.001` in the network's native unit) and prints the transaction hash and a link to the block explorer where you can confirm the transfer.

## Supported Networks

| Network   | Type    | Testnet    | Status      |
|-----------|---------|------------|-------------|
| Ethereum  | EVM     | Sepolia    | Implemented |
| Solana    | non-EVM | Devnet     | Implemented |
| Bitcoin   | non-EVM | Testnet3   | Implemented |

RPC/API endpoints used (all public):

- Ethereum: `https://ethereum-sepolia-rpc.publicnode.com`
- Solana: `https://api.devnet.solana.com`
- Bitcoin: `https://mempool.space/testnet/api`

