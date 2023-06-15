# Testing Upgrades in Local Mode
## Run Instructions
* Before working on the upgrade logic, you'll need to compile the original binary that represent the code before the upgrade takes place. This is done by the following command, where the `old_version` is the version you're upgrading *from*:
``` bash
# e.g. make UPGRADE_OLD_VERSION=v8.0.0 upgrade-build-old-binary
make UPGRADE_OLD_VERSION={VERSION} upgrade-build-old-binary
```
* Then startup the chain, but specify the old tag or commit hash, as well as the upgrade name
```bash
# e.g. make UPGRADE_OLD_VERSION=v8.0.0 UPGRADE_NAME=v9 upgrade-build-old-binary
make UPGRADE_OLD_VERSION={VERSION} UPGRADE_NAME={NAME} start-docker 
```
* The startup script will:
    * Compile the new binary
    * Create the cosmosvisor file structure required for upgrades
    * Rebuild and replace the stride docker image with an image that has both binaries and is running cosmosvisor
        * This image pulls the new binary from the normal docker build that happens at the start of running this make command
* Once the chain is up and running, submit the upgrade by running the following (the upgrade will occur at block 150):
```bash
# e.g. make UPGRADE_NAME=v10 submit-upgrade-immediately
make UPGRADE_NAME={NAME} submit-upgrade-immediately
```
* View the stride logs - you should notice an update occuring at the specified upgrade height.

## Testing Upgrades with Integration Tests
* **WARNING**: The integration tests may change between versions - the following only works if there were not breaking changes. If there are breaking changes, you can replace the GAIA and EVMOS integration test files with those from the old version.
* Compile the old binary
``` bash
# e.g. make UPGRADE_OLD_VERSION=v8.0.0 upgrade-build-old-binary
make UPGRADE_OLD_VERSION={VERSION} upgrade-build-old-binary
```
* Run the following to start the network, run the integration tests on the old binary, and then propose and vote on the upgrade:
```bash
# e.g. make UPGRADE_OLD_VERSION=v8.0.0 UPGRADE_NAME=v10 upgrade-integration-tests-part-1
make UPGRADE_OLD_VERSION={VERSION} UPGRADE_NAME={NAME} upgrade-integration-tests-part-1
```
* Once the integration tests pass and the upgrade has been proposed, wait for the upgrade to occur at block 400. Check the stride logs to confirm the upgrade passes successfully
* Finally, run the remaining integration tests 
```bash
# e.g. make UPGRADE_NAME=v10 finish-upgrade-integration-tests
make UPGRADE_NAME={NAME} finish-upgrade-integration-tests
```
