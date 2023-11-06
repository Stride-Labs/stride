## Community Pool Staking Integration Tests
* Set `HOST_CHAINS` to `(DYDX)` in `config.sh`
* Start dockernet
```bash
make start-docker
```
* Send native tokens to the deposit account to simulate a community pool liquid stake
```bash
bash dockernet/scripts/community-pool-staking/stake.sh
```
* View `logs/balances.log` to watch the funds traverse the different accounts
* To test the redemption flow, run
```bash
bash dockernet/scripts/community-pool-staking/stake.sh
```
* Similarly watch `logs/balances.log` to see the funds move - it will take a few minutes for the unbonding to complete. 
* When you no longer see the pending undelegation, run the claim script to send the native token to the return ICA, and then watch it travel back to the community pool
```bash
bash dockernet/scripts/community-pool-staking/claim.sh
```