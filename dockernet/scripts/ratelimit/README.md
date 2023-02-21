## Rate Limit Integration Tests
* **These tests are not intended to be run with normal CI, they were meant as a comprehensive sanity check before deploying the module and are redundant with the unit tests.** 
* **WARNING**: `STRIDE_HOUR_EPOCH_DURATION` must be at least '90s' in `config.sh`
* `HOST_CHAINS` should be set to `(GAIA JUNO OSMO)` in `config.sh`
* Start dockernet
```
make start-docker
```
* Run the integration tests
```
bash dockernet/scripts/ratelimit/run_all_tests.sh
```
* Each test will print a checkmark or X depending on the status - if there are no X's, the tests passed.
