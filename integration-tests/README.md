# Integration Tests

This design for this integration test framework is heavily inspired by the Cosmology team's [starship](https://github.com/cosmology-tech/starship/tree/main).

## Setup

**Before beginning, ensure you're on node 22**

```bash
nvm install v22
nvm use v22
```

Install javascript and python dependencies.

```bash
make install
```

IMPORTANT: `@cosmjs/*` dependencies must match the versions used by stridejs. To get those versions, run e.g. `pnpm why @cosmjs/amino`.

## Running Tests

Start the network

```bash
make start
```

Run the tests

```bash
make test
```

## Integrating Updated Protos

If the stride proto's change, we need to rebuild stridejs:

- Go to https://github.com/Stride-Labs/stridejs
  - Remove `/dist` from `.gitignore`
  - Update the config in `scripts/clone_repos.ts` to point to the new `stride/cosmos-sdk/ibc-go` version
  - Run `pnpm i`
  - Run `pnpm codegen`
  - Run `pnpm build`
  - Run `git commit...`
  - Run `git push`
  - Get the current `stridejs` commit using `git rev-parse HEAD`
- In the integration tests (this project):
  - Move into the `client` folder (`cd client`)
  - Update the `stridejs` dependency commit hash in `package.json`
  - `pnpm i`

## Debugging (VSCode)

- Open command palette: `Shift + Command + P (Mac) / Ctrl + Shift + P (Windows/Linux)`
- Run the `Debug: Create JavaScript Debug Terminal` command
- Set breakpoints
- Run tests

## Network

## Adding a New Host Zone

- Create a new dockerfile in `dockerfiles/Dockerfile.{chainName}`. You can use one of the existing dockerfiles as a reference, and just modify the `REPO`, `COMMIT_HASH`, and `BINARY` variables.
- Add a makefile entry to build the dockerfile
  ```bash
  build-{chainName}:
    $(call build_and_push_docker,{chainName},.,chains/{chainName}:{chainVersion})
  ```
- Try to build the docker image. You may have to debug here. Use the project's Dockerfile in their repo as a reference.
- [Internal Only] Add a DNS entry in GCP for the RPC and API endpoints
  - Go to GCP Cloud DNS (search DNS in the console)
  - Click on `internal`
  - Grab the IP Address from the exising host zones
  - Click `Add Standard`
  - Set the DNS Name to `{chainName}-api.internal.stridenet.co`
  - Set the IP Address to same IP as the other host zones
- Add the new chain to `network/values.yaml`, including:
  - `chainConfig`
  - `activeChains`
  - `relayers`
- Add the new relayer configs to `network/configs/relayer.yaml` and `network/configs/hermes.toml`
- Then start the network as normal

```bash
make start
```

- If running tests, add the chain config to `client/test/consts.ts` and update the `HOST_CHAIN_NAME` in `client/test/core.test.ts`

### Validator Startup Lifecycle

**initContainer**

- For context, the `initContainer` is a separate container that runs before the main process. In this case, it's used to handle the validator node setup before the main startup loop.
- The _first_ validator runs `init-chain.sh` which creates the `genesis.json` and all the validator keys
- The genesis file and validator keys are uploaded and stored in the API
- Then each validator runs `init-node.sh` which downloads the relevant files and sets up the config files in the validator's home directory
- **NOTE:** Logs from the `initContainer` are not natively viewable via the `kubectl` cli. As a workaround, they are piped to a file in the node and can be viewed by exec'ing into the container with `POD_ID={pod-id} make startup-logs`

**main**

- The main thread simply runs the `binaryd start` command

**postStart**

- As a `postStart` operation (run after the main thread is kicked off), the `create-validator.sh` script is run which runs the appropriate `staking` module transaction to create the validator using the previously acquired keys
- This is run after startup because the validator must sign the tx with their key

## Testing Client

### Debugging (VSCode)

- open command palette: `Shift + Command + P (Mac) / Ctrl + Shift + P (Windows/Linux)`
- run the `Debug: Create JavaScript Debug Terminal` command
- set breakpoints
- run `pnpm test`

### Test new protobuf

- go to https://github.com/Stride-Labs/stridejs
  - remove `/dist` from `.gitignore`
  - update the config in `scripts/clone_repos.ts` to point to the new `stride/cosmos-sdk/ibc-go` version
  - run `pnpm i`
  - run `pnpm codegen`
  - run `pnpm build`
  - run `git commit...`
  - run `git push`
  - get the current `stridejs` commit using `git rev-parse HEAD`
- in the integration tests (this project):
  - update the `stridejs` dependency commit hash in `package.json`
  - `pnpm i`
  - `pnpm test`

## Motivation

TODO

## Design Decisions

### API Service to share files during chain setup

In order to start the network as fast as possible, the chain should be initialized with ICS validators at genesis, rather than performing a switchover. However, in order to build the genesis file, the public keys must be gathered from each validator. This adds the constraint that keys must be consoldiated into a single process responsible for creating the genesis file.

This can be achieved by having a master node creating the genesis.json and keys for each validator, and then having each validator download the files from the master node. Ideally this would be handled by a shared PVC across each validator; however, Kuberentes has a constraint where you cannot mount multiple pods onto the same volume.

This led to the decision to use an API service to act as the intermediary that allows uploading and downloading of files. While at first glance, this smells of overengineering, the fastAPI implementation is actually quite simple (only marginally more code than creating and mounting a volume) and it improves the startup time dramatically since there's no need for the pods to wait for the volume to be mounted. Plus, it's likely that it can be leveraged in the future to help coordinate tasks across the different networks in the setup (e.g. it can store a registry of canonical IBC connections across chains).
