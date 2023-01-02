## Rate Limit Integration Tests
* **These tests are not intended to be run with normal CI, they were meant as a comprehensive sanity check before deploying the module and are redundant with the unit tests.** 
* To setup the tests, modify the following in `config.sh`:
  * Set `HOST_CHAINS` to `(GAIA JUNO OSMO)`
  * Set `STRIDE_HOUR_EPOCH_DURATION` to `90s`
* Start dockernet
```
make start-docker
```
* Run the integration tests
```
bash dockernet/scripts/ratelimit/run_all_tests.sh
```
* Each test will print a checkmark or X depending on the status - if there are no X's, the tests passed.
