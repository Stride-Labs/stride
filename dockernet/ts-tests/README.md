```bash
# build stride locally and run dokcernet
(cd ../.. && make sync && make start-docker build=sgr)

# install deps
pnpm i

# run tests
pnpm test
```

IMPORTANT: `@cosmjs/amino` and `@cosmjs/stargate` vesrions must match the versions used by stridejs. To get those version, run `pnpm why @cosmjs/amino` and `pnpm why @cosmjs/stargate`.
