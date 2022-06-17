SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/testnet_vars.sh $1

# run through init args, if needed
while getopts b flag; do
    case "${flag}" in
        b) ignite chain init ;;
    esac
done

echo "Cleaning state"
rm -rf $STATE
mkdir $STATE
touch $STATE/keys.txt

source ${SCRIPT_DIR}/setup_stride_state.sh
source ${SCRIPT_DIR}/setup_gaia_state.sh $STRIDE_CHAIN
source ${SCRIPT_DIR}/setup_hermes_state.sh $STRIDE_CHAIN
