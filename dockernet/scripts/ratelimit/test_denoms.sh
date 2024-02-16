CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

# We want to cover the following cases:
#
# 1. Send native
# 2. Send non-native (one hop away)
# 3. Send non-native (two hops away)
# 4. Receive sink (one hop away)
# 5. Receive sink (two hops away)
# 6. Receive source native
# 7. Receive source non-native
#
# For each case, we'll simply need to try the transfer and check if the flow updated,
#  if the flow didn't update along with expectations, that means either the denom or channel was wrong

##################################################
# ustrd from Stride to Osmosis then back to Stride
##################################################
__test_denom_send_packet_native_ustrd() {  # send native
    # ustrd sent from Stride to Osmosis
    #   Expected Denom: ustrd
    #   Expected Channel: channel-2
    check_transfer_status STRIDE OSMO channel-2 channel-2 10000000 ustrd ustrd true
}

__test_denom_receive_packet_native_ustrd() { # receive source native
    # ustrd sent from Osmosis to Stride
    #   Expected Denom: ustrd
    #   Expected Channel: channel-2
    ustrd_on_osmo='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
    check_transfer_status OSMO STRIDE channel-0 channel-2 10000000 $ustrd_on_osmo ustrd true
}

test_denom_ustrd() {
    print_header "TESTING DENOMS - USTRD"
    wait_until_epoch_end

    __test_denom_send_packet_native_ustrd
    __test_denom_receive_packet_native_ustrd
}

##############################################
# ujuno from Juno to Stride, then back to Juno
##############################################
__test_denom_receive_packet_non_native() { # receive sink (one hop)
    # ujuno sent from Juno to Stride
    #   Expected Denom: ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5
    #   Expected Channel: channel-1
    juno_on_stride='ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5'
    check_transfer_status JUNO STRIDE channel-0 channel-1 10000000 ujuno $juno_on_stride true
}

__test_denom_send_packet_non_native() { # send non native (one hop)
    # ujuno sent from Stride to Juno
    #   Expected Denom: ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5
    #   Expected Channel: channel-1
    juno_on_stride='ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5'
    check_transfer_status STRIDE JUNO channel-1 channel-1 10000000 $juno_on_stride $juno_on_stride true
}

test_denom_ujuno() {
    print_header "TESTING DENOMS - UJUNO"
    wait_until_epoch_end

    __test_denom_receive_packet_non_native
    __test_denom_send_packet_non_native
}

#####################################################
# stujuno from Stride to Osmosis, then back to Stride
#####################################################
__test_denom_send_packet_native_sttoken() { # send native
    # stujuno sent from Stride to Osmosis
    #   Expected Denom: stujuno
    #   Expected Channel: channel-2
    check_transfer_status STRIDE OSMO channel-2 channel-2 10000000 stujuno stujuno true
}

__test_denom_receive_packet_native_sttoken() { # receive source native
    # stujuno sent from Osmosis to Stride
    #   Expected Denom: stujuno
    #   Expected Channel: channel-2
    stujuno_on_osmo='ibc/C4385BAF25938E02B0EA90D512CE43BFACA892F7FAD81D63CC82BD8EBFA21857'
    check_transfer_status OSMO STRIDE channel-0 channel-2 10000000 $stujuno_on_osmo stujuno true
}

test_denom_sttoken() {
    print_header "TESTING DENOMS - STUJUNO"
    wait_until_epoch_end

    __test_denom_send_packet_native_sttoken
    __test_denom_receive_packet_native_sttoken
}

########################################################################
# ujuno sent to Osmosis then to Stride, then to Juno then back to Stride 
########################################################################
__test_denom_receive_packet_sink_two_hops() {  # receive sink two hops
    # ujuno sent from Juno to Osmosis to Stride
    #   Expected Denom: ibc/FDB2394AA02EA9AC7DF68BE86BBE54846065EB967165FE78262601FBCAFB1A8F
    #                   (transfer/channel-2(juno)/transfer/channel-7(osmo)/ujuno)
    #   Expected Channel: channel-2
    juno_on_osmo='ibc/E5FD4F5963AA3CA00908DEA9BF29D35BA84183BBC0783A1224022BF55D348112'
    traveler_juno_on_stride='ibc/FDB2394AA02EA9AC7DF68BE86BBE54846065EB967165FE78262601FBCAFB1A8F'

    printf "\n>>> Transferring ujuno from Juno to Osmosis\n"
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-7 $(OSMO_ADDRESS) 10000000ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 10

    # Then transfer from osmo to stride 
    check_transfer_status OSMO STRIDE channel-0 channel-2 10000000 $juno_on_osmo $traveler_juno_on_stride true
}

__test_denom_send_packet_non_native_two_hops() { # send non-native (two hops)
    # ujuno (through Osmosis) sent from Stride to Juno  
    #  Expected Denom: ibc/FDB2394AA02EA9AC7DF68BE86BBE54846065EB967165FE78262601FBCAFB1A8F
    #                  (transfer/channel-2(juno)/transfer/channel-7(osmo)/ujuno)
    #  Expected Channel: channel-1
    traveler_juno_on_stride='ibc/FDB2394AA02EA9AC7DF68BE86BBE54846065EB967165FE78262601FBCAFB1A8F'
    check_transfer_status STRIDE JUNO channel-1 channel-1 10000000 $traveler_juno_on_stride $traveler_juno_on_stride true
    sleep 10
}

__test_denom_receive_packet_source_non_native() { # receive source non-native
    # ujuno (through Osmosis, then Stride) sent from Juno to Stride
    #  Expected Denom: ibc/FDB2394AA02EA9AC7DF68BE86BBE54846065EB967165FE78262601FBCAFB1A8F 
    #                  (transfer/channel-2(juno)/transfer/channel-7(osmo)/ujuno)
    #  Expected Channel: channel-1
    traveler_juno_on_stride='ibc/FDB2394AA02EA9AC7DF68BE86BBE54846065EB967165FE78262601FBCAFB1A8F'
    traveler_juno_on_juno='ibc/39A2ED50225EBC20C2C39737A8BB7CEAE0FF9F006C9A22DFF705126EA8E9EA4C'
    check_transfer_status JUNO STRIDE channel-0 channel-1 10000000 $traveler_juno_on_juno $traveler_juno_on_stride true
}

test_denom_juno_traveler() {
    print_header "TESTING DENOMS - TRAVELER JUNO"
    wait_until_epoch_end

    __test_denom_receive_packet_sink_two_hops
    __test_denom_send_packet_non_native_two_hops
    __test_denom_receive_packet_source_non_native
}
