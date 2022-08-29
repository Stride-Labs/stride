#!/bin/bash

echo "Restoring Hermes Accounts"

echo "HERMES_STRIDE_MNEMONIC" > mnemonic.txt
hermes keys add --key-name rly1 --chain STRIDE_CHAIN_ID --mnemonic-file mnemonic.txt --overwrite

echo "HERMES_GAIA_MNEMONIC" > mnemonic.txt
hermes keys add --key-name rly2 --chain GAIA --mnemonic-file mnemonic.txt --overwrite

rm -f mnemonic.txt