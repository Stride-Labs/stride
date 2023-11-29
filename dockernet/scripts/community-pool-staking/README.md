## Community Pool Staking Integration Tests
### Liquid Staking and Redemptions
* To test only liquid staking and redemptions from the community pool (without reinvestment), the setup is much simpler
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
bash dockernet/scripts/community-pool-staking/redeem.sh
```
* Similarly watch `logs/balances.log` to see the funds move - it will take a few minutes for the unbonding to complete. 
* When you no longer see the pending undelegation, run the claim script to send the native token to the return ICA, and then watch it travel back to the community pool
```bash
bash dockernet/scripts/community-pool-staking/claim.sh
```
* To test starting from a gov prop instead of a direct transfer to the deposit account, run
```bash
bash dockernet/scripts/community-pool-staking/stake_proposal.sh
```

### Reinvestment
* To test reinvestment, you must start up noble and osmosis as well
    * Set `HOST_CHAINS` to `(DYDX)` in `config.sh`
    * Set `ACCESSORY_CHAINS` to `(NOBLE OSMO)` in `config.sh
* Start the network
```bash
make start-docker
```
* Start relayers between dydx, noble and osmosis
```bash
bash dockernet/scripts/community-pool-staking/start_relayers.sh
```
* Create a pool on osmosis to allow trades between dydx and noble
```bash
bash dockernet/scripts/community-pool-staking/create_pool.sh
```
* Finally, test the reinvestment flow by sending USDC to the withdrawal address
```bash
bash dockernet/scripts/community-pool-staking/reinvest.sh
```