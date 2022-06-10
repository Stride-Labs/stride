main_config=~/.stride/config/genesis.json
rm $main_config
ignite chain init --home ~/.stride/
jq '.app_state.epochs.epochs[2].duration = $newVal' --arg newVal "3s" $main_config > json.tmp && mv json.tmp $main_config
strided start --home ~/.stride/ 

# strided tx stakeibc register-host-zone connection-0 ATOM stATOM --chain-id stride --home ~/.stride/ --from bob --gas 500000 -y


# BOBADDR=$(strided keys show bob -a)
# strided q bank balances $BOBADDR
# strided q bank balances stride16mgpddhu25qh8q7e0geej594l5080z0eplydww
# strided q stakeibc module-address stakeibc
# ibcaddr=$(strided q stakeibc module-address stakeibc | awk '{print $NF}') 
# strided tx stakeibc liquid-stake 1000 ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --home ~/.stride/ --from bob --chain-id stride -y