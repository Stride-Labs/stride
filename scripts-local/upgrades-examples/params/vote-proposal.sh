
PROPOSAL_ID=1
build/strided --home ./scripts-local/state/stride tx gov vote $PROPOSAL_ID yes --from val1 --keyring-backend test --chain-id STRIDE -y
build/strided --home ./scripts-local/state/stride tx gov vote $PROPOSAL_ID yes --from rly1 --keyring-backend test --chain-id STRIDE -y
build/strided --home ./scripts-local/state/stride tx gov vote $PROPOSAL_ID yes --from icq1 --keyring-backend test --chain-id STRIDE -y