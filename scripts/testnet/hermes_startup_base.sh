
echo "Restoring Hermes Accounts"
hermes -c /tmp/hermes.toml keys restore --mnemonic "RLY_1_KEY" STRIDE_CHAIN
hermes -c /tmp/hermes.toml keys restore --mnemonic "RLY_2_KEY" GAIA_CHAIN

echo "Creating hermes identifiers"
hermes -c /tmp/hermes.toml tx raw create-client STRIDE_CHAIN GAIA_CHAIN
hermes -c /tmp/hermes.toml tx raw conn-init STRIDE_CHAIN GAIA_CHAIN 07-tendermint-0 07-tendermint-0

echo "Creating connection STRIDE_CHAIN <> GAIA_CHAIN"
hermes -c /tmp/hermes.toml create connection STRIDE_CHAIN GAIA_CHAIN

echo "Creating transfer channel"
hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer GAIA_CHAIN connection-0 
