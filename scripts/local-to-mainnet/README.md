# Connecting to Mainnet
## Dependency
* The fleet must be up and running for the host since we need the websocket endpoint

## Setup 
* Ensure you have dockernet setup properly including all submodules up to date and the `STRIDE_ADMIN_MNEMONIC` enviornment variable set 
* Fund three hot wallets and set the mnemonics as environment variables (`HOT_WALLET_1_MNEMONIC`, `HOT_WALLET_2_MNEMONIC`, `HOT_WALLET_3_MNEMONIC`)
    * They all must have a non-zero balance on the host
    * Wallet #1 should have enough to fund each liquid stake
    * Wallet #2 and Wallet #3 only need enough to relay on the host (~1 native token)
* Update the variables at the top of `start.sh`

## Start Stride Local
* Build stride and the relayers
```
make build-docker build=srh
```
* Start a local stride and setup all the commands needed to test the flow
```
bash scripts/local-to-mainnet/start.sh
```

## Walk through Flow
* Step through the commands in `commands.sh` one by one and copy them into the terminal
* In the future, we can automate this more but since this interacts with mainnet, I think it's safer to run these manually for now