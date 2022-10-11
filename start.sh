rm -rf $HOME/.stride
make install
strided init val1 --chain-id=testing --home=$HOME/.stride
strided keys add validator --keyring-backend=test
strided add-genesis-account $(strided keys show validator --keyring-backend=test --home=$HOME/.stride -a) 1000000000000000000ustrd --home=$HOME/.stride
strided gentx validator 10000000000ustrd --keyring-backend=test --chain-id=testing
strided collect-gentxs 
sed -i '' 's/"voting_period": "172800s"/"voting_period": "20s"/g' $HOME/.stride/config/genesis.json
sed -i '' 's/"stake"/"ustrd"/g' $HOME/.stride/config/genesis.json

strided start --home=$HOME/.stride