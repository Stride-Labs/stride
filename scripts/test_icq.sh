#!/bin/bash

echo "Testing icq to read gaia address ${GAIA_ADDRESS_1} from stride"
# strided tx interchainquery query-balance GAIA_1 cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv uatom --connection-id connection-1  --home /stride/.strided --keyring-backend test --chain-id STRIDE_1 -y --from val1
strided tx bank send val1 stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4 2500000ustrd --home /stride/.strided --keyring-backend test --chain-id STRIDE_1 -y
sleep 5
TXHASH=$(strided tx interchainquery query-balance GAIA_1 $GAIA_ADDRESS_1 uatom --connection-id connection-1  --home /stride/.strided --keyring-backend test --chain-id STRIDE_1 -y --from val1 | tail -c 65)
strided q tx $TXHASH
sleep 9

