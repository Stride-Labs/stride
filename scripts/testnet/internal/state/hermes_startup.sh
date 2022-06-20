#!/bin/bash

sleep 120

echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "share woman nominee cloud film memory pull funny base card student yard phone forget easily breeze unaware six rose meadow harbor sausage order orient" internal
hermes -c /tmp/hermes.toml keys restore --mnemonic "void quality prevent ketchup mesh hidden faith bird lonely limb album assist indicate lens loop regular pitch clump coach almost mango useless strong peasant" GAIA_internal

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

wait