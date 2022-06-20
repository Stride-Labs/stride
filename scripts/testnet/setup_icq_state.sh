SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo '\n\nInitializing Hermes...'

# import dependencies
source ${SCRIPT_DIR}/testnet_vars.sh $1

ICQ_DIR=${STATE}/icq
mkdir $ICQ_DIR

RLY_1_KEY=$(GETKEY rly1)
RLY_2_KEY=$(GETRLY2)

ICQ_STARTUP_FILE="${STATE}/icq_startup.sh"
cp ${SCRIPT_DIR}/icq_startup_base.sh $ICQ_STARTUP_FILE
sed -i -E "s|STRIDE_CHAIN|$STRIDE_CHAIN|g" $ICQ_STARTUP_FILE
sed -i -E "s|GAIA_CHAIN|$GAIA_CHAIN|g" $ICQ_STARTUP_FILE
sed -i -E "s|ICQ_STRIDE_KEY|$RLY_1_KEY|g" $ICQ_STARTUP_FILE
sed -i -E "s|ICQ_GAIA_KEY|$RLY_2_KEY|g" $ICQ_STARTUP_FILE

rm "${ICQ_STARTUP_FILE}-e"