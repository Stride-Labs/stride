#!/bin/bash
rm -rf $HOME/.strided/


# make four stride directories
mkdir $HOME/.strided
mkdir $HOME/.strided/validator1
mkdir $HOME/.strided/validator2
mkdir $HOME/.strided/validator3

# init all three validators
strided init --chain-id=testing validator1 --home=$HOME/.strided/validator1
strided init --chain-id=testing validator2 --home=$HOME/.strided/validator2
strided init --chain-id=testing validator3 --home=$HOME/.strided/validator3
# create keys for all three validators
strided keys add validator1 --keyring-backend=test --home=$HOME/.strided/validator1
strided keys add validator2 --keyring-backend=test --home=$HOME/.strided/validator2
strided keys add validator3 --keyring-backend=test --home=$HOME/.strided/validator3

# change staking denom to uosmo
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# create validator node with tokens to transfer to the three other nodes
strided add-genesis-account $(strided keys show validator1 -a --keyring-backend=test --home=$HOME/.strided/validator1) 100000000000uosmo,100000000000stake --home=$HOME/.strided/validator1
strided gentx validator1 500000000uosmo --keyring-backend=test --home=$HOME/.strided/validator1 --chain-id=testing
strided collect-gentxs --home=$HOME/.strided/validator1


# update staking genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="240s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# update crisis variable to uosmo
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# udpate gov genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="60s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# update epochs genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["epochs"]["epochs"][1]["duration"]="60s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# update poolincentives genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# update incentives genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# update mint genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="uosmo"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="day"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json

# update gamm genesis
cat $HOME/.strided/validator1/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"' > $HOME/.strided/validator1/config/tmp_genesis.json && mv $HOME/.strided/validator1/config/tmp_genesis.json $HOME/.strided/validator1/config/genesis.json


# port key (validator1 uses default ports)
# validator1 1317, 9090, 9091, 26658, 26657, 26656, 6060
# validator2 1316, 9088, 9089, 26655, 26654, 26653, 6061
# validator3 1315, 9086, 9087, 26652, 26651, 26650, 6062
# validator4 1314, 9084, 9085, 26649, 26648, 26647, 6063


# change app.toml values

# validator2
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1316|g' $HOME/.strided/validator2/config/app.toml
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9088|g' $HOME/.strided/validator2/config/app.toml
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9089|g' $HOME/.strided/validator2/config/app.toml

# validator3
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1315|g' $HOME/.strided/validator3/config/app.toml
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9086|g' $HOME/.strided/validator3/config/app.toml
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9087|g' $HOME/.strided/validator3/config/app.toml


# change config.toml values

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $HOME/.strided/validator1/config/config.toml
# validator2
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $HOME/.strided/validator2/config/config.toml
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26654|g' $HOME/.strided/validator2/config/config.toml
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $HOME/.strided/validator2/config/config.toml
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $HOME/.strided/validator3/config/config.toml
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $HOME/.strided/validator2/config/config.toml
# validator3
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26652|g' $HOME/.strided/validator3/config/config.toml
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26651|g' $HOME/.strided/validator3/config/config.toml
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $HOME/.strided/validator3/config/config.toml
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $HOME/.strided/validator3/config/config.toml
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $HOME/.strided/validator3/config/config.toml


# copy validator1 genesis file to validator2-3
cp $HOME/.strided/validator1/config/genesis.json $HOME/.strided/validator2/config/genesis.json
cp $HOME/.strided/validator1/config/genesis.json $HOME/.strided/validator3/config/genesis.json


# copy tendermint node id of validator1 to persistent peers of validator2-3
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(strided tendermint show-node-id --home=$HOME/.strided/validator1)@$(curl -4 icanhazip.com):26656\"|g" $HOME/.strided/validator2/config/config.toml
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(strided tendermint show-node-id --home=$HOME/.strided/validator1)@$(curl -4 icanhazip.com):26656\"|g" $HOME/.strided/validator3/config/config.toml


# start all three validators
tmux new -s validator1 -d strided start --home=$HOME/.strided/validator1
tmux new -s validator2 -d strided start --home=$HOME/.strided/validator2
tmux new -s validator3 -d strided start --home=$HOME/.strided/validator3


# send uosmo from first validator to second validator
sleep 7
strided tx bank send validator1 $(strided keys show validator2 -a --keyring-backend=test --home=$HOME/.strided/validator2) 500000000uosmo --keyring-backend=test --home=$HOME/.strided/validator1 --chain-id=testing --yes
sleep 7
strided tx bank send validator1 $(strided keys show validator3 -a --keyring-backend=test --home=$HOME/.strided/validator3) 400000000uosmo --keyring-backend=test --home=$HOME/.strided/validator1 --chain-id=testing --yes

# create second validator
sleep 7
strided tx staking create-validator --amount=500000000uosmo --from=validator2 --pubkey=$(strided tendermint show-validator --home=$HOME/.strided/validator2) --moniker="validator2" --chain-id="testing" --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.strided/validator2 --yes
sleep 7
strided tx staking create-validator --amount=400000000uosmo --from=validator3 --pubkey=$(strided tendermint show-validator --home=$HOME/.strided/validator3) --moniker="validator3" --chain-id="testing" --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="400000000" --keyring-backend=test --home=$HOME/.strided/validator3 --yes
