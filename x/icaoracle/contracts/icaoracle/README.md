# ICA Oracle CW Contract
## Instructions for Testing Locally
* Start dockernet from Stride repo home directory
```bash
make start-docker
```
* Get intergration tests running in background so redemption rate updates
```
make test-integration-docker
```
* Navigate to this contract
```
cd x/icaoracle/contracts/icaoracle
```
* Build the contract
```
make build-optimized
```
* Upload the contract
```
make store-contract 
```
* Add the oracle and instantiate the contract
```
make add-oracle
```
* Query the metrics and watch the redemption rate grow
```
make query-metrics
```
* See makefile for additional commands