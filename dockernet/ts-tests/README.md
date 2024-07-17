```bash
# build stride locally and run dokcernet
(cd ../.. && make sync && make start-docker build=sgr)

# install deps
pnpm i

# run tests
pnpm test
```
