---
title: "Autopilot"
excerpt: ""
category: 6392913957c533007128548e
---

# The Autopilot Module

The `Autopilot` module is to route the ibc transfer packets so that it can reduce the steps that users take to use Stride.

With current implementation of Autopilot module, it supports:

- Liquid staking as part of IBC transfer if it has functional part of LiquidStaking

Note: This will support more functions that can reduce number of users' operations.

## Memo

### Format

```json
{
  "autopilot": {
    "receiver": "strideXXX",
    "{module_name}": { "{additiional_field}": "{value}" }
  }
}
```

### Example (1-Click Liquid Stake)

```json
{
  "autopilot": {
    "receiver": "strideXXX",
    "stakeibc": {
      "action": "LiquidStake"
    }
  }
}
```

### Example (Update Airdrop Address)

```json
{
  "autopilot": {
    "receiver": "strideXXX",
    "claim": {}
  }
}
```

### A Note on Parsing

Since older versions of IBC do not have a `Memo` field, they must pass the routing information in the `Receiver` attribute of the IBC packet. To make autopilot backwards compatible with all older IBC versions, the receiver address must be specified in the JSON string. Before passing the packet down the stack to the transfer module, the address in the JSON string will replace the `Receiver` field in the packet data, regardless of the IBC version.

The module also enforces a maximum length for both the `Memo` and `Receiver` fields of 4000 and 100 characters respectively.

## Params

```
StakeibcActive (default bool = false)
ClaimActive (default bool = false)
```

## Keeper functions

- `TryLiquidStaking()`: Try liquid staking on IBC transfer packet
