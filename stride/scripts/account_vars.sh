# Accounts and exec commands
#############################################################################################################################
# import dependencies
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
echo "Getting relevant addresses..."

# Stride
STRIDE_ADDRESS_1=$($BASE_RUN keys show val1 --home $STATE/STRIDE_1 --keyring-backend test -a)
STRIDE_ADDRESS_2=$($BASE_RUN keys show val2 --home $STATE/STRIDE_2 --keyring-backend test -a)
STRIDE_ADDRESS_3=$($BASE_RUN keys show val3 --home $STATE/STRIDE_3 --keyring-backend test -a)

# Gaia
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"
GAIA2_EXEC="docker-compose --ansi never exec -T gaia2 gaiad --home /gaia/.gaiad"
GAIA3_EXEC="docker-compose --ansi never exec -T gaia3 gaiad --home /gaia/.gaiad"
GAIA_ADDRESS_1=$($GAIA1_EXEC keys show gval1 --keyring-backend test -a --home=/gaia/.gaiad)
GAIA_ADDRESS_2=$($GAIA2_EXEC keys show gval2 --keyring-backend test -a --home=/gaia/.gaiad)
GAIA_ADDRESS_3=$($GAIA3_EXEC keys show gval3 --keyring-backend test -a --home=/gaia/.gaiad)

# Relayers
# NOTE: using $main_cmd and $main_gaia_cmd here ONLY works because they rly1 and rly2
# keys are on stride1 and gaia1, respectively
RLY_ADDRESS_1=$($main_cmd keys show rly1 --keyring-backend test -a)
RLY_ADDRESS_2=$($main_gaia_cmd keys show rly2 --keyring-backend test -a)

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE_1"
STR2_EXEC="docker-compose --ansi never exec -T stride2 strided --home /stride/.strided --chain-id STRIDE_1"
STR3_EXEC="docker-compose --ansi never exec -T stride3 strided --home /stride/.strided --chain-id STRIDE_1"

echo "Grabbed all data, running tests..."