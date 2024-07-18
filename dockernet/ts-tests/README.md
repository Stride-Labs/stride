```bash
# build stride locally and run dokcernet
(cd ../.. && make sync && make start-docker build=sgr)

# install deps
pnpm i

# run tests
pnpm test
```

IMPORTANT: `@cosmjs/*` dependencies must match the versions used by stridejs. To get those versions, run e.g. `pnpm why @cosmjs/amino`.
