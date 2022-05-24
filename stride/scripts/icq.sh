#!/bin/bash

echo "Testing icq to read gaia address ${GAIA_ADDRESS_1} from stride"
# strided query interchainquery query-balance connection-0 $GAIA_ADDRESS_1 --home /stride/.strided --keyring-backend test --from val1 --chain-id STRIDE_1 -y
strided tx interchainquery query-balance GAIA_1 cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv uatom --connection-id connection-1 --from val1 --home /stride/.strided --keyring-backend test --chain-id STRIDE_1
sleep 9

