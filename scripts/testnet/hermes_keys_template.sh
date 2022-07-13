#!/bin/bash

echo "Restoring Hermes Accounts"
hermes keys restore -m "HERMES_STRIDE_MNEMONIC" STRIDE
hermes keys restore -m "HERMES_GAIA_MNEMONIC" GAIA
