
# PARAMS: [name] (--upgrade-height [height]) (--upgrade-info [info]) [flags]
build/strided --home ./scripts-local/state/stride tx gov submit-proposal software-upgrade test1 --title="test1 title" --description "test description" --upgrade-height 100 --from val1 --keyring-backend test --chain-id STRIDE
            # --from val1 \
            # --from "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7" \
            # --keyring-backend test \
            # --chain-id STRIDE

# example: 
# simd tx gov submit-legacy-proposal software-upgrade v2 --title="Test Proposal" --description="testing, testing, 1, 2, 3" --upgrade-height 1000000 --from cosmos1..

