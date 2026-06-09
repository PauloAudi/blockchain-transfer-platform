# Cross-Chain Transfer Platform

A modular and reusable platform for transferring digital assets between different blockchain networks, built as part of an MBA Software Engineering thesis.

## Architecture

The platform uses the **Adapter Pattern** to abstract each blockchain behind a common interface. Adding a new network only requires creating a new adapter — no changes to the core orchestration logic.

```
core/adapter.go          → BlockchainAdapter interface
chains/eth/adapter.go    → Ethereum implementation
chains/sol/adapter.go    → Solana implementation (coming soon)
chains/btc/adapter.go    → Bitcoin implementation (coming soon)
main.go                  → Orchestrator
keygen/main.go           → Wallet key pair generator
```

## Requirements

- [Go 1.21+](https://go.dev/dl/)
- A funded wallet on the target network's testnet

## Setup

### 1. Generate wallets

```bash
go run keygen/main.go
```

This prints two wallet addresses and their private keys. Save the output — you will need it in the next step.

### 2. Create the environment file

Create a `.env` file in the project root:

```
PRIVATE_KEY=<private key of the sender wallet, without 0x>
TO_ADDRESS=<address of the recipient wallet>
```

This file is listed in `.gitignore` and will never be committed.

### 3. Fund the sender wallet

Get testnet ETH for the sender wallet address on the [Sepolia faucet](https://cloud.google.com/application/web3/faucet/ethereum/sepolia).

### 4. Run the transfer

```bash
export $(cat .env) && go run main.go
```

The program prints the transaction hash and a link to the block explorer where you can confirm the transfer.

## Supported Networks

| Network   | Type    | Status      |
|-----------|---------|-------------|
| Ethereum  | EVM     | Implemented |
| Solana    | non-EVM | Coming soon |
| Bitcoin   | non-EVM | Coming soon |
