# Integration Tests

This design for this integration test framework is heavily inspired by the Cosmology team's [starship](https://github.com/cosmology-tech/starship/tree/main).

## Setup

TODO

## Motivation

TODO

## Network

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

TODO

## Design Decisions

### API Service to share files during chain setup

In order to start the network as fast as possible, the chain should be initialized with ICS validators at genesis, rather than performing a switchover. However, in order to build the genesis file, the public keys must be gathered from each validator. This adds the constraint that keys must be consoldiated into a single process responsible for creating the genesis file.

This can be achieved by having a master node creating the genesis.json and keys for each validator, and then having each validator download the files from the master node. Ideally this would be handled by a shared PVC across each validator; however, Kuberentes has a constraint where you cannot mount multiple pods onto the same volume.

This led to the decision to use an API service to act as the intermediary that allows uploading and downloading of files. While at first glance, this smells of overengineering, the fastAPI implementation is actually quite simple (only marginally more code than creating and mounting a volume) and it improves the startup time dramatically since there's no need for the pods to wait for the volume to be mounted. Plus, it's likely that it can be leveraged in the future to help coordinate tasks across the different networks in the setup (e.g. it can store a registry of canonical IBC connections across chains).
