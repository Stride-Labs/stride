#!/bin/bash

sleep 120

echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "lift stairs armed gentle brisk asthma page palace input ice wet impulse thunder betray kangaroo adapt walnut grief twenty cluster large fan hole aisle" internal
hermes -c /tmp/hermes.toml keys restore --mnemonic "surround speed ramp carpet noise suggest enjoy decline rally library potato tray cradle light ostrich announce neutral armed blast split effort sad off elite" GAIA_internal

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
