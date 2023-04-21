## Airdrop Integration Tests
Each airdrop testing script (1 through 4) tests different aspects of the airdrop. 

### Overview 
**NOTE**: You must update the airdrop timing params before running parts 3 and 4 (see next section)
* **Part 1**: Tests basic airdrop claims and actions
* **Part 2**: Tests claiming through autopilot (ibc-go v3)
* **Part 3**: Tests that the airdrop resets properly at the epoch
* **Part 4**: Tests two airdrops running side by side and staggered

### Instructions
* If running Part 3 or 4: Before building stride, you must update the following airdrop timing parameters in `x/claim/types/params.go`:
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
