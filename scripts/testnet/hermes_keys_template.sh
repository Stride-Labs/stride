#!/bin/bash

echo "Restoring Hermes Accounts"
hermes keys restore -m "HERMES_STRIDE_MNEMONIC" STRIDE_CHAIN_ID
hermes keys restore -m "HERMES_GAIA_MNEMONIC" GAIA
