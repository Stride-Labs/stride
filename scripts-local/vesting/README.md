## Vesting Module

### Custom Implementations

Evmos x/vesting (from Agoric)
- Osmosis' clawback vesting, in the _auth_ module https://github.com/osmosis-labs/cosmos-sdk/tree/osmosis-main/x/auth/vesting
- https://docs.evmos.org/modules/vesting/ 
- https://github.com/evmos/evmos/blob/main/x/vesting/spec/README.md (removes all vesting account types other than clawback type)
- https://docs.cosmos.network/main/modules/auth/05_vesting.html
- https://github.com/Agoric/agoric-sdk/issues/4085
- https://github.com/agoric-labs/cosmos-sdk/tree/Agoric/x/auth/vesting/cmd/vestcalc
- Misc Terra vesting upgrades to cosmos-sdk/x/vesting https://github.com/terra-money/core/blob/beeff192329372e2bb993f897b8c866fd8be917d/app/upgrade.go#L24


### Cosmos SDK Vesting Module

The Cosmos Hub's vesting account needs to be initialized during genesis with a starting balance `X` and a vesting end time `ET`. Optional parameters are vesting start time `ST` and number of vesting periods `P`.
Owners of vesting accounts can freely delegate and undelegate from validators but they cannot transfer the unvested tokens to another account. 
There are a few types of vesting accounts:
- Delayed vesting: all coins vest once `ET` is reached. Can be created after genesis.
- Continuous vesting, where coins vest linearly between `ST` and `ET`. Can be created after genesis.
- Periodic vesting, where coins vest periodically based on some period between `ST` and `ET` (in batches). Must be created AT GENESIS, or as part of a manual network upgrade.
The current spec does not allow for clawbacks or conditional vesting (e.g. stop the vest if x condition triggers)


## Testing Delayed Vesting

To test delayed vesting, let's vest `1000strd` to the `val1` account with a 2 minute cliff and 5 minute linear continuous vest. 

The vesting setup needs to occur during genesis. In `scripts-local/init_stride.sh:36`, replace the current `add-genesis-account` command with the following commands, which contain flags to set up delayed vesting to the account.

```
VESTING_START_TIME=$(($(date +%s)+120)) # <= unix time start of vesting period (2 minutes from now)
VESTING_END_TIME=$((VESTING_START_TIME+300)) # <= unix time end of vesting period (7 minutes from now)
VESTING_AMT="1000000000ustrd" # <= amount of tokens to vest to the account
$STRIDE_CMD add-genesis-account ${val_addr} 500000000000ustrd --vesting-start-time $VESTING_START_TIME --vesting-end-time $VESTING_END_TIME --vesting-amount $VESTING_AMT
```

_Note that the amount vested is taken out of the total amount genesis'd to the account, it is NOT additional to the amount genesis'd (in this case, that means the account will have `500000000000ustrd` total tokens, of which `1000000000ustrd` will be vesting)_

A simple way to check how much has vested is to look at the account on the block explorer (with our local ping.pub, setup from https://github.com/Stride-Labs/explorer, that's http://localhost:8080/stride/account/stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7). Another way is to attempt to bank send the full acct balance to another addr before the vest has completed. Inspecting that failed tx will show you how much is available to send (which tells you how much has vested thus far)

```
build/strided --home ./scripts-local/state/stride tx bank send val1 stride1ft20pydau82pgesyl9huhhux307s9h3078692y 499000000000ustrd --chain-id STRIDE --keyring-backend test -y
build/strided --home ./scripts-local/state/stride q tx <TX_HASH_FROM_PREV_TX>
```

Once vesting completes 7 min from genesis, bank sending the full account balance should succeed!


## Our approach

We want delayed vesting for all vesting accounts. We simply need to, for each acct 
1. calc the amount to vest
2. have the recipient set up an addr on testnet, save their seed phrase and send us the address
3. set up vesting for that account with the following params 
    ```
    VESTING_START_TIME=1685592000 # June 1st 2023
    VESTING_END_TIME=1780286400 # June 1st 2026
    VESTING_AMT='XXXustrd'
    VESTER_ADDR='YYY'
    $STRIDE_CMD add-genesis-account $VESTER_ADDR $VESTING_AMT --vesting-start-time $VESTING_START_TIME --vesting-end-time $VESTING_END_TIME --vesting-amount $VESTING_AMT
    ```


Open questions: 
- Should we switch to using Osmosis' version?
- Do we want/need clawbacks?