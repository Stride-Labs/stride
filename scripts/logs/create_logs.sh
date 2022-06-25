SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 15 ; done`)
TMP="temp"
docker compose logs >> $SCRIPT_DIR/$TMP/all.log
docker compose logs stride1 >> $SCRIPT_DIR/$TMP/stride.log
docker compose logs gaia1 >> $SCRIPT_DIR/$TMP/gaia.log
docker compose logs icq >> $SCRIPT_DIR/$TMP/icq.log
docker compose logs hermes >> $SCRIPT_DIR/$TMP/hermes.log

# transactions logs
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
$STR1_EXEC strided q txs --events message.module=interchainquery --limit=100000 > $SCRIPT_DIR/$TMP/interchainquery.log
$STR1_EXEC strided q txs --events message.module=stakeibc --limit=100000 > $SCRIPT_DIR/$TMP/stakeibc.log

# build/gaiad --home scripts-local/state/gaia tx ibc-transfer transfer transfer channel-0 $(build/strided --home scripts-local/state/stride keys show val1 --keyring-backend test -a) 100000uatom --from gval1 --chain-id GAIA --keyring-backend test

# accounts
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"
GAIA_DELEGATE="cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu"
GAIA_WITHDRAWAL="cosmos1lcnmjwjy2lnqged5pnrc0cstz0r88rttunla4zxv84mee30g2q3q48fm53"
STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

N_VALIDATORS_STRIDE=$($STR1_EXEC strided q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
N_VALIDATORS_GAIA=$($GAIA1_EXEC q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
echo "STRIDE @ $($STR1_EXEC strided q  tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" > $SCRIPT_DIR/$TMP/accounts.log
echo "GAIA   @ $($GAIA1_EXEC q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_GAIA VALS" >> $SCRIPT_DIR/$TMP/accounts.log

echo "\nBALANCES STRIDE" >> $SCRIPT_DIR/$TMP/accounts.log
$STR1_EXEC strided q bank balances $STRIDE_ADDRESS >> $SCRIPT_DIR/$TMP/accounts.log
echo "\nBALANCES GAIA" >> $SCRIPT_DIR/$TMP/accounts.log
$GAIA1_EXEC q bank balances $GAIA_DELEGATE >> $SCRIPT_DIR/$TMP/accounts.log
echo "\nDELEGATIONS GAIA" >> $SCRIPT_DIR/$TMP/accounts.log
$GAIA1_EXEC q staking delegations $GAIA_DELEGATE >> $SCRIPT_DIR/$TMP/accounts.log
echo "\nLIST-HOST-ZONES STRIDE" >> $SCRIPT_DIR/$TMP/accounts.log
$STR1_EXEC strided q stakeibc list-host-zone | head -n 50 >> $SCRIPT_DIR/$TMP/accounts.log
echo "\nLIST-CONTROLLER-BALANCES" >> $SCRIPT_DIR/$TMP/accounts.log
$STR1_EXEC strided q stakeibc list-controller-balances >> $SCRIPT_DIR/$TMP/accounts.log

mv $SCRIPT_DIR/$TMP/*.log $SCRIPT_DIR

