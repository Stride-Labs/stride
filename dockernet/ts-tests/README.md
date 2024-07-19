### running the tests

```bash
# build stride locally and run dokcernet
(cd ../.. && make sync && make start-docker build=sgr)

# install deps
pnpm i

# run tests
pnpm test
```

IMPORTANT: `@cosmjs/*` dependencies must match the versions used by stridejs. To get those versions, run e.g. `pnpm why @cosmjs/amino`.

### debugging (vscode)

- open command palette: `Shift + Command + P (Mac) / Ctrl + Shift + P (Windows/Linux)`
- run the `Debug: Create JavaScript Debug Terminal` command
- set breakpoints
- run `pnpm test`

### test new protobufs

- go to https://github.com/Stride-Labs/stridejs
- cd into the `stride` directory
- `git checkout` to the version with the new protobufs
- build and run a dockernet of that version `make sync && make start-docker build=sgr`
- `cd ..` into the root project directory
- `npm i`
- `npm run codegen`
- `git commit`
- `git push`
- get the current `stridejs` commit: `git rev-parse HEAD`
- in the integration tests (this project) `package.json` file, update the `stridejs` dependency commit hash
- `pnpm i`
- `pnpm test`
