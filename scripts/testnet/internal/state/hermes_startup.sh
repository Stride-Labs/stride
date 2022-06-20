#!/bin/bash

sleep 120

echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "drink turn model style increase suspect short dilemma funny senior van idle treat observe lamp pause melt female chief permit kick tower hold wrap" internal
hermes -c /tmp/hermes.toml keys restore --mnemonic "jewel wedding now river obey cross roast team core subject assault blush check oxygen abuse reward estate alarm ski hub unique plastic attract liquid" GAIA_internal

hermes -c /tmp/hermes.toml start &
sleep 30

echo "Creating hermes identifiers"
hermes -c /tmp/hermes.toml tx raw create-client internal GAIA_internal
sleep 15 

hermes -c /tmp/hermes.toml tx raw conn-init internal GAIA_internal 07-tendermint-0 07-tendermint-0
sleep 15

echo "Creating connection internal <> GAIA_internal"
hermes -c /tmp/hermes.toml create connection internal GAIA_internal
sleep 15

echo "Creating transfer channel"
hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer GAIA_internal connection-0 
