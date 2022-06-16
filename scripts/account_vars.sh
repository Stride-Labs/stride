# Accounts and exec commands
#############################################################################################################################
# import dependencies
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
echo "Getting relevant addresses..."

# Stride
STRIDE_ADDRESS_1=$(${STRIDE_RUN_CMDS[0]} keys show ${STRIDE_VAL_ACCTS[0]} --keyring-backend test -a)
STRIDE_ADDRESS_2=$(${STRIDE_RUN_CMDS[1]} keys show ${STRIDE_VAL_ACCTS[1]} --keyring-backend test -a)
STRIDE_ADDRESS_3=$(${STRIDE_RUN_CMDS[2]} keys show ${STRIDE_VAL_ACCTS[2]} --keyring-backend test -a)

# Gaia
GAIA_ADDRESS_1=$($GAIA1_EXEC keys show ${GAIA_VAL_ACCTS[0]} --keyring-backend test -a --home=/gaia/.gaiad)
GAIA_ADDRESS_2=$($GAIA2_EXEC keys show ${GAIA_VAL_ACCTS[1]} --keyring-backend test -a --home=/gaia/.gaiad)
GAIA_ADDRESS_3=$($GAIA3_EXEC keys show ${GAIA_VAL_ACCTS[2]} --keyring-backend test -a --home=/gaia/.gaiad)

# Relayers
# NOTE: using $STRIDE_MAIN_CMD and $GAIA_MAIN_CMD here ONLY works because they rly1 and rly2
# keys are on stride1 and gaia1, respectively
RLY_ADDRESS_1=$($STRIDE_MAIN_CMD keys show rly1 --keyring-backend test -a)
RLY_ADDRESS_2=$($GAIA_MAIN_CMD keys show rly2 --keyring-backend test -a)

echo "Grabbed all data, running tests..."