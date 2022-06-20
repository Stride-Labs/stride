SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo '\n\nInitializing Hermes...'

# import dependencies
source ${SCRIPT_DIR}/testnet_vars.sh $1

HERMES_FILE="${STATE}/hermes.toml"
cp ${SCRIPT_DIR}/hermes_base.toml $HERMES_FILE
sed -i -E "s|127.0.0.1|0.0.0.0|g" $HERMES_FILE
sed -i -E "s|STRIDE_CHAIN|$STRIDE_CHAIN|g" $HERMES_FILE
sed -i -E "s|STRIDE_ADDR|$STRIDE_ENDPOINTS|g" $HERMES_FILE
sed -i -E "s|GAIA_CHAIN|$GAIA_CHAIN|g" $HERMES_FILE
sed -i -E "s|GAIA_ADDR|$GAIA_ENDPOINT|g" $HERMES_FILE
sed -i -E "s|trusting_period = \'150s\'|trusting_period = \'14days\'|g" $HERMES_FILE

RLY_1_KEY=$(GETKEY rly1)
RLY_2_KEY=$(GETRLY2)

HERMES_STARTUP_FILE="${STATE}/hermes_startup.sh"
cp ${SCRIPT_DIR}/hermes_startup_base.sh $HERMES_STARTUP_FILE
sed -i -E "s|STRIDE_CHAIN|$STRIDE_CHAIN|g" $HERMES_STARTUP_FILE
sed -i -E "s|GAIA_CHAIN|$GAIA_CHAIN|g" $HERMES_STARTUP_FILE
sed -i -E "s|RLY_1_KEY|$RLY_1_KEY|g" $HERMES_STARTUP_FILE
sed -i -E "s|RLY_2_KEY|$RLY_2_KEY|g" $HERMES_STARTUP_FILE

rm "${HERMES_FILE}-e"
rm "${HERMES_STARTUP_FILE}-e"

# $HERMES_CMD keys restore --mnemonic "$RLY_1_KEY" $STRIDE_CHAIN
# $HERMES_CMD keys restore --mnemonic "$RLY_2_KEY" $GAIA_CHAIN
