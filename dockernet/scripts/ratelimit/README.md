## Rate Limit Integration Tests
* **These tests are not intended to be run with normal CI, they were meant as a comprehensive sanity check before deploying the module and are redundant with the unit tests.** 
* To run the tests, first modify the HOST_CHAINS variables in `config.sh` to `HOST_CHAINS=(GAIA JUNO OSMO)`, then run
```
make start-docker && bash dockernet/scripts/ratelimit/run_all_tests.sh
```
* Each test will print a checkmark or X depending on the status - if there are no X's, the tests passed.
* These tests will not work once the transactions become governance gated.