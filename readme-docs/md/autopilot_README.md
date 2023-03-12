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

## Params

```
Active (default bool = true)
```

## Keeper functions

- `TryLiquidStaking()`: Try liquid staking on IBC transfer packet
