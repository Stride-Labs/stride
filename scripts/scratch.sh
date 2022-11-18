### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

## IBC ATOM from GAIA to STRIDE
# $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
# sleep 10
# $STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS)

# stride airdrop: stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l

# send funds to airdrop address
# $STRIDE_MAIN_CMD tx bank send val1 stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
# sleep 5
# $STRIDE_MAIN_CMD q tx DFE1C5F517EBB84EA830A91C839B301CE2D6833C0B4D9D91B7F5F3CF096D0713
# send funds to claim address
# $STRIDE_MAIN_CMD tx bank send val1 stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d 1ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
# sleep 5

# $STRIDE_MAIN_CMD q bank balances stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l
# $STRIDE_MAIN_CMD q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d

$STRIDE_MAIN_CMD # version # tx claim claim-free-amount # --from test --chain-id STRIDE -y --keyring-backend test
#$STRIDE_MAIN_CMD tx bank send val1 stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test


build/strided tx claim claim-free-amount stride --from airdrop --home scripts/state/stride1 --node http://localhost:26657

build/strided q tx BA7E0FC2CFBD67E625AD703FEBAFBDF6BB2E88C190B5A5813F171D2DFC214CAF --node http://localhost:26657

build/strided q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d --node http://localhost:26657

build/strided tx staking delegate stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm 1ustrd --from airdrop --home scripts/state/stride1 --node http://localhost:26657



build/strided q staking validators --node http://localhost:26657

build/strided q bank balances stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l --node http://localhost:26657
build/strided q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d --node http://localhost:26657
build/strided q tx 748F78633A641021F06F379288E4D45119F014B300B620F9877F4E2F10257FD5 --node http://localhost:26657

