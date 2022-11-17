# Claims module

## Abstract

The Stride claims module has users claim higher percentages as they perform certain tasks on-chain.
Furthermore, these claimable assets 'expire' if not claimed.
Users have three months (`AirdropDuration`) to claim their full airdrop amount
After three months from launch, all unclaimed tokens get sent to the community pool.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Events](03_events.md)**
4. **[Keeper](04_keeper.md)**  
5. **[React Hooks](05_react_hooks.md)**  
6. **[Queries](06_queries.md)**  
7. **[Params](07_params.md)**

## Genesis State

## Actions

There are 2 types of actions, each of which release another 50% of the airdrop allocation.
The 2 actions are as follows:

```golang
ActionLiquidStake  Action = 0
ActionDelegateStake Action = 1
```

These actions are monitored by registring claim **hooks** to the stakeibc, and staking modules.
This means that when you perform an action, the claims module will immediately unlock those coins if they are applicable.
These actions can be performed in any order.

The code is structured by separating out a segment of the tokens as "claimable", indexed by each action type.
So if Alice delegates tokens, the claims module will move the 50% of the claimables associated with staking to her liquid balance.
If she delegates again, there will not be additional tokens given, as the relevant action has already been performed.
Every action must be performed to claim the full amount.

## ClaimRecords

A claim record is a struct that contains data about the claims process of each airdrop recipient.

It contains an address, the initial claimable airdrop amount, and an array of bools representing 
whether each action has been completed. The position in the array refers to enum number of the action.

So for example, `[true, true]` means that `ActionLiquidStake` and `ActionDelegateStake` are completed.

```golang
type ClaimRecord struct {
	// address of claim user
	Address string
	// weight that represent the portion from total allocation
	Weight sdk.Dec
	// true if action is completed
	// index of bool in array refers to action enum #
	ActionCompleted []bool
}

```


## Params

The airdrop logic has 4 parameters:

```golang
type Params struct {
    // Time that marks the beginning of the airdrop disbursal,
    // should be set to chain launch time.
    AirdropStartTime   time.Time
    AirdropDuration time.Duration
    // denom of claimable asset
    ClaimDenom string
    // address of distributor account
    DistributorAddress string
}
```
