# Testing Upgrades in Local Mode
## Run Instructions
* Before working on the upgrade logic, compile the original binary and place it in `dockernet/upgrades/binaries/` named `strided1`
* **This binary should represent the code before the upgrade changes. You'll likely want to checkout to the main branch to compile this.**
```
git checkout {OLD_COMMIT_HASH}
make build-docker build=s
mkdir -p dockernet/upgrades/binaries
rm -f dockernet/upgrades/binaries/strided1
cp build/strided dockernet/upgrades/binaries/strided1
git checkout {UPDATED_BRANCH}
```
* Then switch the code back to the most recent version 
* Enter the commit hash of the old binary (built above) as `UPGRADE_OLD_COMMIT_HASH` in `dockernet/config.sh`
* Enter upgrade name as `UPGRADE_NAME` in `dockernet/config.sh`
* Then startup the chain as normal and rebuild stride
```
make start-docker build=s
```
* The startup script will:
    * Compile the new binary
    * Create the cosmosvisor file structure required for upgrades
    * Rebuild and replace the stride docker image with an image that has both binaries and is running cosmosvisor
        * This image pulls the new binary from the normal docker build that happens at the start of running this make command
* Once the chain is up and running, set the upgrade height (`UPGRADE_HEIGHT` in `dockernet/submit_upgrade.sh`) and run the upgrade script to propose and vote on an upgrade
```
bash dockernet/upgrades/submit_upgrade.sh
```
* View the stride logs - you should notice an update occuring at the specified upgrade height.
* After the upgrade has occured, check a post-upgrade condition using `dockernet/upgrades/binaries/strided2`

## Testing Upgrades with Integration Tests
* **WARNING**: The integration tests may change between versions - the following only works if there were not breaking changes. If there are breaking changes, you can replace the GAIA and OSMO integration test files with those from the old version.
* Start the network, run the integration tests on the old binary, and then propose and vote on the upgrade:
```
make start-docker-all && make upgrade-init && make upgrade-submit
```
* Once the integration tests pass and the upgrade has been proposed, wait for the upgrade to occur at block 400. Check the stride logs to confirm the upgrade passes successfully
* Finally, run the remaining integration tests 
```
make upgrade-validate
```
