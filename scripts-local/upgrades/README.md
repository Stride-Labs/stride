# Testing Upgrades in Local Mode
## Run Instructions
* Before working on the upgrade logic, compile the original binary and place it in `scripts-local/upgrades/binaries/` named `strided1`
```
make build-local build=s
mkdir -p scripts-local/upgrades/binaries
cp build/strided scripts-local/upgrades/binaries/strided1
```
* Then upgrade code as needed
* Optionally update timing parameters to your liking (defined at the top of `scripts-local/start_upgrade.sh`)   
    * MAX_DEPOSIT_PERIOD
    * VOTING_PERIOD
    * PROPOSAL_NAME
    * UPGRADE_HEIGHT
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
* While the chain is running, now is a good time to test your pre-upgrade condition (i.e. some validation that indicates you are on the old binary). When doing so, use `scripts-local/upgrades/binaries/strided1`
* View the stride logs which are tailed at the end of the startup script - you should notice an update occuring at the specified upgrade height.
* After the upgrade has occured, check a post-upgrade condition using `scripts-local/upgrades/binaries/strided2`
