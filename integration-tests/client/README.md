### running the tests

```bash
# Start the network in k8s
(cd .. && make start)

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
  - update the config in `scripts/clone_repos.ts` to point to the new `stride/cosmos-sdk/ibc-go` version
  - run `pnpm i`
  - run `pnpm codegen`
  - run `git commit...`
  - run `git push`
  - get the current `stridejs` commit using `git rev-parse HEAD`
- in the integration tests (this project):
  - update the `stridejs` dependency commit hash in `package.json`
  - `pnpm i`
  - `pnpm test`
