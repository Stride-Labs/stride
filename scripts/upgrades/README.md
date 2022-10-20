# Testing Upgrades in Local Mode
## Run Instructions
* Before working on the upgrade logic, compile the original binary and place it in `scripts/upgrades/binaries/` named `strided1`
* **This binary should represent the code before the upgrade changes. You'll likely want to checkout to the main branch to compile this.**
```
git checkout {OLD_COMMIT_HASH}
make build-docker build=s
mkdir -p scripts/upgrades/binaries
cp build/strided scripts/upgrades/binaries/strided1
git checkout {UPDATED_BRANCH}
```
* Then switch the code back to the most recent version 
* Enter the commit hash of the old binary (built above) as `UPGRADE_OLD_COMMIT_HASH` in `scripts/vars.sh`
* Enter upgrade name as `UPGRADE_NAME` in `scripts/vars.sh` and `PROPOSAL_NAME` in `scripts/submit_upgrade.sh`
* Optionally update timing parameters the upgrade height (`UPGRADE_HEIGHT` in `scripts/submit_upgrade.sh`)
* Then startup the chain as normal and rebuild stride
```
make start-docker build=s
```
* The startup script will:
    * Compile the new binary
    * Create the cosmosvisor file structure required for upgrades
    * Rebuild and replace the stride docker image with an image that has both binaries and is running cosmosvisor
        * This image pulls the new binary from the normal docker build that happens at the start of running this make command
* Once the chain is up and running, run the upgrade script to propose and vote on an upgrade
```
bash scripts/upgrades/submit_upgrade.sh
```
* To view the status of the proposal (and confirm it has passed), open up a new terminal window and run
```
bash scripts/upgrades/view_proposal_status.sh
```
    * It will first print the time at which voting ends, and then continuously probe the status. You must see "PROPOSAL_STATUS_PASSED" before the upgrade height in order for the upgrade to go through.
* While the chain is running, now is a good time to test your pre-upgrade condition (i.e. some validation that indicates you are on the old binary). When doing so, use `scripts/upgrades/binaries/strided1`
* View the stride logs - you should notice an update occuring at the specified upgrade height.
* After the upgrade has occured, check a post-upgrade condition using `scripts/upgrades/binaries/strided2`
