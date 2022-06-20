#!/bin/bash

sleep 120

echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "kid drift stool kangaroo kid force usual art fringe aerobic fun avoid honey ten since math conduct town sting onion catalog account junk dinner" internal
hermes -c /tmp/hermes.toml keys restore --mnemonic "differ elite buzz carpet true awful win confirm august bird enjoy ready core jar logic radar disorder six trouble excuse filter trim turtle attitude" GAIA_internal

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