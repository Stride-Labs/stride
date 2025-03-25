## Airdrop Integration Tests
Each airdrop testing script (1 through 4) tests different aspects of the airdrop. 

### Overview 
**NOTE**: You must update the airdrop timing params before running parts 4 and 5 (see next section)
* **Part 1: Standard**: Tests basic airdrop claims and actions
* **Part 2: Autopilot**: Tests claiming through autopilot on GAIA (ibc-go v3)
* **Part 3: Evmos**: Tests claiming through autopilot on EVMOS (ibc-go v5)
* **Part 4: Resets**: Tests that the airdrop resets properly at the epoch
* **Part 5: Staggered**: Tests two airdrops running side by side and staggered

### Instructions
* If running part 3, change the `HOST_CHAINS` variable in `config.sh` to run only evmos.
* If running Part 4 or 5: Before building stride, you must update the following airdrop timing parameters in `x/claim/types/params.go`:
    * `DefaultEpochDuration` to `time.Second * 60`
    * `DefaultVestingInitialPeriod` to `time.Second * 120`
* Only the GAIA host zone is required. Start dockernet with:
```bash
make start-docker build=sgr
```
* Run the corresponding script
```bash
bash dockernet/scripts/airdrop/airdrop{1/2/3/4}.sh
```
* **NOTE**: Each script must be run independently, meaning you must restart dockernet between runs (`make start-docker build=sgr`)
