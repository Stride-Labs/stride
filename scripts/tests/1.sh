### IBC OVER TOKENS
GAIA1_EXEC="build/gaiad --home scripts-local/state/gaia"
STR1_EXEC="build/strided --home scripts-local/state/stride"
SCRIPT_DIR="./scripts-local/logs/"
STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

#  ibc over atoms to stride
$GAIA1_EXEC tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
SLEEP 5
$STR1_EXEC q bank balances $STRIDE_ADDRESS

