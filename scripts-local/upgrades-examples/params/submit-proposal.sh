

build/strided --home ./scripts-local/state/stride tx gov submit-proposal param-change \
            ./scripts-local/upgrades-examples/params/param-slashing.json \
            --from val1 \
            --keyring-backend test \
            --chain-id STRIDE