#!/bin/bash

sleep 120

echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "metal true farm crucial april accuse social slice clarify tourist magnet error depend arm bright rail idle leopard hair patrol now fossil core theme" internal
hermes -c /tmp/hermes.toml keys restore --mnemonic "scorpion promote supreme try legend must foil cannon laundry cattle whale tip leopard target fiber festival ball athlete thing fatal sing world dash quarter" GAIA_internal

hermes -c /tmp/hermes.toml start &

echo "Creating hermes identifiers"
hermes -c /tmp/hermes.toml tx raw create-client internal GAIA_internal
hermes -c /tmp/hermes.toml tx raw conn-init internal GAIA_internal 07-tendermint-0 07-tendermint-0

echo "Creating connection internal <> GAIA_internal"
hermes -c /tmp/hermes.toml create connection internal GAIA_internal

echo "Creating transfer channel"
hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer GAIA_internal connection-0 
