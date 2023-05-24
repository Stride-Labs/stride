# Connecting to Mainnet
## Dependency
* The fleet must be up and running for the host since we need the websocket endpoint

## Setup 
* Ensure you have dockernet setup properly including all submodules up to date and the `STRIDE_ADMIN_MNEMONIC` enviornment variable set 
* Fund three hot wallets and set the mnemonics as environment variables (`HOT_WALLET_1_MNEMONIC`, `HOT_WALLET_2_MNEMONIC`, `HOT_WALLET_3_MNEMONIC`)
    * They all must have a non-zero balance on the host
    * Wallet #1 should have enough to fund each liquid stake (~5 native token per attempt)
    * Wallet #2 only needs enough to create clients and connections (~0.20 native token)
    * Wallet #3 only needs enough to relayer on the host (~0.50 native token)
* Update the variables at the top of `start.sh`

## Start Stride Local
* Build stride and the relayers
```bash 
make build-docker build=srh{n} # where n is the new host zone that was just added
```
* Start a local stride instance and setup all the commands needed to test the flow
```
make start-local-to-main
```

## Walk through Flow
* Step through the commands in `local-to-mainnet/commands.sh` one by one and copy them into the terminal
* In the future, we can automate this more but since this interacts with mainnet, I think it's safer to run these manually for now