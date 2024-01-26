## Staketia Integration Tests
* Use the default host chain settings, but shorten the day and stride epochs in `config.sh`
```
STRIDE_DAY_EPOCH_DURATION="40s"
STRIDE_EPOCH_EPOCH_DURATION="10s"
```
* Start dockernet
```bash
make start-docker
```
* As you go through the below flow, watch the `balances.log` and `state.log` files
* Run the setup script to transfer native tokens to Stride and set the withdrawal address
```bash
bash dockernet/scripts/staketia/setup.sh
```
* Run the liquid stake script. Watch the stToken appear in the validator account and a delegation record be created during the next epoch.
```bash
bash dockernet/scripts/staketia/liquid_stake.sh
```
* Delegate on the host zone and confirm on stride. Watch the delegated balance increase and the delegation record be removed.
```bash
bash dockernet/scripts/staketia/delegate.sh
```
* Redeem the stTokens. Watch the stTokens move into the redemption account and the accumulation unbonding record be incremented. A redemption record should also be created.
```bash
bash dockernet/scripts/staketia/redeem_stake.sh
```
* Wait for the next 4 day epoch and see the unbonding record change to status `UNBONDING_QUEUE`. This may take a few minutes.
* Unbond from the host zone and submit the confirm tx back to stride
```bash
bash dockernet/scripts/staketia/undelegate.sh
```
* Wait for the unbonding record's status to change to `UNBONDED`, after the tokens have finished unbonding. This will take a couple minutes.
* Sweep the tokens back to stride and confirm the tx. During the next epoch, the native tokens should be returned the redeemer and the redemption record should be removed.
```bash
bash dockernet/scripts/staketia/sweep.sh
```

