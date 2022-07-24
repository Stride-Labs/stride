# Testing Upgrades in Local Mode
## Run Instructions
* Before working on the upgrade logic, compile the original binary and place it in `scripts-local/upgrades/binaries/` named `strided1`
* **This binary should represent the code before the upgrade changes. You'll likely want to checkout to the main branch to compile this.**
```
make build-local build=s
mkdir -p scripts-local/upgrades/binaries
cp build/strided scripts-local/upgrades/binaries/strided1
```
* Then upgrade code as needed
* Optionally update timing parameters to your liking
    * `MAX_DEPOSIT_PERIOD` (defined in `scripts-local/vars.sh`)   
    * `VOTING_PERIOD` (defined in `scripts-local/vars.sh`)   
    * `PROPOSAL_NAME` (defined in `scripts-local/submit_upgrade.sh`)   
    * `UPGRADE_HEIGHT` (defined in `scripts-local/submit_upgrade.sh`)   
* Then startup the chain in upgrade mode 
```
bash scripts-local/start_upgrade.sh
```
* The startup script will:
    * Compile the new binary
    * Instanitate the chain using accelerated governance parameters
    * Setup the cosmosvisor file structure required for upgrades
    * Start the chain in local mode using cosmosvisor
    * Propose and vote on an upgrade
    * Tail the stride logs so you can view the upgrade taking place
* To view the status of the proposal (and confirm it has passed), open up a new terminal window and run
```
bash scripts-local/upgrades/view_proposal_status.sh
```
    * It will first print the time at which voting ends, and then continuously probe the status. You must see "PROPOSAL_STATUS_PASSED" before the upgrade height in order for the upgrade to go through.
* While the chain is running, now is a good time to test your pre-upgrade condition (i.e. some validation that indicates you are on the old binary). When doing so, use `scripts-local/upgrades/binaries/strided1`
* View the stride logs which are tailed at the end of the startup script - you should notice an update occuring at the specified upgrade height.
* After the upgrade has occured, check a post-upgrade condition using `scripts-local/upgrades/binaries/strided2`
