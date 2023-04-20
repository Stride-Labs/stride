## Airdrop Integration Tests
Each airdrop testing script (1 through 4) tests differnet aspects of the airdrop. 

### Overview 
* **Part 1**: Tests basic airdrop claims and actions
* **Part 2**: Tests claiming through autopilot (ibc-go v3)
* **Part 3**: Tests that the airdrop resets properly at the epoch
* **Part 4**:

### Instructions
* Only the GAIA host zone is required. Start dockernet with:
```bash
make start-docker build=sgr
```
* Run the corresponding script
```bash
bash dockernet/scripts/airdrop/airdrop{1/2/3/4}.sh
```
* **NOTE**: Each script must be run independently, meaning you must restart dockernet between runs (`make start-docker build=sgr`)
