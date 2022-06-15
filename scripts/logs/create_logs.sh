SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 15 ; done`)

rm $SCRIPT_DIR/all.log
docker compose logs >> $SCRIPT_DIR/all.log

rm $SCRIPT_DIR/stride.log
docker compose logs stride1 >> $SCRIPT_DIR/stride.log

rm $SCRIPT_DIR/gaia.log
docker compose logs gaia1 >> $SCRIPT_DIR/gaia.log

rm $SCRIPT_DIR/icq.log
docker compose logs icq >> $SCRIPT_DIR/icq.log

rm $SCRIPT_DIR/hermes.log
docker compose logs hermes >> $SCRIPT_DIR/hermes.log

# transactions logs
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

rm $SCRIPT_DIR/interchainquery.log
$STR1_EXEC strided q txs --events message.module=interchainquery --limit=100000 > $SCRIPT_DIR/interchainquery.log

rm $SCRIPT_DIR/stakeibc.log
$STR1_EXEC strided q txs --events message.module=stakeibc --limit=100000 > $SCRIPT_DIR/stakeibc.log


# accounts
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"
GAIA_DELEGATE="cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu"
GAIA_WITHDRAWAL="cosmos1lcnmjwjy2lnqged5pnrc0cstz0r88rttunla4zxv84mee30g2q3q48fm53"
STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

rm $SCRIPT_DIR/accounts.log
echo "BALANCES STRIDE" > $SCRIPT_DIR/accounts.log
$STR1_EXEC strided q bank balances $STRIDE_ADDRESS >> $SCRIPT_DIR/accounts.log
echo "\nBALANCES GAIA" >> $SCRIPT_DIR/accounts.log
$GAIA1_EXEC q bank balances $GAIA_DELEGATE >> $SCRIPT_DIR/accounts.log
echo "\nDELEGATIONS GAIA" >> $SCRIPT_DIR/accounts.log
$GAIA1_EXEC q staking delegations $GAIA_DELEGATE >> $SCRIPT_DIR/accounts.log
echo "\nLIST-HOST-ZONES" >> $SCRIPT_DIR/accounts.log
$STR1_EXEC strided q stakeibc list-host-zone >> $SCRIPT_DIR/accounts.log
