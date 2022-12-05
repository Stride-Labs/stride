#!/bin/sh

CHAIN_ID=localstride
STRIDE_HOME=$HOME/.stride
CONFIG_FOLDER=$STRIDE_HOME/config
MONIKER=val

MNEMONIC="deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment"

edit_genesis () {

    GENESIS=$CONFIG_FOLDER/genesis.json

    # Update staking module
    dasel put string -f $GENESIS '.app_state.staking.params.bond_denom' 'ustrd'
    dasel put string -f $GENESIS '.app_state.staking.params.unbonding_time' '240s'

    # Update crisis module
    dasel put string -f $GENESIS '.app_state.crisis.constant_fee.denom' 'ustrd'

    # Udpate gov module
    dasel put string -f $GENESIS '.app_state.gov.voting_params.voting_period' '60s'
    dasel put string -f $GENESIS '.app_state.gov.deposit_params.min_deposit.[0].denom' 'ustrd'

    # Update epochs module
    dasel put string -f $GENESIS '.app_state.epochs.epochs.(.identifier=day).duration' '120s'
    dasel put string -f $GENESIS '.app_state.epochs.epochs.(.identifier=stride_epoch).duration' '120s'

    # Update mint module
    dasel put string -f $GENESIS '.app_state.mint.params.mint_denom' 'ustrd'
    dasel put string -f $GENESIS '.app_state.mint.params.epoch_identifier' 'mint'

}

add_genesis_accounts () {

    strided add-genesis-account stride1wal8dgs7whmykpdaz0chan2f54ynythkz0cazc 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1u9klnra0d4zq9ffalpnr3nhz5859yc7ckdk9wt 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1kwax6g0q2nwny5n43fswexgefedge033t5g95j 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1dv0ecm36ywdyc6zjftw0q62zy6v3mndrwxde03 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1z3dj2tvqpzy2l5shx98f9k5486tleah5a00fay 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride14khzkfs8luaqymdtplrt5uwzrghrndeh4235am 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1qym804u6sa2gvxedfy96c0v9jc0ww7593uechw 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1et8cdkxl69yrtmpjhxwey52d88kflwzn5xp4xn 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride1tcrlyn05q9j590uauncywf26ptfn8se65dvfz6 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride14ugekxs6f4rfleg6wj8k0wegv69khfpxkt8yn4 100000000000ustrd --home $STRIDE_HOME
    strided add-genesis-account stride18htv32r83q2wn2knw5wp9m4nkp4xuzyfhmwpqs 100000000000ustrd --home $STRIDE_HOME

    echo $MNEMONIC | strided keys add $MONIKER --recover --keyring-backend=test --home $STRIDE_HOME
    strided gentx $MONIKER 500000000ustrd --keyring-backend=test --chain-id=$CHAIN_ID --home $STRIDE_HOME

    strided collect-gentxs --home $STRIDE_HOME
}

edit_config () {
    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
}

if [[ ! -d $CONFIG_FOLDER ]]
then
    echo $MNEMONIC | strided init -o --chain-id=$CHAIN_ID --home $STRIDE_HOME --recover $MONIKER
    edit_genesis
    add_genesis_accounts
    edit_config
fi

strided start --home $STRIDE_HOME
