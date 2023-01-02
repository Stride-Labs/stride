CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/common.sh
source ${CURRENT_DIR}/setup.sh
source ${CURRENT_DIR}/test_epoch_reset.sh
source ${CURRENT_DIR}/test_tx_reset.sh
source ${CURRENT_DIR}/test_tx_remove.sh
source ${CURRENT_DIR}/test_quota_update.sh
source ${CURRENT_DIR}/test_bidirectional_flow.sh
source ${CURRENT_DIR}/test_denoms.sh

setup_juno_osmo_channel
setup_channel_value
setup_rate_limits

test_epoch_reset_atom_from_gaia_to_stride
test_epoch_reset_atom_from_stride_to_gaia
test_tx_reset_atom_from_gaia_to_stride
test_tx_reset_atom_from_stride_to_gaia
test_quota_update
test_bidirectional
test_denom_ustrd
test_denom_ujuno
test_denom_sttoken
test_denom_juno_traveler
test_remove_rate_limit
