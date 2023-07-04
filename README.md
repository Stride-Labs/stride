<img src="https://drive.google.com/uc?id=1c4DbV3lQ2Ely9VwEur4EmeHIea1my2AC" width="400">

# Multichain Liquid Staking

[Twitter](https://twitter.com/stride_zone) | [Discord](http://stride.zone/discord) | [Website](https://stride.zone/)
## What is Stride?


Stride is a blockchain ("zone") that provides liquidity for staked assets. Using Stride, you can earn *both* staking *and* DeFi yields across the Cosmos IBC ecosystem. Read our ["Meet Stride" blog post](https://stride.zone/blog/meet-stride) to learn more about why we built Stride. 

**Stride** is built using Cosmos SDK and Tendermint. Stride allows users to liquid stake any IBC-compatible cosmos SDK native appchain token. Under the hood, Stride leverages the [Inter-Blockchain Communication protocol](https://ibc.cosmos.network/), [Interchain Accounts](https://blog.cosmos.network/interchain-accounts-take-cosmos-interoperability-to-the-next-level-39c9a8aad4ad) and [Interchain Queries](https://github.com/schnetzlerjoe/interchain-query-spec).

## How does Multichain Liquid Staking work?

![](https://drive.google.com/uc?id=1RuK2YeMH7O6P9-8ro_ybOg5n79ySwFjN)

## Running a Mainnet Node
If you want to setup your node for the Stride mainnet, find the relevant files and instructions [here](https://github.com/Stride-Labs/mainnet/blob/main/mainnet/README.md)

## Getting Started as a Developer
### Developing on Stride

Developers who wish to develop on Stride can easily spin up 3 Stride nodes, 3 Gaia nodes, 1 Hermes relayer and 1 interchain-queries relayer. The Gaia nodes simulate the Cosmos Hub (Gaia) zone in local development node, and the relayers allow Stride zone to interact with that instance of Gaia.

The fastest way to develop on Stride is local development mode.

#### Set up local development mode 
Install the required git submodule dependencies (various chains, relayers, bats). 
```
git submodule update --init --recursive
```

Build executables, initialize state, and start the network with
```
make start-docker build=sgjotr
```
You can optionally pass build arguments to specify which binary to rebuild
1. `s` This will re-build the Stride binary (default)
2. `g` This will re-build the Gaia binary
3. `j` This will re-build the Juno binary
4. `o` This will re-build the Osmo binary
5. `t` This will re-build the Stargaze binary
6. `t` This will re-build the Evmos binary
7. `r` This will re-build the Go Relayer binary
8. `h` This will re-build the Hermes binary

Example: `make start-docker build=sg`, this will:
- Rebuild the Stride and Gaia binaries
- Start 1 Stride and 1 Gaia node in the docker
- Start Relayers

To bring down the chain, execute:
```
make stop-docker
```

To test the chain with a mnemonic that has tokens (e.g. sending Stride transactions), you can use the VAL_MNEMONIC_1, which is
```
close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly
```
This mnemonic will have tokens on every chain running locally.

#### Running integration tests
Ensure submoules are updated
```
git submodule update --init --recursive
```

Build Stride, Gaia, Evmos, and the go relayer
```
make start-docker build=sger
```

Run integration tests
```
make test-integration-docker
```

### Making changes to this repository
###### Summary
Add summary of the pull request here (*E.g. This pull request adds XYZ feature to the x/ABC module and associated unit tests.*)
###### Unit tests

To run unit tests for the whole project, execute:
`make unit-test`
To run unit tests for a particular module (e.g. the stakeibc module), execute:
`make unit-test path=stakeibc`
To run unit tests for a particular package (e.g. the stakeibc module), execute:
`make unit-test path=stakeibc/types`
To inspect unit test coverage, execute:
`make test-cover`

#### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Starport docs](https://docs.starport.com).

#### Release

To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

## Stride's Technical Architecture

Users stake their tokens on Stride from any Cosmos chain. Rewards accumulate in real time. No minimum. They will receive staked tokens immediately when they liquid stake. These staked tokens can be freely traded, and can be redeemed with Stride at any time to receive your original tokens plus staking rewards.

<img src="https://drive.google.com/uc?id=1CjEkovV3tHspZxlgP8b_th5NPb78T9nB" width="920">

On the backend, Stride permissionly stakes these tokens on the host chain and compounds user rewards. Stride lets users use your staked assets to compound their yields. Continue to earn staking yield, and earn additional yield by lending, LPing, and more. They can set their own risk tolerance in Cosmos DeFi.  

<img src="https://drive.google.com/uc?id=11kJZE93BdhNjkaNig3DGYnTnuSNClP8Y" width="900">

Users can always redeem from Stride. When they select "redeem" on the Stride website, Stride will initiate unbonding on the host zone. Once the unbonding period elapses, the users will receive native tokens in their wallets. 

<img src="https://drive.google.com/uc?id=1rtFiUwziiKjeUkJcJ9YuT1AN3JUtSVVr" width="900">

## Attribution

Stride is proud to be an open-source project, and we welcome all other projects to use our repo. We use modules from the cosmos-sdk and other open source projects.

We operate under the Apache 2.0 License, and have used the following modules from fellow Cosmos projects. Huge thank you to these projects!

We use the following modules from [Osmosis](https://github.com/osmosis-labs/osmosis) provided  under [this License](https://github.com/osmosis-labs/osmosis/blob/main/LICENSE):
```
x/epochs
x/mint
x/ratelimit
```
We use the following module (marketed as public infra) from [Quicksilver](https://github.com/ingenuity-build/quicksilver) provided under [this License](https://github.com/ingenuity-build/quicksilver/blob/main/LICENSE): 
```
x/interchainqueries
```

Relevant licenses with full attribution can be found in the subdirectories.
