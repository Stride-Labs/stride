# Integration Tests

This design for this integration test framework is heavily inspired by the Cosmology team's [starship](https://github.com/cosmology-tech/starship/tree/main).

## Setup

TODO

## Motivation

TODO

## Network

TODO

## Testing Client

TODO

## Design Decisions

### API Service to share files during chain setup

In order to start the network as fast as possible, the chain should be initialized with ICS validators at genesis, rather than performing a switchover. However, in order to build the genesis file, the public keys must be gathered from each validator. This adds the constraint that keys must be consolidated into a single process responsible for creating the genesis file.

This can be achieved by having a master node creating the genesis.json and keys for each validator, and then having each validator download the files from the master node. Ideally this would be handled by a shared PVC across each validator; however, Kubernetes has a constraint where you cannot mount multiple pods onto the same volume.

This led to the decision to use an API service to act as the intermediary that allows uploading and downloading of files. While at first glance, this appears overengineered, the fastAPI implementation is actually quite simple (only marginally more code than creating and mounting a volume) and it improves the startup time dramatically since there's no need for the pods to wait for the volume to be mounted. Plus, it's likely that it can be leveraged in the future to help coordinate tasks across the different networks in the setup (e.g. it can store a registry of canonical IBC connections across chains).
