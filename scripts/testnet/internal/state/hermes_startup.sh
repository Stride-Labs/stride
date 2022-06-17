
echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "brass identify file ability snack entry ahead motion hedgehog topic clock edge salt stand alone drop vote final outer brown nerve sadness host brick" internal
hermes -c /tmp/hermes.toml keys restore --mnemonic "cube page approve source empty grape chicken praise scare minor engine jaguar opera install arrive real supply neck file fever turkey volume spike infant" GAIA_internal

hermes -c /tmp/hermes.toml start &

sleep 60

echo "Creating hermes identifiers"
hermes -c /tmp/hermes.toml tx raw create-client internal GAIA_internal
hermes -c /tmp/hermes.toml tx raw conn-init internal GAIA_internal 07-tendermint-0 07-tendermint-0

echo "Creating connection internal <> GAIA_internal"
hermes -c /tmp/hermes.toml create connection internal GAIA_internal

echo "Creating transfer channel"
hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer GAIA_internal connection-0 
