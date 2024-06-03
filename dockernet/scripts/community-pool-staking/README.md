## Community Pool Staking Integration Tests
### Liquid Staking and Redemptions
* To test only liquid staking and redemptions from the community pool (without reinvestment), the setup is much simpler
* Set `HOST_CHAINS` to either `(DYDX)` or `(GAIA)` in `config.sh`
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
    * Set `HOST_CHAINS` to either `(DYDX)` or `(GAIA)` in `config.sh`
    * Set `ACCESSORY_CHAINS` to `(NOBLE OSMO)` in `config.sh
* Start the network
```bash
make start-docker
```
* Start relayers between dydx, noble and osmosis
```bash
bash dockernet/scripts/community-pool-staking/setup_relayers.sh
```
* Create a pool on osmosis to allow trades between dydx and noble
```bash
bash dockernet/scripts/community-pool-staking/create_pool.sh
```
* Register the trade route to configure the conversion of reward tokens to host tokens
```bash
bash dockernet/scripts/community-pool-staking/add_trade_route.sh
```
* Finally, test the reinvestment flow by sending USDC to the withdrawal address. View `logs/balances.log` to watch the funds traverse the different accounts
```bash
bash dockernet/scripts/community-pool-staking/reinvest_reward.sh
```

### Rebate
* Use the default host zone setup of just `GAIA`
* To test sending a rebate to the treasury, there is no need to change anything.
* If you want to send a rebate to the community pool instead, comment out the `GAIA_TREASURY_ADDRESS` in `config.sh`
* Start the network
```bash
make start-docker
```
* Liquid stake to create TVL 
```bash
bash dockernet/scripts/community-pool-staking/stake.sh
```
* Register a rebate
```bash
bash dockernet/scripts/community-pool-staking/rebate.sh
```
* Watch `balances.log` to verify the rewards were distributed correctly. The reinvestment cycle will kick off automatically.
* Notice the rewards start in the withdrawal account, then 0.25% get sent to the community pool, 9.75% get sent to the fee account and 90% get sent to the delegation account.
* Depending on the setup from above, the community pool portion will either go to the main community pool or the treasury. Note: if it goes to the main community pool, it is difficult to distinguish the rebate tokens from the tokens that were already in the account.
