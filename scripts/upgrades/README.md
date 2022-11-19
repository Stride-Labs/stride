# Testing Upgrades in Local Mode
## Run Instructions
* Before working on the upgrade logic, compile the original binary and place it in `scripts/upgrades/binaries/` named `strided1`
* **This binary should represent the code before the upgrade changes. You'll likely want to checkout to the main branch to compile this.**
```
git checkout {OLD_COMMIT_HASH}
make build-docker build=s
mkdir -p scripts/upgrades/binaries
rm -f scripts/upgrades/binaries/strided1
cp build/strided scripts/upgrades/binaries/strided1
git checkout {UPDATED_BRANCH}
```
* Then switch the code back to the most recent version 
* Enter the commit hash of the old binary (built above) as `UPGRADE_OLD_COMMIT_HASH` in `scripts/vars.sh`
* Enter upgrade name as `UPGRADE_NAME` in `scripts/vars.sh`
* Then startup the chain as normal and rebuild stride
```
make start-docker build=s
```
* The startup script will:
    * Compile the new binary
    * Create the cosmosvisor file structure required for upgrades
    * Rebuild and replace the stride docker image with an image that has both binaries and is running cosmosvisor
        * This image pulls the new binary from the normal docker build that happens at the start of running this make command
* Once the chain is up and running, set the upgrade height (`UPGRADE_HEIGHT` in `scripts/submit_upgrade.sh`) and run the upgrade script to propose and vote on an upgrade
```
bash scripts/upgrades/submit_upgrade.sh
```
* View the stride logs - you should notice an update occuring at the specified upgrade height.
* After the upgrade has occured, check a post-upgrade condition using `scripts/upgrades/binaries/strided2`

## Testing Upgrades with Integration Tests
* **WARNING**: The integration tests may change between versions - the following only works if there were not breaking changes. If there are breaking changes, you can replace the GAIA and JUNO integration test files with those from the old version.
* Follow the instructions above to start the network but stop before submitting the proposal
* Run integration tests for GAIA and JUNO (comment out OSMO and STARS in `scripts/tests/run_all_tests.sh`)
* Once the tests pass, grab the current block height, modify `scripts/upgrades/submit_upgrade.sh` to have an upgrade height ~50 blocks in the future, and run the script
* Check the stride logs to confirm the upgrade passes successfully
* Modify `STRIDE_CMD` in `vars.sh` to point to the **new** binary (`STRIDE_CMD="$SCRIPT_DIR/upgrades/binaries/strided2"`)
* Finally, run integration tests for OSMO and STARS (comment out GAIA and JUNO in `scripts/tests/run_all_tests.sh`)
