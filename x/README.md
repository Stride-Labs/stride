---
title: "Stride Modules"
excerpt: ""
category: 62c5c5ff03a5bf069004def2
---

## Stride Modules

`stakeibc` - Manages minting and burning of stAssets, staking and unstaking of native assets across chains.
`icacallbacks` - Callbacks for interchain accounts.
`records` - IBC middleware wrapping the transfer module, does record keeping on IBC transfers and ICA calls
`claim` - airdrop logic for Stride's rolling, task-based airdrop
`interchainquery` - Issues queries between IBC chains, verifies state proof and executes callbacks.
`epochs` - Makes on-chain timers which other modules can execute code during.
`mint` - Controls token supply emissions, and what modules they are directed to.

### Attribution

Stride is proud to be an open-source project, and we welcome all other projects to use our repo. We use modules from the cosmos-sdk and other open source projects.

We operate under the Apache 2.0 License, and have used the following modules from fellow Cosmos projects. Huge thank you to these projects!

We use the following modules from [Osmosis](https://github.com/osmosis-labs/osmosis) provided  under [this License](https://github.com/osmosis-labs/osmosis/blob/main/LICENSE):

```
x/epochs
x/mint
```

We use the following module (marketed as public infra) from [Quicksilver](https://github.com/ingenuity-build/quicksilver) provided under [this License](https://github.com/ingenuity-build/quicksilver/blob/main/LICENSE):

```
x/interchainqueries
```

Relevant licenses with full attribution can be found in the relevant repo subdirectories.
