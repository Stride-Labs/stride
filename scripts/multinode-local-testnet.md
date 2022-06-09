# Multi Node Local Testnet Script

This script creates a multi node local testnet with three validator nodes on a single machine. Note: The default weights of these validators is 5:5:4 respectively. That means in order to keep the chain running, at a minimum Validator1 and Validator2 must be running in order to keep greater than 66% power online.

## Instructions

Clone the stride repo

Checkout the branch you are looking to test

Make install / reload profile

Give the script permission with `chmod +x scripts/multinode-local-testnet.sh`

Run with `./scripts/multinode-local-testnet.sh` (allow ~45 seconds to run, required sleep commands due to multiple transactions)

## Logs

Validator1: `tmux a -t validator1`

Validator2: `tmux a -t validator2`

Validator3: `tmux a -t validator3`

## Directories

Validator1: `$HOME/.strided/validator1`

Validator2: `$HOME/.strided/validator2`

Validator3: `$HOME/.strided/validator3`

## Ports

Validator1: `1317, 9090, 9091, 26658, 26657, 26656, 6060`

Validator2: `1316, 9088, 9089, 26655, 26654, 26653, 6061`

Validator3: `1315, 9086, 9087, 26652, 26651, 26650, 6062`

Ensure to include the `--home` flag or `--node` flag when using a particular node.

## Examples

Validator2: `strided status --node "tcp://localhost:26654"`

Validator3: `strided status --node "tcp://localhost:26651"`

or

Validator1: `strided keys list --keyring-backend test --home $HOME/.strided/validator1`

Validator2: `strided keys list --keyring-backend test --home $HOME/.strided/validator2`

## Commands

strided q bank balances cosmos1d5nl74gmghpxp690va2mxrs4az74upgcgjqmwf // --home=$HOME/.strided/validator1

# send

strided tx bank send cosmos1d5nl74gmghpxp690va2mxrs4az74upgcgjqmwf cosmos1pgtvvls6qtu6lpa0jy4j6xxc678tzmnhuwee03 100ustrd // --home=$HOME/.strided/validator1

# transfer tokens tx

strided tx bank send validator1 cosmos1pgtvvls6qtu6lpa0jy4j6xxc678tzmnhuwee03 1000ustrd --keyring-backend test --home=$HOME/.strided/validator1 --chain-id=testing

# query tx

strided q tx 44E80B728F4448E941A458E42751FD19264A604F4C50D330316EE23DC2A2BA95 --home=$HOME/.strided/validator1 --node "tcp://localhost:26657"

## manual signing

# test generate tx

strided tx bank send cosmos1d5nl74gmghpxp690va2mxrs4az74upgcgjqmwf cosmos1pgtvvls6qtu6lpa0jy4j6xxc678tzmnhuwee03 1000ustrd --chain-id testing --home=$HOME/.strided/validator1 --keyring-backend test --generate-only

# test sign tx

strided tx sign test_tx.json --chain-id testing --keyring-backend test --home=$HOME/.strided/validator1 --from validator1

# broadcast

strided tx broadcast tx_signed.json --keyring-backend test --home=$HOME/.strided/validator1 --node "tcp://localhost:26657"
