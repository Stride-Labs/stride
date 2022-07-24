
cd /stride 
git clone https://github.com/tendermint/faucet.git
cd faucet 
make build 
make install
/stride/go/bin/faucet --cli-name strided --denoms ustrd,ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --account-name val2 --keyring-backend test --credit-amount 10000000 --max-credit 5000000001 --home /stride/.stride --port 27935

# /stride/go/bin/faucet --cli-name strided --denoms ustrd --account-name val2 --keyring-backend test --home /stride/.stride
# /stride/go/bin/faucet --cli-name /usr/local/bin/strided --denoms ustrd --account-name val2 --keyring-backend test --home /stride/.stride
# curl -X POST -d '{"address": "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"}' stride-node2.testnet-vishal.stridenet.co:8000
# /stride/go/bin/faucet --cli-name strided --denoms ustrd,ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --account-name val2 --keyring-backend test --credit-amount 10000000 --max-credit 5000000001 --home /stride/.stride --port 27935 & 